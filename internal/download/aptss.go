package download

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/Xynrin/spark-store-tui/internal/domain"
	"github.com/Xynrin/spark-store-tui/internal/system"
)

// APTSSService delegates Debian package retrieval to Spark Store's official
// aptss backend. aptss/aria2 owns mirror selection, repository digest
// verification and resume state; the TUI only persists the surrounding task.
type APTSSService struct {
	store       TaskWriter
	downloadDir string
	command     string
}

func NewAPTSSService(store TaskWriter, downloadDir string) *APTSSService {
	return &APTSSService{store: store, downloadDir: downloadDir}
}

func (s *APTSSService) Download(ctx context.Context, app domain.App) (domain.DownloadTask, error) {
	process, task, err := s.PrepareDownload(app)
	if err != nil {
		return task, err
	}
	command := exec.CommandContext(ctx, process.Path, process.Args[1:]...)
	command.Dir = process.Dir
	command.Env = process.Env
	err = command.Run()
	return s.FinishDownload(task, err)
}

// PrepareDownload creates a live aptss process suitable for tea.ExecProcess,
// allowing aria2's progress and prompts to remain visible in the terminal.
func (s *APTSSService) PrepareDownload(app domain.App) (*exec.Cmd, domain.DownloadTask, error) {
	if app.Filename == "" || filepath.Base(app.Filename) != app.Filename {
		return nil, domain.DownloadTask{}, fmt.Errorf("%s 没有安全的 Debian 软件包文件名", app.Name)
	}
	if err := validateArchitecture(app.Architecture, app.Filename); err != nil {
		return nil, domain.DownloadTask{}, err
	}
	packageName, err := system.PackageName(system.Host{Family: "deb"}, app)
	if err != nil {
		return nil, domain.DownloadTask{}, err
	}
	packageSpec := packageName
	if safePackageVersion(app.Version) {
		packageSpec += "=" + app.Version
	}
	commandPath := s.command
	if commandPath == "" {
		commandPath, err = exec.LookPath("aptss")
		if err != nil {
			return nil, domain.DownloadTask{}, fmt.Errorf("找不到 aptss；请安装 Spark Store/aptss 后重试：%w", err)
		}
	}
	if err := os.MkdirAll(s.downloadDir, 0o755); err != nil {
		return nil, domain.DownloadTask{}, err
	}

	task := s.newTask(app, packageSpec)
	if err := s.persist(task); err != nil {
		return nil, task, err
	}
	process := exec.Command(commandPath, "download", packageSpec)
	process.Dir = s.downloadDir
	return process, task, nil
}

// FinishDownload reconciles aptss/aria2 output with the durable TUI task.
func (s *APTSSService) FinishDownload(task domain.DownloadTask, runErr error) (domain.DownloadTask, error) {
	now := time.Now().UTC()
	info, statErr := os.Stat(task.Destination)
	controlExists := regularFileExists(task.Destination + ".aria2")
	if runErr == nil && statErr == nil && !info.IsDir() && !controlExists {
		task.Status = domain.TaskCompleted
		task.BytesCompleted = info.Size()
		task.BytesTotal = info.Size()
		task.Error = ""
		task.UpdatedAt = now
		return task, s.persist(task)
	}

	if runErr == nil {
		runErr = fmt.Errorf("aptss 未生成预期文件 %s", filepath.Base(task.Destination))
	}
	task.Status = domain.TaskFailed
	if (statErr == nil && !info.IsDir()) || controlExists {
		task.Status = domain.TaskInterrupted
	}
	if statErr == nil && !info.IsDir() {
		task.BytesCompleted = info.Size()
	}
	task.Error = runErr.Error()
	task.UpdatedAt = now
	if err := s.persist(task); err != nil {
		return task, err
	}
	return task, fmt.Errorf("aptss 下载失败：%w；可执行 sudo aptss ssupdate 后按 D 续传", runErr)
}

// RecoverInterrupted never starts network work. A completed aria2 output is
// accepted; a file with an .aria2 control file remains resumable.
func (s *APTSSService) RecoverInterrupted() ([]domain.DownloadTask, error) {
	reader, ok := s.store.(TaskReader)
	if !ok {
		return nil, nil
	}
	tasks, err := reader.Load()
	if err != nil {
		return nil, err
	}
	recovered := make([]domain.DownloadTask, 0)
	for _, task := range tasks {
		if task.Status != domain.TaskQueued && task.Status != domain.TaskRunning && task.Status != domain.TaskVerifying {
			continue
		}
		info, statErr := os.Stat(task.Destination)
		if statErr == nil && !info.IsDir() && !regularFileExists(task.Destination+".aria2") {
			task.Status = domain.TaskCompleted
			task.BytesCompleted = info.Size()
			task.BytesTotal = info.Size()
			task.Error = ""
		} else {
			task.Status = domain.TaskInterrupted
			task.Error = "应用已重启；按 D 让 aptss/aria2 继续下载"
			for _, partial := range []string{task.Destination, task.Destination + ".part"} {
				if partialInfo, partialErr := os.Stat(partial); partialErr == nil && !partialInfo.IsDir() {
					task.BytesCompleted = partialInfo.Size()
					break
				}
			}
		}
		task.UpdatedAt = time.Now().UTC()
		if err := s.persist(task); err != nil {
			return recovered, err
		}
		recovered = append(recovered, task)
	}
	return recovered, nil
}

func (s *APTSSService) ExistingPath(app domain.App) string {
	if app.Filename == "" || filepath.Base(app.Filename) != app.Filename {
		return ""
	}
	path := filepath.Join(s.downloadDir, app.Filename)
	if !regularFileExists(path) || regularFileExists(path+".aria2") {
		return ""
	}
	return path
}

func (s *APTSSService) newTask(app domain.App, packageSpec string) domain.DownloadTask {
	destination := filepath.Clean(filepath.Join(s.downloadDir, app.Filename))
	if reader, ok := s.store.(TaskReader); ok {
		if tasks, err := reader.Load(); err == nil {
			for index := len(tasks) - 1; index >= 0; index-- {
				task := tasks[index]
				if task.AppID == app.ID && filepath.Clean(task.Destination) == destination && task.Status != domain.TaskCompleted {
					task.Status = domain.TaskRunning
					task.URL = "aptss:" + packageSpec
					task.MirrorID = "aptss"
					task.Error = ""
					task.UpdatedAt = time.Now().UTC()
					return task
				}
			}
		}
	}
	now := time.Now().UTC()
	return domain.DownloadTask{
		ID:          newTaskID(),
		AppID:       app.ID,
		SourceID:    app.SourceID,
		MirrorID:    "aptss",
		URL:         "aptss:" + packageSpec,
		Destination: destination,
		Status:      domain.TaskRunning,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func (s *APTSSService) persist(task domain.DownloadTask) error {
	if s.store == nil {
		return nil
	}
	return s.store.Upsert(task)
}

func safePackageVersion(value string) bool {
	return value != "" && !strings.HasPrefix(value, "-") && strings.IndexFunc(value, func(character rune) bool {
		return unicode.IsSpace(character) || unicode.IsControl(character)
	}) == -1
}

func regularFileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

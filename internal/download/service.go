package download

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Xynrin/spark-store-tui/internal/domain"
)

// Service converts source metadata into durable download tasks. It falls back
// through Metalink mirrors one at a time and retains failed task history.
type Service struct {
	client      *http.Client
	manager     *Manager
	store       TaskWriter
	downloadDir string
}

// TaskReader is implemented by the durable task store. Keeping it optional
// lets callers use an in-memory TaskWriter in focused tests.
type TaskReader interface {
	Load() ([]domain.DownloadTask, error)
}

func NewService(client *http.Client, store TaskWriter, downloadDir string) *Service {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Minute}
	}
	return &Service{client: client, manager: NewManager(client, store), store: store, downloadDir: downloadDir}
}

func DefaultDownloadDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "downloads"
	}
	return filepath.Join(home, "Downloads")
}

func (s *Service) Download(ctx context.Context, app domain.App) (domain.DownloadTask, error) {
	if app.MetalinkURL == "" {
		return domain.DownloadTask{}, fmt.Errorf("%s does not expose a Spark Metalink asset", app.Name)
	}
	asset, err := ResolveMetalink(ctx, s.client, app.MetalinkURL)
	if err != nil {
		return domain.DownloadTask{}, err
	}
	task := s.newTask(app, asset.Filename)
	for _, url := range asset.URLs {
		task.URL = url
		task.MirrorID = "metalink-auto"
		completed, downloadErr := s.manager.Run(ctx, task)
		if downloadErr == nil {
			return completed, nil
		}
		task = completed
		err = downloadErr
	}
	return domain.DownloadTask{}, fmt.Errorf("all Metalink mirrors failed: %w", err)
}

func (s *Service) newTask(app domain.App, filename string) domain.DownloadTask {
	destination := filepath.Clean(filepath.Join(s.downloadDir, filename))
	if reader, ok := s.store.(TaskReader); ok {
		if tasks, err := reader.Load(); err == nil {
			for index := len(tasks) - 1; index >= 0; index-- {
				task := tasks[index]
				if task.AppID == app.ID && filepath.Clean(task.Destination) == destination && task.Status != domain.TaskCompleted {
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
		MirrorID:    "metalink-auto",
		Destination: destination,
		Status:      domain.TaskQueued,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// RecoverInterrupted reconciles durable task state with files left on disk
// after an application crash or forced exit. It never starts network work;
// the user can press D to resume any interrupted task.
func (s *Service) RecoverInterrupted() ([]domain.DownloadTask, error) {
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
		if info, statErr := os.Stat(task.Destination); statErr == nil && !info.IsDir() {
			task.Status = domain.TaskCompleted
			task.BytesCompleted = info.Size()
			task.BytesTotal = info.Size()
			task.Error = ""
		} else {
			if part, partErr := os.Stat(task.Destination + ".part"); partErr == nil && !part.IsDir() {
				task.BytesCompleted = part.Size()
			}
			task.Status = domain.TaskInterrupted
			task.Error = "application restarted; press D to resume"
		}
		task.UpdatedAt = time.Now().UTC()
		if err := s.manager.persist(task); err != nil {
			return recovered, err
		}
		recovered = append(recovered, task)
	}
	return recovered, nil
}

// ExistingPath returns the previously downloaded package for an application.
// The filename comes from signed/official metadata rather than a UI label.
func (s *Service) ExistingPath(app domain.App) string {
	if app.Filename == "" {
		return ""
	}
	path := filepath.Join(s.downloadDir, app.Filename)
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return ""
	}
	return path
}

func newTaskID() string {
	bytes := make([]byte, 12)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("task-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

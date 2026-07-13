package ui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Xynrin/spark-store-tui/internal/domain"
	"github.com/Xynrin/spark-store-tui/internal/preview"
	"github.com/Xynrin/spark-store-tui/internal/provider"
	"github.com/Xynrin/spark-store-tui/internal/system"
)

type focusPane int

const (
	focusSources focusPane = iota
	focusApps
)

type screen int

const (
	screenHome screen = iota
	screenBrowse
)

type appsLoadedMsg struct {
	sourceID string
	apps     []domain.App
	err      error
}

type categoriesLoadedMsg struct {
	sourceID   string
	categories []domain.Category
	err        error
}

type previewLoadedMsg struct {
	appID   string
	image   string
	request int
}

type previewDueMsg struct {
	appID   string
	url     string
	request int
}

type downloadFinishedMsg struct {
	task domain.DownloadTask
	err  error
}

type uninstallFinishedMsg struct{ err error }
type installFinishedMsg struct{ err error }
type updateFinishedMsg struct{ err error }

type packageStatusMsg struct {
	appID  string
	status system.AppPackageStatus
	err    error
}

// Downloader is kept narrow so UI tests and alternate task runners do not
// need to depend on the download implementation.
type Downloader interface {
	Download(context.Context, domain.App) (domain.DownloadTask, error)
	ExistingPath(domain.App) string
}

// ProcessDownloader yields terminal ownership while aptss/aria2 is running so
// users can see native progress and the operation cannot look silently stuck.
type ProcessDownloader interface {
	PrepareDownload(domain.App) (*exec.Cmd, domain.DownloadTask, error)
	FinishDownload(domain.DownloadTask, error) (domain.DownloadTask, error)
}

// Model receives normalized providers rather than source-specific URLs. The
// UI uses one live Spark Store provider and delegates package operations to the
// distribution's official aptss/APM backend.
type Model struct {
	sources              []domain.CatalogSource
	loaders              map[string]provider.CatalogProvider
	downloader           Downloader
	apps                 []domain.App
	host                 system.Host
	screen               screen
	focus                focusPane
	selectedSource       int
	selectedApp          int
	category             string
	categories           []domain.Category
	categoryIndex        int
	width                int
	height               int
	status               string
	loadError            string
	loading              bool
	imageReady           bool
	brandImage           string
	previewImage         string
	previewRequest       int
	searching            bool
	query                string
	descriptionExpanded  bool
	pendingUninstall     bool
	pendingExisting      bool
	pendingInstall       bool
	pendingPackage       string
	deleteAfterUninstall bool
	downloading          bool
	checkingPackage      bool
	pendingUpdate        bool
	packageStatus        system.AppPackageStatus
	packageStatusAppID   string
}

var (
	accent       = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF9D2E"))
	muted        = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C8799"))
	selected     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#243B55")).Bold(true)
	panel        = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#3D4D63")).Padding(1, 2)
	focusedPanel = panel.Copy().BorderForeground(lipgloss.Color("#FF9D2E"))
	header       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	good         = lipgloss.NewStyle().Foreground(lipgloss.Color("#78D64B"))
)

func New(sources []domain.CatalogSource, loaders map[string]provider.CatalogProvider, host system.Host, downloader Downloader) Model {
	return Model{
		sources:    sources,
		loaders:    loaders,
		downloader: downloader,
		host:       host,
		category:   "development",
		width:      120,
		height:     38,
		status:     "按 Enter 加载真实星火商店目录",
		imageReady: system.HasImagePreview(),
		brandImage: preview.Brand(),
	}
}

// WithRecoveredDownloads surfaces restart recovery without automatically
// restarting network work. Any interrupted package can be resumed with D.
func (m Model) WithRecoveredDownloads(tasks []domain.DownloadTask) Model {
	if len(tasks) == 0 {
		return m
	}
	completed := 0
	interrupted := 0
	for _, task := range tasks {
		switch task.Status {
		case domain.TaskCompleted:
			completed++
		case domain.TaskInterrupted:
			interrupted++
		}
	}
	parts := make([]string, 0, 2)
	if interrupted > 0 {
		parts = append(parts, fmt.Sprintf("已恢复 %d 个中断下载，按 D 可继续", interrupted))
	}
	if completed > 0 {
		parts = append(parts, fmt.Sprintf("已确认 %d 个已完成安装包", completed))
	}
	if len(parts) > 0 {
		m.status = strings.Join(parts, "；")
	}
	return m
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = message.Width, message.Height
	case categoriesLoadedMsg:
		if message.sourceID != m.selectedSourceValue().ID {
			return m, nil
		}
		if message.err != nil {
			m.loading = false
			m.loadError = message.err.Error()
			m.status = "分类加载失败"
			return m, nil
		}
		m.categories = message.categories
		m.categoryIndex = categoryIndex(m.categories, m.category)
		if len(m.categories) > 0 {
			m.category = m.categories[m.categoryIndex].ID
		}
		m.status = "正在加载真实应用目录…"
		return m, m.loadApps()
	case appsLoadedMsg:
		if message.sourceID != m.selectedSourceValue().ID {
			return m, nil
		}
		m.loading = false
		m.apps = message.apps
		m.selectedApp = 0
		m.clearPackageStatus()
		m.previewRequest++
		if message.err != nil {
			m.loadError = message.err.Error()
			m.status = "目录加载失败"
			return m, nil
		}
		m.loadError = ""
		m.status = fmt.Sprintf("已加载 %d 个真实应用（%s）", len(m.apps), m.category)
		return m, m.schedulePreview()
	case previewDueMsg:
		if message.request != m.previewRequest || message.appID != m.selectedAppValue().ID {
			return m, nil
		}
		return m, m.loadPreview(message.appID, message.url, message.request)
	case previewLoadedMsg:
		if message.request == m.previewRequest && m.selectedAppValue().ID == message.appID {
			m.previewImage = message.image
		}
	case downloadFinishedMsg:
		m.downloading = false
		if message.err != nil {
			m.status = "下载失败：" + message.err.Error()
		} else {
			m.pendingPackage = message.task.Destination
			m.pendingInstall = true
			m.status = "下载完成：I 或 Enter 安装；Esc 保留安装包"
		}
	case uninstallFinishedMsg:
		if message.err != nil {
			m.status = "卸载失败：" + message.err.Error()
		} else {
			m.status = "卸载完成"
			if m.deleteAfterUninstall && m.pendingPackage != "" {
				if err := os.Remove(m.pendingPackage); err != nil {
					m.status = "已卸载应用，但删除安装包失败：" + err.Error()
				} else {
					m.status = "已卸载应用并删除安装包"
				}
			}
		}
		m.deleteAfterUninstall = false
		m.pendingPackage = ""
	case installFinishedMsg:
		m.pendingInstall = false
		if message.err != nil {
			m.status = "安装失败：" + message.err.Error()
		} else if m.pendingPackage == "" {
			m.status = "安装完成"
		} else {
			m.status = "安装完成；安装包保留在 " + m.pendingPackage
		}
	case packageStatusMsg:
		if message.appID != m.selectedAppValue().ID {
			return m, nil
		}
		m.checkingPackage = false
		m.packageStatusAppID = message.appID
		if message.err != nil {
			m.packageStatusAppID = ""
			m.status = "检查应用更新失败：" + message.err.Error()
			return m, nil
		}
		m.packageStatus = message.status
		if !message.status.Installed {
			m.status = "当前应用尚未安装；按 D 下载并安装"
		} else if message.status.UpdateAvailable {
			m.pendingUpdate = true
			m.status = fmt.Sprintf("发现更新 %s → %s：P 或 Enter 升级；Esc 取消", message.status.InstalledVersion, message.status.CandidateVersion)
		} else {
			m.status = "当前应用已是软件源中的最新版本：" + message.status.InstalledVersion
		}
	case updateFinishedMsg:
		m.pendingUpdate = false
		if message.err != nil {
			m.status = "应用更新失败：" + message.err.Error()
		} else {
			m.packageStatus.Installed = true
			m.packageStatus.InstalledVersion = m.packageStatus.CandidateVersion
			m.packageStatus.UpdateAvailable = false
			m.status = "应用更新完成"
		}
	case tea.KeyPressMsg:
		return m.handleKey(message.String())
	}
	return m, nil
}

func (m Model) handleKey(key string) (tea.Model, tea.Cmd) {
	if key == "q" || key == "ctrl+c" {
		return m, tea.Quit
	}
	if m.searching {
		switch key {
		case "esc":
			m.searching = false
		case "enter":
			m.searching = false
		case "backspace", "ctrl+h":
			m.query = trimLastRune(m.query)
			m.selectedApp = 0
			m.clearPackageStatus()
		default:
			if key == "space" {
				key = " "
			}
			if utf8.RuneCountInString(key) == 1 {
				m.query += key
				m.selectedApp = 0
				m.clearPackageStatus()
			}
		}
		m.previewImage = ""
		m.previewRequest++
		return m, m.schedulePreview()
	}
	if m.pendingUpdate {
		switch key {
		case "p", "enter":
			process, err := system.UpdateProcess(m.host, m.selectedAppValue())
			m.pendingUpdate = false
			if err != nil {
				m.status = "无法更新：" + err.Error()
				return m, nil
			}
			m.status = "正在通过 aptss/apm 更新应用…"
			return m, tea.ExecProcess(process, func(err error) tea.Msg { return updateFinishedMsg{err: err} })
		case "esc":
			m.pendingUpdate = false
			m.status = "已取消应用更新"
			return m, nil
		default:
			m.status = "发现可用更新：P 或 Enter 升级；Esc 取消"
			return m, nil
		}
	}
	if m.pendingUninstall {
		switch key {
		case "u", "enter", "a":
			process, err := system.UninstallProcess(m.host, m.selectedAppValue())
			m.pendingUninstall = false
			if err != nil {
				m.status = "无法卸载：" + err.Error()
				return m, nil
			}
			m.status = "正在执行卸载…"
			m.deleteAfterUninstall = key == "a"
			return m, tea.ExecProcess(process, func(err error) tea.Msg { return uninstallFinishedMsg{err: err} })
		case "x":
			m.pendingUninstall = false
			if m.pendingPackage == "" {
				m.status = "没有本地安装包可删除"
				return m, nil
			}
			if err := os.Remove(m.pendingPackage); err != nil {
				m.status = "删除安装包失败：" + err.Error()
			} else {
				m.status = "已删除安装包"
			}
			m.pendingPackage = ""
			return m, nil
		case "esc":
			m.pendingUninstall = false
			m.status = "已取消卸载"
			return m, nil
		default:
			m.status = "再按 U 或 Enter 确认卸载；Esc 取消"
			return m, nil
		}
	}
	if m.pendingExisting {
		switch key {
		case "i", "enter":
			m.pendingExisting = false
			m.pendingInstall = true
			return m, m.installPackage()
		case "d":
			m.pendingExisting = false
			m.status = "正在重新下载…"
			m.downloading = true
			return m, m.downloadApp()
		case "x":
			m.pendingExisting = false
			if err := os.Remove(m.pendingPackage); err != nil {
				m.status = "删除安装包失败：" + err.Error()
			} else {
				m.status = "已删除安装包，按 D 重新下载"
			}
			m.pendingPackage = ""
			return m, nil
		case "esc":
			m.pendingExisting = false
			m.status = "已取消操作，保留安装包"
			return m, nil
		default:
			m.status = "已存在安装包：I 安装 · D 重新下载 · X 删除 · Esc 取消"
			return m, nil
		}
	}
	if m.pendingInstall {
		switch key {
		case "i", "enter":
			return m, m.installPackage()
		case "esc":
			m.pendingInstall = false
			m.status = "已保留安装包，可稍后按 D 再选择安装"
			return m, nil
		default:
			m.status = "下载完成：I 或 Enter 安装；Esc 保留安装包"
			return m, nil
		}
	}
	if m.screen == screenHome {
		switch key {
		case "enter", "b":
			m.screen = screenBrowse
			m.status = "正在加载星火商店分类…"
			m.loading = true
			return m, m.loadCategories()
		case "?":
			m.status = "Enter/B 浏览应用 · q 退出 · 浏览页中按 Esc 返回简介"
		}
		return m, nil
	}

	switch key {
	case "esc", "h":
		m.screen = screenHome
		m.status = "已返回星火商店简介"
	case "tab", "right", "left":
		if m.focus == focusSources {
			m.focus = focusApps
		} else {
			m.focus = focusSources
		}
	case "up", "k", "down", "j":
		delta := -1
		if key == "down" || key == "j" {
			delta = 1
		}
		if m.focus == focusSources && len(m.sources) > 0 {
			m.selectedSource = wrap(m.selectedSource+delta, len(m.sources))
			m.selectedApp = 0
			m.apps = nil
			m.previewImage = ""
			m.descriptionExpanded = false
			m.clearPackageStatus()
			m.previewRequest++
			m.loading = true
			m.loadError = ""
			m.categories = nil
			m.status = "正在切换目录源…"
			return m, m.loadCategories()
		}
		if m.focus == focusApps && len(m.filteredApps()) > 0 {
			m.selectedApp = wrap(m.selectedApp+delta, len(m.filteredApps()))
			m.previewImage = ""
			m.descriptionExpanded = false
			m.clearPackageStatus()
			m.previewRequest++
			return m, m.schedulePreview()
		}
	case "/":
		m.searching = true
		m.status = "输入关键词过滤应用；Enter 确认，Esc 取消"
	case "r":
		m.loading = true
		m.loadError = ""
		m.status = "正在刷新目录…"
		m.clearPackageStatus()
		return m, m.loadApps()
	case "[", "]":
		if len(m.categories) == 0 {
			m.status = "当前目录源不提供分类"
			return m, nil
		}
		delta := -1
		if key == "]" {
			delta = 1
		}
		m.categoryIndex = wrap(m.categoryIndex+delta, len(m.categories))
		m.category = m.categories[m.categoryIndex].ID
		m.apps = nil
		m.previewImage = ""
		m.descriptionExpanded = false
		m.clearPackageStatus()
		m.previewRequest++
		m.loading = true
		m.status = "正在切换分类…"
		return m, m.loadApps()
	case "enter":
		app := m.selectedAppValue()
		if app.ID == "" {
			m.status = "当前没有可下载应用"
		} else {
			m.status = fmt.Sprintf("已选择 %s；下载任务接入完成后将从详情资产创建", app.Name)
		}
	case "d":
		if m.downloading {
			m.status = "下载任务正在运行"
			return m, nil
		}
		if m.host.Family == "arch" || m.host.Family == "rpm" || m.host.Family == "suse" {
			if m.selectedAppValue().PackageName == "" {
				m.status = "当前应用没有可用于 APM 的包名"
				return m, nil
			}
			m.pendingInstall = true
			m.status = "将通过 Amber APM 安装：I 或 Enter 确认；Esc 取消"
			return m, nil
		}
		if m.selectedAppValue().PackageName == "" || m.selectedAppValue().Filename == "" {
			m.status = "当前应用没有可交给 aptss 的软件包元数据"
			return m, nil
		}
		if m.downloader != nil {
			if existing := m.downloader.ExistingPath(m.selectedAppValue()); existing != "" {
				m.pendingExisting = true
				m.pendingPackage = existing
				m.status = "已存在安装包：I 安装 · D 重新下载 · X 删除 · Esc 取消"
				return m, nil
			}
		}
		m.downloading = true
		m.status = "正在通过 aptss/aria2 下载或续传…"
		return m, m.downloadApp()
	case "p":
		if m.checkingPackage {
			m.status = "正在查询所选应用的安装与更新状态"
			return m, nil
		}
		if m.selectedAppValue().ID == "" {
			m.status = "当前没有可检查更新的应用"
			return m, nil
		}
		m.checkingPackage = true
		m.status = "正在通过 aptss/apm 检查所选应用更新…"
		return m, m.checkPackageStatus()
	case "u":
		if _, err := system.UninstallProcess(m.host, m.selectedAppValue()); err != nil {
			m.status = "无法卸载：" + err.Error()
			return m, nil
		}
		m.pendingUninstall = true
		if m.downloader != nil {
			m.pendingPackage = m.downloader.ExistingPath(m.selectedAppValue())
		}
		m.status = "U/Enter 卸载应用 · X 删除安装包 · A 两者都删除 · Esc 取消"
	case "e":
		m.descriptionExpanded = !m.descriptionExpanded
	case "i":
		if m.imageReady {
			m.status = "图片预览已启用；使用应用图标，截图会在详情扩展后显示"
		} else if command, err := system.ImagePreviewCommand(m.host.Family); err == nil {
			m.status = "执行 --bootstrap-images 安装图片预览依赖：" + command
		} else {
			m.status = "当前发行版尚未配置图片预览依赖；请安装 chafa"
		}
	case "?":
		m.status = "↑↓ 选择 · / 搜索 · D 下载 · P 更新 · U 卸载 · E 简介 · [ ] 分类 · R 刷新 · q 退出"
	}
	return m, nil
}

func (m Model) downloadApp() tea.Cmd {
	app := m.selectedAppValue()
	downloader := m.downloader
	if processDownloader, ok := downloader.(ProcessDownloader); ok {
		process, task, err := processDownloader.PrepareDownload(app)
		if err != nil {
			return func() tea.Msg { return downloadFinishedMsg{task: task, err: err} }
		}
		return tea.ExecProcess(process, func(runErr error) tea.Msg {
			completed, finishErr := processDownloader.FinishDownload(task, runErr)
			return downloadFinishedMsg{task: completed, err: finishErr}
		})
	}
	return func() tea.Msg {
		if downloader == nil {
			return downloadFinishedMsg{err: fmt.Errorf("download service is not configured")}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
		defer cancel()
		task, err := downloader.Download(ctx, app)
		return downloadFinishedMsg{task: task, err: err}
	}
}

func (m Model) checkPackageStatus() tea.Cmd {
	app := m.selectedAppValue()
	host := m.host
	return func() tea.Msg {
		status, err := system.QueryPackageStatus(host, app)
		return packageStatusMsg{appID: app.ID, status: status, err: err}
	}
}

func (m *Model) clearPackageStatus() {
	m.checkingPackage = false
	m.pendingUpdate = false
	m.packageStatus = system.AppPackageStatus{}
	m.packageStatusAppID = ""
}

func (m Model) installPackage() tea.Cmd {
	path := m.pendingPackage
	process, err := system.InstallProcess(m.host, m.selectedAppValue(), path)
	if err != nil {
		return func() tea.Msg { return installFinishedMsg{err: err} }
	}
	return tea.ExecProcess(process, func(err error) tea.Msg { return installFinishedMsg{err: err} })
}

func (m Model) loadApps() tea.Cmd {
	source := m.selectedSourceValue()
	loader := m.loaders[source.ID]
	category := m.category
	architecture := m.host.SparkArchPath
	return func() tea.Msg {
		if loader == nil {
			return appsLoadedMsg{sourceID: source.ID, err: fmt.Errorf("%s 尚未配置仓库", source.Name)}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		apps, err := loader.ListApps(ctx, provider.Query{Architecture: architecture, Category: category, Limit: 40})
		return appsLoadedMsg{sourceID: source.ID, apps: apps, err: err}
	}
}

func (m Model) loadCategories() tea.Cmd {
	source := m.selectedSourceValue()
	loader := m.loaders[source.ID]
	return func() tea.Msg {
		categoryLoader, ok := loader.(provider.CategoryProvider)
		if !ok {
			return categoriesLoadedMsg{sourceID: source.ID, err: fmt.Errorf("%s 尚未配置分类或仓库", source.Name)}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		categories, err := categoryLoader.ListCategories(ctx)
		return categoriesLoadedMsg{sourceID: source.ID, categories: categories, err: err}
	}
}

func (m Model) schedulePreview() tea.Cmd {
	if !m.imageReady || m.width < 80 {
		return nil
	}
	app := m.selectedAppValue()
	url := app.IconURL
	if url == "" && len(app.ScreenshotURLs) > 0 {
		url = app.ScreenshotURLs[0]
	}
	if app.ID == "" || url == "" {
		return nil
	}
	return tea.Tick(350*time.Millisecond, func(time.Time) tea.Msg {
		return previewDueMsg{appID: app.ID, url: url, request: m.previewRequest}
	})
}

func (m Model) loadPreview(appID, url string, request int) tea.Cmd {
	return func() tea.Msg {
		return previewLoadedMsg{appID: appID, image: preview.RenderURL(url), request: request}
	}
}

func wrap(value, count int) int {
	value %= count
	if value < 0 {
		value += count
	}
	return value
}

func (m Model) View() tea.View { return tea.NewView(m.render()) }

func (m Model) render() string {
	if m.screen == screenHome {
		return m.renderHome()
	}
	return m.renderBrowse()
}

func (m Model) renderHome() string {
	previewStatus := muted.Render("终端图片预览：按 I 查看依赖补给方式")
	if m.imageReady {
		previewStatus = good.Render("终端图片预览：已启用（chafa）")
	}
	intro := strings.Join([]string{
		accent.Render("✦ 星火终端助手"),
		header.Render("在终端中浏览、下载和管理 Linux 软件"),
		"公开 metadata 用于展示；Debian 应用交给 aptss 管理。",
		"aptss 负责搜索软件源、镜像下载、校验、安装与更新。",
		"",
		good.Render("Enter  加载并浏览星火商店"),
		muted.Render("B 浏览应用    ? 帮助    q 退出"),
	}, "\n")
	host := fmt.Sprintf("系统：%s\n发行版族：%s\n架构：%s\nSpark 路径：%s", m.host.Distro, m.host.Family, m.host.Architecture, m.host.SparkArchPath)
	introPanel := panel.Width(max(44, m.width-8)).Render(intro)
	if m.brandImage != "" && m.width >= 120 {
		brandPanel := panel.Copy().Width(24).Height(13).Render(accent.Render("星火标识") + "\n\n" + m.brandImage)
		textPanel := panel.Copy().Width(max(58, m.width-36)).Height(13).Render(intro)
		introPanel = lipgloss.JoinHorizontal(lipgloss.Top, brandPanel, " ", textPanel)
	}
	hostPanel := panel.Width(max(44, m.width-8)).Render(accent.Render("本机自动识别") + "\n" + host + "\n\n" + previewStatus)
	return lipgloss.JoinVertical(lipgloss.Left, introPanel, "", hostPanel, "", muted.Render("启动时不会安装软件；明确执行 --bootstrap-images 后才会调用系统包管理器。"), "", accent.Render(m.status))
}

func (m Model) renderBrowse() string {
	title := header.Render("✦ Spark Store TUI") + "  " + muted.Render("真实目录 · 分类："+m.selectedCategoryName()) + "  " + good.Render("● "+m.host.Architecture)
	if m.width < 72 {
		return m.renderNarrowBrowse()
	}
	var content string
	if m.width < 132 {
		content = m.renderCompactBrowse()
	} else {
		content = m.renderWideBrowse()
	}
	footer := muted.Render("↑↓ 导航  / 搜索  D 下载  P 更新  U 卸载  E 简介  [ ] 分类  R 刷新  Esc 首页  q 退出") + "  " + accent.Render(m.status)
	return lipgloss.JoinVertical(lipgloss.Left, title, "", content, "", footer)
}

func (m Model) renderCompactBrowse() string {
	names := make([]string, 0, len(m.sources))
	for index, source := range m.sources {
		name := source.Name
		if index == m.selectedSource {
			name = selected.Render(" " + name + " ")
		}
		names = append(names, name)
	}
	source := m.selectedSourceValue()
	sourcePanel := panel.Width(max(44, m.width-8)).Render(accent.Render("软件源") + "  " + strings.Join(names, "  ") + "\n" + muted.Render(source.Description))
	return lipgloss.JoinVertical(lipgloss.Left, sourcePanel, "", m.renderApps(max(44, m.width-8)), "", m.renderDetails(max(44, m.width-8)))
}

func (m Model) renderNarrowBrowse() string {
	apps := m.filteredApps()
	lines := []string{
		header.Render("✦ Spark Store TUI"),
		muted.Render(m.selectedSourceValue().Name + " · " + m.selectedCategoryName() + " · " + m.host.Architecture),
		m.searchLine(),
		accent.Render(fmt.Sprintf("应用 %d/%d", min(m.selectedApp+1, len(apps)), len(apps))),
	}
	for index, app := range m.windowedApps(apps) {
		actualIndex := m.windowStart(len(apps)) + index
		line := truncate(app.Name+" · "+valueOr(app.Version, "未知"), 28)
		if actualIndex == m.selectedApp {
			line = selected.Render("› " + line)
		}
		lines = append(lines, line)
	}
	app := m.selectedAppValue()
	if app.ID != "" {
		lines = append(lines, "", accent.Render("详情"), truncate(valueOr(app.Description, "暂无简介"), 30))
	}
	lines = append(lines, muted.Render("/ 搜索  D 下载  P 更新  U 卸载  E 简介  [ ] 分类  q 退出"), accent.Render(m.status))
	return strings.Join(lines, "\n")
}

func (m Model) renderWideBrowse() string {
	leftWidth := max(28, m.width/4)
	rightWidth := max(34, m.width/3)
	centerWidth := max(44, m.width-leftWidth-rightWidth-8)
	return lipgloss.JoinHorizontal(lipgloss.Top, m.renderSources(leftWidth), " ", m.renderApps(centerWidth), " ", m.renderDetails(rightWidth))
}

func (m Model) renderSources(width int) string {
	style := panel.Width(width - 4)
	if m.focus == focusSources {
		style = focusedPanel.Width(width - 4)
	}
	lines := []string{accent.Render("软件源")}
	for index, source := range m.sources {
		line := "  " + source.Name
		if index == m.selectedSource {
			line = selected.Width(width - 8).Render("› " + source.Name)
		}
		lines = append(lines, line)
	}
	lines = append(lines, "", accent.Render("自动筛选"), "  分类  "+m.selectedCategoryName(), "  架构  "+m.host.Architecture, "  Spark  "+m.host.SparkArchPath, "", muted.Render("[ / ] 切换分类"), "", muted.Render("镜像策略"), "  Auto / 国内优先 / 海外优先")
	return style.Render(strings.Join(lines, "\n"))
}

func (m Model) renderApps(width int) string {
	style := panel.Width(width - 4)
	if m.focus == focusApps {
		style = focusedPanel.Width(width - 4)
	}
	lines := []string{accent.Render("应用列表")}
	if m.loading {
		lines = append(lines, muted.Render("正在请求公开 metadata…"))
	}
	if m.loadError != "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Render(m.loadError))
	}
	apps := m.filteredApps()
	lines = append(lines, m.searchLine())
	for index, app := range m.windowedApps(apps) {
		actualIndex := m.windowStart(len(apps)) + index
		line := truncate(app.Name+"  "+valueOr(app.Version, "未知")+" · "+valueOr(app.PackageFormat, "包")+" · "+sourceLabel(app.SourceID), max(18, width-10))
		if actualIndex == m.selectedApp {
			line = selected.Width(width - 8).Render("› " + app.Name + "  " + valueOr(app.Version, "未知"))
		}
		lines = append(lines, line)
	}
	if !m.loading && m.loadError == "" && len(apps) == 0 {
		lines = append(lines, muted.Render("当前目录没有可显示的应用。"))
	}
	if len(apps) > 0 {
		lines = append(lines, muted.Render(fmt.Sprintf("%d / %d", m.selectedApp+1, len(apps))))
	}
	return style.Render(strings.Join(lines, "\n"))
}

func (m Model) renderDetails(width int) string {
	app := m.selectedAppValue()
	image := muted.Render("图片预览：等待应用图标")
	if !m.imageReady {
		image = muted.Render("图片预览：按 I 查看 chafa 自动补给命令")
	} else if m.previewImage != "" {
		image = m.previewImage
	}
	description := descriptionForDisplay(valueOr(app.Description, "真实目录接入后显示完整简介。"), width, m.descriptionExpanded)
	descriptionHint := muted.Render("按 E 展开简介")
	if m.descriptionExpanded {
		descriptionHint = muted.Render("按 E 收起简介")
	}
	downloadHint := muted.Render("D 下载 · P 检查更新 · U 卸载")
	if m.downloading {
		downloadHint = accent.Render("下载任务运行中…")
	}
	lines := []string{accent.Render("应用详情"), header.Render(valueOr(app.Name, "加载中")), "版本：" + valueOr(app.Version, "待加载"), "包格式：" + valueOr(app.PackageFormat, "待加载"), "大小：" + valueOr(app.Size, "待加载"), "来源：" + sourceLabel(app.SourceID), m.packageStatusLine(app), "", description, descriptionHint, "", image, "", downloadHint, muted.Render("Debian：aptss/aria2 负责镜像选择、续传和软件源摘要校验")}
	return panel.Width(width - 4).Render(strings.Join(lines, "\n"))
}

func (m Model) packageStatusLine(app domain.App) string {
	if m.checkingPackage && m.packageStatusAppID == "" {
		return "安装状态：正在查询…"
	}
	if app.ID == "" || m.packageStatusAppID != app.ID {
		return "安装状态：按 P 检查更新"
	}
	if !m.packageStatus.Installed {
		return "安装状态：未安装"
	}
	if m.packageStatus.UpdateAvailable {
		return fmt.Sprintf("安装状态：%s，可更新至 %s", m.packageStatus.InstalledVersion, m.packageStatus.CandidateVersion)
	}
	return "安装状态：已是最新 " + m.packageStatus.InstalledVersion
}

func (m Model) selectedAppValue() domain.App {
	apps := m.filteredApps()
	if len(apps) == 0 {
		return domain.App{}
	}
	index := m.selectedApp
	if index >= len(apps) {
		index = 0
	}
	return apps[index]
}

func (m Model) filteredApps() []domain.App {
	if m.query == "" {
		return m.apps
	}
	query := strings.ToLower(m.query)
	apps := make([]domain.App, 0, len(m.apps))
	for _, app := range m.apps {
		content := strings.ToLower(strings.Join([]string{app.Name, app.Description, app.Version, app.PackageFormat}, " "))
		if strings.Contains(content, query) {
			apps = append(apps, app)
		}
	}
	return apps
}

func (m Model) searchLine() string {
	if m.searching {
		return accent.Render("搜索：") + m.query + "_"
	}
	if m.query != "" {
		return accent.Render("筛选：") + m.query + muted.Render("  （按 / 修改）")
	}
	return muted.Render("按 / 搜索应用")
}

func (m Model) windowedApps(apps []domain.App) []domain.App {
	if len(apps) == 0 {
		return nil
	}
	start := m.windowStart(len(apps))
	end := min(start+m.visibleRows(), len(apps))
	return apps[start:end]
}

func (m Model) windowStart(count int) int {
	rows := m.visibleRows()
	if count <= rows {
		return 0
	}
	start := m.selectedApp - rows/2
	if start < 0 {
		return 0
	}
	if start+rows > count {
		return count - rows
	}
	return start
}

func (m Model) visibleRows() int {
	return max(4, min(12, m.height-16))
}

func trimLastRune(value string) string {
	if value == "" {
		return ""
	}
	_, size := utf8.DecodeLastRuneInString(value)
	return value[:len(value)-size]
}

func truncate(value string, limit int) string {
	if limit <= 1 || utf8.RuneCountInString(value) <= limit {
		return value
	}
	return string([]rune(value)[:limit-1]) + "…"
}

func descriptionForDisplay(value string, width int, expanded bool) string {
	value = strings.Join(strings.Fields(value), " ")
	if expanded {
		return truncate(value, 1600)
	}
	limit := max(90, min(260, width*4))
	if utf8.RuneCountInString(value) <= limit {
		return value
	}
	return truncate(value, limit) + " " + muted.Render("…")
}

func (m Model) selectedSourceValue() domain.CatalogSource {
	if len(m.sources) == 0 || m.selectedSource >= len(m.sources) {
		return domain.CatalogSource{}
	}
	return m.sources[m.selectedSource]
}

func (m Model) selectedCategoryName() string {
	if len(m.categories) > 0 && m.categoryIndex < len(m.categories) {
		return m.categories[m.categoryIndex].Name
	}
	return m.category
}

func categoryIndex(categories []domain.Category, id string) int {
	for index, category := range categories {
		if category.ID == id {
			return index
		}
	}
	return 0
}

func sourceLabel(id string) string {
	if id == "" {
		return "-"
	}
	return "Spark"
}

func valueOr(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

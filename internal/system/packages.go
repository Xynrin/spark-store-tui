package system

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/Xynrin/spark-store-tui/internal/domain"
)

// UninstallCommand returns a native package-manager command for installed
// packages. The package name is validated before it can reach a shell.
func UninstallCommand(host Host, app domain.App) (string, error) {
	packageName, err := PackageName(host, app)
	if err != nil {
		return "", err
	}
	switch host.Family {
	case "deb":
		return "sudo aptss remove " + packageName + " -y", nil
	case "arch", "rpm", "suse":
		return "sudo apm remove " + packageName + " -y", nil
	default:
		return "", fmt.Errorf("uninstall is not configured for %s", host.Family)
	}
}

// InstallProcess constructs the native, local-file installation process. It
// does not start the process; the TUI always asks for confirmation first.
func InstallProcess(host Host, app domain.App, packagePath string) (*exec.Cmd, error) {
	switch host.Family {
	case "arch", "rpm", "suse":
		packageName, err := PackageName(host, app)
		if err != nil {
			return nil, err
		}
		if _, err := exec.LookPath("apm"); err != nil {
			return nil, fmt.Errorf("此发行版安装星火应用需要 Amber APM（apm）：%w", err)
		}
		return packageProcess("apm", "install", packageName, "-y"), nil
	}

	if packagePath == "" {
		return nil, fmt.Errorf("package path is required")
	}
	format := strings.ToLower(filepath.Ext(packagePath))
	switch host.Family {
	case "deb":
		if format != ".deb" {
			return nil, fmt.Errorf("Debian family cannot install %s automatically", format)
		}
		aptss, err := exec.LookPath("aptss")
		if err != nil {
			return nil, fmt.Errorf("Debian 系安装星火应用需要 aptss 或 Spark Store：%w", err)
		}
		return packageProcess(aptss, "install", packagePath, "-y"), nil
	case "rpm":
		if format != ".rpm" {
			return nil, fmt.Errorf("RPM family cannot install %s automatically", format)
		}
		return packageProcess("dnf", "install", "-y", packagePath), nil
	case "suse":
		if format != ".rpm" {
			return nil, fmt.Errorf("openSUSE cannot install %s automatically", format)
		}
		return packageProcess("zypper", "--non-interactive", "install", packagePath), nil
	default:
		return nil, fmt.Errorf("installation is not configured for %s", host.Family)
	}
}

func UninstallProcess(host Host, app domain.App) (*exec.Cmd, error) {
	packageName, err := PackageName(host, app)
	if err != nil {
		return nil, err
	}
	switch host.Family {
	case "deb":
		aptss, err := exec.LookPath("aptss")
		if err != nil {
			return nil, fmt.Errorf("Debian 系安装和卸载星火应用需要 aptss 或 Spark Store：%w", err)
		}
		return packageProcess(aptss, "remove", packageName, "-y"), nil
	case "arch", "rpm", "suse":
		if _, err := exec.LookPath("apm"); err != nil {
			return nil, fmt.Errorf("此发行版卸载星火应用需要 Amber APM（apm）：%w", err)
		}
		return packageProcess("apm", "remove", packageName, "-y"), nil
	default:
		return nil, fmt.Errorf("uninstall is not configured for %s", host.Family)
	}
}

// AppPackageStatus is resolved through the same package backend used for
// installation. CandidateVersion is the backend's currently preferred
// version, not merely the version cached in the visual catalog.
type AppPackageStatus struct {
	PackageName      string
	InstalledVersion string
	CandidateVersion string
	Installed        bool
	UpdateAvailable  bool
}

func QueryPackageStatus(host Host, app domain.App) (AppPackageStatus, error) {
	packageName, err := PackageName(host, app)
	if err != nil {
		return AppPackageStatus{}, err
	}
	backend, err := backendPath(host)
	if err != nil {
		return AppPackageStatus{}, err
	}
	command := exec.Command(backend, "policy", packageName)
	command.Env = append(packageEnvironment(), "LC_ALL=C", "LANGUAGE=C")
	output, err := command.CombinedOutput()
	if err != nil {
		detail := strings.TrimSpace(string(output))
		if detail != "" {
			return AppPackageStatus{}, fmt.Errorf("查询 %s 版本失败：%w：%s", packageName, err, detail)
		}
		return AppPackageStatus{}, fmt.Errorf("查询 %s 版本失败：%w", packageName, err)
	}
	return parsePolicyStatus(packageName, string(output)), nil
}

func UpdateProcess(host Host, app domain.App) (*exec.Cmd, error) {
	packageName, err := PackageName(host, app)
	if err != nil {
		return nil, err
	}
	backend, err := backendPath(host)
	if err != nil {
		return nil, err
	}
	return packageProcess(backend, "install", "--only-upgrade", packageName, "-y"), nil
}

func backendPath(host Host) (string, error) {
	name := ""
	switch host.Family {
	case "deb":
		name = "aptss"
	case "arch", "rpm", "suse":
		name = "apm"
	default:
		return "", fmt.Errorf("package backend is not configured for %s", host.Family)
	}
	path, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("找不到应用管理后端 %s：%w", name, err)
	}
	return path, nil
}

func parsePolicyStatus(packageName, output string) AppPackageStatus {
	status := AppPackageStatus{PackageName: packageName}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if value, found := strings.CutPrefix(line, "Installed:"); found {
			status.InstalledVersion = strings.TrimSpace(value)
		}
		if value, found := strings.CutPrefix(line, "Candidate:"); found {
			status.CandidateVersion = strings.TrimSpace(value)
		}
	}
	if status.InstalledVersion == "(none)" {
		status.InstalledVersion = ""
	}
	if status.CandidateVersion == "(none)" {
		status.CandidateVersion = ""
	}
	status.Installed = status.InstalledVersion != ""
	status.UpdateAvailable = status.Installed && status.CandidateVersion != "" && status.CandidateVersion != status.InstalledVersion
	return status
}

// PackageName returns the package-manager name. Spark directory names are not
// authoritative: for example the "vscode" directory publishes a Debian
// package named "code". The filename prefix is therefore preferred.
func PackageName(host Host, app domain.App) (string, error) {
	packageName := app.PackageName
	if host.Family == "deb" || host.Family == "arch" || host.Family == "rpm" || host.Family == "suse" {
		packageName = publishedPackageName(app)
	}
	if packageName == "" {
		return "", fmt.Errorf("%s does not expose a package name", app.Name)
	}
	if !safePackageName(packageName) {
		return "", fmt.Errorf("invalid package name %q", packageName)
	}
	return packageName, nil
}

// Spark's directory name is not always the Debian package name used by APM.
// For example, the vscode directory publishes code_<version>_<arch>.deb and
// Amber exposes it as "code". Debian filenames provide the authoritative name.
func publishedPackageName(app domain.App) string {
	filename := filepath.Base(app.Filename)
	if strings.EqualFold(filepath.Ext(filename), ".deb") {
		if candidate, _, found := strings.Cut(filename, "_"); found && safePackageName(candidate) {
			return candidate
		}
	}
	return app.PackageName
}

// privilegedCommand avoids invoking sudo from a root shell. Besides being
// unnecessary, newer sudo implementations can reject malformed inherited
// desktop-session variables before the package manager is reached.
func privilegedCommand(name string, args ...string) *exec.Cmd {
	return privilegedCommandForEUID(os.Geteuid(), name, args...)
}

func packageProcess(name string, args ...string) *exec.Cmd {
	command := privilegedCommand(name, args...)
	command.Env = packageEnvironment()
	return command
}

// WSL normally appends Windows directories to PATH. Amber APM currently
// forwards PATH through shell/bwrap code that can split entries such as
// "Program Files", so package operations receive Linux paths only. Other
// environment variables remain intact for authentication and locale handling.
func packageEnvironment() []string {
	environment := os.Environ()
	for index, entry := range environment {
		if !strings.HasPrefix(entry, "PATH=") {
			continue
		}
		parts := filepath.SplitList(strings.TrimPrefix(entry, "PATH="))
		linuxPaths := make([]string, 0, len(parts))
		for _, part := range parts {
			if !strings.HasPrefix(part, "/") || strings.HasPrefix(part, "/mnt/") || strings.Contains(part, "\\") {
				continue
			}
			linuxPaths = append(linuxPaths, part)
		}
		if len(linuxPaths) == 0 {
			linuxPaths = []string{"/usr/local/sbin", "/usr/local/bin", "/usr/sbin", "/usr/bin", "/sbin", "/bin"}
		}
		environment[index] = "PATH=" + strings.Join(linuxPaths, string(os.PathListSeparator))
		break
	}
	return environment
}

func privilegedCommandForEUID(euid int, name string, args ...string) *exec.Cmd {
	if euid == 0 {
		return exec.Command(name, args...)
	}
	return exec.Command("sudo", append([]string{name}, args...)...)
}

func safePackageName(value string) bool {
	return value != "" && strings.IndexFunc(value, func(character rune) bool {
		return !(unicode.IsLetter(character) || unicode.IsDigit(character) || strings.ContainsRune(".+:_-", character))
	}) == -1
}

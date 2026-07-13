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
	packageName := packageNameForHost(host, app)
	if packageName == "" {
		return "", fmt.Errorf("%s does not expose a package name", app.Name)
	}
	if !safePackageName(packageName) {
		return "", fmt.Errorf("invalid package name %q", packageName)
	}
	switch host.Family {
	case "deb":
		return "sudo aptss remove " + packageName + " -y", nil
	case "arch", "rpm", "suse":
		return "sudo apm remove -y " + packageName, nil
	default:
		return "", fmt.Errorf("uninstall is not configured for %s", host.Family)
	}
}

// InstallProcess constructs the native, local-file installation process. It
// does not start the process; the TUI always asks for confirmation first.
func InstallProcess(host Host, app domain.App, packagePath string) (*exec.Cmd, error) {
	switch host.Family {
	case "arch", "rpm", "suse":
		packageName := amberPackageName(app)
		if packageName == "" || !safePackageName(packageName) {
			return nil, fmt.Errorf("invalid APM package name %q", packageName)
		}
		if _, err := exec.LookPath("apm"); err != nil {
			return nil, fmt.Errorf("此发行版安装星火应用需要 Amber APM（apm）：%w", err)
		}
		return privilegedCommand("apm", "install", packageName, "-y"), nil
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
		return debianInstallProcess(packagePath, exec.LookPath, defaultSSInstallPaths)
	case "rpm":
		if format != ".rpm" {
			return nil, fmt.Errorf("RPM family cannot install %s automatically", format)
		}
		return privilegedCommand("dnf", "install", "-y", packagePath), nil
	case "suse":
		if format != ".rpm" {
			return nil, fmt.Errorf("openSUSE cannot install %s automatically", format)
		}
		return privilegedCommand("zypper", "--non-interactive", "install", packagePath), nil
	default:
		return nil, fmt.Errorf("installation is not configured for %s", host.Family)
	}
}

func UninstallProcess(host Host, app domain.App) (*exec.Cmd, error) {
	packageName := packageNameForHost(host, app)
	if packageName == "" || !safePackageName(packageName) {
		return nil, fmt.Errorf("invalid package name %q", packageName)
	}
	switch host.Family {
	case "deb":
		aptss, err := exec.LookPath("aptss")
		if err != nil {
			return nil, fmt.Errorf("Debian 系安装和卸载星火应用需要 aptss 或 Spark Store：%w", err)
		}
		return privilegedCommand(aptss, "remove", packageName, "-y"), nil
	case "arch", "rpm", "suse":
		if _, err := exec.LookPath("apm"); err != nil {
			return nil, fmt.Errorf("此发行版卸载星火应用需要 Amber APM（apm）：%w", err)
		}
		return privilegedCommand("apm", "remove", packageName, "-y"), nil
	default:
		return nil, fmt.Errorf("uninstall is not configured for %s", host.Family)
	}
}

var defaultSSInstallPaths = []string{
	"/usr/bin/ssinstall",
	"/usr/local/bin/ssinstall",
	"/opt/durapps/spark-store/bin/ssinstall",
}

// Spark Store's current Debian flow downloads the package from its Metalink
// and delegates the local .deb to ssinstall. ssinstall handles Spark-specific
// dependency and desktop integration; older installations that only expose
// aptss remain supported as a fallback.
func debianInstallProcess(packagePath string, lookPath func(string) (string, error), ssinstallPaths []string) (*exec.Cmd, error) {
	if ssinstall, err := lookPath("ssinstall"); err == nil {
		return privilegedCommand(ssinstall, packagePath), nil
	}
	for _, candidate := range ssinstallPaths {
		info, err := os.Stat(candidate)
		if err == nil && info.Mode().IsRegular() && info.Mode().Perm()&0o111 != 0 {
			return privilegedCommand(candidate, packagePath), nil
		}
	}
	if aptss, err := lookPath("aptss"); err == nil {
		return privilegedCommand(aptss, "install", packagePath, "-y"), nil
	}
	return nil, fmt.Errorf("Debian 系安装星火应用需要 aptss 或 Spark Store；请先安装官方 aptss 后重试")
}

func packageNameForHost(host Host, app domain.App) string {
	if host.Family == "arch" || host.Family == "rpm" || host.Family == "suse" {
		return amberPackageName(app)
	}
	return app.PackageName
}

// Spark's directory name is not always the Debian package name used by APM.
// For example, the vscode directory publishes code_<version>_<arch>.deb and
// Amber exposes it as "code". Debian filenames provide the authoritative name.
func amberPackageName(app domain.App) string {
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

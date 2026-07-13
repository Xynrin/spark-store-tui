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
	packageName := app.PackageName
	if packageName == "" {
		return "", fmt.Errorf("%s does not expose a package name", app.Name)
	}
	if !safePackageName(packageName) {
		return "", fmt.Errorf("invalid package name %q", packageName)
	}
	switch host.Family {
	case "deb":
		return "sudo apt-get remove -y " + packageName, nil
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
		if app.PackageName == "" || !safePackageName(app.PackageName) {
			return nil, fmt.Errorf("invalid APM package name %q", app.PackageName)
		}
		if _, err := exec.LookPath("apm"); err != nil {
			return nil, fmt.Errorf("此发行版安装星火应用需要 Amber APM（apm）：%w", err)
		}
		return privilegedCommand("apm", "install", app.PackageName, "-y"), nil
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
		return privilegedCommand("apt-get", "install", "-y", packagePath), nil
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
	packageName := app.PackageName
	if packageName == "" || !safePackageName(packageName) {
		return nil, fmt.Errorf("invalid package name %q", packageName)
	}
	switch host.Family {
	case "deb":
		return privilegedCommand("apt-get", "remove", "-y", packageName), nil
	case "arch", "rpm", "suse":
		if _, err := exec.LookPath("apm"); err != nil {
			return nil, fmt.Errorf("此发行版卸载星火应用需要 Amber APM（apm）：%w", err)
		}
		return privilegedCommand("apm", "remove", packageName, "-y"), nil
	default:
		return nil, fmt.Errorf("uninstall is not configured for %s", host.Family)
	}
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

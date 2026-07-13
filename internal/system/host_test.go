package system

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Xynrin/spark-store-tui/internal/domain"
)

func TestFamilyAndArchitecture(t *testing.T) {
	if got := Family("ubuntu", "debian"); got != "deb" {
		t.Fatalf("family = %q, want deb", got)
	}
	if got := ArchitectureFromGOARCH("arm64"); got != "aarch64" {
		t.Fatalf("architecture = %q, want aarch64", got)
	}
	if got := ArchitectureFromGOARCH("loong64"); got != "loongarch64" {
		t.Fatalf("architecture = %q, want loongarch64", got)
	}
	if got := StorePath("store", "aarch64"); got != "arm64-store" {
		t.Fatalf("store path = %q", got)
	}
	if got := StorePath("store", "loongarch64"); got != "loong64-store" {
		t.Fatalf("store path = %q, want loong64-store", got)
	}
}

func TestUninstallCommand(t *testing.T) {
	command, err := UninstallCommand(Host{Family: "deb"}, domain.App{Name: "Code", PackageName: "code"})
	if err != nil || command != "sudo aptss remove code -y" {
		t.Fatalf("command = %q, err = %v", command, err)
	}
	if _, err := UninstallCommand(Host{Family: "deb"}, domain.App{PackageName: "code;rm"}); err == nil {
		t.Fatal("expected invalid package name error")
	}
}

func TestInstallProcess(t *testing.T) {
	aptss := fakeExecutable(t, "aptss")
	process, err := InstallProcess(Host{Family: "deb"}, domain.App{PackageName: "code"}, "/tmp/code.deb")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{aptss, "install", "/tmp/code.deb", "-y"}
	if os.Geteuid() != 0 {
		want = append([]string{"sudo"}, want...)
	}
	if got := process.Args; !reflect.DeepEqual(got, want) {
		t.Fatalf("arguments = %q, want %q", got, want)
	}
}

func TestInstallProcessUsesAPMForArch(t *testing.T) {
	fakeExecutable(t, "apm")
	process, err := InstallProcess(Host{Family: "arch"}, domain.App{PackageName: "vscode"}, "")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"apm", "install", "vscode", "-y"}
	if os.Geteuid() != 0 {
		want = append([]string{"sudo"}, want...)
	}
	if got := process.Args; !reflect.DeepEqual(got, want) {
		t.Fatalf("arguments = %q, want %q", got, want)
	}
}

func TestInstallProcessUsesDebianPackageNameForAPM(t *testing.T) {
	fakeExecutable(t, "apm")
	app := domain.App{
		PackageName: "vscode",
		Filename:    "code_1.128.0-1783465401_amd64.deb",
	}
	process, err := InstallProcess(Host{Family: "rpm"}, app, "")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"apm", "install", "code", "-y"}
	if os.Geteuid() != 0 {
		want = append([]string{"sudo"}, want...)
	}
	if got := process.Args; !reflect.DeepEqual(got, want) {
		t.Fatalf("arguments = %q, want %q", got, want)
	}
	if command, err := UninstallCommand(Host{Family: "rpm"}, app); err != nil || command != "sudo apm remove code -y" {
		t.Fatalf("uninstall command = %q, err = %v", command, err)
	}
}

func TestPackageNameUsesDebianFilename(t *testing.T) {
	name, err := PackageName(Host{Family: "deb"}, domain.App{
		PackageName: "vscode", Filename: "code_1.128.0-1783465401_amd64.deb",
	})
	if err != nil || name != "code" {
		t.Fatalf("package name = %q, err = %v", name, err)
	}
}

func TestPrivilegedCommandOnlyUsesSudoForNonRoot(t *testing.T) {
	root := privilegedCommandForEUID(0, "apt-get", "remove", "-y", "vlc")
	if got, want := root.Args, []string{"apt-get", "remove", "-y", "vlc"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("root command = %q, want %q", got, want)
	}

	user := privilegedCommandForEUID(1000, "apt-get", "remove", "-y", "vlc")
	if got, want := user.Args, []string{"sudo", "apt-get", "remove", "-y", "vlc"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("user command = %q, want %q", got, want)
	}
}

func TestDebianInstallProcessRequiresAPTSS(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	process, err := InstallProcess(Host{Family: "deb"}, domain.App{PackageName: "code"}, "/tmp/code.deb")
	if err == nil || process != nil {
		t.Fatalf("process = %+v, err = %v", process, err)
	}
}

func TestUpdateProcessUsesOnlyUpgrade(t *testing.T) {
	aptss := fakeExecutable(t, "aptss")
	process, err := UpdateProcess(Host{Family: "deb"}, domain.App{
		PackageName: "vscode", Filename: "code_1.128.0_amd64.deb",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{aptss, "install", "--only-upgrade", "code", "-y"}
	if os.Geteuid() != 0 {
		want = append([]string{"sudo"}, want...)
	}
	if !reflect.DeepEqual(process.Args, want) {
		t.Fatalf("arguments = %q, want %q", process.Args, want)
	}
}

func TestParsePolicyStatus(t *testing.T) {
	status := parsePolicyStatus("code", "code:\n  Installed: 1.0\n  Candidate: 1.1\n")
	if !status.Installed || !status.UpdateAvailable || status.InstalledVersion != "1.0" || status.CandidateVersion != "1.1" {
		t.Fatalf("status = %+v", status)
	}
	missing := parsePolicyStatus("code", "Installed: (none)\nCandidate: 1.1\n")
	if missing.Installed || missing.UpdateAvailable {
		t.Fatalf("missing status = %+v", missing)
	}
}

func TestPackageEnvironmentDropsWindowsWSLPaths(t *testing.T) {
	t.Setenv("PATH", "/usr/local/bin:/mnt/c/Program Files/PowerShell/7:/usr/bin:C:\\Tools")
	for _, entry := range packageEnvironment() {
		if entry == "PATH=/usr/local/bin:/usr/bin" {
			return
		}
	}
	t.Fatalf("sanitized PATH not found in %q", packageEnvironment())
}

func fakeExecutable(t *testing.T, name string) string {
	t.Helper()
	directory := t.TempDir()
	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", directory)
	return path
}

func TestImagePreviewCommand(t *testing.T) {
	command, err := ImagePreviewCommand("deb")
	if err != nil || command != "sudo apt-get update && sudo apt-get install -y chafa" {
		t.Fatalf("command = %q, err = %v", command, err)
	}
}

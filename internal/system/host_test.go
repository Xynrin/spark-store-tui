package system

import (
	"os"
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
	if err != nil || command != "sudo apt-get remove -y code" {
		t.Fatalf("command = %q, err = %v", command, err)
	}
	if _, err := UninstallCommand(Host{Family: "deb"}, domain.App{PackageName: "code;rm"}); err == nil {
		t.Fatal("expected invalid package name error")
	}
}

func TestInstallProcess(t *testing.T) {
	process, err := InstallProcess(Host{Family: "deb"}, domain.App{}, "/tmp/code.deb")
	if err != nil || process.Path == "" || !contains(process.Args, "install") {
		t.Fatalf("process = %+v, err = %v", process, err)
	}
}

func TestInstallProcessUsesAPMForArch(t *testing.T) {
	bin := t.TempDir()
	apm := bin + "/apm"
	if err := os.WriteFile(apm, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
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
	bin := t.TempDir()
	apm := bin + "/apm"
	if err := os.WriteFile(apm, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
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
	if command, err := UninstallCommand(Host{Family: "rpm"}, app); err != nil || command != "sudo apm remove -y code" {
		t.Fatalf("uninstall command = %q, err = %v", command, err)
	}
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
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

func TestInstallProcessSkipsSudoForCurrentRootUser(t *testing.T) {
	process, err := InstallProcess(Host{Family: "deb"}, domain.App{}, "/tmp/code.deb")
	if err != nil {
		t.Fatal(err)
	}
	if os.Geteuid() == 0 && process.Args[0] == "sudo" {
		t.Fatalf("root process must not invoke sudo: %q", process.Args)
	}
}

func TestImagePreviewCommand(t *testing.T) {
	command, err := ImagePreviewCommand("deb")
	if err != nil || command != "sudo apt-get update && sudo apt-get install -y chafa" {
		t.Fatalf("command = %q, err = %v", command, err)
	}
}

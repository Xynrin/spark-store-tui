package ui

import (
	"strings"
	"testing"

	"github.com/Xynrin/spark-store-tui/internal/domain"
	"github.com/Xynrin/spark-store-tui/internal/system"
)

func TestFilteredAppsMatchesNameAndDescription(t *testing.T) {
	model := Model{
		apps: []domain.App{
			{Name: "Visual Studio Code", Description: "Code editor"},
			{Name: "VLC", Description: "Media player"},
		},
		query: "editor",
	}
	apps := model.filteredApps()
	if len(apps) != 1 || apps[0].Name != "Visual Studio Code" {
		t.Fatalf("filtered apps: %+v", apps)
	}
}

func TestWindowedAppsCentersSelection(t *testing.T) {
	model := Model{height: 22, selectedApp: 8}
	apps := make([]domain.App, 20)
	for index := range apps {
		apps[index].Name = string(rune('a' + index))
	}
	window := model.windowedApps(apps)
	if len(window) != 6 || window[0].Name != "f" {
		t.Fatalf("window: %+v", window)
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("abcdefgh", 5); got != "abcd…" {
		t.Fatalf("truncate = %q", got)
	}
}

func TestRecoveredDownloadsLeaveModelOperable(t *testing.T) {
	model := New(nil, nil, system.Host{}, nil).WithRecoveredDownloads([]domain.DownloadTask{
		{Status: domain.TaskInterrupted},
		{Status: domain.TaskCompleted},
	})
	if !strings.Contains(model.status, "按 D 可继续") || model.downloading {
		t.Fatalf("recovery status = %q, downloading = %v", model.status, model.downloading)
	}
}

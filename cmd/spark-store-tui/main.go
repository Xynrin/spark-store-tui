package main

import (
	"flag"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/Xynrin/spark-store-tui/internal/download"
	"github.com/Xynrin/spark-store-tui/internal/provider"
	"github.com/Xynrin/spark-store-tui/internal/state"
	"github.com/Xynrin/spark-store-tui/internal/system"
	"github.com/Xynrin/spark-store-tui/internal/ui"
)

func main() {
	bootstrapImages := flag.Bool("bootstrap-images", false, "install the optional terminal image preview dependency for this distribution")
	flag.Parse()

	host := system.Detect()
	if *bootstrapImages {
		if system.HasImagePreview() {
			fmt.Println("terminal image preview dependency is already available")
			return
		}
		if err := system.RunImagePreviewBootstrap(host); err != nil {
			fmt.Fprintln(os.Stderr, "sparkstore:", err)
			os.Exit(1)
		}
		return
	}

	sources := provider.BuiltinSources()
	loaders := map[string]provider.CatalogProvider{
		"spark-store": provider.SparkMetadataProvider{
			Catalog:         sources[0],
			DefaultArchPath: host.SparkArchPath,
		},
	}
	// Bubble Tea v2 owns rendering and terminal-mode negotiation. Keeping the
	// default renderer also makes the first milestone usable in basic terminals.
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	tasks := state.NewTaskStore(configDir + string(os.PathSeparator) + "sparkstore" + string(os.PathSeparator) + "tasks.json")
	downloader := download.NewService(nil, tasks, download.DefaultDownloadDir())
	recovered, recoveryErr := downloader.RecoverInterrupted()
	if recoveryErr != nil {
		fmt.Fprintln(os.Stderr, "sparkstore: could not recover download state:", recoveryErr)
	}
	program := tea.NewProgram(ui.New(sources, loaders, host, downloader).WithRecoveredDownloads(recovered))

	if _, err := program.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "sparkstore:", err)
		os.Exit(1)
	}
}

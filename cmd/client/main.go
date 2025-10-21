package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/cmrd-a/GophKeeper/client"
	"github.com/cmrd-a/GophKeeper/cmd/client/tui"
)

var (
	// Build information (set via ldflags during build)
	version   = "dev"
	buildDate = "unknown"

	// CLI flags
	showVersion = flag.Bool("version", false, "Show version information")
)

func main() {
	flag.Parse()

	// Show version and exit if requested
	if *showVersion {
		fmt.Printf("GophKeeper Client\n")
		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Build Date: %s\n", buildDate)
		os.Exit(0)
	}
	serverAddr := "localhost:8082"
	if addr := os.Getenv("GOPHKEEPER_SERVER"); addr != "" {
		serverAddr = addr
	}

	slog.Info("Starting GophKeeper client", "serverAddr", serverAddr)
	slog.Info("Server should be running", "command", "go run ./cmd/server")

	// Create client configuration
	config := &client.ClientConfig{
		ServerAddr: serverAddr,
	}

	// Create client using the new client package
	gophClient, err := client.NewClient(config)
	if err != nil {
		slog.Error("Failed to create client", "error", err)
		os.Exit(1)
	}
	defer gophClient.Close()

	// Create and run the TUI application
	app := tui.NewApp(gophClient)
	app.SetBuildInfo(version, buildDate)

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		slog.Error("Failed to run TUI", "error", err)
		os.Exit(1)
	}
}

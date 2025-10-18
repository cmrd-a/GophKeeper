package main

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/cmrd-a/GophKeeper/client"
	"github.com/cmrd-a/GophKeeper/cmd/client/tui"
)

func main() {
	serverAddr := "localhost:8082"
	if addr := os.Getenv("GOPHKEEPER_SERVER"); addr != "" {
		serverAddr = addr
	}

	log.Printf("Starting GophKeeper client, connecting to server at %s", serverAddr)
	log.Printf("Make sure the server is running with: go run ./cmd/server")

	// Create client configuration
	config := &client.ClientConfig{
		ServerAddr: serverAddr,
	}

	// Create client using the new client package
	gophClient, err := client.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer gophClient.Close()

	// Create and run the TUI application
	app := tui.NewApp(gophClient)

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("Failed to run TUI: %v", err)
	}
}

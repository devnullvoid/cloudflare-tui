package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cloudflare/cloudflare-go"
	"github.com/devnullvoid/cloudflare-tui/internal/ui"
)

func main() {
	// 1. Authentication Check
	token := os.Getenv("CLOUDFLARE_API_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, ui.ErrStyle.Render("Error: CLOUDFLARE_API_TOKEN environment variable is required."))
		os.Exit(1)
	}

	// 2. Initialize Cloudflare Client
	api, err := cloudflare.NewWithAPIToken(token)
	if err != nil {
		log.Fatalf("Failed to initialize Cloudflare client: %v", err)
	}

	// 3. Setup Initial Bubble Tea Model
	m := ui.InitialModel(api)

	// 4. Run the Program
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

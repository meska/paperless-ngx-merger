package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/meska/paperless-merger/internal/config"
	"github.com/meska/paperless-merger/internal/ui"
)

var version = "dev"

func main() {
	// Carica o crea la configurazione
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Errore nel caricamento della configurazione: %v\n", err)
		os.Exit(1)
	}

	// Se la configurazione non esiste, mostra il setup iniziale
	if cfg.BaseURL == "" || cfg.APIKey == "" {
		p := tea.NewProgram(ui.NewSetupModel(cfg))
		if _, err := p.Run(); err != nil {
			fmt.Printf("Errore: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Avvia l'applicazione principale
	p := tea.NewProgram(ui.NewMainModel(cfg))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Errore: %v\n", err)
		os.Exit(1)
	}
}

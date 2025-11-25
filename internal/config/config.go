package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config rappresenta la configurazione dell'applicazione
type Config struct {
	BaseURL  string `json:"base_url"`
	APIKey   string `json:"api_key"`
	Language string `json:"language"` // "auto", "en", "it"
}

// GetConfigPath restituisce il percorso del file di configurazione
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	
	configDir := filepath.Join(homeDir, ".config", "paperless-merger")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}
	
	return filepath.Join(configDir, "config.json"), nil
}

// Load carica la configurazione dal file
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// Se il file non esiste, restituisce una configurazione vuota
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("errore nella lettura del file di configurazione: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("errore nel parsing del file di configurazione: %w", err)
	}

	return &cfg, nil
}

// Save salva la configurazione nel file
func (c *Config) Save() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("errore nella serializzazione della configurazione: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("errore nel salvataggio della configurazione: %w", err)
	}

	return nil
}

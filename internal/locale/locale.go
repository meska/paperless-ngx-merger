package locale

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed *.json
var localeFS embed.FS

// Localizer gestisce le traduzioni dell'applicazione
type Localizer struct {
	bundle    *i18n.Bundle
	localizer *i18n.Localizer
	lang      string
}

// New crea un nuovo Localizer con la lingua specificata
// lang può essere "en", "it", o "auto" per detection automatica
func New(lang string) (*Localizer, error) {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// Carica i file di traduzione embedded
	if _, err := bundle.LoadMessageFileFS(localeFS, "en.json"); err != nil {
		return nil, fmt.Errorf("failed to load en.json: %w", err)
	}
	if _, err := bundle.LoadMessageFileFS(localeFS, "it.json"); err != nil {
		return nil, fmt.Errorf("failed to load it.json: %w", err)
	}

	// Determina la lingua da usare
	actualLang := lang
	if lang == "auto" {
		actualLang = detectLanguage()
	}

	localizer := i18n.NewLocalizer(bundle, actualLang)

	return &Localizer{
		bundle:    bundle,
		localizer: localizer,
		lang:      actualLang,
	}, nil
}

// T traduce una chiave nella lingua configurata
func (l *Localizer) T(key string) string {
	msg, err := l.localizer.Localize(&i18n.LocalizeConfig{
		MessageID: key,
	})
	if err != nil {
		// Fallback: ritorna la chiave se la traduzione non esiste
		return key
	}
	return msg
}

// GetLanguage ritorna la lingua attualmente in uso
func (l *Localizer) GetLanguage() string {
	return l.lang
}

// detectLanguage rileva la lingua del sistema operativo
func detectLanguage() string {
	// Su macOS, prova prima con le impostazioni di sistema
	if lang := detectMacOSLanguage(); lang != "" {
		return lang
	}
	
	// Controlla LANG environment variable
	envLang := os.Getenv("LANG")
	if envLang == "" {
		envLang = os.Getenv("LC_ALL")
	}
	if envLang == "" {
		envLang = os.Getenv("LC_MESSAGES")
	}

	// Parse della lingua (es. "it_IT.UTF-8" -> "it")
	if len(envLang) >= 2 {
		langCode := envLang[:2]
		// Supportiamo solo en e it per ora
		if langCode == "it" {
			return "it"
		}
		if langCode == "en" {
			return "en"
		}
	}

	// Default a inglese
	return "en"
}

// detectMacOSLanguage rileva la lingua su macOS usando defaults
func detectMacOSLanguage() string {
	// Esegue: defaults read -g AppleLanguages
	// Output tipico:
	// (
	//     "it-IT",
	//     "en-IT"
	// )
	cmd := exec.Command("defaults", "read", "-g", "AppleLanguages")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Cerca la prima lingua nella lista
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Rimuovi virgolette e spazi
		line = strings.Trim(line, `",; `)
		
		// Controlla se inizia con it
		if strings.HasPrefix(strings.ToLower(line), "it") {
			return "it"
		}
		// Controlla se inizia con en (ma non è it-EN o simili)
		if strings.HasPrefix(strings.ToLower(line), "en") && !strings.Contains(strings.ToLower(line), "it") {
			return "en"
		}
	}
	
	return ""
}

// GetAvailableLanguages ritorna la lista delle lingue disponibili
func GetAvailableLanguages() []string {
	return []string{"en", "it"}
}

package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/meska/paperless-merger/internal/config"
	"github.com/meska/paperless-merger/internal/locale"
	"github.com/meska/paperless-merger/internal/paperless"
)

// SetupModel rappresenta il modello per la configurazione iniziale
type SetupModel struct {
	config       *config.Config
	localizer    *locale.Localizer
	inputs       []textinput.Model
	focused      int
	langCursor   int  // cursore per selezione lingua
	err          error
	quitting     bool
	returnToMain bool // indica se tornare al main menu invece di uscire
}

// NewSetupModel crea un nuovo modello di setup
func NewSetupModel(cfg *config.Config) SetupModel {
	// Se non c'è una lingua configurata, usa auto-detect di default
	if cfg.Language == "" {
		cfg.Language = "auto"
	}
	
	// Crea localizer con la lingua configurata o auto-detect
	loc, err := locale.New(cfg.Language)
	if err != nil {
		// Fallback a inglese in caso di errore
		loc, _ = locale.New("en")
	}
	
	inputs := make([]textinput.Model, 2)

	// Input per URL
	inputs[0] = textinput.New()
	inputs[0].Placeholder = loc.T("setup.url_placeholder")
	inputs[0].Focus()
	inputs[0].CharLimit = 200
	inputs[0].Width = 50
	if cfg.BaseURL != "" {
		inputs[0].SetValue(cfg.BaseURL)
	}

	// Input per API Key
	inputs[1] = textinput.New()
	inputs[1].Placeholder = loc.T("setup.apikey_placeholder")
	inputs[1].CharLimit = 200
	inputs[1].Width = 50
	inputs[1].EchoMode = textinput.EchoPassword
	inputs[1].EchoCharacter = '•'
	if cfg.APIKey != "" {
		inputs[1].SetValue(cfg.APIKey)
	}

	// Determina langCursor basato sulla lingua configurata
	langCursor := 0
	switch cfg.Language {
	case "auto":
		langCursor = 0
	case "en":
		langCursor = 1
	case "it":
		langCursor = 2
	}
	
	// Se c'è già una config, significa che stiamo modificando settings
	returnToMain := cfg.BaseURL != "" && cfg.APIKey != ""

	return SetupModel{
		config:       cfg,
		localizer:    loc,
		inputs:       inputs,
		focused:      0,
		langCursor:   langCursor,
		returnToMain: returnToMain,
	}
}

func (m SetupModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m SetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit

		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Se siamo nel campo lingua (focused == 2)
			if m.focused == 2 {
				if s == "up" || s == "k" {
					if m.langCursor > 0 {
						m.langCursor--
					}
					return m, nil
				} else if s == "down" || s == "j" {
					if m.langCursor < 2 { // 3 opzioni: auto, en, it
						m.langCursor++
					}
					return m, nil
				} else if s == "enter" {
					// Salva la lingua selezionata e passa al salvataggio
					return m.saveAndQuit()
				} else if s == "shift+tab" || s == "up" {
					// Torna al campo precedente
					m.focused = 1
					return m, m.inputs[1].Focus()
				}
				return m, nil
			}

			// Se premiamo enter sull'ultimo input, passiamo alla selezione lingua
			if s == "enter" && m.focused == len(m.inputs)-1 {
				m.focused = 2 // campo lingua
				// Sfoca l'ultimo input
				m.inputs[m.focused-2].Blur()
				return m, nil
			}

			// Altrimenti navighiamo tra i campi
			if s == "up" || s == "shift+tab" {
				m.focused--
			} else {
				m.focused++
			}

			// Limita al range degli input
			if m.focused > len(m.inputs)-1 {
				m.focused = 0
			} else if m.focused < 0 {
				m.focused = len(m.inputs) - 1
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focused {
					cmds[i] = m.inputs[i].Focus()
				} else {
					m.inputs[i].Blur()
				}
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Aggiorna l'input corrente solo se non siamo nel campo lingua
	if m.focused < 2 {
		cmd := m.updateInputs(msg)
		return m, cmd
	}
	
	return m, nil
}

func (m *SetupModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m SetupModel) View() string {
	if m.quitting {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true)

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)

	var s string
	s += titleStyle.Render(m.localizer.T("setup.title")) + "\n\n"
	s += labelStyle.Render(m.localizer.T("setup.welcome")) + "\n\n"

	s += labelStyle.Render(m.localizer.T("setup.url_label")) + "\n"
	s += m.inputs[0].View() + "\n\n"

	s += labelStyle.Render(m.localizer.T("setup.apikey_label")) + "\n"
	s += m.inputs[1].View() + "\n\n"

	// Selezione lingua
	s += labelStyle.Render(m.localizer.T("setup.language_label")) + "\n"
	
	langOptions := []string{
		m.localizer.T("setup.language_auto"),
		m.localizer.T("setup.language_en"),
		m.localizer.T("setup.language_it"),
	}
	
	for i, opt := range langOptions {
		cursor := " "
		if m.focused == 2 && i == m.langCursor {
			cursor = ">"
			s += selectedStyle.Render(fmt.Sprintf("%s %s", cursor, opt)) + "\n"
		} else {
			s += labelStyle.Render(fmt.Sprintf("%s %s", cursor, opt)) + "\n"
		}
	}
	s += "\n"

	if m.err != nil {
		s += errorStyle.Render(fmt.Sprintf(m.localizer.T("setup.error"), m.err)) + "\n\n"
	}

	s += labelStyle.Render(m.localizer.T("setup.help")) + "\n"

	return s
}

func (m SetupModel) saveAndQuit() (tea.Model, tea.Cmd) {
	m.config.BaseURL = m.inputs[0].Value()
	m.config.APIKey = m.inputs[1].Value()
	
	// Salva la lingua selezionata
	switch m.langCursor {
	case 0:
		m.config.Language = "auto"
	case 1:
		m.config.Language = "en"
	case 2:
		m.config.Language = "it"
	}

	// Testa la connessione
	client := paperless.NewClient(m.config.BaseURL, m.config.APIKey)
	if err := client.TestConnection(); err != nil {
		m.err = fmt.Errorf(m.localizer.T("setup.connection_failed"), err)
		return m, nil
	}

	// Salva la configurazione
	if err := m.config.Save(); err != nil {
		m.err = err
		return m, nil
	}

	// Se dobbiamo tornare al main menu, ricarica il localizer con la nuova lingua
	if m.returnToMain {
		loc, err := locale.New(m.config.Language)
		if err != nil {
			loc, _ = locale.New("en")
		}
		mainModel := NewMainModel(m.config)
		mainModel.localizer = loc
		// Ricarica le scelte con la nuova lingua
		mainModel.choices = []string{
			loc.T("main.entity_tags"),
			loc.T("main.entity_correspondents"),
			loc.T("main.entity_doctypes"),
		}
		return mainModel, nil
	}

	m.quitting = true
	return m, tea.Quit
}

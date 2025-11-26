package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/meska/paperless-merger/internal/config"
	"github.com/meska/paperless-merger/internal/locale"
)

// EntityType rappresenta il tipo di entità da gestire
type EntityType int

const (
	EntityTags EntityType = iota
	EntityCorrespondents
	EntityDocumentTypes
)

// MergeMode rappresenta la modalità di merge
type MergeMode int

const (
	ModeSemiAutomatic MergeMode = iota
	ModeManual
)

// MainModel rappresenta il modello principale dell'applicazione
type MainModel struct {
	config       *config.Config
	localizer    *locale.Localizer
	cursor       int
	choices      []string
	selected     EntityType
	mergeMode    MergeMode
	quitting     bool
	showList     bool
	showModeMenu bool
	listModel    *ListModel
}

// NewMainModel crea un nuovo modello principale
func NewMainModel(cfg *config.Config) MainModel {
	// Inizializza localizer con la lingua configurata
	loc, err := locale.New(cfg.Language)
	if err != nil {
		// Fallback a inglese
		loc, _ = locale.New("en")
	}
	
	return MainModel{
		config:    cfg,
		localizer: loc,
		cursor:    0,
		choices: []string{
			loc.T("main.entity_tags"),
			loc.T("main.entity_correspondents"),
			loc.T("main.entity_doctypes"),
		},
		showModeMenu: true,
	}
}

func (m MainModel) Init() tea.Cmd {
	return nil
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.showList && m.listModel != nil {
		// Delega alla vista lista
		newModel, cmd := m.listModel.Update(msg)
		
		if listModel, ok := newModel.(ListModel); ok {
			m.listModel = &listModel
			if listModel.quitting {
				m.showList = false
				m.listModel = nil
				return m, nil
			}
			return m, cmd
		}
		
		return newModel, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit

		case "s", "S":
			// Apri settings (torna al setup)
			if m.showModeMenu {
				// Solo dalla prima schermata (menu modalità)
				setupModel := NewSetupModel(m.config)
				return setupModel, setupModel.Init()
			}

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case "enter", " ":
			if m.showModeMenu {
				// Menu principale - prima scelta
				if m.cursor == 0 {
					// 🤖 Semi-automatica - vai a selezione entità
					m.mergeMode = ModeSemiAutomatic
					m.showModeMenu = false
					m.cursor = 0
				} else if m.cursor == 1 {
					// ✋ Manuale - vai a selezione entità
				m.mergeMode = ModeManual
				m.showModeMenu = false
				m.cursor = 0
			}
			} else {
				// Seleziona entità per merge
				m.selected = EntityType(m.cursor)
				m.showList = true
				listModel := NewListModel(m.config, m.localizer, m.selected, m.mergeMode)
				m.listModel = &listModel
				return m, listModel.Init()
			}
		}
	}

	return m, nil
}

func (m MainModel) View() string {
	if m.showList && m.listModel != nil {
		return m.listModel.View()
	}
	
	if m.quitting {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		MarginBottom(1)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	s := titleStyle.Render(m.localizer.T("main.title")) + "\n\n"

	if m.showModeMenu {
		// Menu principale
		s += normalStyle.Render(m.localizer.T("main.select_mode")) + "\n\n"
		
		modeChoices := []string{
			m.localizer.T("main.mode_semiauto"),
			m.localizer.T("main.mode_manual"),
		}
		
		for i, choice := range modeChoices {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
				s += selectedStyle.Render(fmt.Sprintf("%s %s", cursor, choice)) + "\n"
			} else {
				s += normalStyle.Render(fmt.Sprintf("%s %s", cursor, choice)) + "\n"
			}
		}
	} else {
		// Menu selezione entità
		var modeDesc string
		if m.mergeMode == ModeSemiAutomatic {
			modeDesc = fmt.Sprintf(m.localizer.T("main.mode_current"), m.localizer.T("main.mode_semiauto_label"))
		} else {
			modeDesc = fmt.Sprintf(m.localizer.T("main.mode_current"), m.localizer.T("main.mode_manual_label"))
		}
		s += normalStyle.Render(modeDesc) + "\n\n"
		s += normalStyle.Render(m.localizer.T("main.select_entity")) + "\n\n"

		for i, choice := range m.choices {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
				s += selectedStyle.Render(fmt.Sprintf("%s %s", cursor, choice)) + "\n"
			} else {
				s += normalStyle.Render(fmt.Sprintf("%s %s", cursor, choice)) + "\n"
			}
		}
	}

	s += "\n" + normalStyle.Render(m.localizer.T("main.help")) + "\n"

	return s
}

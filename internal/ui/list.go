package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/meska/paperless-merger/internal/config"
	"github.com/meska/paperless-merger/internal/locale"
	"github.com/meska/paperless-merger/internal/paperless"
	"github.com/meska/paperless-merger/internal/similarity"
)

// ListModel rappresenta il modello per la lista di elementi
type ListModel struct {
	config        *config.Config
	localizer     *locale.Localizer
	entityType    EntityType
	mergeMode     MergeMode
	client        *paperless.Client
	groups        []similarity.SimilarityGroup
	allItems      []similarity.SimilarItem // Tutti gli elementi (per modalità manuale)
	filteredItems []similarity.SimilarItem // Elementi filtrati dalla search
	cursor        int
	groupCursor   int
	selectedMap   map[int]bool // ID -> selezionato
	loading       bool
	merging       bool          // Stato durante il merge
	mergeStatus   string        // Messaggio di stato del merge
	mergeProgress float64       // Progresso merge (0.0-1.0)
	mergeTotal    int           // Numero totale operazioni
	mergeCurrent  int           // Operazione corrente
	progressChan  chan tea.Msg  // Canale per aggiornamenti progress
	err           error
	quitting      bool
	mode          string // "browse", "select", "merge", "manual" (per modalità manuale)
	mergeInput    textinput.Model
	searchInput   textinput.Model // Per filtrare nella modalità manuale
	progress      progress.Model
	currentGroup  *similarity.SimilarityGroup
	width         int // Larghezza del terminale
	height        int // Altezza del terminale
}

type loadedMsg struct {
	groups   []similarity.SimilarityGroup
	allItems []similarity.SimilarItem
	err      error
}

type mergeCompleteMsg struct {
	err error
}

type mergeProgressMsg struct {
	current int
	total   int
	status  string
}

// waitForProgress è un comando che legge dal canale di progress
func waitForProgress(progressChan chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return <-progressChan
	}
}

// NewListModel crea un nuovo modello lista
func NewListModel(cfg *config.Config, loc *locale.Localizer, entityType EntityType, mergeMode MergeMode) ListModel {
	client := paperless.NewClient(cfg.BaseURL, cfg.APIKey)
	
	input := textinput.New()
	input.Placeholder = loc.T("list.merge_input_placeholder")
	input.CharLimit = 200
	input.Width = 50

	searchInput := textinput.New()
	searchInput.Placeholder = loc.T("list.merge_search_placeholder")
	searchInput.CharLimit = 100
	searchInput.Width = 50

	prog := progress.New(progress.WithDefaultGradient())
	prog.Width = 50

	initialMode := "browse"
	if mergeMode == ModeManual {
		initialMode = "manual"
	}

	return ListModel{
		config:      cfg,
		localizer:   loc,
		entityType:  entityType,
		mergeMode:   mergeMode,
		client:      client,
		selectedMap: make(map[int]bool),
		loading:     true,
		mode:        initialMode,
		mergeInput:  input,
		searchInput: searchInput,
		progress:    prog,
		width:       80,  // Valori di default ragionevoli
		height:      24,  // Saranno aggiornati dal WindowSizeMsg
	}
}

func (m ListModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadData,
	)
}

func (m ListModel) loadData() tea.Msg {
	var items []similarity.SimilarItem
	var err error

	switch m.entityType {
	case EntityTags:
		tags, err := m.client.GetTags()
		if err != nil {
			return loadedMsg{err: err}
		}
		items = make([]similarity.SimilarItem, len(tags))
		for i, tag := range tags {
			items[i] = similarity.SimilarItem{ID: tag.ID, Name: tag.Name}
		}

	case EntityCorrespondents:
		correspondents, err := m.client.GetCorrespondents()
		if err != nil {
			return loadedMsg{err: err}
		}
		items = make([]similarity.SimilarItem, len(correspondents))
		for i, corr := range correspondents {
			items[i] = similarity.SimilarItem{ID: corr.ID, Name: corr.Name}
		}

	case EntityDocumentTypes:
		docTypes, err := m.client.GetDocumentTypes()
		if err != nil {
			return loadedMsg{err: err}
		}
		items = make([]similarity.SimilarItem, len(docTypes))
		for i, dt := range docTypes {
			items[i] = similarity.SimilarItem{ID: dt.ID, Name: dt.Name}
		}
	}

	if err != nil {
		return loadedMsg{err: err}
	}

	// Salva tutti gli elementi per modalità manuale
	allItems := items

	// Trova gruppi simili (soglia 0.7 = 70% similarità) solo per modalità semi-automatica
	var groups []similarity.SimilarityGroup
	if len(items) > 0 {
		groups = similarity.FindSimilarGroups(items, 0.7)
	}

	return loadedMsg{groups: groups, allItems: allItems}
}

func (m ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case loadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.groups = msg.groups
		m.allItems = msg.allItems
		m.filteredItems = msg.allItems // Inizialmente tutti visibili
		if m.mergeMode == ModeManual {
			m.searchInput.Focus()
		}
		return m, nil

	case mergeProgressMsg:
		m.mergeCurrent = msg.current
		m.mergeTotal = msg.total
		m.mergeStatus = msg.status
		if msg.total > 0 {
			m.mergeProgress = float64(msg.current) / float64(msg.total)
		}
		// Continua ad ascoltare aggiornamenti dal canale
		return m, waitForProgress(m.progressChan)

	case mergeCompleteMsg:
		m.merging = false
		m.mergeStatus = ""
		m.mergeProgress = 0
		m.mergeCurrent = 0
		m.mergeTotal = 0
		if msg.err != nil {
			m.err = msg.err
			m.mode = "select" // Torna indietro in caso di errore
			return m, nil
		}
		// Merge completato con successo, ricarica i dati
		if m.mergeMode == ModeManual {
			m.mode = "manual"
		} else {
			m.mode = "browse"
		}
		m.selectedMap = make(map[int]bool)
		m.currentGroup = nil
		m.loading = true
		return m, m.loadData

	case tea.KeyMsg:
		// Non processare input durante il merge
		if m.merging {
			return m, nil
		}
		
		if m.mode == "merge" {
			return m.updateMergeMode(msg)
		} else if m.mode == "select" {
			return m.updateSelectMode(msg)
		} else if m.mode == "manual" {
			return m.updateManualMode(msg)
		}
		return m.updateBrowseMode(msg)
	}

	return m, nil
}

func (m ListModel) updateBrowseMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	
	case "q", "esc":
		// Torna al menu principale invece di uscire
		m.quitting = true
		return m, nil

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.groups)-1 {
			m.cursor++
		}

	case "enter", " ":
		if len(m.groups) > 0 {
			m.mode = "select"
			m.currentGroup = &m.groups[m.cursor]
			m.groupCursor = 0
			// Pre-seleziona tutti gli elementi del gruppo
			for _, item := range m.currentGroup.Items {
				m.selectedMap[item.ID] = true
			}
		}
	}

	return m, nil
}

func (m ListModel) updateSelectMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "esc":
		// Torna alla modalità browse
		m.mode = "browse"
		m.selectedMap = make(map[int]bool)
		m.currentGroup = nil
		return m, nil

	case "up", "k":
		if m.groupCursor > 0 {
			m.groupCursor--
		}

	case "down", "j":
		if m.currentGroup != nil && m.groupCursor < len(m.currentGroup.Items)-1 {
			m.groupCursor++
		}

	case " ":
		if m.currentGroup != nil {
			item := m.currentGroup.Items[m.groupCursor]
			m.selectedMap[item.ID] = !m.selectedMap[item.ID]
		}

	case "enter":
		// Passa alla modalità merge
		m.mode = "merge"
		// Imposta il nome dell'elemento attualmente selezionato come valore di default
		if m.currentGroup != nil && m.groupCursor < len(m.currentGroup.Items) {
			m.mergeInput.SetValue(m.currentGroup.Items[m.groupCursor].Name)
		}
		return m, m.mergeInput.Focus()
	}

	return m, nil
}

func (m ListModel) updateMergeMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "esc":
		// Torna alla modalità precedente
		if m.mergeMode == ModeManual {
			m.mode = "manual"
		} else {
			m.mode = "select"
		}
		m.mergeInput.Blur()
		return m, nil

	case "enter":
		// Esegui il merge
		m.mergeInput.Blur()
		m.merging = true
		m.mergeStatus = m.localizer.T("merge.status_start")
		m.mergeProgress = 0
		m.mergeCurrent = 0
		m.mergeTotal = 0
		
		// Crea canale per progress
		m.progressChan = make(chan tea.Msg, 10)
		
		// Avvia merge in goroutine
		go func() {
			result := m.executeMerge(m.progressChan)
			m.progressChan <- result
			close(m.progressChan)
		}()
		
		// Inizia ad ascoltare il canale
		return m, waitForProgress(m.progressChan)
	}

	// Aggiorna il textinput (ma l'Enter viene gestito sopra)
	var cmd tea.Cmd
	m.mergeInput, cmd = m.mergeInput.Update(msg)
	return m, cmd
}

func (m ListModel) updateManualMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "esc":
		// Se la search è attiva e non vuota, la svuota
		// Altrimenti torna al menu principale
		if m.searchInput.Focused() && m.searchInput.Value() != "" {
			m.searchInput.SetValue("")
			m.filteredItems = m.allItems
			m.cursor = 0
			return m, nil
		}
		m.quitting = true
		return m, nil

	case "up", "k":
		if !m.searchInput.Focused() {
			if m.cursor > 0 {
				m.cursor--
			}
		}

	case "down", "j":
		if !m.searchInput.Focused() {
			if m.cursor < len(m.filteredItems)-1 {
				m.cursor++
			}
		}

	case "tab":
		// Toggle focus tra search e lista
		if m.searchInput.Focused() {
			m.searchInput.Blur()
		} else {
			return m, m.searchInput.Focus()
		}

	case " ":
		if !m.searchInput.Focused() && len(m.filteredItems) > 0 {
			item := m.filteredItems[m.cursor]
			m.selectedMap[item.ID] = !m.selectedMap[item.ID]
		}

	case "enter":
		if m.searchInput.Focused() {
			// Se nella search, passa alla lista
			m.searchInput.Blur()
			return m, nil
		}
		// Altrimenti vai al merge
		selectedCount := 0
		for _, selected := range m.selectedMap {
			if selected {
				selectedCount++
			}
		}
		if selectedCount >= 2 {
			m.mode = "merge"
			// Imposta l'elemento attualmente sotto il cursore come nome di default
			if m.cursor < len(m.filteredItems) {
				m.mergeInput.SetValue(m.filteredItems[m.cursor].Name)
			}
			return m, m.mergeInput.Focus()
		}
	}

	// Aggiorna il search input
	if m.searchInput.Focused() {
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		
		// Filtra gli elementi in base alla query
		query := strings.ToLower(m.searchInput.Value())
		if query == "" {
			m.filteredItems = m.allItems
		} else {
			m.filteredItems = m.filterItems(query)
		}
		
		// Resetta il cursore se fuori range
		if m.cursor >= len(m.filteredItems) {
			m.cursor = 0
		}
		
		return m, cmd
	}

	return m, nil
}

func (m ListModel) filterItems(query string) []similarity.SimilarItem {
	var filtered []similarity.SimilarItem
	for _, item := range m.allItems {
		if strings.Contains(strings.ToLower(item.Name), query) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func (m ListModel) executeMerge(progressChan chan<- tea.Msg) tea.Msg {
	finalName := m.mergeInput.Value()
	if finalName == "" {
		return mergeCompleteMsg{err: fmt.Errorf(m.localizer.T("merge.error_empty_name"))}
	}

	// Raccogli gli ID selezionati (deduplicati)
	selectedIDs := make([]int, 0)
	seenIDs := make(map[int]bool)
	
	if m.mergeMode == ModeManual {
		// In modalità manuale, usa allItems
		for _, item := range m.allItems {
			if m.selectedMap[item.ID] && !seenIDs[item.ID] {
				selectedIDs = append(selectedIDs, item.ID)
				seenIDs[item.ID] = true
			}
		}
	} else {
		// In modalità semi-automatica, usa currentGroup
		if m.currentGroup == nil {
			return mergeCompleteMsg{err: fmt.Errorf(m.localizer.T("merge.error_no_group"))}
		}
		for _, item := range m.currentGroup.Items {
			if m.selectedMap[item.ID] && !seenIDs[item.ID] {
				selectedIDs = append(selectedIDs, item.ID)
				seenIDs[item.ID] = true
			}
		}
	}

	if len(selectedIDs) < 2 {
		return mergeCompleteMsg{err: fmt.Errorf(m.localizer.T("merge.error_min_items"))}
	}

	// Trova se esiste già un elemento con il nome finale tra quelli selezionati
	var mainID int
	var toDeleteIDs []int
	nameExists := false
	
	// Cerca in allItems per la modalità manuale
	itemsToCheck := m.allItems
	if m.mergeMode == ModeSemiAutomatic && m.currentGroup != nil {
		itemsToCheck = m.currentGroup.Items
	}
	
	for _, item := range itemsToCheck {
		if m.selectedMap[item.ID] && item.Name == finalName {
			mainID = item.ID
			nameExists = true
			break
		}
	}
	
	// Se il nome esiste già, usalo come principale ed elimina tutti gli altri
	if nameExists {
		for _, id := range selectedIDs {
			if id != mainID {
				toDeleteIDs = append(toDeleteIDs, id)
			}
		}
	} else {
		// Altrimenti il primo diventa il principale
		mainID = selectedIDs[0]
		toDeleteIDs = selectedIDs[1:]
	}

	// Calcola il numero totale di operazioni granulari
	// Per ogni elemento da eliminare: recupero docs + N aggiornamenti docs + eliminazione
	// Stimiamo basandoci sul numero di elementi (useremo aggiornamenti in tempo reale)
	var err error
	currentOp := 0.0
	totalOps := 0.0
	
	// Stima iniziale: 1 op per preparazione (se serve) + 3 ops per elemento (get, update, delete)
	if !nameExists {
		totalOps++ // preparazione nome temporaneo
	}
	totalOps += float64(len(toDeleteIDs)) * 3 // get + update docs + delete per elemento
	if !nameExists {
		totalOps++ // aggiornamento finale nome
	}

	// Se il nome non esiste, aggiorna prima l'elemento principale con un nome temporaneo
	needsRename := !nameExists
	if needsRename {
		currentOp++
		progressChan <- mergeProgressMsg{
			current: int(currentOp),
			total:   int(totalOps),
			status:  m.localizer.T("merge.status_prepare"),
		}

		tempName := fmt.Sprintf("__MERGING_%d_%s", mainID, finalName)
		
		switch m.entityType {
		case EntityTags:
			err = m.client.UpdateTag(mainID, tempName)
		case EntityCorrespondents:
			err = m.client.UpdateCorrespondent(mainID, tempName)
		case EntityDocumentTypes:
			err = m.client.UpdateDocumentType(mainID, tempName)
		}

		if err != nil {
			return mergeCompleteMsg{err: fmt.Errorf(m.localizer.T("merge.error_temp_update"), err)}
		}
	}

	// Per ogni elemento da eliminare, migra i documenti e poi eliminalo
	for idx, oldID := range toDeleteIDs {
		// Step 1: Recupero documenti
		currentOp++
		progressChan <- mergeProgressMsg{
			current: int(currentOp),
			total:   int(totalOps),
			status:  fmt.Sprintf(m.localizer.T("merge.status_get_docs"), idx+1, len(toDeleteIDs)),
		}

		var docs []paperless.Document
		
		switch m.entityType {
		case EntityTags:
			docs, err = m.client.GetDocumentsByTag(oldID)
			if err != nil {
				return mergeCompleteMsg{err: fmt.Errorf(m.localizer.T("merge.error_get_docs"), err)}
			}

		case EntityCorrespondents:
			docs, err = m.client.GetDocumentsByCorrespondent(oldID)
			if err != nil {
				return mergeCompleteMsg{err: fmt.Errorf(m.localizer.T("merge.error_get_docs"), err)}
			}

		case EntityDocumentTypes:
			docs, err = m.client.GetDocumentsByType(oldID)
			if err != nil {
				return mergeCompleteMsg{err: fmt.Errorf(m.localizer.T("merge.error_get_docs"), err)}
			}
		}

		// Step 2: Aggiorna documenti
		if len(docs) > 0 {
			currentOp++
			progressChan <- mergeProgressMsg{
				current: int(currentOp),
				total:   int(totalOps),
				status:  fmt.Sprintf(m.localizer.T("merge.status_update_docs"), len(docs), idx+1, len(toDeleteIDs)),
			}

			switch m.entityType {
			case EntityTags:
				for _, doc := range docs {
					if err := m.client.UpdateDocumentTags(doc.ID, oldID, mainID); err != nil {
						return mergeCompleteMsg{err: fmt.Errorf(m.localizer.T("merge.error_update_doc"), doc.ID, err)}
					}
				}

			case EntityCorrespondents:
				for _, doc := range docs {
					if err := m.client.UpdateDocumentCorrespondent(doc.ID, mainID); err != nil {
						return mergeCompleteMsg{err: fmt.Errorf(m.localizer.T("merge.error_update_doc"), doc.ID, err)}
					}
				}

			case EntityDocumentTypes:
				for _, doc := range docs {
					if err := m.client.UpdateDocumentTypeForDoc(doc.ID, mainID); err != nil {
						return mergeCompleteMsg{err: fmt.Errorf(m.localizer.T("merge.error_update_doc"), doc.ID, err)}
					}
				}
			}
		} else {
			currentOp++ // Skip se non ci sono documenti
		}

		// Step 3: Elimina elemento vecchio
		currentOp++
		progressChan <- mergeProgressMsg{
			current: int(currentOp),
			total:   int(totalOps),
			status:  fmt.Sprintf(m.localizer.T("merge.status_delete"), idx+1, len(toDeleteIDs)),
		}

		switch m.entityType {
		case EntityTags:
			if err := m.client.DeleteTag(oldID); err != nil && !strings.Contains(err.Error(), "404") {
				return mergeCompleteMsg{err: fmt.Errorf("errore nell'eliminazione tag %d: %w", oldID, err)}
			}

		case EntityCorrespondents:
			if err := m.client.DeleteCorrespondent(oldID); err != nil && !strings.Contains(err.Error(), "404") {
				return mergeCompleteMsg{err: fmt.Errorf("errore nell'eliminazione corrispondente %d: %w", oldID, err)}
			}

		case EntityDocumentTypes:
			if err := m.client.DeleteDocumentType(oldID); err != nil && !strings.Contains(err.Error(), "404") {
				return mergeCompleteMsg{err: fmt.Errorf("errore nell'eliminazione tipo documento %d: %w", oldID, err)}
			}
		}
	}

	// Se necessario, aggiorna il nome dell'elemento principale al nome finale
	// (ora non ci sono più conflitti perché tutti gli altri elementi sono stati eliminati)
	if needsRename {
		currentOp++
		progressChan <- mergeProgressMsg{
			current: int(currentOp),
			total:   int(totalOps),
			status:  m.localizer.T("merge.status_final_name"),
		}

		switch m.entityType {
		case EntityTags:
			err = m.client.UpdateTag(mainID, finalName)
		case EntityCorrespondents:
			err = m.client.UpdateCorrespondent(mainID, finalName)
		case EntityDocumentTypes:
			err = m.client.UpdateDocumentType(mainID, finalName)
		}

		if err != nil {
			return mergeCompleteMsg{err: fmt.Errorf(m.localizer.T("merge.error_final_update"), err)}
		}
	}

	// Merge completato con successo
	return mergeCompleteMsg{err: nil}
}

func (m ListModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		MarginBottom(1)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)

	var entityName string
	switch m.entityType {
	case EntityTags:
		entityName = m.localizer.T("entity.tags")
	case EntityCorrespondents:
		entityName = m.localizer.T("entity.correspondents")
	case EntityDocumentTypes:
		entityName = m.localizer.T("entity.doctypes")
	}

	s := titleStyle.Render(fmt.Sprintf(m.localizer.T("list.title"), entityName)) + "\n\n"

	if m.loading {
		s += normalStyle.Render(m.localizer.T("list.loading")) + "\n"
		return s
	}

	if m.merging {
		s += normalStyle.Render(m.localizer.T("list.merging")) + "\n\n"
		if m.mergeStatus != "" {
			s += normalStyle.Render(m.mergeStatus) + "\n\n"
		}
		s += m.progress.ViewAs(m.mergeProgress) + "\n\n"
		if m.mergeTotal > 0 {
			s += normalStyle.Render(fmt.Sprintf(m.localizer.T("merge.progress_operation"), m.mergeCurrent, m.mergeTotal)) + "\n"
		}
		return s
	}

	if m.err != nil {
		s += errorStyle.Render(fmt.Sprintf(m.localizer.T("list.error"), m.err)) + "\n\n"
		s += normalStyle.Render(m.localizer.T("list.error_back")) + "\n"
		return s
	}

	if m.mode == "merge" {
		s += normalStyle.Render(m.localizer.T("list.merge_input_label")) + "\n\n"
		s += m.mergeInput.View() + "\n\n"
		
		var selected []string
		itemsToCheck := m.allItems
		if m.mergeMode == ModeSemiAutomatic && m.currentGroup != nil {
			itemsToCheck = m.currentGroup.Items
		}
		
		for _, item := range itemsToCheck {
			if m.selectedMap[item.ID] {
				selected = append(selected, item.Name)
			}
		}
		s += normalStyle.Render(fmt.Sprintf(m.localizer.T("list.merge_items_to_merge"), len(selected))) + "\n"
		s += normalStyle.Render(strings.Join(selected, " → ")) + "\n\n"
		s += normalStyle.Render(m.localizer.T("list.merge_help")) + "\n"
		return s
	}

	if m.mode == "manual" {
		// Modalità manuale: mostra tutti gli elementi con search
		selectedCount := 0
		for _, selected := range m.selectedMap {
			if selected {
				selectedCount++
			}
		}
		
		s += normalStyle.Render(fmt.Sprintf(m.localizer.T("list.manual_title"), len(m.allItems), selectedCount)) + "\n\n"
		s += normalStyle.Render(m.localizer.T("list.manual_search")) + m.searchInput.View() + "\n\n"
		
		if len(m.filteredItems) == 0 {
			s += normalStyle.Render(m.localizer.T("list.manual_no_results")) + "\n"
		} else {
			// Calcola dinamicamente il numero di elementi visibili in base all'altezza del terminale
			// Sottrai 10 righe per header, search, help, ecc.
			maxVisible := m.height - 10
			if maxVisible < 5 {
				maxVisible = 5 // Minimo 5 elementi visibili
			}
			halfVisible := maxVisible / 2
			
			var startIdx, endIdx int
			
			// Se ci sono meno elementi del massimo visibile, mostra tutti
			if len(m.filteredItems) <= maxVisible {
				startIdx = 0
				endIdx = len(m.filteredItems)
			} else {
				// Scroll centrato: mantieni il cursore al centro quando possibile
				startIdx = m.cursor - halfVisible
				
				// Correggi se troppo in alto
				if startIdx < 0 {
					startIdx = 0
				}
				
				// Correggi se troppo in basso
				if startIdx > len(m.filteredItems)-maxVisible {
					startIdx = len(m.filteredItems) - maxVisible
				}
				
				endIdx = startIdx + maxVisible
				if endIdx > len(m.filteredItems) {
					endIdx = len(m.filteredItems)
				}
			}
			
			if startIdx > 0 {
				s += normalStyle.Render(fmt.Sprintf(m.localizer.T("list.manual_above"), startIdx)) + "\n"
			}
			
			for i := startIdx; i < endIdx; i++ {
				item := m.filteredItems[i]
				cursor := " "
				checkbox := "[ ]"
				if m.selectedMap[item.ID] {
					checkbox = "[✓]"
				}
				
				line := fmt.Sprintf("%s %s %s", cursor, checkbox, item.Name)
				
				if i == m.cursor {
					cursor = ">"
					s += selectedStyle.Render(cursor + " " + checkbox + " " + item.Name) + "\n"
				} else {
					s += normalStyle.Render(line) + "\n"
				}
			}
			
			if endIdx < len(m.filteredItems) {
				s += normalStyle.Render(fmt.Sprintf(m.localizer.T("list.manual_below"), len(m.filteredItems)-endIdx)) + "\n"
			}
		}
		
		s += "\n"
		if selectedCount >= 2 {
			s += normalStyle.Render(m.localizer.T("list.manual_help_merge")) + "\n"
		} else {
			s += normalStyle.Render(m.localizer.T("list.manual_help")) + "\n"
		}
		return s
	}

	if m.mode == "select" && m.currentGroup != nil {
		s += normalStyle.Render(fmt.Sprintf(m.localizer.T("list.select_group"), m.currentGroup.Representative)) + "\n"
		s += normalStyle.Render(fmt.Sprintf(m.localizer.T("list.select_label"), 
			len(m.selectedMap), len(m.currentGroup.Items))) + "\n\n"

		for i, item := range m.currentGroup.Items {
			cursor := " "
			checkbox := "[ ]"
			if m.selectedMap[item.ID] {
				checkbox = "[✓]"
			}
			
			line := fmt.Sprintf("%s %s %s", cursor, checkbox, item.Name)
			
			if i == m.groupCursor {
				cursor = ">"
				s += selectedStyle.Render(cursor + " " + checkbox + " " + item.Name) + "\n"
			} else {
				s += normalStyle.Render(line) + "\n"
			}
		}

		s += "\n" + normalStyle.Render(m.localizer.T("list.select_help")) + "\n"
		return s
	}

	// Modalità browse
	if len(m.groups) == 0 {
		s += normalStyle.Render(m.localizer.T("list.browse_no_duplicates")) + "\n\n"
		s += normalStyle.Render(m.localizer.T("list.browse_back")) + "\n"
		return s
	}

	s += normalStyle.Render(fmt.Sprintf(m.localizer.T("list.browse_found"), len(m.groups))) + "\n\n"

	// Calcola dinamicamente il numero di gruppi visibili in base all'altezza del terminale
	// Sottrai 8 righe per header, help, ecc.
	maxVisible := m.height - 8
	if maxVisible < 5 {
		maxVisible = 5 // Minimo 5 gruppi visibili
	}
	halfVisible := maxVisible / 2
	
	var startIdx, endIdx int
	
	// Se ci sono meno gruppi del massimo visibile, mostra tutti
	if len(m.groups) <= maxVisible {
		startIdx = 0
		endIdx = len(m.groups)
	} else {
		// Scroll centrato: mantieni il cursore al centro quando possibile
		startIdx = m.cursor - halfVisible
		
		// Correggi se troppo in alto
		if startIdx < 0 {
			startIdx = 0
		}
		
		// Correggi se troppo in basso
		if startIdx > len(m.groups)-maxVisible {
			startIdx = len(m.groups) - maxVisible
		}
		
		endIdx = startIdx + maxVisible
		if endIdx > len(m.groups) {
			endIdx = len(m.groups)
		}
	}
	
	if startIdx > 0 {
		s += normalStyle.Render(fmt.Sprintf("... (%d gruppi sopra) ...", startIdx)) + "\n"
	}
	
	for i := startIdx; i < endIdx; i++ {
		group := m.groups[i]
		cursor := " "
		line := fmt.Sprintf("%s [%d] %s", cursor, len(group.Items), group.Representative)
		
		if i == m.cursor {
			cursor = ">"
			s += selectedStyle.Render(cursor + " " + line[2:]) + "\n"
		} else {
			s += normalStyle.Render(line) + "\n"
		}
	}
	
	if endIdx < len(m.groups) {
		s += normalStyle.Render(fmt.Sprintf("... (%d gruppi sotto) ...", len(m.groups)-endIdx)) + "\n"
	}

	s += "\n" + normalStyle.Render(m.localizer.T("list.browse_help")) + "\n"

	return s
}

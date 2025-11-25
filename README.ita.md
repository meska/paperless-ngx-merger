# Paperless-ngx Merger

Un'applicazione TUI (Text User Interface) per gestire e unire tags, corrispondenti e tipi di documento duplicati o simili in Paperless-ngx.

## 🚀 Funzionalità

- **Connessione sicura a Paperless-ngx**: Configurazione iniziale interattiva con salvataggio delle credenziali
- **Rilevamento intelligente di duplicati**: Algoritmo di similarità basato sulla distanza di Levenshtein per trovare elementi con testo simile
- **Gestione completa di**:
  - Tags
  - Corrispondenti
  - Tipi di documento
- **Merge interattivo**: 
  - Visualizzazione di gruppi di elementi simili
  - Selezione degli elementi da unire
  - Aggiornamento automatico di tutti i documenti collegati
  - Eliminazione degli elementi obsoleti

## 📋 Prerequisiti

- Go 1.21 o superiore
- Accesso a un'istanza Paperless-ngx
- API Key di Paperless-ngx (generabile dalle impostazioni utente)

## 🔧 Installazione

### Da sorgente

```bash
# Clona il repository
git clone https://github.com/meska/paperless-merger.git
cd paperless-merger

# Installa le dipendenze
go mod download

# Compila
go build -o paperless-merger ./cmd/paperless-merger

# Esegui
./paperless-merger
```

### Installazione diretta

```bash
go install github.com/meska/paperless-merger/cmd/paperless-merger@latest
```

## 📖 Utilizzo

### Primo avvio

Al primo avvio, l'applicazione ti chiederà:
1. **URL del server Paperless-ngx** (es. `https://paperless.example.com`)
2. **API Key** per l'autenticazione

Le credenziali verranno salvate in `~/.config/paperless-merger/config.json` e non verranno mai condivise.

### Utilizzo principale

1. **Seleziona il tipo di entità** da gestire:
   - Tags
   - Corrispondenti
   - Tipi di Documento

2. **Visualizza i gruppi di elementi simili**: L'applicazione mostrerà automaticamente i gruppi di elementi con testo simile (soglia di similarità: 70%)

3. **Gestisci un gruppo**:
   - Seleziona gli elementi da unire (Space per selezionare/deselezionare)
   - Premi Enter per procedere

4. **Esegui il merge**:
   - Inserisci il nome finale che vuoi dare agli elementi uniti
   - Conferma con Enter
   - L'applicazione:
     - Aggiornerà tutti i documenti che usano gli elementi selezionati
     - Eliminerà gli elementi obsoleti
     - Manterrà solo l'elemento principale con il nuovo nome

## ⌨️ Comandi

### Menu principale
- `↑/↓` o `j/k`: Naviga tra le opzioni
- `Enter`: Seleziona un'opzione
- `q` o `Esc`: Esci

### Lista elementi simili
- `↑/↓` o `j/k`: Naviga tra i gruppi
- `Enter`: Gestisci un gruppo
- `Esc`: Torna al menu principale

### Selezione elementi
- `↑/↓` o `j/k`: Naviga tra gli elementi
- `Space`: Seleziona/Deseleziona un elemento
- `Enter`: Procedi al merge
- `Esc`: Torna alla lista gruppi

### Merge
- `Enter`: Conferma il merge
- `Esc`: Annulla

## 🔒 Sicurezza

- Le credenziali sono salvate in `~/.config/paperless-merger/config.json` con permessi `0600` (leggibile solo dall'utente)
- Il file di configurazione è automaticamente ignorato da git
- L'API Key è nascosta durante l'inserimento

## 🏗️ Struttura del progetto

```
paperless-merger/
├── cmd/
│   └── paperless-merger/    # Entrypoint dell'applicazione
│       └── main.go
├── internal/
│   ├── config/              # Gestione configurazione
│   │   └── config.go
│   ├── paperless/           # Client API Paperless-ngx
│   │   └── client.go
│   ├── similarity/          # Algoritmo di similarità
│   │   └── similarity.go
│   └── ui/                  # Interfaccia Bubbletea
│       ├── setup.go         # Setup iniziale
│       ├── main.go          # Menu principale
│       └── list.go          # Lista e merge
├── go.mod
├── go.sum
├── .gitignore
└── README.md
```

## 🛠️ Tecnologie utilizzate

- **Go**: Linguaggio di programmazione
- **Bubbletea**: Framework TUI per interfacce terminale interattive
- **Lipgloss**: Libreria per lo styling del terminale
- **Bubbles**: Componenti riutilizzabili per Bubbletea

## 📝 API Paperless-ngx utilizzate

- `GET /api/tags/`: Recupero tags
- `GET /api/correspondents/`: Recupero corrispondenti
- `GET /api/document_types/`: Recupero tipi di documento
- `GET /api/documents/`: Recupero documenti filtrati
- `PATCH /api/tags/{id}/`: Aggiornamento tag
- `PATCH /api/correspondents/{id}/`: Aggiornamento corrispondente
- `PATCH /api/document_types/{id}/`: Aggiornamento tipo documento
- `PATCH /api/documents/{id}/`: Aggiornamento documento
- `DELETE /api/tags/{id}/`: Eliminazione tag
- `DELETE /api/correspondents/{id}/`: Eliminazione corrispondente
- `DELETE /api/document_types/{id}/`: Eliminazione tipo documento

## 🤝 Contribuire

I contributi sono benvenuti! Sentiti libero di:
- Aprire issue per bug o richieste di funzionalità
- Proporre pull request

## 📄 Licenza

MIT

## ⚠️ Note importanti

- **Backup**: Si consiglia di fare un backup del database di Paperless-ngx prima di utilizzare questa applicazione
- **Test**: L'applicazione è stata testata con Paperless-ngx v1.17+
- **Permessi**: Assicurati che l'API Key abbia i permessi necessari per modificare tags, corrispondenti e documenti

## 🐛 Troubleshooting

### Errore di connessione
- Verifica che l'URL di Paperless-ngx sia corretto e accessibile
- Controlla che l'API Key sia valida
- Assicurati che non ci siano firewall che bloccano la connessione

### Errore di aggiornamento documenti
- Verifica i permessi dell'API Key
- Controlla i log di Paperless-ngx per eventuali errori server-side

### L'applicazione non trova duplicati
- La soglia di similarità è impostata al 70%
- Gli elementi devono avere almeno il 70% di caratteri in comune per essere considerati simili
- Puoi modificare la soglia nel file `internal/ui/list.go` (riga con `FindSimilarGroups`)

## 📧 Contatti

Per domande o supporto, apri un issue su GitHub.

# Paperless-ngx Merger

A TUI (Text User Interface) application to manage and merge duplicate or similar tags, correspondents, and document types in Paperless-ngx.

[🇮🇹 Versione Italiana](README.ita.md)

## 🚀 Features

- **Secure connection to Paperless-ngx**: Interactive initial setup with credential storage
- **Intelligent duplicate detection**: Similarity algorithm based on Levenshtein distance to find items with similar text
- **Complete management of**:
  - Tags
  - Correspondents
  - Document Types
- **Interactive merge**: 
  - Display groups of similar items
  - Selection of items to merge
  - Automatic update of all linked documents
  - Deletion of obsolete items

## 📋 Prerequisites

- Go 1.21 or higher
- Access to a Paperless-ngx instance
- Paperless-ngx API Key (can be generated from user settings)

## 🔧 Installation

### From source

```bash
# Clone the repository
git clone https://github.com/meska/paperless-merger.git
cd paperless-merger

# Install dependencies
go mod download

# Build
go build -o paperless-merger ./cmd/paperless-merger

# Run
./paperless-merger
```

### Direct installation

```bash
go install github.com/meska/paperless-merger/cmd/paperless-merger@latest
```

## 📖 Usage

### First run

On first run, the application will ask for:
1. **Paperless-ngx server URL** (e.g. `https://paperless.example.com`)
2. **API Key** for authentication
3. **Language preference** (auto-detect, English, or Italian)

Credentials will be saved in `~/.config/paperless-merger/config.json` and will never be shared.

### Main usage

1. **Select the entity type** to manage:
   - Tags
   - Correspondents
   - Document Types

2. **View similar item groups**: The application will automatically show groups of items with similar text (similarity threshold: 70%)

3. **Manage a group**:
   - Select items to merge (Space to select/deselect)
   - Press Enter to proceed

4. **Execute merge**:
   - Enter the final name for the merged items
   - Confirm with Enter
   - The application will:
     - Update all documents using the selected items
     - Delete obsolete items
     - Keep only the main item with the new name

## ⌨️ Commands

### Main menu
- `↑/↓` or `j/k`: Navigate between options
- `Enter`: Select an option
- `q` or `Esc`: Exit

### Similar items list
- `↑/↓` or `j/k`: Navigate between groups
- `Enter`: Manage a group
- `Esc`: Return to main menu

### Item selection
- `↑/↓` or `j/k`: Navigate between items
- `Space`: Select/Deselect an item
- `Enter`: Proceed to merge
- `Esc`: Return to group list

### Merge
- `Enter`: Confirm merge
- `Esc`: Cancel

## 🔒 Security

- Credentials are saved in `~/.config/paperless-merger/config.json` with `0600` permissions (readable only by the user)
- Configuration file is automatically ignored by git
- API Key is hidden during input

## 🏗️ Project structure

```
paperless-merger/
├── cmd/
│   └── paperless-merger/    # Application entrypoint
│       └── main.go
├── internal/
│   ├── config/              # Configuration management
│   │   └── config.go
│   ├── locale/              # Internationalization
│   │   └── locale.go
│   ├── paperless/           # Paperless-ngx API client
│   │   └── client.go
│   ├── similarity/          # Similarity algorithm
│   │   └── similarity.go
│   └── ui/                  # Bubbletea interface
│       ├── setup.go         # Initial setup
│       ├── main.go          # Main menu
│       └── list.go          # List and merge
├── locales/
│   ├── en.json              # English translations
│   └── it.json              # Italian translations
├── go.mod
├── go.sum
├── .gitignore
└── README.md
```

## 🛠️ Technologies used

- **Go**: Programming language
- **Bubbletea**: TUI framework for interactive terminal interfaces
- **Lipgloss**: Terminal styling library
- **Bubbles**: Reusable components for Bubbletea
- **go-i18n**: Internationalization and localization

## 📝 Paperless-ngx APIs used

- `GET /api/tags/`: Retrieve tags
- `GET /api/correspondents/`: Retrieve correspondents
- `GET /api/document_types/`: Retrieve document types
- `GET /api/documents/`: Retrieve filtered documents
- `PATCH /api/tags/{id}/`: Update tag
- `PATCH /api/correspondents/{id}/`: Update correspondent
- `PATCH /api/document_types/{id}/`: Update document type
- `PATCH /api/documents/{id}/`: Update document
- `DELETE /api/tags/{id}/`: Delete tag
- `DELETE /api/correspondents/{id}/`: Delete correspondent
- `DELETE /api/document_types/{id}/`: Delete document type

## 🤝 Contributing

Contributions are welcome! Feel free to:
- Open issues for bugs or feature requests
- Submit pull requests

## 📄 License

MIT

## ⚠️ Important notes

- **Backup**: It is recommended to backup your Paperless-ngx database before using this application
- **Testing**: The application has been tested with Paperless-ngx v1.17+
- **Permissions**: Ensure the API Key has the necessary permissions to modify tags, correspondents, and documents

## 🐛 Troubleshooting

### Connection error
- Verify that the Paperless-ngx URL is correct and accessible
- Check that the API Key is valid
- Make sure there are no firewalls blocking the connection

### Document update error
- Verify API Key permissions
- Check Paperless-ngx logs for server-side errors

### Application doesn't find duplicates
- The similarity threshold is set at 70%
- Items must have at least 70% of characters in common to be considered similar
- You can modify the threshold in `internal/ui/list.go` (line with `FindSimilarGroups`)

## 📧 Contact

For questions or support, open an issue on GitHub.

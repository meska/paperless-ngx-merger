package paperless

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client rappresenta il client per l'API di Paperless-ngx
type Client struct {
	BaseURL string
	APIKey  string
	client  *http.Client
}

// Tag rappresenta un tag di Paperless
type Tag struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Color string `json:"colour"`
	Match string `json:"match"`
}

// Correspondent rappresenta un corrispondente di Paperless
type Correspondent struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Match string `json:"match"`
}

// DocumentType rappresenta un tipo di documento di Paperless
type DocumentType struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Match string `json:"match"`
}

// Document rappresenta un documento di Paperless
type Document struct {
	ID               int    `json:"id"`
	Title            string `json:"title"`
	Correspondent    *int   `json:"correspondent"`
	DocumentType     *int   `json:"document_type"`
	Tags             []int  `json:"tags"`
}

// ListResponse rappresenta la risposta paginata dell'API
type ListResponse struct {
	Count    int             `json:"count"`
	Next     *string         `json:"next"`
	Previous *string         `json:"previous"`
	Results  json.RawMessage `json:"results"`
}

// NewClient crea un nuovo client per Paperless-ngx
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		client:  &http.Client{},
	}
}

// makeRequest esegue una richiesta HTTP all'API
func (c *Client) makeRequest(method, endpoint string, body io.Reader) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.BaseURL, endpoint)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Token %s", c.APIKey))
	req.Header.Set("Content-Type", "application/json")

	return c.client.Do(req)
}

// GetTags recupera tutti i tags
func (c *Client) GetTags() ([]Tag, error) {
	resp, err := c.makeRequest("GET", "/api/tags/?page_size=100", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("errore API: %d - %s", resp.StatusCode, string(body))
	}

	var listResp ListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, err
	}

	var tags []Tag
	if err := json.Unmarshal(listResp.Results, &tags); err != nil {
		return nil, err
	}

	return tags, nil
}

// GetCorrespondents recupera tutti i corrispondenti
func (c *Client) GetCorrespondents() ([]Correspondent, error) {
	resp, err := c.makeRequest("GET", "/api/correspondents/?page_size=100", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("errore API: %d - %s", resp.StatusCode, string(body))
	}

	var listResp ListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, err
	}

	var correspondents []Correspondent
	if err := json.Unmarshal(listResp.Results, &correspondents); err != nil {
		return nil, err
	}

	return correspondents, nil
}

// GetDocumentTypes recupera tutti i tipi di documento
func (c *Client) GetDocumentTypes() ([]DocumentType, error) {
	resp, err := c.makeRequest("GET", "/api/document_types/?page_size=100", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("errore API: %d - %s", resp.StatusCode, string(body))
	}

	var listResp ListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, err
	}

	var docTypes []DocumentType
	if err := json.Unmarshal(listResp.Results, &docTypes); err != nil {
		return nil, err
	}

	return docTypes, nil
}

// UpdateTag aggiorna un tag
func (c *Client) UpdateTag(id int, name string) error {
	body := strings.NewReader(fmt.Sprintf(`{"name": "%s"}`, name))
	resp, err := c.makeRequest("PATCH", fmt.Sprintf("/api/tags/%d/", id), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("errore nell'aggiornamento del tag: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// UpdateCorrespondent aggiorna un corrispondente
func (c *Client) UpdateCorrespondent(id int, name string) error {
	body := strings.NewReader(fmt.Sprintf(`{"name": "%s"}`, name))
	resp, err := c.makeRequest("PATCH", fmt.Sprintf("/api/correspondents/%d/", id), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("errore nell'aggiornamento del corrispondente: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// UpdateDocumentType aggiorna un tipo di documento
func (c *Client) UpdateDocumentType(id int, name string) error {
	body := strings.NewReader(fmt.Sprintf(`{"name": "%s"}`, name))
	resp, err := c.makeRequest("PATCH", fmt.Sprintf("/api/document_types/%d/", id), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("errore nell'aggiornamento del tipo documento: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// DeleteTag elimina un tag
func (c *Client) DeleteTag(id int) error {
	resp, err := c.makeRequest("DELETE", fmt.Sprintf("/api/tags/%d/", id), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("errore nell'eliminazione del tag: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// DeleteCorrespondent elimina un corrispondente
func (c *Client) DeleteCorrespondent(id int) error {
	resp, err := c.makeRequest("DELETE", fmt.Sprintf("/api/correspondents/%d/", id), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("errore nell'eliminazione del corrispondente: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// DeleteDocumentType elimina un tipo di documento
func (c *Client) DeleteDocumentType(id int) error {
	resp, err := c.makeRequest("DELETE", fmt.Sprintf("/api/document_types/%d/", id), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("errore nell'eliminazione del tipo documento: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// GetDocumentsByTag recupera tutti i documenti che hanno un certo tag
func (c *Client) GetDocumentsByTag(tagID int) ([]Document, error) {
	endpoint := fmt.Sprintf("/api/documents/?tags__id__in=%d&page_size=100", tagID)
	return c.getDocuments(endpoint)
}

// GetDocumentsByCorrespondent recupera tutti i documenti di un corrispondente
func (c *Client) GetDocumentsByCorrespondent(correspondentID int) ([]Document, error) {
	endpoint := fmt.Sprintf("/api/documents/?correspondent__id=%d&page_size=100", correspondentID)
	return c.getDocuments(endpoint)
}

// GetDocumentsByType recupera tutti i documenti di un tipo
func (c *Client) GetDocumentsByType(typeID int) ([]Document, error) {
	endpoint := fmt.Sprintf("/api/documents/?document_type__id=%d&page_size=100", typeID)
	return c.getDocuments(endpoint)
}

// getDocuments è un helper per recuperare documenti
func (c *Client) getDocuments(endpoint string) ([]Document, error) {
	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("errore API: %d - %s", resp.StatusCode, string(body))
	}

	var listResp ListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, err
	}

	var documents []Document
	if err := json.Unmarshal(listResp.Results, &documents); err != nil {
		return nil, err
	}

	return documents, nil
}

// UpdateDocumentTags aggiorna i tags di un documento
func (c *Client) UpdateDocumentTags(docID int, oldTagID, newTagID int) error {
	// Prima recuperiamo il documento per avere tutti i suoi tag
	doc, err := c.getDocument(docID)
	if err != nil {
		return err
	}

	// Sostituiamo il vecchio tag con il nuovo
	newTags := make([]int, 0)
	for _, tagID := range doc.Tags {
		if tagID == oldTagID {
			newTags = append(newTags, newTagID)
		} else {
			newTags = append(newTags, tagID)
		}
	}

	// Aggiorniamo il documento
	tagsJSON, _ := json.Marshal(newTags)
	body := strings.NewReader(fmt.Sprintf(`{"tags": %s}`, string(tagsJSON)))
	resp, err := c.makeRequest("PATCH", fmt.Sprintf("/api/documents/%d/", docID), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("errore nell'aggiornamento del documento: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// UpdateDocumentCorrespondent aggiorna il corrispondente di un documento
func (c *Client) UpdateDocumentCorrespondent(docID, newCorrespondentID int) error {
	body := strings.NewReader(fmt.Sprintf(`{"correspondent": %d}`, newCorrespondentID))
	resp, err := c.makeRequest("PATCH", fmt.Sprintf("/api/documents/%d/", docID), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("errore nell'aggiornamento del documento: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// UpdateDocumentType aggiorna il tipo di un documento
func (c *Client) UpdateDocumentTypeForDoc(docID, newTypeID int) error {
	body := strings.NewReader(fmt.Sprintf(`{"document_type": %d}`, newTypeID))
	resp, err := c.makeRequest("PATCH", fmt.Sprintf("/api/documents/%d/", docID), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("errore nell'aggiornamento del documento: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// getDocument recupera un singolo documento
func (c *Client) getDocument(docID int) (*Document, error) {
	resp, err := c.makeRequest("GET", fmt.Sprintf("/api/documents/%d/", docID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("errore API: %d - %s", resp.StatusCode, string(body))
	}

	var doc Document
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, err
	}

	return &doc, nil
}

// TestConnection verifica la connessione all'API
func (c *Client) TestConnection() error {
	resp, err := c.makeRequest("GET", "/api/", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("connessione fallita: status code %d", resp.StatusCode)
	}

	return nil
}

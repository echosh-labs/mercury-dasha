package axismundi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Client handles HTTP and MCP communication with an Axis Mundi instance.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient returns a new configured Axis Mundi client.
func NewClient(baseURL, apiKey string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8088"
	}
	return &Client{
		baseURL: baseURL,
		apiKey:  strings.TrimSpace(apiKey),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Ping checks whether the Axis Mundi service is reachable.
func (c *Client) Ping(ctx context.Context) (bool, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/health", nil)
	if err != nil {
		return false, "", err
	}
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("unexpected health status code: %d", resp.StatusCode)
	}

	var body struct {
		Status  string `json:"status"`
		Service string `json:"service"`
		Mode    string `json:"mode"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)

	return true, body.Mode, nil
}

// RawRegistryItem represents the shape returned by Axis Mundi's /api/registry endpoint.
type RawRegistryItem struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
	Status  string `json:"status,omitempty"`
}

// FetchRegistry retrieves workspace items from Axis Mundi via /api/registry or MCP fallback.
func (c *Client) FetchRegistry(ctx context.Context) ([]WorkspaceItem, error) {
	// 1. Try standard REST registry endpoint first
	items, err := c.fetchViaREST(ctx)
	if err == nil {
		return items, nil
	}

	// 2. Fall back to MCP Streamable HTTP endpoint
	mcpItems, mcpErr := c.fetchViaMCP(ctx)
	if mcpErr == nil {
		return mcpItems, nil
	}

	return nil, fmt.Errorf("failed to fetch registry: rest error: %v; mcp error: %v", err, mcpErr)
}

func (c *Client) fetchViaREST(ctx context.Context) ([]WorkspaceItem, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/registry", nil)
	if err != nil {
		return nil, err
	}
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("REST status %d", resp.StatusCode)
	}

	var rawItems []RawRegistryItem
	if err := json.NewDecoder(resp.Body).Decode(&rawItems); err != nil {
		return nil, err
	}

	items := make([]WorkspaceItem, 0, len(rawItems))
	now := time.Now()
	for _, raw := range rawItems {
		wType := normalizeType(raw.Type)
		items = append(items, WorkspaceItem{
			ID:          raw.ID,
			Type:        wType,
			Title:       raw.Title,
			Snippet:     raw.Snippet,
			Status:      raw.Status,
			Source:      string(wType),
			FirstSeenAt: now,
			LastSeenAt:  now,
		})
	}
	return items, nil
}

func (c *Client) fetchViaMCP(ctx context.Context) ([]WorkspaceItem, error) {
	rpcReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "resources/list",
		"params":  map[string]any{},
	}
	payload, err := json.Marshal(rpcReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/mcp", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MCP status %d", resp.StatusCode)
	}

	var rpcResp struct {
		Result struct {
			Resources []struct {
				URI         string `json:"uri"`
				Name        string `json:"name"`
				Description string `json:"description"`
				MimeType    string `json:"mimeType"`
			} `json:"resources"`
		} `json:"result"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return nil, err
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("MCP RPC error: %s", rpcResp.Error.Message)
	}

	items := make([]WorkspaceItem, 0, len(rpcResp.Result.Resources))
	now := time.Now()
	for _, r := range rpcResp.Result.Resources {
		// URI scheme: keep://notes/{id}, docs://documents/{id}, sheets://spreadsheets/{id}, gmail://threads/{id}
		itemType := parseURIType(r.URI)
		items = append(items, WorkspaceItem{
			ID:          r.URI,
			Type:        itemType,
			Title:       r.Name,
			Snippet:     r.Description,
			Source:      "mcp",
			FirstSeenAt: now,
			LastSeenAt:  now,
		})
	}
	return items, nil
}

func normalizeType(t string) WorkspaceType {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "keep", "note", "notes":
		return TypeKeep
	case "gmail", "mail", "email", "thread":
		return TypeGmail
	case "doc", "docs", "document":
		return TypeDoc
	case "sheet", "sheets", "spreadsheet":
		return TypeSheet
	case "calendar", "event":
		return TypeCalendar
	default:
		return WorkspaceType(strings.ToLower(t))
	}
}

func parseURIType(uri string) WorkspaceType {
	parts := strings.SplitN(uri, "://", 2)
	if len(parts) > 0 {
		return normalizeType(parts[0])
	}
	return TypeKeep
}

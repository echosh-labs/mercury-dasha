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

// FetchRegistry retrieves workspace items from Axis Mundi.
// Prioritizes the standard MCP protocol (resources/list) over ad-hoc REST,
// with graceful fallback to /api/registry if MCP is unavailable.
func (c *Client) FetchRegistry(ctx context.Context) ([]WorkspaceItem, error) {
	// 1. Prioritize standard MCP protocol (resources/list)
	mcpItems, mcpErr := c.fetchViaMCP(ctx)
	if mcpErr == nil {
		return mcpItems, nil
	}

	// 2. Graceful fallback to REST registry endpoint
	items, restErr := c.fetchViaREST(ctx)
	if restErr == nil {
		return items, nil
	}

	return nil, fmt.Errorf("failed to fetch registry: mcp error: %v; rest error: %v", mcpErr, restErr)
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
		itemType := parseURIType(r.URI)
		status, snippet := parseStatusAndSnippet(r.Description)
		items = append(items, WorkspaceItem{
			ID:          r.URI,
			Type:        itemType,
			Title:       r.Name,
			Snippet:     snippet,
			Status:      status,
			Source:      "mcp",
			FirstSeenAt: now,
			LastSeenAt:  now,
		})
	}
	return items, nil
}

// ReadResource fetches resource contents from Axis Mundi via MCP resources/read.
func (c *Client) ReadResource(ctx context.Context, uri string) (string, error) {
	rpcReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "resources/read",
		"params": map[string]any{
			"uri": uri,
		},
	}
	payload, err := json.Marshal(rpcReq)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/mcp", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("MCP status %d", resp.StatusCode)
	}

	var rpcResp struct {
		Result struct {
			Contents []struct {
				URI      string `json:"uri"`
				MimeType string `json:"mimeType"`
				Text     string `json:"text"`
			} `json:"contents"`
		} `json:"result"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return "", err
	}
	if rpcResp.Error != nil {
		return "", fmt.Errorf("MCP RPC error: %s", rpcResp.Error.Message)
	}
	if len(rpcResp.Result.Contents) == 0 {
		return "", fmt.Errorf("empty resource content returned")
	}

	return rpcResp.Result.Contents[0].Text, nil
}

// CallTool invokes a tool on Axis Mundi via MCP tools/call.
func (c *Client) CallTool(ctx context.Context, name string, args map[string]any) (string, error) {
	if args == nil {
		args = make(map[string]any)
	}
	rpcReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      name,
			"arguments": args,
		},
	}
	payload, err := json.Marshal(rpcReq)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/mcp", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("MCP status %d", resp.StatusCode)
	}

	var rpcResp struct {
		Result struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
			IsError bool `json:"isError"`
		} `json:"result"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return "", err
	}
	if rpcResp.Error != nil {
		return "", fmt.Errorf("MCP RPC error: %s", rpcResp.Error.Message)
	}
	if rpcResp.Result.IsError {
		errMsg := "tool call failed"
		if len(rpcResp.Result.Content) > 0 {
			errMsg = rpcResp.Result.Content[0].Text
		}
		return "", fmt.Errorf("%s", errMsg)
	}
	if len(rpcResp.Result.Content) == 0 {
		return "", nil
	}
	return rpcResp.Result.Content[0].Text, nil
}

func parseStatusAndSnippet(desc string) (string, string) {
	desc = strings.TrimSpace(desc)
	if strings.HasPrefix(desc, "[") {
		idx := strings.Index(desc, "]")
		if idx != -1 {
			status := desc[1:idx]
			snippet := strings.TrimSpace(desc[idx+1:])
			return status, snippet
		}
	}
	return "Pending", desc
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

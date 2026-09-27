package grecorder

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

// Client is a type-safe JSON-RPC 2.0 client for Google Recorder MCP server.
type Client struct {
	mcpURL     string
	httpClient *http.Client
	requestID  uint64
}

// NewClient creates a new MCP client for the given Streamable HTTP endpoint.
func NewClient(mcpURL string) *Client {
	if mcpURL == "" {
		mcpURL = "http://localhost:8091/mcp"
	}
	return &Client{
		mcpURL: mcpURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Allow long-running operations like full library sync
		},
	}
}

// Ping checks if the MCP server endpoint is reachable.
func (c *Client) Ping(ctx context.Context) bool {
	ctxShort, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctxShort, http.MethodGet, c.mcpURL, nil)
	if err != nil {
		return false
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// CallTool invokes a named tool on the MCP server via JSON-RPC 2.0.
func (c *Client) CallTool(ctx context.Context, name string, args map[string]any) (string, error) {
	id := atomic.AddUint64(&c.requestID, 1)
	rpcReq := RPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  "tools/call",
		Params: ToolCallParams{
			Name:      name,
			Arguments: args,
		},
	}

	rawBody, err := json.Marshal(rpcReq)
	if err != nil {
		return "", fmt.Errorf("failed to encode tool call request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.mcpURL, bytes.NewReader(rawBody))
	if err != nil {
		return "", fmt.Errorf("failed to create http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send mcp request to %s: %w", c.mcpURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("mcp server returned status %d: %s", resp.StatusCode, string(body))
	}

	var rpcResp RPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return "", fmt.Errorf("failed to parse mcp response: %w", err)
	}

	if rpcResp.Error != nil {
		return "", fmt.Errorf("mcp rpc error [%d]: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	var toolRes ToolResult
	if err := json.Unmarshal(rpcResp.Result, &toolRes); err != nil {
		return "", fmt.Errorf("failed to parse tool result: %w", err)
	}

	var combinedText string
	for _, item := range toolRes.Content {
		if item.Type == "text" {
			combinedText += item.Text
		}
	}

	if toolRes.IsError {
		return combinedText, errors.New(combinedText)
	}

	return combinedText, nil
}

// AuthStatus checks session validity and authentication status.
func (c *Client) AuthStatus(ctx context.Context) (bool, error) {
	out, err := c.CallTool(ctx, "auth_status", map[string]any{})
	if err != nil {
		return false, err
	}
	var res struct {
		Authenticated bool   `json:"authenticated"`
		AuthValid     bool   `json:"auth_valid"`
		Status        string `json:"status"`
	}
	if err := json.Unmarshal([]byte(out), &res); err != nil {
		// If text response contains "valid" or "Authenticated"
		return true, nil
	}
	return res.Authenticated || res.AuthValid, nil
}

// ListRecordings queries recordings with metadata and IDs.
func (c *Client) ListRecordings(ctx context.Context, limit int, query string, all bool) ([]Recording, error) {
	args := map[string]any{
		"limit": limit,
		"all":   all,
	}
	if query != "" {
		args["query"] = query
	}

	out, err := c.CallTool(ctx, "list_recordings", args)
	if err != nil {
		return nil, err
	}

	// 1. Direct array attempt
	var direct []Recording
	if errDirect := json.Unmarshal([]byte(out), &direct); errDirect == nil {
		return direct, nil
	}

	// 2. Look for JSON array beginning '['
	if idx := strings.Index(out, "["); idx != -1 {
		if errArray := json.Unmarshal([]byte(out[idx:]), &direct); errArray == nil {
			return direct, nil
		}
	}

	// 3. Structured object attempt
	var res struct {
		Recordings []Recording `json:"recordings"`
		Total      int         `json:"total"`
	}
	if errObj := json.Unmarshal([]byte(out), &res); errObj == nil && len(res.Recordings) > 0 {
		return res.Recordings, nil
	}

	// 4. Look for JSON object beginning '{'
	if idx := strings.Index(out, "{"); idx != -1 {
		if errObj := json.Unmarshal([]byte(out[idx:]), &res); errObj == nil {
			return res.Recordings, nil
		}
	}

	return nil, fmt.Errorf("failed to parse recordings json: output: %s", out)
}

// GetRecording retrieves full metadata for a recording ID.
func (c *Client) GetRecording(ctx context.Context, recordingID string) (*Recording, error) {
	out, err := c.CallTool(ctx, "get_recording", map[string]any{
		"recording_id": recordingID,
	})
	if err != nil {
		return nil, err
	}

	var rec Recording
	if err := json.Unmarshal([]byte(out), &rec); err == nil {
		return &rec, nil
	}

	if idx := strings.Index(out, "{"); idx != -1 {
		if errObj := json.Unmarshal([]byte(out[idx:]), &rec); errObj == nil {
			return &rec, nil
		}
	}

	return nil, fmt.Errorf("failed to parse recording: %s", out)
}

// GetTranscript retrieves formatted text or structured JSON transcript.
func (c *Client) GetTranscript(ctx context.Context, recordingID string, format string) (string, error) {
	if format == "" {
		format = "json"
	}
	return c.CallTool(ctx, "get_transcript", map[string]any{
		"recording_id": recordingID,
		"format":       format,
	})
}

// DownloadRecording downloads a single recording side-by-side into outputDir.
func (c *Client) DownloadRecording(ctx context.Context, recordingID string, outputDir string, force bool) error {
	_, err := c.CallTool(ctx, "download_recording", map[string]any{
		"recording_id":        recordingID,
		"output_dir":          outputDir,
		"download_audio":      true,
		"download_transcript": true,
		"transcript_format":   "both",
		"download_metadata":   true,
		"force":               force,
	})
	return err
}

// SyncRecordings triggers an incremental library synchronization into outputDir.
func (c *Client) SyncRecordings(ctx context.Context, outputDir string, opts SyncOptions) (string, error) {
	args := map[string]any{
		"output_dir":          outputDir,
		"download_audio":      opts.DownloadAudio,
		"download_transcript": opts.DownloadTranscript,
		"transcript_format":   opts.TranscriptFormat,
		"download_metadata":   opts.DownloadMetadata,
		"force":               opts.Force,
		"all":                 opts.All,
	}
	if opts.Limit > 0 {
		args["limit"] = opts.Limit
	}
	if opts.Query != "" {
		args["query"] = opts.Query
	}

	return c.CallTool(ctx, "sync_recordings", args)
}

// GetSyncStatus retrieves sync status or folder manifest stats.
func (c *Client) GetSyncStatus(ctx context.Context, outputDir string) (*SyncStatus, error) {
	out, err := c.CallTool(ctx, "get_sync_status", map[string]any{
		"output_dir": outputDir,
	})
	if err != nil {
		return nil, err
	}

	var status SyncStatus
	if err := json.Unmarshal([]byte(out), &status); err != nil {
		return nil, fmt.Errorf("failed to parse sync status: %w", err)
	}
	return &status, nil
}

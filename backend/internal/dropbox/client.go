package dropbox

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var (
	ErrNotConfigured    = errors.New("dropbox integration is not configured")
	ErrTokenExpired     = errors.New("dropbox access token expired")
	ErrPermissionDenied = errors.New("dropbox permission denied or scope missing")
)

type AccountInfo struct {
	AccountID      string `json:"account_id"`
	DisplayName    string `json:"display_name"`
	Email          string `json:"email"`
	AccountType    string `json:"account_type"`
	Country        string `json:"country"`
	UsedBytes      uint64 `json:"used_bytes,omitempty"`
	AllocatedBytes uint64 `json:"allocated_bytes,omitempty"`
}

type FileEntry struct {
	Tag            string    `json:".tag"`
	Name           string    `json:"name"`
	PathLower      string    `json:"path_lower"`
	PathDisplay    string    `json:"path_display"`
	Size           int64     `json:"size,omitempty"`
	ServerModified time.Time `json:"server_modified,omitempty"`
}

type Client struct {
	accessToken  string
	refreshToken string
	appKey       string
	appSecret    string
	basePath     string
	httpClient   *http.Client
	mu           sync.RWMutex
}

func NewClient(accessToken, refreshToken, appKey, appSecret, basePath string) *Client {
	if basePath == "" {
		basePath = "/MercuryDasha/backups"
	}
	// Ensure basePath starts with / and doesn't end with /
	if !strings.HasPrefix(basePath, "/") {
		basePath = "/" + basePath
	}
	basePath = strings.TrimSuffix(basePath, "/")

	return &Client{
		accessToken:  strings.TrimSpace(accessToken),
		refreshToken: strings.TrimSpace(refreshToken),
		appKey:       strings.TrimSpace(appKey),
		appSecret:    strings.TrimSpace(appSecret),
		basePath:     basePath,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *Client) IsConfigured() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.accessToken != "" || (c.refreshToken != "" && c.appKey != "" && c.appSecret != "")
}

func (c *Client) GetBasePath() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.basePath
}

func (c *Client) GetAccessToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.accessToken
}

func (c *Client) SetAccessToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.accessToken = strings.TrimSpace(token)
}

// RefreshToken exchanges the refresh_token for a new short-lived access token.
func (c *Client) RefreshToken(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.refreshToken == "" || c.appKey == "" || c.appSecret == "" {
		return errors.New("cannot refresh token: refresh_token, app_key, or app_secret missing")
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", c.refreshToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.dropboxapi.com/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.appKey, c.appSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute refresh request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("dropbox refresh failed (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var res struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return fmt.Errorf("failed to parse refresh response: %w", err)
	}

	if res.AccessToken == "" {
		return errors.New("empty access token received from refresh endpoint")
	}

	c.accessToken = res.AccessToken
	return nil
}

func (c *Client) doRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	token := c.GetAccessToken()
	if token == "" {
		if err := c.RefreshToken(ctx); err != nil {
			return nil, ErrNotConfigured
		}
		token = c.GetAccessToken()
	}

	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// If 401 Unauthorized, try refreshing token once if refresh credentials exist
	if resp.StatusCode == http.StatusUnauthorized && c.refreshToken != "" && c.appKey != "" {
		resp.Body.Close()
		if refreshErr := c.RefreshToken(ctx); refreshErr == nil {
			// Retry request with new token
			req.Header.Set("Authorization", "Bearer "+c.GetAccessToken())
			return c.httpClient.Do(req)
		}
	}

	return resp, nil
}

// GetAccountInfo retrieves the authenticated user's profile and space usage.
func (c *Client) GetAccountInfo(ctx context.Context) (*AccountInfo, error) {
	if !c.IsConfigured() {
		return nil, ErrNotConfigured
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.dropboxapi.com/2/users/get_current_account", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get account info (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var raw struct {
		AccountID   string `json:"account_id"`
		AccountType struct {
			Tag string `json:".tag"`
		} `json:"account_type"`
		Country string `json:"country"`
		Email   string `json:"email"`
		Name    struct {
			DisplayName string `json:"display_name"`
		} `json:"name"`
	}

	if err := json.Unmarshal(bodyBytes, &raw); err != nil {
		return nil, fmt.Errorf("failed to decode account info: %w", err)
	}

	info := &AccountInfo{
		AccountID:   raw.AccountID,
		DisplayName: raw.Name.DisplayName,
		Email:       raw.Email,
		AccountType: raw.AccountType.Tag,
		Country:     raw.Country,
	}

	// Try fetching space usage (optional)
	spaceReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.dropboxapi.com/2/users/get_space_usage", nil)
	if err == nil {
		if spaceResp, err := c.doRequest(ctx, spaceReq); err == nil {
			defer spaceResp.Body.Close()
			if spaceResp.StatusCode == http.StatusOK {
				var space struct {
					Used uint64 `json:"used"`
					Allocation struct {
						Tag       string `json:".tag"`
						Allocated uint64 `json:"allocated"`
					} `json:"allocation"`
				}
				if json.NewDecoder(spaceResp.Body).Decode(&space) == nil {
					info.UsedBytes = space.Used
					info.AllocatedBytes = space.Allocation.Allocated
				}
			}
		}
	}

	return info, nil
}

// UploadFile uploads a stream directly to the target Dropbox path.
func (c *Client) UploadFile(ctx context.Context, dropboxPath string, content io.Reader) (*FileEntry, error) {
	if !c.IsConfigured() {
		return nil, ErrNotConfigured
	}

	if !strings.HasPrefix(dropboxPath, "/") {
		dropboxPath = c.basePath + "/" + dropboxPath
	}

	argJSON, err := json.Marshal(map[string]any{
		"path":            dropboxPath,
		"mode":            "overwrite",
		"autorename":      false,
		"mute":            false,
		"strict_conflict": false,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://content.dropboxapi.com/2/files/upload", content)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Dropbox-API-Arg", string(argJSON))
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upload failed (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var entry FileEntry
	if err := json.Unmarshal(bodyBytes, &entry); err != nil {
		return nil, fmt.Errorf("failed to decode upload response: %w", err)
	}

	return &entry, nil
}

// ListFolder lists files in the given folder path. If path is empty, uses configured basePath.
func (c *Client) ListFolder(ctx context.Context, folderPath string) ([]FileEntry, error) {
	if !c.IsConfigured() {
		return nil, ErrNotConfigured
	}

	if folderPath == "" {
		folderPath = c.basePath
	} else if !strings.HasPrefix(folderPath, "/") {
		folderPath = c.basePath + "/" + folderPath
	}

	// Dropbox root is empty string ""
	if folderPath == "/" {
		folderPath = ""
	}

	payload, err := json.Marshal(map[string]any{
		"path":                                folderPath,
		"recursive":                           false,
		"include_media_info":                  false,
		"include_deleted":                     false,
		"include_has_explicit_shared_members": false,
		"include_mounted_folders":             true,
		"limit":                               100,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.dropboxapi.com/2/files/list_folder", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		if strings.Contains(string(bodyBytes), "not_found") {
			return []FileEntry{}, nil
		}
		return nil, fmt.Errorf("list_folder failed (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var res struct {
		Entries []FileEntry `json:"entries"`
		HasMore bool        `json:"has_more"`
		Cursor  string      `json:"cursor"`
	}
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, fmt.Errorf("failed to decode list_folder response: %w", err)
	}

	return res.Entries, nil
}

// GetTemporaryLink generates a short-lived download link for a Dropbox file.
func (c *Client) GetTemporaryLink(ctx context.Context, dropboxPath string) (string, error) {
	if !c.IsConfigured() {
		return "", ErrNotConfigured
	}

	if !strings.HasPrefix(dropboxPath, "/") {
		dropboxPath = c.basePath + "/" + dropboxPath
	}

	payload, err := json.Marshal(map[string]string{
		"path": dropboxPath,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.dropboxapi.com/2/files/get_temporary_link", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get_temporary_link failed (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var res struct {
		Link string `json:"link"`
	}
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return "", fmt.Errorf("failed to decode temporary link: %w", err)
	}

	return res.Link, nil
}

type SearchMatch struct {
	Tag            string    `json:".tag"`
	Name           string    `json:"name"`
	PathDisplay    string    `json:"path_display"`
	Size           int64     `json:"size,omitempty"`
	ServerModified time.Time `json:"server_modified,omitempty"`
	Highlight      string    `json:"highlight,omitempty"`
}

// SearchFiles searches for files/folders matching query within folderPath.
func (c *Client) SearchFiles(ctx context.Context, query, folderPath string, maxResults int) ([]SearchMatch, error) {
	if !c.IsConfigured() {
		return nil, ErrNotConfigured
	}

	if maxResults <= 0 || maxResults > 100 {
		maxResults = 25
	}

	options := map[string]any{
		"max_results": maxResults,
	}
	if folderPath != "" {
		if !strings.HasPrefix(folderPath, "/") {
			folderPath = "/" + folderPath
		}
		if folderPath != "/" {
			options["path"] = folderPath
		}
	}

	payload, err := json.Marshal(map[string]any{
		"query":   query,
		"options": options,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.dropboxapi.com/2/files/search_v2", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var res struct {
		Matches []struct {
			Metadata struct {
				Metadata FileEntry `json:"metadata"`
			} `json:"metadata"`
			HighlightSpans []struct {
				HighlightStr string `json:"highlight_str"`
			} `json:"highlight_spans"`
		} `json:"matches"`
	}
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return nil, fmt.Errorf("failed to decode search results: %w", err)
	}

	var results []SearchMatch
	for _, m := range res.Matches {
		entry := m.Metadata.Metadata
		highlight := ""
		if len(m.HighlightSpans) > 0 {
			highlight = m.HighlightSpans[0].HighlightStr
		}
		results = append(results, SearchMatch{
			Tag:            entry.Tag,
			Name:           entry.Name,
			PathDisplay:    entry.PathDisplay,
			Size:           entry.Size,
			ServerModified: entry.ServerModified,
			Highlight:      highlight,
		})
	}

	return results, nil
}

// DownloadFile streams a file's content directly to the provided writer.
func (c *Client) DownloadFile(ctx context.Context, dropboxPath string, w io.Writer) (*FileEntry, error) {
	if !c.IsConfigured() {
		return nil, ErrNotConfigured
	}

	if !strings.HasPrefix(dropboxPath, "/") {
		dropboxPath = c.basePath + "/" + dropboxPath
	}

	argJSON, err := json.Marshal(map[string]string{
		"path": dropboxPath,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://content.dropboxapi.com/2/files/download", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Dropbox-API-Arg", string(argJSON))

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download failed (%d): %s", resp.StatusCode, string(errBody))
	}

	metaHeader := resp.Header.Get("Dropbox-API-Result")
	var entry FileEntry
	if metaHeader != "" {
		_ = json.Unmarshal([]byte(metaHeader), &entry)
	}

	if _, err := io.Copy(w, resp.Body); err != nil {
		return nil, fmt.Errorf("failed writing downloaded stream: %w", err)
	}

	return &entry, nil
}

// GetFileMetadata retrieves detailed metadata for a file or folder.
func (c *Client) GetFileMetadata(ctx context.Context, dropboxPath string) (*FileEntry, error) {
	if !c.IsConfigured() {
		return nil, ErrNotConfigured
	}

	if !strings.HasPrefix(dropboxPath, "/") {
		dropboxPath = c.basePath + "/" + dropboxPath
	}

	payload, err := json.Marshal(map[string]any{
		"path":               dropboxPath,
		"include_media_info": true,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.dropboxapi.com/2/files/get_metadata", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get_metadata failed (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var entry FileEntry
	if err := json.Unmarshal(bodyBytes, &entry); err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %w", err)
	}

	return &entry, nil
}

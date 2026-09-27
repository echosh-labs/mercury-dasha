package syncthing

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client handles authenticated HTTP REST interactions with the local Syncthing daemon.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	pollClient *http.Client
}

// NewClient creates a new Syncthing REST API client.
func NewClient(baseURL, apiKey string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8384"
	}
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		pollClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	reqURL := fmt.Sprintf("%s%s", c.baseURL, path)
	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return nil, err
	}
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}
	req.Header.Set("Accept", "application/json")
	return req, nil
}

// Ping checks if the Syncthing daemon is online and responding.
func (c *Client) Ping(ctx context.Context) (bool, string, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/rest/system/version", nil)
	if err != nil {
		return false, "", err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("syncthing returned status %d", resp.StatusCode)
	}

	var data struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return false, "", err
	}
	return true, data.Version, nil
}

// GetDevices returns configured and active remote devices.
func (c *Client) GetDevices(ctx context.Context) ([]DeviceConfig, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/rest/config/devices", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var devices []DeviceConfig
	if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
		return nil, err
	}

	// Enrich with live connection status
	connReq, err := c.newRequest(ctx, http.MethodGet, "/rest/system/connections", nil)
	if err == nil {
		connResp, err := c.httpClient.Do(connReq)
		if err == nil {
			defer connResp.Body.Close()
			var connData struct {
				Connections map[string]struct {
					Connected bool   `json:"connected"`
					Address   string `json:"address"`
					At        string `json:"at"`
				} `json:"connections"`
			}
			if err := json.NewDecoder(connResp.Body).Decode(&connData); err == nil {
				for i := range devices {
					if conn, ok := connData.Connections[devices[i].DeviceID]; ok {
						devices[i].Connected = conn.Connected
						devices[i].Address = conn.Address
						if conn.At != "" {
							if t, err := time.Parse(time.RFC3339, conn.At); err == nil {
								devices[i].LastSeen = t
							}
						}
					}
				}
			}
		}
	}

	return devices, nil
}

// GetFolders returns configured sync folders and their current stats.
func (c *Client) GetFolders(ctx context.Context) ([]FolderConfig, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/rest/config/folders", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var folders []FolderConfig
	if err := json.NewDecoder(resp.Body).Decode(&folders); err != nil {
		return nil, err
	}

	for i := range folders {
		stReq, err := c.newRequest(ctx, http.MethodGet, fmt.Sprintf("/rest/db/status?folder=%s", url.QueryEscape(folders[i].ID)), nil)
		if err == nil {
			stResp, err := c.httpClient.Do(stReq)
			if err == nil {
				var status FolderDBStatus
				if err := json.NewDecoder(stResp.Body).Decode(&status); err == nil {
					folders[i].State = status.State
					folders[i].GlobalFiles = status.GlobalFiles
					folders[i].GlobalBytes = status.GlobalBytes
				}
				_ = stResp.Body.Close()
			}
		}
	}

	return folders, nil
}

// RescanFolder triggers an immediate rescan of a folder.
func (c *Client) RescanFolder(ctx context.Context, folderID string) error {
	path := "/rest/db/scan"
	if folderID != "" {
		path = fmt.Sprintf("/rest/db/scan?folder=%s", url.QueryEscape(folderID))
	}
	req, err := c.newRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("rescan returned status %d", resp.StatusCode)
	}
	return nil
}

// GetEvents polls Syncthing's SSE event bus starting from `since` event ID.
func (c *Client) GetEvents(ctx context.Context, since int64, limit int, timeoutSec int) ([]Event, error) {
	if timeoutSec <= 0 {
		timeoutSec = 30
	}
	if limit <= 0 {
		limit = 100
	}

	u := fmt.Sprintf("/rest/events?since=%d&limit=%d&timeout=%d", since, limit, timeoutSec)
	req, err := c.newRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.pollClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("events returned status %d", resp.StatusCode)
	}

	var events []Event
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, err
	}
	return events, nil
}

package dropbox

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestDropboxClient_NotConfigured(t *testing.T) {
	client := NewClient("", "", "", "", "")
	if client.IsConfigured() {
		t.Fatalf("expected client to not be configured")
	}

	_, err := client.GetAccountInfo(context.Background())
	if err != ErrNotConfigured {
		t.Fatalf("expected ErrNotConfigured, got %v", err)
	}

	_, err = client.UploadFile(context.Background(), "test.db", strings.NewReader("data"))
	if err != ErrNotConfigured {
		t.Fatalf("expected ErrNotConfigured, got %v", err)
	}
}

func TestDropboxClient_Configured(t *testing.T) {
	client := NewClient("sl.u.mocktoken", "", "", "", "/Custom/Backups")
	if !client.IsConfigured() {
		t.Fatalf("expected client to be configured")
	}
	if client.GetBasePath() != "/Custom/Backups" {
		t.Fatalf("expected basePath /Custom/Backups, got %s", client.GetBasePath())
	}
}

func TestDropboxClient_MockAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		switch r.URL.Path {
		case "/2/users/get_current_account":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"account_id": "dbid:12345",
				"email":      "test@example.com",
				"country":    "CA",
				"account_type": map[string]string{
					".tag": "pro",
				},
				"name": map[string]string{
					"display_name": "Test User",
				},
			})
		case "/2/files/list_folder":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"entries": []map[string]any{
					{
						".tag":         "file",
						"name":         "mercury-backup-1.db",
						"path_lower":   "/mercurydasha/backups/mercury-backup-1.db",
						"path_display": "/MercuryDasha/backups/mercury-backup-1.db",
						"size":         1024,
					},
				},
				"has_more": false,
				"cursor":   "cursor-123",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClient("test-token", "", "", "", "/MercuryDasha/backups")
	client.httpClient = server.Client()

	// Account info test
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, server.URL+"/2/users/get_current_account", nil)
	resp, err := client.doRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
}

func TestDropboxClient_LiveRefreshIntegration(t *testing.T) {
	refreshToken := os.Getenv("DROPBOX_REFRESH_TOKEN")
	appKey := os.Getenv("DROPBOX_APP_KEY")
	appSecret := os.Getenv("DROPBOX_APP_SECRET")

	if refreshToken == "" || appKey == "" || appSecret == "" {
		t.Skip("skipping live Dropbox integration test: credentials not in environment")
	}

	client := NewClient("", refreshToken, appKey, appSecret, "/MercuryDasha/backups")
	account, err := client.GetAccountInfo(context.Background())
	if err != nil {
		t.Fatalf("live GetAccountInfo failed with auto-refresh: %v", err)
	}

	if account.Email == "" {
		t.Errorf("expected non-empty account email, got %q", account.Email)
	}
}


package axismundi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_FetchRegistryREST(t *testing.T) {
	mockItems := []RawRegistryItem{
		{ID: "keep-1", Type: "keep", Title: "Weekly Alchemy Directives", Snippet: "Review transits", Status: "Pending"},
		{ID: "doc-1", Type: "doc", Title: "Foundations Volume III", Snippet: "Chapter 4 drafting", Status: "Active"},
		{ID: "sheet-1", Type: "sheet", Title: "Q3 Treasury Ledger", Snippet: "Burn rates", Status: "Review"},
		{ID: "gmail-1", Type: "gmail", Title: "Project Falcon Launch Update", Snippet: "All systems go", Status: "Complete"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/health" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "UP", "mode": "AUTO"})
			return
		}
		if r.URL.Path == "/api/registry" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(mockItems)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "")
	live, mode, err := client.Ping(context.Background())
	if err != nil || !live || mode != "AUTO" {
		t.Fatalf("Ping failed: live=%v mode=%s err=%v", live, mode, err)
	}

	items, err := client.FetchRegistry(context.Background())
	if err != nil {
		t.Fatalf("FetchRegistry error: %v", err)
	}

	if len(items) != 4 {
		t.Fatalf("Expected 4 items, got %d", len(items))
	}

	// Verify categorization
	typesFound := map[WorkspaceType]bool{}
	for _, it := range items {
		typesFound[it.Type] = true
	}
	if !typesFound[TypeKeep] || !typesFound[TypeDoc] || !typesFound[TypeSheet] || !typesFound[TypeGmail] {
		t.Errorf("Missing expected types: %+v", typesFound)
	}
}

func TestListener_ProcessItemsAndAlerts(t *testing.T) {
	listener := NewListener("http://localhost:8088", "", nil, nil)

	batch1 := []WorkspaceItem{
		{ID: "keep-1", Type: TypeKeep, Title: "Transmutation Notes", Snippet: "Gold synthesis"},
		{ID: "gmail-1", Type: TypeGmail, Title: "Security Advisory", Snippet: "Patch applied"},
	}

	alerts1 := listener.ProcessItems(batch1)
	if len(alerts1) != 2 {
		t.Fatalf("Expected 2 alerts on initial discovery, got %d", len(alerts1))
	}

	// Subsequent poll with same items should NOT generate new alerts
	alerts2 := listener.ProcessItems(batch1)
	if len(alerts2) != 0 {
		t.Fatalf("Expected 0 alerts for known items, got %d", len(alerts2))
	}

	// Incoming new Google Doc
	batch2 := append(batch1, WorkspaceItem{
		ID: "doc-1", Type: TypeDoc, Title: "Kybalion Commentary", Snippet: "The Principle of Polarity",
	})

	alerts3 := listener.ProcessItems(batch2)
	if len(alerts3) != 1 {
		t.Fatalf("Expected 1 new alert for newly added doc, got %d", len(alerts3))
	}
	if alerts3[0].ItemID != "doc-1" {
		t.Errorf("Expected alert for doc-1, got %s", alerts3[0].ItemID)
	}

	// Verify Feed
	feed := listener.GetFeed("", false)
	if feed.Total != 3 {
		t.Errorf("Expected feed total 3, got %d", feed.Total)
	}
	if feed.Status.Counts.KeepNotes != 1 || feed.Status.Counts.Gmail != 1 || feed.Status.Counts.Docs != 1 {
		t.Errorf("Unexpected feed counts: %+v", feed.Status.Counts)
	}
}

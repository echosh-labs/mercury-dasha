package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/api"
	"github.com/echosh-labs/mercury-dasha/internal/config"
	"github.com/echosh-labs/mercury-dasha/internal/dasha"
	"github.com/echosh-labs/mercury-dasha/internal/db"
)

func setupTestMCPServer(t *testing.T) (*Server, *Handler, db.StorageEngine, func()) {
	tmpDir, err := os.MkdirTemp("", "mcp-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open test db: %v", err)
	}

	// Seed profile
	start := time.Date(2009, 4, 13, 0, 0, 0, 0, time.UTC)
	end := time.Date(2028, 4, 13, 0, 0, 0, 0, time.UTC)
	mockTimeline := []dasha.DashaPeriod{
		{
			Level:        dasha.LevelMahadasha,
			Planet:       dasha.PlanetSaturn,
			PlanetName:   "Saturn",
			SanskritName: "Shani",
			StartDate:    start,
			EndDate:      end,
			SubPeriods: []dasha.DashaPeriod{
				{
					Level:        dasha.LevelAntardasha,
					Planet:       dasha.PlanetJupiter,
					PlanetName:   "Jupiter",
					SanskritName: "Guru",
					StartDate:    time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC),
					EndDate:      end,
				},
			},
		},
	}
	prof := dasha.DashaProfile{
		ID:       "profile:sovereign-genesis",
		Timeline: mockTimeline,
	}
	raw, _ := json.Marshal(prof)
	_ = store.SaveProfile("profile:sovereign-genesis", raw)

	cfg := &config.Config{
		DropboxLocalPath: tmpDir,
		Port:             "8080",
	}

	apiH := api.NewHandler(cfg, store)
	handler := NewHandler(apiH, store, cfg)
	srv := NewServer(handler, "")

	cleanup := func() {
		_ = store.Close()
		_ = os.RemoveAll(tmpDir)
	}

	return srv, handler, store, cleanup
}

func TestMCPInitialize(t *testing.T) {
	_, handler, _, cleanup := setupTestMCPServer(t)
	defer cleanup()

	req := Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  map[string]any{},
	}
	reqBytes, _ := json.Marshal(req)

	resp := handler.HandleRequest(context.Background(), reqBytes)
	if resp == nil || resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp)
	}

	resultMap, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("Expected map result, got %T", resp.Result)
	}

	if resultMap["protocolVersion"] != ProtocolVersion {
		t.Errorf("Expected protocol version %s, got %v", ProtocolVersion, resultMap["protocolVersion"])
	}
}

func TestMCPToolsListAndCall(t *testing.T) {
	_, handler, _, cleanup := setupTestMCPServer(t)
	defer cleanup()

	// 1. tools/list
	reqList := Request{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}
	rawList, _ := json.Marshal(reqList)
	respList := handler.HandleRequest(context.Background(), rawList)
	if respList == nil || respList.Error != nil {
		t.Fatalf("Unexpected tools/list error: %v", respList)
	}

	resMap := respList.Result.(map[string]any)
	toolsList := resMap["tools"].([]Tool)
	if len(toolsList) < 7 {
		t.Fatalf("Expected at least 7 tools, got %d", len(toolsList))
	}

	// 2. tools/call: get_active_cosmic_pulse
	reqPulse := Request{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params: map[string]any{
			"name":      "get_active_cosmic_pulse",
			"arguments": map[string]any{},
		},
	}
	rawPulse, _ := json.Marshal(reqPulse)
	respPulse := handler.HandleRequest(context.Background(), rawPulse)
	if respPulse == nil || respPulse.Error != nil {
		t.Fatalf("get_active_cosmic_pulse call failed: %v", respPulse)
	}

	tResult := respPulse.Result.(*ToolResult)
	if len(tResult.Content) == 0 {
		t.Fatalf("Expected tool content result")
	}

	var pulseData map[string]any
	if err := json.Unmarshal([]byte(tResult.Content[0].Text), &pulseData); err != nil {
		t.Fatalf("Failed to parse tool result JSON: %v", err)
	}

	if pulseData["active_mahadasha"] != "Saturn" {
		t.Errorf("Expected Saturn, got %v", pulseData["active_mahadasha"])
	}
}

func TestMCPResourcesListAndRead(t *testing.T) {
	_, handler, _, cleanup := setupTestMCPServer(t)
	defer cleanup()

	// 1. resources/list
	reqList := Request{
		JSONRPC: "2.0",
		ID:      4,
		Method:  "resources/list",
	}
	rawList, _ := json.Marshal(reqList)
	respList := handler.HandleRequest(context.Background(), rawList)
	if respList == nil || respList.Error != nil {
		t.Fatalf("resources/list error: %v", respList)
	}

	// 2. resources/read: mercury://ancestral
	reqRead := Request{
		JSONRPC: "2.0",
		ID:      5,
		Method:  "resources/read",
		Params: map[string]any{
			"uri": "mercury://ancestral",
		},
	}
	rawRead, _ := json.Marshal(reqRead)
	respRead := handler.HandleRequest(context.Background(), rawRead)
	if respRead == nil || respRead.Error != nil {
		t.Fatalf("resources/read error: %v", respRead)
	}

	readMap := respRead.Result.(map[string]any)
	contents := readMap["contents"].([]ResourceContent)
	if len(contents) == 0 || !strings.Contains(contents[0].Text, "Vernon Douglas Wood") {
		t.Errorf("Expected Vernon Douglas Wood in ancestral resource, got %+v", contents)
	}
}

func TestMCPHTTPTransport(t *testing.T) {
	srv, _, _, cleanup := setupTestMCPServer(t)
	defer cleanup()

	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)

	// Test GET /mcp (discovery info)
	reqGet := httptest.NewRequest("GET", "/mcp", nil)
	recGet := httptest.NewRecorder()
	mux.ServeHTTP(recGet, reqGet)

	if recGet.Code != http.StatusOK {
		t.Fatalf("Expected 200 from GET /mcp, got %d", recGet.Code)
	}

	// Test POST /mcp (JSON-RPC initialize)
	initReq := Request{
		JSONRPC: "2.0",
		ID:      100,
		Method:  "initialize",
	}
	initBody, _ := json.Marshal(initReq)
	reqPost := httptest.NewRequest("POST", "/mcp", bytes.NewReader(initBody))
	reqPost.Header.Set("Content-Type", "application/json")
	recPost := httptest.NewRecorder()
	mux.ServeHTTP(recPost, reqPost)

	if recPost.Code != http.StatusOK {
		t.Fatalf("Expected 200 from POST /mcp, got %d: %s", recPost.Code, recPost.Body.String())
	}

	var resp Response
	if err := json.NewDecoder(recPost.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("Unexpected JSON-RPC error: %v", resp.Error)
	}
}

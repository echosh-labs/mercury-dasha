package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/db"
)

func TestUnifiedVisualCatalogHandler(t *testing.T) {
	h, mux, _, cleanup := setupTestRecorderHandler(t)
	defer cleanup()

	// Seed index with Pixel 7 photo, Dropbox photo, and Dropbox video
	entries := []db.IndexEntry{
		{
			ID:        "media/pixel7/PXL_20240223_080000.jpg",
			Category:  "photos",
			Path:      "pixel7/PXL_20240223_080000.jpg",
			FullPath:  "/home/justin/media/pixel7/PXL_20240223_080000.jpg",
			FileName:  "PXL_20240223_080000.jpg",
			Extension: ".jpg",
			SizeBytes: 2048,
			ModTime:   time.Date(2024, 2, 23, 8, 0, 0, 0, time.UTC),
			Tags:      []string{"photos", "jpg", "phone", "year:2024"},
			Metadata: map[string]any{
				"album":            "pixel7",
				"photo_date":       "2024-02-23",
				"year":             2024,
				"dasha_mahadasha":  "Saturn",
				"dasha_antardasha": "Rahu",
				"sacred_metal":     "Lead (Plumbum)",
				"device_model":     "Pixel 7",
			},
		},
		{
			ID:        "Pictures/2023/vacation.png",
			Category:  "photos",
			Path:      "Pictures/2023/vacation.png",
			FullPath:  "/home/justin/Dropbox/Pictures/2023/vacation.png",
			FileName:  "vacation.png",
			Extension: ".png",
			SizeBytes: 4096,
			ModTime:   time.Date(2023, 7, 10, 14, 0, 0, 0, time.UTC),
			Tags:      []string{"photos", "png", "year:2023"},
			Metadata: map[string]any{
				"album":            "Pictures",
				"photo_date":       "2023-07-10",
				"year":             2023,
				"dasha_mahadasha":  "Jupiter",
				"dasha_antardasha": "Moon",
				"sacred_metal":     "Tin (Stannum)",
			},
		},
		{
			ID:        "video/clip_2021.mp4",
			Category:  "video",
			Path:      "video/clip_2021.mp4",
			FullPath:  "/home/justin/Dropbox/video/clip_2021.mp4",
			FileName:  "clip_2021.mp4",
			Extension: ".mp4",
			SizeBytes: 8192,
			ModTime:   time.Date(2021, 5, 1, 10, 0, 0, 0, time.UTC),
			Tags:      []string{"video", "mp4", "year:2021"},
			Metadata: map[string]any{
				"album":            "video",
				"duration_sec":     45.0,
				"resolution":       "1920x1080",
				"photo_date":       "2021-05-01",
				"year":             2021,
				"dasha_mahadasha":  "Saturn",
				"dasha_antardasha": "Saturn",
				"sacred_metal":     "Lead (Plumbum)",
			},
		},
	}

	if err := h.store.BatchPutIndexEntries(entries); err != nil {
		t.Fatalf("failed to seed index entries: %v", err)
	}

	h.InvalidateVisualCache()

	// 1. Fetch entire catalog
	req := httptest.NewRequest("GET", "/api/v1/media/catalog", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp UnifiedVisualCatalogResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Total != 3 {
		t.Errorf("expected 3 total items, got %d", resp.Total)
	}
	if resp.Counts.Pixel7 != 1 {
		t.Errorf("expected 1 pixel7 item, got %d", resp.Counts.Pixel7)
	}
	if resp.Counts.DropboxPhotos != 1 {
		t.Errorf("expected 1 dropbox photos item, got %d", resp.Counts.DropboxPhotos)
	}
	if resp.Counts.DropboxVideos != 1 {
		t.Errorf("expected 1 dropbox videos item, got %d", resp.Counts.DropboxVideos)
	}

	// 2. Filter by source=pixel7
	pReq := httptest.NewRequest("GET", "/api/v1/media/catalog?source=pixel7", nil)
	pRR := httptest.NewRecorder()
	mux.ServeHTTP(pRR, pReq)
	var pResp UnifiedVisualCatalogResponse
	_ = json.NewDecoder(pRR.Body).Decode(&pResp)
	if pResp.Total != 1 || len(pResp.Items) != 1 || pResp.Items[0].Source != "pixel7" {
		t.Errorf("expected 1 pixel7 item, got %d", pResp.Total)
	}

	// 3. Filter by category=video
	vReq := httptest.NewRequest("GET", "/api/v1/media/catalog?category=video", nil)
	vRR := httptest.NewRecorder()
	mux.ServeHTTP(vRR, vReq)
	var vResp UnifiedVisualCatalogResponse
	_ = json.NewDecoder(vRR.Body).Decode(&vResp)
	if vResp.Total != 1 || vResp.Items[0].Category != "video" {
		t.Errorf("expected 1 video item, got %d", vResp.Total)
	}

	// 4. Filter by dasha=Jupiter
	dReq := httptest.NewRequest("GET", "/api/v1/media/catalog?dasha=Jupiter", nil)
	dRR := httptest.NewRecorder()
	mux.ServeHTTP(dRR, dReq)
	var dResp UnifiedVisualCatalogResponse
	_ = json.NewDecoder(dRR.Body).Decode(&dResp)
	if dResp.Total != 1 || dResp.Items[0].DashaMahadasha != "Jupiter" {
		t.Errorf("expected 1 Jupiter item, got %d", dResp.Total)
	}

	// 5. Filter by year=2024
	yReq := httptest.NewRequest("GET", "/api/v1/media/catalog?year=2024", nil)
	yRR := httptest.NewRecorder()
	mux.ServeHTTP(yRR, yReq)
	var yResp UnifiedVisualCatalogResponse
	_ = json.NewDecoder(yRR.Body).Decode(&yResp)
	if yResp.Total != 1 || yResp.Items[0].Year != 2024 {
		t.Errorf("expected 1 2024 item, got %d", yResp.Total)
	}
}

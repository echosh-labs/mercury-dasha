package api

import (
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

type Router struct {
	mux     *http.ServeMux
	handler *Handler
}

func NewRouter(handler *Handler) *Router {
	mux := http.NewServeMux()

	// Health & Telemetry
	mux.HandleFunc("GET /healthz", handler.HealthzHandler)
	mux.HandleFunc("GET /api/health", handler.HealthzHandler)
	mux.HandleFunc("GET /api/telemetry", handler.TelemetryHandler)

	// Mercury Dasha Domain Endpoints
	mux.HandleFunc("GET /api/dasha/overview", handler.DashaOverviewHandler)
	mux.HandleFunc("GET /api/dasha/nakshatras", handler.DashaNakshatrasHandler)
	mux.HandleFunc("GET /api/dasha/planets", handler.DashaPlanetsHandler)
	mux.HandleFunc("GET /api/dasha/calculate", handler.DashaCalculateHandler)
	mux.HandleFunc("POST /api/dasha/calculate", handler.DashaCalculateHandler)
	mux.HandleFunc("GET /api/dasha/hora", handler.DashaHoraHandler)
	mux.HandleFunc("GET /api/dasha/alchemy", handler.DashaAlchemyHandler)
	mux.HandleFunc("GET /api/v1/hora/active", handler.HoraActiveHandler)
	mux.HandleFunc("GET /api/v1/hora/schedule", handler.HoraScheduleHandler)

	// Administrative System Settings
	mux.HandleFunc("GET /api/v1/settings", handler.GetSettingsHandler)
	mux.HandleFunc("POST /api/v1/settings", handler.SaveSettingsHandler)
	mux.HandleFunc("PUT /api/v1/settings", handler.SaveSettingsHandler)

	// Chrono-Pulse Real-Time Event Stream
	mux.HandleFunc("GET /api/v1/stream/pulse", handler.ChronoPulseHandler)

	// Sovereign Profile Persistence & Astrological-Alchemical Resonance
	mux.HandleFunc("GET /api/v1/profiles", handler.ListProfilesHandler)
	mux.HandleFunc("POST /api/v1/profiles", handler.CreateProfileHandler)
	mux.HandleFunc("GET /api/v1/profiles/{id...}", handler.GetProfileHandler)
	mux.HandleFunc("POST /api/v1/profiles/{id...}", handler.SaveProfileHandler)
	mux.HandleFunc("PUT /api/v1/profiles/{id...}", handler.SaveProfileHandler)
	mux.HandleFunc("DELETE /api/v1/profiles/{id...}", handler.DeleteProfileHandler)

	// Sovereign Characters & Audiovisual Chronicles Studio
	mux.HandleFunc("GET /api/v1/characters", handler.ListCharactersHandler)
	mux.HandleFunc("POST /api/v1/characters", handler.CreateProfileHandler)
	mux.HandleFunc("GET /api/v1/characters/{id...}", handler.GetCharacterHandler)
	mux.HandleFunc("POST /api/v1/characters/{id...}", handler.SaveCharacterHandler)
	mux.HandleFunc("PUT /api/v1/characters/{id...}", handler.SaveCharacterHandler)
	mux.HandleFunc("DELETE /api/v1/characters/{id...}", handler.DeleteProfileHandler)

	// Meta / JSON Document APIs
	mux.HandleFunc("GET /api/v1/meta", handler.ListMetaHandler)
	mux.HandleFunc("GET /api/v1/meta/{key...}", handler.GetMetaHandler)
	mux.HandleFunc("POST /api/v1/meta/{key...}", handler.PutMetaHandler)
	mux.HandleFunc("PUT /api/v1/meta/{key...}", handler.PutMetaHandler)
	mux.HandleFunc("DELETE /api/v1/meta/{key...}", handler.DeleteMetaHandler)

	// Database Backup Snapshot
	mux.HandleFunc("GET /api/v1/backup", handler.BackupHandler)

	// Dropbox Cloud Endpoints
	mux.HandleFunc("GET /api/v1/dropbox/status", handler.DropboxStatusHandler)
	mux.HandleFunc("POST /api/v1/dropbox/backup", handler.DropboxBackupHandler)
	mux.HandleFunc("GET /api/v1/dropbox/files", handler.DropboxFilesHandler)
	mux.HandleFunc("GET /api/v1/dropbox/link", handler.DropboxLinkHandler)
	mux.HandleFunc("GET /api/v1/dropbox/search", handler.DropboxSearchHandler)
	mux.HandleFunc("GET /api/v1/dropbox/content", handler.DropboxContentHandler)
	mux.HandleFunc("GET /api/v1/dropbox/meta", handler.DropboxMetaHandler)

	// Dropbox Content Indexer & Hybrid Streaming
	mux.HandleFunc("POST /api/v1/index/scan", handler.IndexScanHandler)
	mux.HandleFunc("GET /api/v1/index/status", handler.IndexStatusHandler)
	mux.HandleFunc("GET /api/v1/index/search", handler.IndexSearchHandler)
	mux.HandleFunc("GET /api/v1/index/entry", handler.IndexEntryHandler)
	mux.HandleFunc("GET /api/v1/index/content", handler.IndexContentHandler)

	// Syncthing Media Ingestion Subsystem & Vault Streaming
	mux.HandleFunc("GET /api/v1/syncthing/status", handler.SyncthingStatusHandler)
	mux.HandleFunc("GET /api/v1/syncthing/feed", handler.SyncthingFeedHandler)
	mux.HandleFunc("POST /api/v1/syncthing/rescan", handler.SyncthingRescanHandler)
	mux.HandleFunc("GET /api/v1/media/stream", handler.MediaStreamHandler)
	mux.HandleFunc("GET /api/v1/media/thumbnail", handler.MediaThumbnailHandler)
	mux.HandleFunc("GET /api/v1/media/catalog", handler.UnifiedVisualCatalogHandler)

	// High-Fidelity Audio Input & D&D Sonic Chronicle
	mux.HandleFunc("GET /api/v1/sessions", handler.ListSessionsHandler)
	mux.HandleFunc("POST /api/v1/sessions", handler.CreateSessionHandler)
	mux.HandleFunc("POST /api/v1/sessions/upload", handler.UploadSessionHandler)
	mux.HandleFunc("GET /api/v1/sessions/{id...}", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/stream") {
			handler.SessionStreamHandler(w, r)
		} else if strings.HasSuffix(path, "/slice") {
			handler.PrecisionSliceHandler(w, r)
		} else {
			handler.GetSessionHandler(w, r)
		}
	})
	mux.HandleFunc("POST /api/v1/sessions/{id...}", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/chunk") {
			handler.SessionChunkHandler(w, r)
		} else if strings.HasSuffix(path, "/complete") {
			handler.CompleteSessionHandler(w, r)
		} else if strings.HasSuffix(path, "/markers") {
			handler.AddMarkerHandler(w, r)
		} else {
			handler.CreateSessionHandler(w, r)
		}
	})
	mux.HandleFunc("DELETE /api/v1/sessions/{id...}", handler.DeleteSessionHandler)

	// Google Recorder MCP Audio & Synchronized Transcripts
	mux.HandleFunc("GET /api/v1/recorder/status", handler.RecorderStatusHandler)
	mux.HandleFunc("POST /api/v1/recorder/sync", handler.RecorderSyncHandler)
	mux.HandleFunc("GET /api/v1/recorder/sync/status", handler.RecorderSyncStatusHandler)
	mux.HandleFunc("GET /api/v1/recorder/recordings", handler.RecorderRecordingsHandler)
	mux.HandleFunc("GET /api/v1/recorder/recordings/{id...}", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/audio") {
			handler.RecorderAudioHandler(w, r)
		} else {
			handler.RecorderTranscriptHandler(w, r)
		}
	})

	// Unified Audio Chronicle Portal
	mux.HandleFunc("GET /api/v1/audio/catalog", handler.UnifiedAudioCatalogHandler)
	mux.HandleFunc("POST /api/v1/audio/transcribe", handler.AudioTranscribePlaceholderHandler)

	// Scribe's Sanctuary & Sovereign Text Matrix
	mux.HandleFunc("GET /api/v1/text/catalog", handler.TextCatalogHandler)
	mux.HandleFunc("GET /api/v1/text/document", handler.TextDocumentHandler)
	mux.HandleFunc("GET /api/v1/text/sanctuaries", handler.TextSanctuariesHandler)
	mux.HandleFunc("GET /api/v1/text/timeline", handler.TextTimelineHandler)
	mux.HandleFunc("GET /api/v1/timeline/correspondence", handler.TimelineCorrespondenceHandler)

	// Axis Mundi Workspace Ingestion & Feed Hub
	mux.HandleFunc("GET /api/v1/axis-mundi/status", handler.AxisMundiStatusHandler)
	mux.HandleFunc("GET /api/v1/axis-mundi/feed", handler.AxisMundiFeedHandler)
	mux.HandleFunc("POST /api/v1/axis-mundi/sync", handler.AxisMundiSyncHandler)
	mux.HandleFunc("POST /api/v1/axis-mundi/events", handler.AxisMundiEventHandler)
	mux.HandleFunc("POST /api/v1/axis-mundi/discoveries", handler.AxisMundiDiscoveriesHandler)

	// Foundations Storytelling Synthesis & Video Manifest Bridge
	mux.HandleFunc("POST /api/v1/foundations/seeds/from-text", handler.FoundationsSeedFromTextHandler)
	mux.HandleFunc("GET /api/v1/foundations/seeds", handler.FoundationsListSeedsHandler)
	mux.HandleFunc("GET /api/v1/foundations/seeds/resonant", handler.FoundationsResonantSeedsHandler)
	mux.HandleFunc("GET /api/v1/foundations/seeds/{id}", handler.FoundationsGetSeedHandler)
	mux.HandleFunc("DELETE /api/v1/foundations/seeds/{id}", handler.FoundationsDeleteSeedHandler)
	mux.HandleFunc("POST /api/v1/foundations/seeds/{id}/manifest", handler.FoundationsSeedManifestHandler)
	mux.HandleFunc("POST /api/v1/foundations/manifest/from-seed", handler.FoundationsSeedManifestHandler)

	// Hermetic Slideshows & Audiovisual Chronicles
	mux.HandleFunc("GET /api/v1/slideshows", handler.ListSlideshowsHandler)
	mux.HandleFunc("POST /api/v1/slideshows", handler.SaveSlideshowHandler)
	mux.HandleFunc("GET /api/v1/slideshows/{id...}", handler.GetSlideshowHandler)
	mux.HandleFunc("POST /api/v1/slideshows/{id...}", handler.SlideshowManifestHandler)

	// YouTube Sovereign Video Uploader & Studio
	mux.HandleFunc("GET /api/v1/youtube/status", handler.YouTubeStatusHandler)
	mux.HandleFunc("GET /api/v1/youtube/auth/url", handler.YouTubeAuthURLHandler)
	mux.HandleFunc("GET /api/v1/youtube/auth/callback", handler.YouTubeAuthCallbackHandler)
	mux.HandleFunc("POST /api/v1/youtube/auth/disconnect", handler.YouTubeDisconnectHandler)
	mux.HandleFunc("POST /api/v1/youtube/upload", handler.YouTubeUploadHandler)
	mux.HandleFunc("GET /api/v1/youtube/jobs", handler.YouTubeJobsHandler)
	mux.HandleFunc("GET /api/v1/youtube/jobs/{id...}", handler.YouTubeJobDetailHandler)
	mux.HandleFunc("POST /api/v1/youtube/jobs/{id...}", handler.YouTubeJobDetailHandler)
	mux.HandleFunc("GET /api/v1/youtube/videos", handler.YouTubeVideosHandler)
	mux.HandleFunc("GET /api/v1/youtube/analytics", handler.YouTubeAnalyticsHandler)
	mux.HandleFunc("GET /api/v1/youtube/finance", handler.YouTubeFinanceHandler)
	mux.HandleFunc("POST /api/v1/youtube/finance/sync-amra", handler.YouTubeSyncAmraHandler)

	// AMRA Financial Core, Ledger & Sovereign Treasury Bridge
	mux.HandleFunc("GET /api/v1/amra/plans", handler.amraHandler.ListPlansHandler)
	mux.HandleFunc("POST /api/v1/amra/checkout", handler.amraHandler.CheckoutHandler)
	mux.HandleFunc("POST /api/v1/amra/webhook/{provider}", handler.amraHandler.WebhookHandler)
	mux.HandleFunc("GET /api/v1/amra/subscriptions/{id}", handler.amraHandler.GetSubscriptionHandler)
	mux.HandleFunc("GET /api/v1/amra/ledger", handler.amraHandler.LedgerHandler)
	mux.HandleFunc("GET /api/v1/amra/metrics", handler.amraHandler.MetricsHandler)
	mux.HandleFunc("GET /api/v1/amra/geometry", handler.amraHandler.GeometryHandler)
	mux.HandleFunc("POST /api/v1/amra/geometry", handler.amraHandler.GeometryHandler)
	mux.HandleFunc("GET /api/v1/amra/gcloud/status", handler.amraHandler.GCloudStatusHandler)
	mux.HandleFunc("GET /api/v1/amra/gcloud/billing", handler.amraHandler.GCloudBillingHandler)
	mux.HandleFunc("POST /api/v1/amra/gcloud/sync-ledger", handler.amraHandler.GCloudSyncLedgerHandler)
	mux.HandleFunc("POST /api/v1/treasury/events", handler.TreasuryEventHandler)
	mux.HandleFunc("GET /api/v1/treasury/status", handler.TreasuryStatusHandler)

	// Esoteric Foundations & Philosophical Texts (BoltDB Served)
	mux.HandleFunc("GET /api/v1/esoteric", handler.GetEsotericCatalogHandler)
	mux.HandleFunc("GET /api/v1/esoteric/arishadvarga", handler.GetArishadvargaHandler)
	mux.HandleFunc("GET /api/v1/esoteric/{key...}", handler.GetEsotericDocHandler)

	// System & Agent Mission Control
	mux.HandleFunc("GET /api/v1/system/overview", handler.SystemOverviewHandler)
	mux.HandleFunc("GET /api/v1/system/routes", handler.APICatalogHandler)
	mux.HandleFunc("GET /api/v1/agent/conversations", handler.AgentConversationsHandler)
	mux.HandleFunc("GET /api/v1/agent/artifact", handler.AgentArtifactHandler)
	mux.HandleFunc("GET /api/v1/agent/bicameral", handler.AgentBicameralHandler)

	// Autonomous Self-Healing & Remote Service Restart
	mux.HandleFunc("GET /api/v1/system/health", handler.SystemHealthHandler)
	mux.HandleFunc("POST /api/v1/system/self-healing/check", handler.SelfHealingCheckHandler)
	mux.HandleFunc("POST /api/v1/system/self-healing/remediate", handler.SelfHealingRemediateHandler)
	mux.HandleFunc("GET /api/v1/system/self-healing/incidents", handler.SelfHealingIncidentsHandler)
	mux.HandleFunc("POST /api/v1/system/restart", handler.RestartHandler)
	mux.HandleFunc("GET /api/v1/system/restart/history", handler.RestartHistoryHandler)

	return &Router{
		mux:     mux,
		handler: handler,
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handler := r.recoveryMiddleware(r.loggingMiddleware(r.corsMiddleware(r.mux)))
	handler.ServeHTTP(w, req)
}

func (r *Router) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-Admin-Token")

		if req.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, req)
	})
}

func (r *Router) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, req)
		log.Printf("%s %s in %v", req.Method, req.URL.Path, time.Since(start))
	})
}

func (r *Router) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("Panic recovered: %v", rec)
				if r.handler != nil && r.handler.SelfHealing() != nil {
					r.handler.SelfHealing().RecordPanic(rec, debug.Stack(), req.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":"internal server error: panic recovered and logged to self-healing incident journal"}`))
			}
		}()
		next.ServeHTTP(w, req)
	})
}

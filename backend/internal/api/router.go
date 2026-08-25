package api

import (
	"log"
	"net/http"
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
	mux.HandleFunc("GET /api/dasha/alchemy", handler.DashaAlchemyHandler)

	// Meta / JSON Document APIs
	mux.HandleFunc("GET /api/v1/meta", handler.ListMetaHandler)
	mux.HandleFunc("GET /api/v1/meta/{key...}", handler.GetMetaHandler)
	mux.HandleFunc("POST /api/v1/meta/{key...}", handler.PutMetaHandler)
	mux.HandleFunc("PUT /api/v1/meta/{key...}", handler.PutMetaHandler)
	mux.HandleFunc("DELETE /api/v1/meta/{key...}", handler.DeleteMetaHandler)

	// Database Backup Snapshot
	mux.HandleFunc("GET /api/v1/backup", handler.BackupHandler)

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
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

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
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, req)
	})
}

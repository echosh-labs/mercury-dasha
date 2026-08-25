package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/echosh-labs/mercury-dasha/internal/api"
	"github.com/echosh-labs/mercury-dasha/internal/config"
	"github.com/echosh-labs/mercury-dasha/internal/db"
)

//go:embed all:frontend_out/*
var frontendFS embed.FS

func main() {
	cfg := config.Load()
	log.Printf("Starting [%s] in %s mode...", cfg.ServiceName, cfg.Environment)

	// 1. Initialize BoltDB
	store, err := db.Open(cfg.BoltDBPath)
	if err != nil {
		log.Fatalf("Fatal: could not open BoltDB at %s: %v", cfg.BoltDBPath, err)
	}
	defer func() {
		log.Println("Closing BoltDB connection...")
		if err := store.Close(); err != nil {
			log.Printf("Error closing BoltDB: %v", err)
		}
	}()

	// 2. Initialize API Router
	handler := api.NewHandler(cfg, store)
	apiRouter := api.NewRouter(handler)

	// 3. Prepare Embedded Frontend FS
	distFS, err := fs.Sub(frontendFS, "frontend_out")
	if err != nil {
		log.Fatalf("Fatal: could not extract embedded frontend FS: %v", err)
	}
	fileServer := http.FileServer(http.FS(distFS))

	// 4. Root Multiplexer
	rootMux := http.NewServeMux()

	// Mount API & Health
	rootMux.Handle("/api/", apiRouter)
	rootMux.Handle("/healthz", apiRouter)

	// Mount Frontend with SPA fallback
	rootMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		// Check if file exists in embedded FS
		f, err := distFS.Open(path)
		if err != nil {
			// SPA / Next.js client-side routing fallback: serve index.html directly
			indexData, err := fs.ReadFile(distFS, "index.html")
			if err != nil {
				http.Error(w, "index.html not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(indexData)
			return
		}
		_ = f.Close()

		fileServer.ServeHTTP(w, r)
	})

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      rootMux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 5. Graceful Shutdown Listener
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("🚀 [%s] serving on http://0.0.0.0:%s (DB: %s)", cfg.ServiceName, cfg.Port, cfg.BoltDBPath)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-stopChan
	log.Printf("Received termination signal from Cloud Run. Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced shutdown: %v", err)
	}

	fmt.Println("Mercury-Dasha engine exited cleanly.")
}

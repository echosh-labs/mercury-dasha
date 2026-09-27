package main

import (
	"context"
	"embed"
	"encoding/json"
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
	"github.com/echosh-labs/mercury-dasha/internal/dasha"
	"github.com/echosh-labs/mercury-dasha/internal/db"
	"github.com/echosh-labs/mercury-dasha/internal/mcp"
	"github.com/echosh-labs/mercury-dasha/internal/portutil"
	"github.com/echosh-labs/mercury-dasha/internal/timeline"
)

//go:embed all:frontend_out/*
var frontendFS embed.FS

func main() {
	cfg := config.Load()
	log.Printf("Starting [%s] in %s mode...", cfg.ServiceName, cfg.Environment)

	// 0. Ensure port availability / remedy contentious port at startup
	if err := portutil.ClearPortIfContentious(cfg.Port); err != nil {
		log.Printf("⚠️ Warning: port contentious remediation: %v", err)
	}

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

	// 1b. Seed Canonical Esoteric Foundations & Sovereign Storehouse
	if err := db.SeedEsotericContent(store); err != nil {
		log.Printf("Warning: failed to seed esoteric content: %v", err)
	}
	if err := dasha.SeedSovereignStorehouse(store); err != nil {
		log.Printf("Warning: failed to seed sovereign storehouse: %v", err)
	}
	if err := dasha.MigrateProfiles(store); err != nil {
		log.Printf("Warning: failed to migrate profiles: %v", err)
	}
	if err := timeline.SeedTempleOfIlluminationSlideshow(store, cfg.DropboxLocalPath); err != nil {
		log.Printf("Warning: failed to seed temple of illumination slideshow: %v", err)
	}

	// 2. Initialize API Router
	handler := api.NewHandler(cfg, store)
	apiRouter := api.NewRouter(handler)

	// 2b. Initialize Native Model Context Protocol (MCP) Server
	mcpHandler := mcp.NewHandler(handler, store, cfg)
	mcpServer := mcp.NewServer(mcpHandler, "")

	// Check if invoked in Stdio MCP mode
	for _, arg := range os.Args[1:] {
		if arg == "mcp" || arg == "--mcp-stdio" || arg == "--stdio" {
			log.Println("⚡ Starting Mercury Dasha in Native Stdio MCP mode...")
			if err := mcpServer.ServeStdio(os.Stdin, os.Stdout); err != nil {
				log.Fatalf("MCP Stdio runtime error: %v", err)
			}
			return
		}
	}

	// 2c. Start Autonomous Self-Healing Watchdog
	if sh := handler.SelfHealing(); sh != nil {
		sh.Start()
		defer sh.Stop()
	}

	// 2d. Start Syncthing Event-Driven Media Vault Subsystem
	if st := handler.Syncthing(); st != nil {
		stCtx, stCancel := context.WithCancel(context.Background())
		defer stCancel()
		st.Start(stCtx)
		defer st.Stop()
	}

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

	// Mount Native Streamable HTTP MCP Server
	mcpServer.RegisterRoutes(rootMux)

	// Mount Frontend with SPA and directory routing fallback
	rootMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Guard API and MCP namespaces: Unmatched API routes must return clean JSON 404
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/mcp") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error":  "API route not found",
				"path":   r.URL.Path,
				"method": r.Method,
				"status": http.StatusNotFound,
			})
			return
		}

		cleanPath := strings.TrimPrefix(r.URL.Path, "/")
		if cleanPath == "" {
			cleanPath = "index.html"
		}

		// Try direct path, then directory index.html, then .html
		targetFile := cleanPath
		f, err := distFS.Open(targetFile)
		if err != nil {
			if fDir, errDir := distFS.Open(strings.TrimSuffix(cleanPath, "/") + "/index.html"); errDir == nil {
				_ = fDir.Close()
				targetFile = strings.TrimSuffix(cleanPath, "/") + "/index.html"
				f = nil
				err = nil
			} else if fHtml, errHtml := distFS.Open(strings.TrimSuffix(cleanPath, "/") + ".html"); errHtml == nil {
				_ = fHtml.Close()
				targetFile = strings.TrimSuffix(cleanPath, "/") + ".html"
				f = nil
				err = nil
			}
		}

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
		if f != nil {
			_ = f.Close()
		}

		if targetFile != cleanPath {
			htmlData, err := fs.ReadFile(distFS, targetFile)
			if err == nil {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(htmlData)
				return
			}
		}

		fileServer.ServeHTTP(w, r)
	})

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      rootMux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 5. Graceful Shutdown & Remote Restart Listener
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("🚀 [%s] serving on http://0.0.0.0:%s (DB: %s)", cfg.ServiceName, cfg.Port, cfg.BoltDBPath)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	restartMgr := handler.RestartManager()

	select {
	case sig := <-stopChan:
		log.Printf("Received termination signal (%v). Shutting down gracefully...", sig)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server forced shutdown: %v", err)
		}

		handler.Close()
		fmt.Println("Mercury-Dasha engine exited cleanly.")

	case req := <-restartMgr.RestartChan():
		log.Printf("🔄 [Self-Healing Engine] Service restart initiated (reason: %q, initiator: %s, delay: %v)", req.Reason, req.Initiator, req.Delay)
		if req.Delay > 0 {
			time.Sleep(req.Delay)
		}

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown warning during restart: %v", err)
		}
		cancel()

		// Gracefully terminate handlers, tickers, and close BoltDB before re-exec
		handler.Close()
		_ = store.Close()

		// Perform sovereign binary re-execution (syscall.Exec on POSIX / exec.Command fallback)
		restartMgr.PerformReexec(req)
	}
}

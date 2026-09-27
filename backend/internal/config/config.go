package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	Port                 string
	BoltDBPath           string
	ServiceName          string
	Environment          string
	DropboxAccessToken   string
	DropboxRefreshToken  string
	DropboxAppKey        string
	DropboxAppSecret     string
	DropboxBackupPath    string
	DropboxLocalPath     string
	YouTubeClientID      string
	YouTubeClientSecret  string
	YouTubeRedirectURL   string
	RecorderMCPURL       string
	RecorderOutputDir    string
	RecorderAutoSyncMins int
	RestartToken         string
	WatchdogIntervalSec  int
	MemoryLimitMB        float64
	SyncthingURL         string
	SyncthingAPIKey      string
	MediaLocalPath       string
	AxisMundiURL         string
	AxisMundiAPIKey      string
}

func loadEnvFile(paths ...string) {
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				k := strings.TrimSpace(parts[0])
				v := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
				if os.Getenv(k) == "" {
					_ = os.Setenv(k, v)
				}
			}
		}
		break
	}
}

func Load() *Config {
	loadEnvFile(".env", "../.env", "../../.env")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("BOLT_DB_PATH")
	if dbPath == "" {
		dbPath = "/var/data/mercury-dasha.db"
	}

	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "mercury-dasha"
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "production"
	}

	dropboxBackupPath := os.Getenv("DROPBOX_BACKUP_PATH")
	if dropboxBackupPath == "" {
		dropboxBackupPath = "/MercuryDasha/backups"
	}

	youtubeRedirectURL := os.Getenv("YOUTUBE_REDIRECT_URL")
	if youtubeRedirectURL == "" {
		youtubeRedirectURL = "http://localhost:" + port + "/api/v1/youtube/auth/callback"
	}

	dropboxLocalPath := os.Getenv("DROPBOX_LOCAL_PATH")
	if dropboxLocalPath == "" {
		dropboxLocalPath = "/home/justin/Dropbox"
	}

	recorderMCPURL := os.Getenv("RECORDER_MCP_URL")
	if recorderMCPURL == "" {
		recorderMCPURL = "http://localhost:8091/mcp"
	}

	recorderOutputDir := os.Getenv("RECORDER_OUTPUT_DIR")
	if recorderOutputDir == "" {
		recorderOutputDir = filepath.Join(dropboxLocalPath, "audio", "recorder")
	}

	autoSyncMins := 30
	if v := os.Getenv("RECORDER_AUTO_SYNC_MINUTES"); v != "" {
		if iv, err := strconv.Atoi(v); err == nil {
			autoSyncMins = iv
		}
	}

	watchdogSec := 15
	if v := os.Getenv("WATCHDOG_INTERVAL_SEC"); v != "" {
		if iv, err := strconv.Atoi(v); err == nil && iv > 0 {
			watchdogSec = iv
		}
	}

	memoryLimitMB := 1024.0
	if v := os.Getenv("MEMORY_LIMIT_MB"); v != "" {
		if fv, err := strconv.ParseFloat(v, 64); err == nil && fv > 0 {
			memoryLimitMB = fv
		}
	}

	syncthingURL := os.Getenv("SYNCTHING_URL")
	if syncthingURL == "" {
		syncthingURL = "http://127.0.0.1:8384"
	}

	syncthingAPIKey := os.Getenv("SYNCTHING_API_KEY")
	if syncthingAPIKey == "" {
		syncthingAPIKey = "okRtYsugpNESuDPeGtaWGE7rp7azkcjK"
	}

	mediaLocalPath := os.Getenv("MEDIA_LOCAL_PATH")
	if mediaLocalPath == "" {
		mediaLocalPath = "/home/justin/media"
	}

	axisMundiURL := os.Getenv("AXIS_MUNDI_URL")
	if axisMundiURL == "" {
		axisMundiURL = "http://127.0.0.1:8088"
	}
	axisMundiAPIKey := os.Getenv("AXIS_MUNDI_API_KEY")
	if axisMundiAPIKey == "" {
		axisMundiAPIKey = os.Getenv("MCP_API_KEY")
	}

	return &Config{
		Port:                 port,
		BoltDBPath:           dbPath,
		ServiceName:          serviceName,
		Environment:          env,
		DropboxAccessToken:   os.Getenv("DROPBOX_ACCESS_TOKEN"),
		DropboxRefreshToken:  os.Getenv("DROPBOX_REFRESH_TOKEN"),
		DropboxAppKey:        os.Getenv("DROPBOX_APP_KEY"),
		DropboxAppSecret:     os.Getenv("DROPBOX_APP_SECRET"),
		DropboxBackupPath:    dropboxBackupPath,
		DropboxLocalPath:     dropboxLocalPath,
		YouTubeClientID:      os.Getenv("YOUTUBE_CLIENT_ID"),
		YouTubeClientSecret:  os.Getenv("YOUTUBE_CLIENT_SECRET"),
		YouTubeRedirectURL:   youtubeRedirectURL,
		RecorderMCPURL:       recorderMCPURL,
		RecorderOutputDir:    recorderOutputDir,
		RecorderAutoSyncMins: autoSyncMins,
		RestartToken:         os.Getenv("MERCURY_RESTART_TOKEN"),
		WatchdogIntervalSec:  watchdogSec,
		MemoryLimitMB:        memoryLimitMB,
		SyncthingURL:         syncthingURL,
		SyncthingAPIKey:      syncthingAPIKey,
		MediaLocalPath:       mediaLocalPath,
		AxisMundiURL:         axisMundiURL,
		AxisMundiAPIKey:      axisMundiAPIKey,
	}
}

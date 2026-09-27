// System Runtime, Agentic Console, and Storehouse Type Definitions

export interface TelemetryData {
  service: string;
  uptime: string;
  alloc_mb: number;
  sys_mb: number;
  goroutines: number;
  timestamp: string;
  num_gc?: number;
}

export interface SystemOverview {
  service: string;
  version: string;
  go_version: string;
  environment: string;
  timestamp: string;
  uptime: string;
  uptime_sec: number;
  runtime_stats: {
    goroutines: number;
    alloc_mb: number;
    total_alloc_mb: number;
    sys_mb: number;
    num_gc: number;
  };
  database: {
    path: string;
    size_bytes: number;
    key_count: number;
    open_time: string;
    allocated_pages: number;
    category_counts: Record<string, number>;
  };
  dropbox: {
    local_path: string;
    local_exists: boolean;
    total_indexed: number;
    cloud_configured: boolean;
    account_email?: string;
  };
}

export interface IndexEntry {
  id: string;
  category: "text" | "audio" | "video" | "code" | "books";
  path: string;
  full_path: string;
  file_name: string;
  extension: string;
  size_bytes: number;
  mod_time: string;
  indexed_at: string;
  tags?: string[];
  snippet?: string;
  metadata?: Record<string, any>;
}

export interface CrawlStatus {
  is_scanning: boolean;
  current_category?: string;
  started_at?: string;
  finished_at?: string;
  total_seen: number;
  total_indexed: number;
  errors_count: number;
}

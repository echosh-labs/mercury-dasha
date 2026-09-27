"use client";

import React, { useState, useEffect, useCallback, useRef } from "react";
import {
  Search,
  BookOpen,
  FileText,
  Music,
  Video,
  Code2,
  RefreshCw,
  FolderTree,
  ExternalLink,
  Copy,
  Check,
  Play,
  FileCheck2,
  Layers,
  Sparkles,
  AlertCircle,
  Cloud,
  UploadCloud,
  HardDrive,
  Download,
  ShieldCheck,
  X,
  Eye,
  ChevronDown,
  ChevronUp,
  FileArchive,
  Volume2,
  VolumeX,
  Volume1,
  Radio
} from "lucide-react";
import { useYouTubeStudio } from "@/lib/youtube-context";

interface IndexEntry {
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

interface CrawlStatus {
  is_scanning: boolean;
  current_category?: string;
  started_at?: string;
  finished_at?: string;
  total_seen: number;
  total_indexed: number;
  errors_count: number;
  category_counts: Record<string, number>;
}

interface DropboxAccount {
  account_id: string;
  display_name: string;
  email: string;
  account_type: string;
  country: string;
  used_bytes?: number;
  allocated_bytes?: number;
}

interface DropboxStatus {
  configured: boolean;
  base_path: string;
  account?: DropboxAccount | null;
  error?: string;
  message?: string;
}

interface DropboxFile {
  name: string;
  path_display: string;
  size?: number;
  server_modified?: string;
}

export default function DropboxSovereignStorehouse() {
  const { openUploader } = useYouTubeStudio();

  // Index & Search State
  const [query, setQuery] = useState("");
  const [selectedCategory, setSelectedCategory] = useState<string>("all");
  const [selectedExt, setSelectedExt] = useState<string>("");
  const [results, setResults] = useState<IndexEntry[]>([]);
  const [totalMatches, setTotalMatches] = useState(0);
  const [loading, setLoading] = useState(false);
  const [status, setStatus] = useState<CrawlStatus | null>(null);
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [activePreview, setActivePreview] = useState<IndexEntry | null>(null);
  const [previewText, setPreviewText] = useState<string | null>(null);
  const [previewLoading, setPreviewLoading] = useState(false);

  // Persistent Audio Volume State across sessions (hydrated in useEffect to prevent SSR mismatch)
  const [audioVolume, setAudioVolume] = useState<number>(0.8);
  const [audioMuted, setAudioMuted] = useState<boolean>(false);
  const audioRef = useRef<HTMLAudioElement | null>(null);

  useEffect(() => {
    try {
      const savedVol = localStorage.getItem("mercury_storehouse_audio_volume");
      if (savedVol !== null) {
        const parsed = parseFloat(savedVol);
        if (!isNaN(parsed) && parsed >= 0 && parsed <= 1) {
          setAudioVolume(parsed);
          if (audioRef.current) audioRef.current.volume = parsed;
        }
      }
      const savedMuted = localStorage.getItem("mercury_storehouse_audio_muted");
      if (savedMuted !== null) {
        const isMuted = savedMuted === "true";
        setAudioMuted(isMuted);
        if (audioRef.current) audioRef.current.muted = isMuted;
      }
    } catch {
      // Ignore storage access errors
    }
  }, []);

  const updateAudioVolume = useCallback((newVol: number) => {
    const clamped = Math.max(0, Math.min(1, newVol));
    setAudioVolume(clamped);
    if (clamped > 0 && audioMuted) {
      setAudioMuted(false);
      if (typeof window !== "undefined") {
        localStorage.setItem("mercury_storehouse_audio_muted", "false");
      }
    }
    if (audioRef.current) {
      audioRef.current.volume = clamped;
      if (clamped > 0 && audioRef.current.muted) {
        audioRef.current.muted = false;
      }
    }
    if (typeof window !== "undefined") {
      localStorage.setItem("mercury_storehouse_audio_volume", clamped.toString());
    }
  }, [audioMuted]);

  const toggleAudioMute = useCallback(() => {
    setAudioMuted((prev) => {
      const next = !prev;
      if (audioRef.current) {
        audioRef.current.muted = next;
      }
      if (typeof window !== "undefined") {
        localStorage.setItem("mercury_storehouse_audio_muted", String(next));
      }
      return next;
    });
  }, []);

  const handleAudioVolumeChange = useCallback((e: React.SyntheticEvent<HTMLAudioElement>) => {
    const el = e.currentTarget;
    setAudioVolume(el.volume);
    setAudioMuted(el.muted);
    if (typeof window !== "undefined") {
      localStorage.setItem("mercury_storehouse_audio_volume", el.volume.toString());
      localStorage.setItem("mercury_storehouse_audio_muted", String(el.muted));
    }
  }, []);

  const syncAudioElement = useCallback((el: HTMLAudioElement | null) => {
    if (el) {
      el.volume = audioVolume;
      el.muted = audioMuted;
    }
  }, [audioVolume, audioMuted]);

  // Cloud API & Backup State
  const [cloudStatus, setCloudStatus] = useState<DropboxStatus | null>(null);
  const [showCloudDrawer, setShowCloudDrawer] = useState(false);
  const [cloudFiles, setCloudFiles] = useState<DropboxFile[]>([]);
  const [backingUp, setBackingUp] = useState(false);
  const [feedback, setFeedback] = useState<{ type: "success" | "error"; text: string } | null>(null);

  // Fetch Index Status
  const fetchStatus = useCallback(async () => {
    try {
      const res = await fetch("/api/v1/index/status");
      if (res.ok) {
        const data = await res.json();
        setStatus(data);
      }
    } catch (err) {
      console.error("Failed to fetch index status:", err);
    }
  }, []);

  // Fetch Cloud Account Status
  const fetchCloudStatus = useCallback(async () => {
    try {
      const res = await fetch("/api/v1/dropbox/status");
      if (res.ok) {
        const data = await res.json();
        setCloudStatus(data);
        if (data.configured && !data.error) {
          fetchCloudFiles();
        }
      }
    } catch (err) {
      console.error("Failed to fetch Dropbox cloud status:", err);
    }
  }, []);

  const fetchCloudFiles = async () => {
    try {
      const res = await fetch("/api/v1/dropbox/files");
      if (res.ok) {
        const data = await res.json();
        setCloudFiles(data.entries || []);
      }
    } catch (err) {
      console.error("Failed to fetch cloud backup files:", err);
    }
  };

  // Search Files across local index
  const searchFiles = useCallback(async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams();
      if (query.trim()) params.set("q", query.trim());
      if (selectedCategory && selectedCategory !== "all") params.set("category", selectedCategory);
      if (selectedExt) params.set("ext", selectedExt);
      params.set("limit", "50");

      const res = await fetch(`/api/v1/index/search?${params.toString()}`);
      if (res.ok) {
        const data = await res.json();
        setResults(data.results || []);
        setTotalMatches(data.total || 0);
      }
    } catch (err) {
      console.error("Search failed:", err);
    } finally {
      setLoading(false);
    }
  }, [query, selectedCategory, selectedExt]);

  useEffect(() => {
    fetchStatus();
    fetchCloudStatus();
    const interval = setInterval(fetchStatus, 5000);
    return () => clearInterval(interval);
  }, [fetchStatus, fetchCloudStatus]);

  useEffect(() => {
    const delayDebounce = setTimeout(() => {
      searchFiles();
    }, 250);
    return () => clearTimeout(delayDebounce);
  }, [searchFiles]);

  // Handle Full Content Preview
  const handleOpenPreview = async (entry: IndexEntry) => {
    setActivePreview(entry);
    setPreviewText(null);

    const isText = [".txt", ".md", ".json", ".yaml", ".yml", ".go", ".ts", ".js", ".py", ".sh", ".sql", ".html", ".css"].includes(entry.extension.toLowerCase());
    if (isText) {
      setPreviewLoading(true);
      try {
        const res = await fetch(`/api/v1/index/content?path=${encodeURIComponent(entry.path)}`);
        if (res.ok) {
          const text = await res.text();
          setPreviewText(text);
        } else {
          setPreviewText(entry.snippet || "Unable to read full content.");
        }
      } catch (err) {
        setPreviewText(entry.snippet || "Error fetching file text.");
      } finally {
        setPreviewLoading(false);
      }
    }
  };

  // Trigger Local Crawl & Index
  const triggerScan = async (category?: string) => {
    try {
      const url = category && category !== "all" 
        ? `/api/v1/index/scan?category=${category}` 
        : "/api/v1/index/scan";
      await fetch(url, { method: "POST" });
      await fetchStatus();
    } catch (err) {
      console.error("Trigger scan failed:", err);
    }
  };

  // Trigger Cloud Backup Snapshot
  const triggerCloudBackup = async () => {
    setBackingUp(true);
    setFeedback(null);
    try {
      const res = await fetch("/api/v1/dropbox/backup", { method: "POST" });
      const data = await res.json();
      if (res.ok) {
        setFeedback({
          type: "success",
          text: `Snapshot archived to Dropbox Cloud: ${data.entry?.name || "backup.db"} (${formatBytes(data.size_bytes || 0)})`
        });
        fetchCloudFiles();
      } else {
        setFeedback({ type: "error", text: data.error || "Cloud backup failed" });
      }
    } catch (err: any) {
      setFeedback({ type: "error", text: err.message || "Network error during backup" });
    } finally {
      setBackingUp(false);
    }
  };

  // Generate Temporary Cloud Stream Link
  const generateCloudLink = async (path: string, id: string) => {
    try {
      const remotePath = path.startsWith("/") ? path : `/${path}`;
      const res = await fetch(`/api/v1/dropbox/link?path=${encodeURIComponent(remotePath)}`);
      if (res.ok) {
        const data = await res.json();
        if (data.link) {
          navigator.clipboard.writeText(data.link);
          setCopiedId(id);
          setTimeout(() => setCopiedId(null), 2000);
        }
      }
    } catch (err) {
      console.error("Failed to generate link:", err);
    }
  };

  const copyPath = (path: string, id: string) => {
    navigator.clipboard.writeText(path);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
  };

  const getCategoryBadge = (cat: string) => {
    switch (cat) {
      case "books":
        return <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium bg-purple-500/20 text-purple-300 border border-purple-500/30"><BookOpen className="w-3 h-3" /> Books</span>;
      case "audio":
        return <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium bg-amber-500/20 text-amber-300 border border-amber-500/30"><Music className="w-3 h-3" /> Audio</span>;
      case "video":
        return <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium bg-rose-500/20 text-rose-300 border border-rose-500/30"><Video className="w-3 h-3" /> Video</span>;
      case "code":
        return <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium bg-emerald-500/20 text-emerald-300 border border-emerald-500/30"><Code2 className="w-3 h-3" /> Code</span>;
      default:
        return <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium bg-cyan-500/20 text-cyan-300 border border-cyan-500/30"><FileText className="w-3 h-3" /> Text</span>;
    }
  };

  const counts = status?.category_counts || {};
  const totalCount = counts.total || 0;

  return (
    <div className="space-y-6">
      {/* Top Banner & Control Bar */}
      <div className="p-6 rounded-2xl bg-gradient-to-r from-zinc-900 via-[#111928] to-zinc-900 border border-zinc-800 shadow-xl backdrop-blur-md">
        <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-6">
          <div>
            <div className="flex items-center gap-3">
              <div className="p-2.5 rounded-xl bg-amber-500/10 text-amber-400 border border-amber-500/20">
                <FolderTree className="w-6 h-6" />
              </div>
              <div>
                <div className="flex items-center gap-2">
                  <h2 className="text-xl font-bold text-white tracking-tight">
                    Dropbox Sovereign Storehouse
                  </h2>
                  <span className="text-xs px-2.5 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 font-mono flex items-center gap-1">
                    <ShieldCheck className="w-3 h-3" />
                    <span>SOVEREIGN POSIX</span>
                  </span>
                </div>
                <p className="text-xs text-zinc-400 mt-1">
                  Direct local access to <code className="text-amber-300 font-mono">/home/justin/Dropbox</code> with BoltDB indexed metadata & automated OAuth cloud sync.
                </p>
              </div>
            </div>
          </div>

          {/* Quick Stats & Cloud Toggle */}
          <div className="flex flex-wrap items-center gap-3">
            {cloudStatus?.configured && cloudStatus.account && (
              <button
                onClick={() => setShowCloudDrawer(!showCloudDrawer)}
                className="flex items-center gap-2 px-3.5 py-2 rounded-xl text-xs font-mono bg-cyan-950/70 text-cyan-300 border border-cyan-800/80 hover:bg-cyan-900/60 transition"
              >
                <Cloud className="w-3.5 h-3.5 text-cyan-400" />
                <span>{cloudStatus.account.display_name} ({cloudStatus.account.account_type.toUpperCase()})</span>
                {showCloudDrawer ? <ChevronUp className="w-3.5 h-3.5" /> : <ChevronDown className="w-3.5 h-3.5" />}
              </button>
            )}

            <button
              onClick={() => triggerScan("all")}
              disabled={status?.is_scanning}
              className={`flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-semibold font-mono transition-all ${
                status?.is_scanning
                  ? "bg-amber-500/20 text-amber-300 border border-amber-500/30 cursor-not-allowed animate-pulse"
                  : "bg-amber-500 text-zinc-950 hover:bg-amber-400 shadow-lg shadow-amber-500/20"
              }`}
            >
              <RefreshCw className={`w-3.5 h-3.5 ${status?.is_scanning ? "animate-spin" : ""}`} />
              {status?.is_scanning ? `Indexing (${status.current_category || "..."})...` : "Crawl & Index All"}
            </button>
          </div>
        </div>

        {/* Expandable Cloud Drawer */}
        {showCloudDrawer && cloudStatus?.configured && cloudStatus.account && (
          <div className="mt-6 pt-5 border-t border-zinc-800/80 grid grid-cols-1 md:grid-cols-3 gap-4 text-xs font-mono">
            <div className="p-3.5 rounded-xl bg-zinc-900/90 border border-zinc-800">
              <div className="text-zinc-500 uppercase text-[10px]">Cloud Account & Quota</div>
              <div className="text-zinc-200 font-medium mt-1">{cloudStatus.account.email}</div>
              <div className="text-zinc-400 text-[11px] mt-1">
                {formatBytes(cloudStatus.account.used_bytes || 0)} used of {formatBytes(cloudStatus.account.allocated_bytes || 0)}
              </div>
              <div className="w-full bg-zinc-800 h-1.5 rounded-full mt-2 overflow-hidden">
                <div 
                  className="bg-cyan-500 h-full rounded-full transition-all"
                  style={{ width: `${Math.min(100, ((cloudStatus.account.used_bytes || 0) / (cloudStatus.account.allocated_bytes || 1)) * 100)}%` }}
                />
              </div>
            </div>

            <div className="p-3.5 rounded-xl bg-zinc-900/90 border border-zinc-800 flex flex-col justify-between">
              <div>
                <div className="text-zinc-500 uppercase text-[10px]">Remote Backup Target</div>
                <div className="text-cyan-300 font-medium mt-1">{cloudStatus.base_path}</div>
                <div className="text-zinc-500 text-[11px] mt-0.5">Non-expiring OAuth2 refresh token active</div>
              </div>
              <button
                onClick={triggerCloudBackup}
                disabled={backingUp}
                className="mt-2 flex items-center justify-center gap-1.5 px-3 py-1.5 rounded-lg bg-cyan-900/40 text-cyan-300 border border-cyan-700 hover:bg-cyan-800/50 transition font-medium"
              >
                <UploadCloud className={`w-3.5 h-3.5 ${backingUp ? "animate-bounce" : ""}`} />
                <span>{backingUp ? "Archiving to Cloud..." : "Upload BoltDB Snapshot"}</span>
              </button>
            </div>

            <div className="p-3.5 rounded-xl bg-zinc-900/90 border border-zinc-800">
              <div className="text-zinc-500 uppercase text-[10px]">Recent Cloud Backups ({cloudFiles.length})</div>
              <div className="mt-1.5 space-y-1 max-h-20 overflow-y-auto text-[11px]">
                {cloudFiles.slice(0, 3).map((f) => (
                  <div key={f.path_display} className="flex items-center justify-between text-zinc-300">
                    <span className="truncate max-w-[140px]">{f.name}</span>
                    <span className="text-zinc-500">{formatBytes(f.size || 0)}</span>
                  </div>
                ))}
                {cloudFiles.length === 0 && <span className="text-zinc-500">No remote snapshots yet.</span>}
              </div>
            </div>
          </div>
        )}

        {/* Feedback Alert */}
        {feedback && (
          <div className={`mt-4 p-3 rounded-xl border text-xs font-mono flex items-center justify-between ${
            feedback.type === "success" 
              ? "bg-emerald-950/60 border-emerald-800 text-emerald-300"
              : "bg-rose-950/60 border-rose-800 text-rose-300"
          }`}>
            <span>{feedback.text}</span>
            <button onClick={() => setFeedback(null)}><X className="w-3.5 h-3.5" /></button>
          </div>
        )}
      </div>

      {/* Category Pills Bar */}
      <div className="flex flex-wrap items-center gap-2">
        <button
          onClick={() => { setSelectedCategory("all"); setSelectedExt(""); }}
          className={`flex items-center gap-2 px-3.5 py-1.5 rounded-xl text-xs font-mono font-medium transition-all ${
            selectedCategory === "all"
              ? "bg-amber-500 text-zinc-950 font-bold shadow-md shadow-amber-500/20"
              : "bg-zinc-900 hover:bg-zinc-800 text-zinc-300 border border-zinc-800"
          }`}
        >
          <FolderTree className="w-3.5 h-3.5" />
          <span>All Documents</span>
          <span className="text-[10px] px-1.5 py-0.2 rounded-full bg-black/30">
            {totalCount.toLocaleString()}
          </span>
        </button>

        <button
          onClick={() => { setSelectedCategory("books"); setSelectedExt(""); }}
          className={`flex items-center gap-2 px-3.5 py-1.5 rounded-xl text-xs font-mono font-medium transition-all ${
            selectedCategory === "books"
              ? "bg-purple-600 text-white font-bold shadow-md shadow-purple-600/20"
              : "bg-zinc-900 hover:bg-zinc-800 text-zinc-300 border border-zinc-800"
          }`}
        >
          <BookOpen className="w-3.5 h-3.5 text-purple-400" />
          <span>Books</span>
          <span className="text-[10px] px-1.5 py-0.2 rounded-full bg-black/30">
            {(counts.books || 0).toLocaleString()}
          </span>
        </button>

        <button
          onClick={() => { setSelectedCategory("code"); setSelectedExt(""); }}
          className={`flex items-center gap-2 px-3.5 py-1.5 rounded-xl text-xs font-mono font-medium transition-all ${
            selectedCategory === "code"
              ? "bg-emerald-600 text-white font-bold shadow-md shadow-emerald-600/20"
              : "bg-zinc-900 hover:bg-zinc-800 text-zinc-300 border border-zinc-800"
          }`}
        >
          <Code2 className="w-3.5 h-3.5 text-emerald-400" />
          <span>Code</span>
          <span className="text-[10px] px-1.5 py-0.2 rounded-full bg-black/30">
            {(counts.code || 0).toLocaleString()}
          </span>
        </button>

        <button
          onClick={() => { setSelectedCategory("audio"); setSelectedExt(""); }}
          className={`flex items-center gap-2 px-3.5 py-1.5 rounded-xl text-xs font-mono font-medium transition-all ${
            selectedCategory === "audio"
              ? "bg-amber-600 text-white font-bold shadow-md shadow-amber-600/20"
              : "bg-zinc-900 hover:bg-zinc-800 text-zinc-300 border border-zinc-800"
          }`}
        >
          <Music className="w-3.5 h-3.5 text-amber-400" />
          <span>Audio</span>
          <span className="text-[10px] px-1.5 py-0.2 rounded-full bg-black/30">
            {(counts.audio || 0).toLocaleString()}
          </span>
        </button>

        <button
          onClick={() => { setSelectedCategory("video"); setSelectedExt(""); }}
          className={`flex items-center gap-2 px-3.5 py-1.5 rounded-xl text-xs font-mono font-medium transition-all ${
            selectedCategory === "video"
              ? "bg-rose-600 text-white font-bold shadow-md shadow-rose-600/20"
              : "bg-zinc-900 hover:bg-zinc-800 text-zinc-300 border border-zinc-800"
          }`}
        >
          <Video className="w-3.5 h-3.5 text-rose-400" />
          <span>Video</span>
          <span className="text-[10px] px-1.5 py-0.2 rounded-full bg-black/30">
            {(counts.video || 0).toLocaleString()}
          </span>
        </button>

        <button
          onClick={() => { setSelectedCategory("text"); setSelectedExt(""); }}
          className={`flex items-center gap-2 px-3.5 py-1.5 rounded-xl text-xs font-mono font-medium transition-all ${
            selectedCategory === "text"
              ? "bg-cyan-600 text-white font-bold shadow-md shadow-cyan-600/20"
              : "bg-zinc-900 hover:bg-zinc-800 text-zinc-300 border border-zinc-800"
          }`}
        >
          <FileText className="w-3.5 h-3.5 text-cyan-400" />
          <span>Text & Notes</span>
          <span className="text-[10px] px-1.5 py-0.2 rounded-full bg-black/30">
            {(counts.text || 0).toLocaleString()}
          </span>
        </button>
      </div>

      {/* Search Input Bar */}
      <div className="relative">
        <Search className="w-4 h-4 text-zinc-500 absolute left-4 top-1/2 -translate-y-1/2" />
        <input
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder={`Search across ${totalCount.toLocaleString()} indexed files by keyword, title, tag, or extension...`}
          className="w-full bg-zinc-900/90 border border-zinc-800 rounded-xl pl-11 pr-24 py-3 text-sm text-zinc-100 placeholder-zinc-500 focus:outline-none focus:border-amber-500 focus:ring-1 focus:ring-amber-500 transition font-mono"
        />
        <div className="absolute right-3 top-1/2 -translate-y-1/2 flex items-center gap-2">
          {loading && <RefreshCw className="w-4 h-4 text-amber-400 animate-spin" />}
          <span className="text-xs font-mono text-zinc-500">
            {totalMatches > 0 ? `${totalMatches.toLocaleString()} matches` : ""}
          </span>
        </div>
      </div>

      {/* Results List */}
      <div className="space-y-3">
        {results.map((entry) => (
          <div
            key={entry.id}
            className="p-4 rounded-xl bg-zinc-900/70 border border-zinc-800 hover:border-zinc-700 transition flex flex-col md:flex-row md:items-center justify-between gap-4 group"
          >
            <div className="space-y-1.5 flex-1 min-w-0">
              <div className="flex items-center gap-2 flex-wrap">
                {getCategoryBadge(entry.category)}
                <span className="text-xs px-2 py-0.5 rounded bg-zinc-800 text-zinc-400 font-mono">
                  {entry.extension}
                </span>
                <span className="text-xs font-mono text-zinc-500">
                  {formatBytes(entry.size_bytes)}
                </span>
              </div>

              <div className="text-sm font-semibold text-zinc-100 truncate group-hover:text-amber-300 transition">
                {entry.file_name}
              </div>

              <div className="text-xs font-mono text-zinc-500 truncate">
                {entry.path}
              </div>

              {/* Text Snippet Preview if present */}
              {entry.snippet && (
                <div className="mt-2 p-2.5 rounded-lg bg-black/40 border border-zinc-800/80 text-xs font-mono text-zinc-400 italic line-clamp-2">
                  "{entry.snippet}"
                </div>
              )}

              {/* Taxonomy Tags if present */}
              {entry.tags && entry.tags.length > 0 && (
                <div className="flex flex-wrap gap-1 mt-1">
                  {entry.tags.map((t) => (
                    <span key={t} className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-zinc-800/90 text-zinc-400">
                      #{t}
                    </span>
                  ))}
                </div>
              )}
            </div>

            {/* Action Buttons */}
            <div className="flex items-center gap-2 self-end md:self-center shrink-0">
              <button
                onClick={() => handleOpenPreview(entry)}
                className="flex items-center gap-1 px-3 py-1.5 rounded-lg text-xs font-medium bg-amber-500/10 text-amber-300 border border-amber-500/20 hover:bg-amber-500/20 transition font-mono"
              >
                <Play className="w-3.5 h-3.5" />
                <span>Play / View</span>
              </button>

              {entry.category === "video" && (
                <button
                  onClick={() =>
                    openUploader({
                      filePath: entry.full_path || `/home/justin/Dropbox/${entry.path}`,
                      title: entry.file_name.replace(/\.[^/.]+$/, ""),
                      description: `Archived from Dropbox Storehouse: ${entry.path}`,
                    })
                  }
                  className="flex items-center gap-1 px-2.5 py-1.5 rounded-lg text-xs font-medium bg-red-600/20 text-red-300 border border-red-500/30 hover:bg-red-600/30 transition font-mono"
                  title="Dispatch to YouTube Sovereign Studio"
                >
                  <Video className="w-3.5 h-3.5 text-red-400" />
                  <span>YouTube</span>
                </button>
              )}

              <button
                onClick={() => copyPath(`/home/justin/Dropbox/${entry.path}`, entry.id)}
                className="p-1.5 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-zinc-300 transition"
                title="Copy POSIX Path"
              >
                {copiedId === entry.id ? <Check className="w-4 h-4 text-emerald-400" /> : <Copy className="w-4 h-4" />}
              </button>

              <a
                href={`/api/v1/index/content?path=${encodeURIComponent(entry.path)}`}
                download={entry.file_name}
                className="p-1.5 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-zinc-300 transition"
                title="Download Local File"
              >
                <Download className="w-4 h-4" />
              </a>

              {cloudStatus?.configured && (
                <button
                  onClick={() => generateCloudLink(entry.path, `cloud-${entry.id}`)}
                  className="p-1.5 rounded-lg bg-cyan-950/60 hover:bg-cyan-900/60 text-cyan-400 border border-cyan-800/80 transition"
                  title="Copy Direct 4h Cloud Stream Link"
                >
                  {copiedId === `cloud-${entry.id}` ? <Check className="w-4 h-4 text-emerald-400" /> : <ExternalLink className="w-4 h-4" />}
                </button>
              )}
            </div>
          </div>
        ))}

        {results.length === 0 && !loading && (
          <div className="p-12 text-center rounded-2xl bg-zinc-900/40 border border-zinc-800 text-zinc-500 font-mono text-xs">
            No files matched your query. Click "Crawl & Index All" to refresh the BoltDB index across /home/justin/Dropbox.
          </div>
        )}
      </div>

      {/* In-App Streaming & Preview Modal */}
      {activePreview && (
        <div className="fixed inset-0 z-50 bg-black/80 backdrop-blur-md flex items-center justify-center p-4">
          <div className="bg-zinc-900 border border-zinc-800 rounded-2xl w-full max-w-3xl overflow-hidden shadow-2xl flex flex-col max-h-[85vh]">
            <div className="px-5 py-4 border-b border-zinc-800 flex items-center justify-between bg-zinc-950/50">
              <div className="flex items-center gap-2 min-w-0">
                {getCategoryBadge(activePreview.category)}
                <span className="text-sm font-bold text-zinc-100 truncate font-mono">
                  {activePreview.file_name}
                </span>
              </div>
              <button
                onClick={() => { setActivePreview(null); setPreviewText(null); }}
                className="p-1 rounded-lg text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800 transition"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="p-6 overflow-y-auto space-y-4">
              {/* Audio Player */}
              {activePreview.category === "audio" && (() => {
                const isAmr = activePreview.extension.toLowerCase() === ".amr" || activePreview.file_name.toLowerCase().endsWith(".amr");
                const contentUrl = `/api/v1/index/content?path=${encodeURIComponent(activePreview.path)}`;
                const rawUrl = `/api/v1/index/content?path=${encodeURIComponent(activePreview.path)}&raw=true`;

                return (
                  <div className="p-6 rounded-xl bg-black/40 border border-zinc-800 space-y-5 text-center">
                    <div className="flex items-center justify-center gap-3">
                      {isAmr ? (
                        <Radio className="w-10 h-10 text-amber-400 animate-pulse" />
                      ) : (
                        <Music className="w-10 h-10 text-amber-400 animate-pulse" />
                      )}
                    </div>

                    <div>
                      <h4 className="text-sm font-semibold text-zinc-200 truncate">{activePreview.file_name}</h4>
                      <div className="text-xs font-mono text-zinc-400 truncate mt-0.5">{activePreview.path}</div>
                    </div>

                    {/* AMR Format Transcoding Badge */}
                    {isAmr && (
                      <div className="flex flex-col items-center gap-1.5 p-3 rounded-lg bg-amber-500/10 border border-amber-500/20 text-xs text-amber-300">
                        <div className="inline-flex items-center gap-1.5 font-medium">
                          <Radio className="w-3.5 h-3.5 text-amber-400" />
                          <span>BlackBerry Voice Memo (AMR-NB) • Transcoded to 16-bit PCM WAV (8 kHz Mono)</span>
                        </div>
                        <p className="text-[11px] text-zinc-400">
                          Transparently decoded by sovereign Go engine with instant HTTP 206 byte-range seeking.
                        </p>
                      </div>
                    )}

                    {/* Persistent Volume Control HUD */}
                    <div className="bg-zinc-900/70 border border-zinc-800 rounded-lg p-3 space-y-2 text-left">
                      <div className="flex items-center justify-between text-xs">
                        <div className="flex items-center gap-2 text-zinc-300">
                          <button
                            type="button"
                            onClick={toggleAudioMute}
                            className="p-1 rounded hover:bg-zinc-800 text-zinc-300 hover:text-amber-400 transition"
                            title={audioMuted ? "Unmute" : "Mute"}
                          >
                            {audioMuted || audioVolume === 0 ? (
                              <VolumeX className="w-4 h-4 text-red-400" />
                            ) : audioVolume < 0.5 ? (
                              <Volume1 className="w-4 h-4 text-amber-400" />
                            ) : (
                              <Volume2 className="w-4 h-4 text-amber-400" />
                            )}
                          </button>
                          <span className="font-medium font-mono">
                            {audioMuted ? "MUTED" : `${Math.round(audioVolume * 100)}%`}
                          </span>
                        </div>
                        <span className="text-[10px] text-zinc-500 font-mono flex items-center gap-1">
                          <span className="w-1.5 h-1.5 rounded-full bg-emerald-400"></span>
                          Session Memory Active
                        </span>
                      </div>

                      <div className="flex items-center gap-3">
                        <input
                          type="range"
                          min="0"
                          max="1"
                          step="0.01"
                          value={audioMuted ? 0 : audioVolume}
                          onChange={(e) => updateAudioVolume(parseFloat(e.target.value))}
                          className="w-full h-1.5 bg-zinc-700 rounded-lg appearance-none cursor-pointer accent-amber-500"
                        />
                      </div>

                      <div className="flex items-center gap-1.5 pt-1">
                        <span className="text-[10px] text-zinc-500 uppercase tracking-wider">Presets:</span>
                        {[0.25, 0.5, 0.75, 1.0].map((preset) => (
                          <button
                            key={preset}
                            type="button"
                            onClick={() => updateAudioVolume(preset)}
                            className={`px-1.5 py-0.5 rounded text-[10px] font-mono transition ${
                              !audioMuted && Math.abs(audioVolume - preset) < 0.05
                                ? "bg-amber-500/20 text-amber-300 border border-amber-500/40"
                                : "bg-zinc-800 text-zinc-400 hover:text-zinc-200"
                            }`}
                          >
                            {Math.round(preset * 100)}%
                          </button>
                        ))}
                      </div>
                    </div>

                    {/* Audio Player Element */}
                    <audio
                      ref={(el) => {
                        audioRef.current = el;
                        syncAudioElement(el);
                      }}
                      controls
                      autoPlay
                      className="w-full mt-2"
                      src={contentUrl}
                      onLoadedMetadata={(e) => syncAudioElement(e.currentTarget)}
                      onCanPlay={(e) => syncAudioElement(e.currentTarget)}
                      onVolumeChange={handleAudioVolumeChange}
                    >
                      Your browser does not support audio playback.
                    </audio>

                    {/* Download options */}
                    <div className="flex items-center justify-center gap-3 pt-2">
                      <a
                        href={contentUrl}
                        download={isAmr ? activePreview.file_name.replace(/\.amr$/i, ".wav") : activePreview.file_name}
                        className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium bg-zinc-800 hover:bg-zinc-700 text-zinc-200 transition"
                      >
                        <Download className="w-3.5 h-3.5" />
                        <span>Download {isAmr ? "PCM WAV" : "Audio"}</span>
                      </a>
                      {isAmr && (
                        <a
                          href={rawUrl}
                          download={activePreview.file_name}
                          className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium bg-zinc-900 border border-zinc-700 hover:bg-zinc-800 text-zinc-300 transition"
                        >
                          <Download className="w-3.5 h-3.5" />
                          <span>Download Original (.amr)</span>
                        </a>
                      )}
                    </div>
                  </div>
                );
              })()}

              {/* Video Player */}
              {activePreview.category === "video" && (
                <div className="space-y-3">
                  <div className="rounded-xl overflow-hidden bg-black border border-zinc-800">
                    <video
                      controls
                      autoPlay
                      className="w-full max-h-[50vh]"
                      src={`/api/v1/index/content?path=${encodeURIComponent(activePreview.path)}`}
                    >
                      Your browser does not support video playback.
                    </video>
                  </div>
                  <div className="flex items-center justify-between p-3 rounded-xl bg-red-950/20 border border-red-900/40">
                    <div className="text-xs text-zinc-300 font-mono">
                      Dispatch this video directly to your authenticated YouTube channel.
                    </div>
                    <button
                      onClick={() => {
                        openUploader({
                          filePath: activePreview.full_path || `/home/justin/Dropbox/${activePreview.path}`,
                          title: activePreview.file_name.replace(/\.[^/.]+$/, ""),
                          description: `Archived from Dropbox Storehouse: ${activePreview.path}`,
                        });
                        setActivePreview(null);
                      }}
                      className="px-3 py-1.5 rounded-lg text-xs font-semibold bg-red-600 hover:bg-red-500 text-white transition flex items-center gap-1.5 font-mono shadow-lg shadow-red-950/40"
                    >
                      <Video className="w-3.5 h-3.5" />
                      <span>Upload to YouTube</span>
                    </button>
                  </div>
                </div>
              )}

              {/* Book / PDF Viewer */}
              {activePreview.category === "books" && (
                <div className="p-8 text-center rounded-xl bg-purple-950/20 border border-purple-900/40 space-y-3">
                  <BookOpen className="w-12 h-12 text-purple-400 mx-auto" />
                  <h3 className="text-base font-bold text-zinc-100">{activePreview.file_name}</h3>
                  <div className="text-xs font-mono text-zinc-400 max-w-lg mx-auto truncate">
                    {activePreview.path}
                  </div>
                  <div className="pt-2 flex justify-center gap-3">
                    <a
                      href={`/api/v1/index/content?path=${encodeURIComponent(activePreview.path)}`}
                      target="_blank"
                      rel="noreferrer"
                      className="px-4 py-2 rounded-xl text-xs font-semibold bg-purple-600 hover:bg-purple-500 text-white transition flex items-center gap-2 font-mono"
                    >
                      <ExternalLink className="w-4 h-4" />
                      <span>Open Document / PDF</span>
                    </a>
                  </div>
                </div>
              )}

              {/* Text / Code Viewer */}
              {(activePreview.category === "text" || activePreview.category === "code") && (
                <div className="space-y-2 font-mono">
                  {previewLoading && (
                    <div className="p-8 text-center text-xs text-zinc-500 flex items-center justify-center gap-2">
                      <RefreshCw className="w-4 h-4 animate-spin text-amber-400" />
                      <span>Streaming document content...</span>
                    </div>
                  )}
                  {previewText && (
                    <pre className="p-4 rounded-xl bg-black/60 border border-zinc-800 text-xs text-zinc-200 overflow-x-auto leading-relaxed max-h-[50vh]">
                      <code>{previewText}</code>
                    </pre>
                  )}
                </div>
              )}
            </div>

            <div className="px-5 py-3 border-t border-zinc-800 bg-zinc-950/50 flex items-center justify-between text-xs font-mono text-zinc-500">
              <span>{formatBytes(activePreview.size_bytes)}</span>
              <button
                onClick={() => copyPath(`/home/justin/Dropbox/${activePreview.path}`, "modal")}
                className="flex items-center gap-1 text-zinc-400 hover:text-zinc-200 transition"
              >
                <Copy className="w-3.5 h-3.5" />
                <span>{copiedId === "modal" ? "Copied!" : "Copy POSIX Path"}</span>
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

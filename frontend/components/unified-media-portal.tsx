"use client";

import React, { useState, useEffect, useRef, useCallback, useMemo } from "react";
import {
  Image as ImageIcon,
  Film,
  Video,
  Play,
  Pause,
  RotateCcw,
  RotateCw,
  Volume2,
  VolumeX,
  Search,
  RefreshCw,
  Clock,
  Calendar,
  ShieldCheck,
  Download,
  Copy,
  Check,
  Sparkles,
  Filter,
  ChevronLeft,
  ChevronRight,
  Maximize2,
  X,
  Smartphone,
  Folder,
  Layers,
  HardDrive,
  Info,
  Sliders,
  ExternalLink,
  Compass,
  Flame,
  Grid,
  List,
  Eye,
  Camera
} from "lucide-react";

export interface UnifiedVisualItem {
  id: string;
  title: string;
  file_name: string;
  source: "pixel7" | "dropbox_photos" | "dropbox_videos" | "vault" | string;
  category: "photos" | "video" | string;
  format: string;
  path: string;
  full_path: string;
  size_bytes: number;
  mod_time: string;
  media_date?: string;
  year?: number;
  duration_sec?: number;
  duration_str?: string;
  resolution?: string;
  thumbnail_url: string;
  stream_url: string;
  tags?: string[];
  dasha_mahadasha?: string;
  dasha_antardasha?: string;
  sacred_metal?: string;
  hora?: string;
  hermetic_axiom?: string;
  device_model?: string;
  location?: string;
  album?: string;
  metadata?: Record<string, any>;
}

interface VisualCatalogResponse {
  items: UnifiedVisualItem[];
  total: number;
  offset: number;
  limit: number;
  counts: {
    total: number;
    photos: number;
    videos: number;
    pixel7: number;
    dropbox_photos: number;
    dropbox_videos: number;
    years: number[];
    dashas: Record<string, number>;
  };
}

interface SyncthingStatus {
  syncthing_live: boolean;
  syncthing_url: string;
  version?: string;
  total_vault_files: number;
  total_vault_bytes: number;
  is_syncing?: boolean;
}

// Planetary Sacred Metal Color Palettes
const PLANET_COLORS: Record<string, { bg: string; text: string; border: string; metal: string }> = {
  Saturn: { bg: "bg-slate-800/80", text: "text-slate-300", border: "border-slate-600/50", metal: "Lead" },
  Jupiter: { bg: "bg-amber-950/80", text: "text-amber-300", border: "border-amber-600/50", metal: "Tin" },
  Mars: { bg: "bg-red-950/80", text: "text-red-300", border: "border-red-600/50", metal: "Iron" },
  Sun: { bg: "bg-yellow-950/80", text: "text-yellow-300", border: "border-yellow-600/50", metal: "Gold" },
  Venus: { bg: "bg-emerald-950/80", text: "text-emerald-300", border: "border-emerald-600/50", metal: "Copper" },
  Mercury: { bg: "bg-cyan-950/80", text: "text-cyan-300", border: "border-cyan-600/50", metal: "Quicksilver" },
  Moon: { bg: "bg-blue-950/80", text: "text-blue-300", border: "border-blue-600/50", metal: "Silver" },
  Rahu: { bg: "bg-purple-950/80", text: "text-purple-300", border: "border-purple-600/50", metal: "Lodestone" },
  Ketu: { bg: "bg-violet-950/80", text: "text-violet-300", border: "border-violet-600/50", metal: "Magnetite" },
};

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
}

export default function UnifiedMediaPortal() {
  // Data State
  const [catalog, setCatalog] = useState<VisualCatalogResponse | null>(null);
  const [syncthingStatus, setSyncthingStatus] = useState<SyncthingStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [rescanLoading, setRescanLoading] = useState(false);
  const [rescanMessage, setRescanMessage] = useState<string | null>(null);

  // Filters & Search
  const [sourceFilter, setSourceFilter] = useState<"all" | "pixel7" | "dropbox_photos" | "dropbox_videos">("all");
  const [categoryFilter, setCategoryFilter] = useState<"all" | "photos" | "video">("all");
  const [selectedYear, setSelectedYear] = useState<string>("all");
  const [selectedDasha, setSelectedDasha] = useState<string>("all");
  const [searchQuery, setSearchQuery] = useState("");
  const [sortBy, setSortBy] = useState<"date_desc" | "date_asc" | "size_desc" | "name_asc">("date_desc");
  const [page, setPage] = useState(1);
  const pageSize = 48;
  const [viewMode, setViewMode] = useState<"grid" | "table">("grid");

  // Lightbox & Cinema State
  const [activeMedia, setActiveMedia] = useState<UnifiedVisualItem | null>(null);
  const [activeMediaIndex, setActiveMediaIndex] = useState<number>(-1);
  const [isCinemaOpen, setIsCinemaOpen] = useState(false);
  const [copiedId, setCopiedId] = useState<string | null>(null);

  // Video Player Controls State
  const videoRef = useRef<HTMLVideoElement | null>(null);
  const [playbackRate, setPlaybackRate] = useState(1.0);
  const [isMuted, setIsMuted] = useState(false);

  // Fetch Catalog
  const fetchCatalog = useCallback(async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams();
      if (sourceFilter !== "all") params.set("source", sourceFilter);
      if (categoryFilter !== "all") params.set("category", categoryFilter);
      if (selectedYear !== "all") params.set("year", selectedYear);
      if (selectedDasha !== "all") params.set("dasha", selectedDasha);
      if (searchQuery.trim()) params.set("q", searchQuery.trim());
      params.set("sort", sortBy);
      params.set("limit", String(pageSize));
      params.set("offset", String((page - 1) * pageSize));

      const res = await fetch(`/api/v1/media/catalog?${params.toString()}`);
      if (res.ok) {
        const data: VisualCatalogResponse = await res.json();
        setCatalog(data);
      }
    } catch (err) {
      console.error("Failed to load visual catalog:", err);
    } finally {
      setLoading(false);
    }
  }, [sourceFilter, categoryFilter, selectedYear, selectedDasha, searchQuery, sortBy, page, pageSize]);

  // Fetch Syncthing Status
  const fetchSyncthingStatus = useCallback(async () => {
    try {
      const res = await fetch("/api/v1/syncthing/status");
      if (res.ok) {
        const data: SyncthingStatus = await res.json();
        setSyncthingStatus(data);
      }
    } catch (err) {
      console.error("Failed to load Syncthing status:", err);
    }
  }, []);

  useEffect(() => {
    fetchCatalog();
  }, [fetchCatalog]);

  useEffect(() => {
    fetchSyncthingStatus();
  }, [fetchSyncthingStatus]);

  // Trigger Syncthing Rescan
  const handleTriggerRescan = async () => {
    setRescanLoading(true);
    setRescanMessage(null);
    try {
      const res = await fetch("/api/v1/syncthing/rescan", { method: "POST" });
      const data = await res.json();
      if (res.ok && data.success) {
        setRescanMessage("Rescan initiated. Refreshing visual cache...");
        setTimeout(() => {
          fetchCatalog();
          fetchSyncthingStatus();
          setRescanMessage(null);
        }, 2000);
      } else {
        setRescanMessage(data.error || "Failed to trigger rescan");
      }
    } catch (err: any) {
      setRescanMessage(err.message || "Error contacting subsystem");
    } finally {
      setRescanLoading(false);
    }
  };

  // Open Media Modal (Video or Photo)
  const openMedia = (item: UnifiedVisualItem, index: number) => {
    setActiveMedia(item);
    setActiveMediaIndex(index);
    setIsCinemaOpen(true);
    setPlaybackRate(1.0);
  };

  const closeMedia = () => {
    setIsCinemaOpen(false);
    setActiveMedia(null);
    setActiveMediaIndex(-1);
  };

  // Navigate Modal Prev / Next
  const navigateMedia = useCallback((direction: "prev" | "next") => {
    if (!catalog?.items || catalog.items.length === 0) return;
    let newIndex = activeMediaIndex;
    if (direction === "prev") {
      newIndex = activeMediaIndex > 0 ? activeMediaIndex - 1 : catalog.items.length - 1;
    } else {
      newIndex = activeMediaIndex < catalog.items.length - 1 ? activeMediaIndex + 1 : 0;
    }
    setActiveMediaIndex(newIndex);
    setActiveMedia(catalog.items[newIndex]);
  }, [catalog?.items, activeMediaIndex]);

  // Keyboard navigation for Lightbox
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (!isCinemaOpen) return;
      if (e.key === "Escape") closeMedia();
      if (e.key === "ArrowLeft") navigateMedia("prev");
      if (e.key === "ArrowRight") navigateMedia("next");
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isCinemaOpen, navigateMedia]);

  // Copy WSL Path
  const copyToClipboard = (text: string, id: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  // Seek Video
  const handleSeekDelta = (seconds: number) => {
    if (videoRef.current) {
      videoRef.current.currentTime = Math.max(0, videoRef.current.currentTime + seconds);
    }
  };

  // Step Video 1 Frame (approx 1/30s)
  const handleStepFrame = (frames: number) => {
    if (videoRef.current) {
      videoRef.current.pause();
      videoRef.current.currentTime = Math.max(0, videoRef.current.currentTime + frames * (1 / 30));
    }
  };

  // Set Playback Speed
  const handlePlaybackRateChange = (rate: number) => {
    setPlaybackRate(rate);
    if (videoRef.current) {
      videoRef.current.playbackRate = rate;
    }
  };

  // Pagination calculation
  const totalItems = catalog?.total || 0;
  const totalPages = Math.ceil(totalItems / pageSize) || 1;

  return (
    <div className="space-y-6">
      {/* 1. Header Banner & Quick Telemetry */}
      <div className="bg-gradient-to-r from-slate-900 via-[#101726] to-slate-900 p-6 rounded-2xl border border-slate-800 shadow-xl space-y-4">
        <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
          <div>
            <div className="flex items-center space-x-3">
              <div className="p-2.5 rounded-xl bg-gradient-to-br from-cyan-500/20 to-indigo-500/20 border border-cyan-500/30 text-cyan-400 shadow-inner">
                <Camera className="w-5 h-5" />
              </div>
              <div>
                <h1 className="text-xl font-bold text-white tracking-tight flex items-center gap-2">
                  <span>Visual Chronicle</span>
                  <span className="text-xs font-mono font-normal px-2.5 py-0.5 rounded-full bg-cyan-950/80 text-cyan-300 border border-cyan-700/50">
                    Cinema &amp; Gallery
                  </span>
                </h1>
                <p className="text-xs text-slate-400 mt-0.5 leading-relaxed">
                  Federated visual archive uniting Syncthing Pixel 7 mobile stream and historical Dropbox photo/video vaults under Vimshottari Dasha governance.
                </p>
              </div>
            </div>
          </div>

          {/* Subsystem Telemetry & Rescan Action */}
          <div className="flex flex-wrap items-center gap-2 text-xs font-mono">
            {syncthingStatus && (
              <div className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-slate-900/90 border border-slate-800 text-slate-300 shadow-sm">
                <span className={`w-2 h-2 rounded-full ${syncthingStatus.syncthing_live ? "bg-emerald-400 animate-pulse" : "bg-red-400"}`} />
                <span>Pixel 7 Dropzone:</span>
                <span className="text-cyan-300 font-semibold">{formatBytes(syncthingStatus.total_vault_bytes)}</span>
              </div>
            )}

            <button
              onClick={handleTriggerRescan}
              disabled={rescanLoading}
              className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 transition disabled:opacity-50"
              title="Rescan Syncthing & Invalidate Visual Cache"
            >
              <RefreshCw className={`w-3.5 h-3.5 ${rescanLoading ? "animate-spin text-cyan-400" : ""}`} />
              <span>{rescanLoading ? "Rescanning..." : "Rescan Vault"}</span>
            </button>
          </div>
        </div>

        {/* Rescan feedback banner */}
        {rescanMessage && (
          <div className="text-xs font-mono px-3 py-1.5 rounded-lg bg-cyan-950/60 border border-cyan-800 text-cyan-300 flex items-center justify-between">
            <span>{rescanMessage}</span>
            <button onClick={() => setRescanMessage(null)} className="text-slate-400 hover:text-white">
              <X className="w-3.5 h-3.5" />
            </button>
          </div>
        )}

        {/* Aggregate Facet Badges */}
        {catalog?.counts && (
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-2 pt-2 border-t border-slate-800/80 font-mono text-xs">
            <div className="bg-slate-900/60 p-2.5 rounded-xl border border-slate-800/60 flex items-center justify-between">
              <span className="text-slate-400 flex items-center gap-1.5">
                <Layers className="w-3.5 h-3.5 text-cyan-400" />
                <span>Total Assets</span>
              </span>
              <span className="font-bold text-white text-sm">{catalog.counts.total.toLocaleString()}</span>
            </div>
            <div className="bg-slate-900/60 p-2.5 rounded-xl border border-slate-800/60 flex items-center justify-between">
              <span className="text-slate-400 flex items-center gap-1.5">
                <Smartphone className="w-3.5 h-3.5 text-cyan-400" />
                <span>Pixel 7 Captures</span>
              </span>
              <span className="font-bold text-cyan-300 text-sm">{catalog.counts.pixel7.toLocaleString()}</span>
            </div>
            <div className="bg-slate-900/60 p-2.5 rounded-xl border border-slate-800/60 flex items-center justify-between">
              <span className="text-slate-400 flex items-center gap-1.5">
                <ImageIcon className="w-3.5 h-3.5 text-amber-400" />
                <span>Dropbox Photos</span>
              </span>
              <span className="font-bold text-amber-300 text-sm">{catalog.counts.dropbox_photos.toLocaleString()}</span>
            </div>
            <div className="bg-slate-900/60 p-2.5 rounded-xl border border-slate-800/60 flex items-center justify-between">
              <span className="text-slate-400 flex items-center gap-1.5">
                <Film className="w-3.5 h-3.5 text-purple-400" />
                <span>Video Vault</span>
              </span>
              <span className="font-bold text-purple-300 text-sm">{catalog.counts.videos.toLocaleString()}</span>
            </div>
          </div>
        )}
      </div>

      {/* 2. Control Ribbon: Source Tabs, Categories, Search, Filters */}
      <div className="bg-slate-900/80 p-4 rounded-xl border border-slate-800 space-y-3">
        <div className="flex flex-wrap items-center justify-between gap-3">
          {/* Source Tabs */}
          <div className="flex items-center gap-1.5 bg-slate-950 p-1 rounded-xl border border-slate-800 text-xs font-mono">
            <button
              onClick={() => { setSourceFilter("all"); setPage(1); }}
              className={`px-3 py-1.5 rounded-lg transition ${
                sourceFilter === "all" ? "bg-slate-800 text-white font-semibold shadow" : "text-slate-400 hover:text-slate-200"
              }`}
            >
              All Sources
            </button>
            <button
              onClick={() => { setSourceFilter("pixel7"); setPage(1); }}
              className={`px-3 py-1.5 rounded-lg transition flex items-center gap-1.5 ${
                sourceFilter === "pixel7" ? "bg-cyan-950/80 text-cyan-300 border border-cyan-700/60 font-semibold shadow" : "text-slate-400 hover:text-cyan-300"
              }`}
            >
              <Smartphone className="w-3.5 h-3.5 text-cyan-400" />
              <span>Pixel 7 ({catalog?.counts.pixel7 ?? 0})</span>
            </button>
            <button
              onClick={() => { setSourceFilter("dropbox_photos"); setPage(1); }}
              className={`px-3 py-1.5 rounded-lg transition flex items-center gap-1.5 ${
                sourceFilter === "dropbox_photos" ? "bg-amber-950/80 text-amber-300 border border-amber-700/60 font-semibold shadow" : "text-slate-400 hover:text-amber-300"
              }`}
            >
              <ImageIcon className="w-3.5 h-3.5 text-amber-400" />
              <span>Dropbox Photos</span>
            </button>
            <button
              onClick={() => { setSourceFilter("dropbox_videos"); setPage(1); }}
              className={`px-3 py-1.5 rounded-lg transition flex items-center gap-1.5 ${
                sourceFilter === "dropbox_videos" ? "bg-purple-950/80 text-purple-300 border border-purple-700/60 font-semibold shadow" : "text-slate-400 hover:text-purple-300"
              }`}
            >
              <Video className="w-3.5 h-3.5 text-purple-400" />
              <span>Video Vault</span>
            </button>
          </div>

          {/* Category Toggle (All / Photos / Video) */}
          <div className="flex items-center gap-1 bg-slate-950 p-1 rounded-xl border border-slate-800 text-xs font-mono">
            <button
              onClick={() => { setCategoryFilter("all"); setPage(1); }}
              className={`px-2.5 py-1.5 rounded-lg transition ${
                categoryFilter === "all" ? "bg-slate-800 text-white font-semibold" : "text-slate-400 hover:text-white"
              }`}
            >
              All Formats
            </button>
            <button
              onClick={() => { setCategoryFilter("photos"); setPage(1); }}
              className={`px-2.5 py-1.5 rounded-lg transition flex items-center gap-1 ${
                categoryFilter === "photos" ? "bg-slate-800 text-amber-300 font-semibold" : "text-slate-400 hover:text-white"
              }`}
            >
              <ImageIcon className="w-3 h-3 text-amber-400" />
              <span>Photos</span>
            </button>
            <button
              onClick={() => { setCategoryFilter("video"); setPage(1); }}
              className={`px-2.5 py-1.5 rounded-lg transition flex items-center gap-1 ${
                categoryFilter === "video" ? "bg-slate-800 text-purple-300 font-semibold" : "text-slate-400 hover:text-white"
              }`}
            >
              <Film className="w-3 h-3 text-purple-400" />
              <span>Videos</span>
            </button>
          </div>

          {/* View Mode Toggle */}
          <div className="flex items-center gap-1 bg-slate-950 p-1 rounded-xl border border-slate-800">
            <button
              onClick={() => setViewMode("grid")}
              className={`p-1.5 rounded-lg transition ${
                viewMode === "grid" ? "bg-slate-800 text-cyan-400" : "text-slate-400 hover:text-white"
              }`}
              title="Grid View"
            >
              <Grid className="w-3.5 h-3.5" />
            </button>
            <button
              onClick={() => setViewMode("table")}
              className={`p-1.5 rounded-lg transition ${
                viewMode === "table" ? "bg-slate-800 text-cyan-400" : "text-slate-400 hover:text-white"
              }`}
              title="Detailed Table View"
            >
              <List className="w-3.5 h-3.5" />
            </button>
          </div>
        </div>

        {/* Secondary Filter Bar: Search, Year, Dasha, Sort */}
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-2 pt-2 border-t border-slate-800/60 text-xs">
          {/* Search Box */}
          <div className="relative">
            <Search className="w-3.5 h-3.5 text-slate-400 absolute left-3 top-3" />
            <input
              type="text"
              placeholder="Search filename, album, tags, dasha..."
              value={searchQuery}
              onChange={(e) => { setSearchQuery(e.target.value); setPage(1); }}
              className="w-full bg-slate-950 border border-slate-800 rounded-lg pl-9 pr-3 py-2 text-white placeholder-slate-500 focus:outline-none focus:border-cyan-500 transition font-mono"
            />
            {searchQuery && (
              <button
                onClick={() => { setSearchQuery(""); setPage(1); }}
                className="absolute right-3 top-2.5 text-slate-500 hover:text-white"
              >
                <X className="w-3.5 h-3.5" />
              </button>
            )}
          </div>

          {/* Year Filter */}
          <div className="flex items-center gap-2">
            <Calendar className="w-3.5 h-3.5 text-slate-400 shrink-0" />
            <select
              value={selectedYear}
              onChange={(e) => { setSelectedYear(e.target.value); setPage(1); }}
              className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-cyan-500 transition font-mono"
            >
              <option value="all">All Years</option>
              {catalog?.counts.years?.map((y) => (
                <option key={y} value={String(y)}>
                  {y}
                </option>
              ))}
            </select>
          </div>

          {/* Vimshottari Mahadasha Filter */}
          <div className="flex items-center gap-2">
            <Compass className="w-3.5 h-3.5 text-slate-400 shrink-0" />
            <select
              value={selectedDasha}
              onChange={(e) => { setSelectedDasha(e.target.value); setPage(1); }}
              className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-cyan-500 transition font-mono"
            >
              <option value="all">All Planetary Rulers (Dasha)</option>
              {Object.keys(PLANET_COLORS).map((planet) => {
                const count = catalog?.counts.dashas?.[planet] || 0;
                return (
                  <option key={planet} value={planet}>
                    {planet} ({PLANET_COLORS[planet].metal}) {count > 0 ? `[${count}]` : ""}
                  </option>
                );
              })}
            </select>
          </div>

          {/* Sort By */}
          <div className="flex items-center gap-2">
            <Sliders className="w-3.5 h-3.5 text-slate-400 shrink-0" />
            <select
              value={sortBy}
              onChange={(e) => { setSortBy(e.target.value as any); setPage(1); }}
              className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-cyan-500 transition font-mono"
            >
              <option value="date_desc">Newest First (Timeline)</option>
              <option value="date_asc">Oldest First</option>
              <option value="size_desc">Largest File Size</option>
              <option value="name_asc">Name (A-Z)</option>
            </select>
          </div>
        </div>
      </div>

      {/* 3. Media Presentation (Grid or Table) */}
      {loading ? (
        <div className="h-72 flex flex-col items-center justify-center space-y-3 bg-slate-900/40 rounded-xl border border-slate-800/80">
          <RefreshCw className="w-8 h-8 text-cyan-400 animate-spin" />
          <p className="text-xs font-mono text-slate-400">Loading sovereign media catalog...</p>
        </div>
      ) : !catalog?.items || catalog.items.length === 0 ? (
        <div className="h-64 flex flex-col items-center justify-center space-y-3 bg-slate-900/40 rounded-xl border border-slate-800/80 text-center p-6">
          <ImageIcon className="w-10 h-10 text-slate-600" />
          <h3 className="text-sm font-semibold text-slate-300">No media matches active criteria</h3>
          <p className="text-xs text-slate-500 max-w-sm">
            Try resetting your search query, source selection, or astrological filters.
          </p>
          <button
            onClick={() => {
              setSourceFilter("all");
              setCategoryFilter("all");
              setSelectedYear("all");
              setSelectedDasha("all");
              setSearchQuery("");
              setPage(1);
            }}
            className="px-3 py-1.5 rounded-lg bg-cyan-950 border border-cyan-800 text-cyan-300 text-xs font-mono hover:bg-cyan-900 transition"
          >
            Reset Filters
          </button>
        </div>
      ) : viewMode === "grid" ? (
        /* Visual Thumbnail Grid */
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-3">
          {catalog.items.map((item, index) => {
            const isVideo = item.category === "video";
            const planetMeta = item.dasha_mahadasha ? PLANET_COLORS[item.dasha_mahadasha] : null;

            return (
              <div
                key={item.id}
                onClick={() => openMedia(item, index)}
                className="group relative aspect-square rounded-xl bg-slate-950 border border-slate-800 hover:border-cyan-500/50 overflow-hidden cursor-pointer shadow-md transition-all hover:shadow-cyan-950/30 hover:scale-[1.02]"
              >
                {/* Media Image / Poster Thumbnail */}
                <img
                  src={item.thumbnail_url || item.stream_url}
                  alt={item.title || item.file_name}
                  loading="lazy"
                  decoding="async"
                  className="w-full h-full object-cover transition-transform duration-300 group-hover:scale-105"
                  onError={(e) => {
                    // Fallback to stream URL if thumbnail 404s
                    const target = e.currentTarget;
                    if (target.src !== window.location.origin + item.stream_url) {
                      target.src = item.stream_url;
                    }
                  }}
                />

                {/* Dark Gradient Overlay for Badges */}
                <div className="absolute inset-0 bg-gradient-to-t from-slate-950/90 via-transparent to-slate-950/40 opacity-80 group-hover:opacity-95 transition-opacity" />

                {/* Top Badges: Source & Astrological Ruler */}
                <div className="absolute top-2 left-2 right-2 flex items-center justify-between gap-1 pointer-events-none">
                  <span
                    className={`text-[10px] font-mono px-1.5 py-0.5 rounded-md flex items-center gap-1 shadow-sm backdrop-blur-md ${
                      item.source === "pixel7"
                        ? "bg-cyan-950/80 text-cyan-300 border border-cyan-700/60"
                        : "bg-slate-900/80 text-slate-300 border border-slate-700/60"
                    }`}
                  >
                    {item.source === "pixel7" ? <Smartphone className="w-2.5 h-2.5 text-cyan-400" /> : <Folder className="w-2.5 h-2.5 text-amber-400" />}
                    <span>{item.source === "pixel7" ? "Pixel 7" : "Vault"}</span>
                  </span>

                  {planetMeta && (
                    <span
                      className={`text-[9px] font-mono px-1.5 py-0.5 rounded-md border backdrop-blur-md shadow-sm ${planetMeta.bg} ${planetMeta.text} ${planetMeta.border}`}
                      title={`${item.dasha_mahadasha} (${planetMeta.metal})`}
                    >
                      {item.dasha_mahadasha}
                    </span>
                  )}
                </div>

                {/* Center Video Play Badge */}
                {isVideo && (
                  <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
                    <div className="w-10 h-10 rounded-full bg-slate-950/70 border border-white/20 text-white flex items-center justify-center shadow-lg group-hover:scale-110 group-hover:bg-cyan-500/80 group-hover:border-cyan-400 transition-all">
                      <Play className="w-4 h-4 fill-white ml-0.5" />
                    </div>
                  </div>
                )}

                {/* Bottom Metadata: Filename, Duration / Date */}
                <div className="absolute bottom-2 left-2 right-2 space-y-0.5 pointer-events-none">
                  <p className="text-[11px] font-medium text-white truncate drop-shadow">
                    {item.title || item.file_name}
                  </p>
                  <div className="flex items-center justify-between text-[10px] font-mono text-slate-300 drop-shadow">
                    <span>{item.media_date || item.mod_time?.substring(0, 10)}</span>
                    {isVideo && item.duration_str && (
                      <span className="px-1.5 py-0.2 rounded bg-black/60 text-purple-300 font-semibold border border-purple-800/40">
                        {item.duration_str}
                      </span>
                    )}
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      ) : (
        /* Detailed Table View */
        <div className="bg-slate-950 rounded-xl border border-slate-800 overflow-x-auto shadow-md">
          <table className="w-full text-left border-collapse text-xs font-mono">
            <thead>
              <tr className="border-b border-slate-800 bg-slate-900/60 text-slate-400">
                <th className="py-2.5 px-3">Asset</th>
                <th className="py-2.5 px-3">Category</th>
                <th className="py-2.5 px-3">Source</th>
                <th className="py-2.5 px-3">Date</th>
                <th className="py-2.5 px-3">Dasha Ruler</th>
                <th className="py-2.5 px-3">Resolution / Duration</th>
                <th className="py-2.5 px-3">Size</th>
                <th className="py-2.5 px-3 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/60">
              {catalog.items.map((item, index) => {
                const isVideo = item.category === "video";
                const planetMeta = item.dasha_mahadasha ? PLANET_COLORS[item.dasha_mahadasha] : null;

                return (
                  <tr
                    key={item.id}
                    onClick={() => openMedia(item, index)}
                    className="hover:bg-slate-900/40 transition cursor-pointer group"
                  >
                    <td className="py-2 px-3 flex items-center gap-2 max-w-xs">
                      <div className="w-8 h-8 rounded-lg bg-slate-900 border border-slate-800 overflow-hidden shrink-0">
                        <img
                          src={item.thumbnail_url || item.stream_url}
                          alt=""
                          className="w-full h-full object-cover"
                          loading="lazy"
                        />
                      </div>
                      <span className="font-medium text-white truncate" title={item.file_name}>
                        {item.file_name}
                      </span>
                    </td>
                    <td className="py-2 px-3 text-slate-300 capitalize">{item.category}</td>
                    <td className="py-2 px-3">
                      <span className={`px-2 py-0.5 rounded text-[10px] ${
                        item.source === "pixel7" ? "bg-cyan-950 text-cyan-300 border border-cyan-800" : "bg-slate-900 text-slate-400"
                      }`}>
                        {item.source === "pixel7" ? "Pixel 7" : "Dropbox"}
                      </span>
                    </td>
                    <td className="py-2 px-3 text-slate-400">{item.media_date || item.mod_time?.substring(0, 10)}</td>
                    <td className="py-2 px-3">
                      {planetMeta ? (
                        <span className={`px-2 py-0.5 rounded text-[10px] border ${planetMeta.bg} ${planetMeta.text} ${planetMeta.border}`}>
                          {item.dasha_mahadasha} ({planetMeta.metal})
                        </span>
                      ) : (
                        <span className="text-slate-600">—</span>
                      )}
                    </td>
                    <td className="py-2 px-3 text-slate-400">
                      {isVideo ? item.duration_str || "Video" : item.resolution || "Photo"}
                    </td>
                    <td className="py-2 px-3 text-slate-400">{formatBytes(item.size_bytes)}</td>
                    <td className="py-2 px-3 text-right" onClick={(e) => e.stopPropagation()}>
                      <button
                        onClick={() => copyToClipboard(item.full_path, item.id)}
                        className="p-1 hover:text-cyan-400 text-slate-400 transition"
                        title="Copy WSL Path"
                      >
                        {copiedId === item.id ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                      </button>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      {/* 4. Pagination Controls */}
      {totalPages > 1 && (
        <div className="flex flex-col sm:flex-row items-center justify-between gap-3 bg-slate-900/60 p-3 rounded-xl border border-slate-800 text-xs font-mono">
          <div className="text-slate-400">
            Showing <span className="text-white font-semibold">{(page - 1) * pageSize + 1}</span> to{" "}
            <span className="text-white font-semibold">{Math.min(page * pageSize, totalItems)}</span> of{" "}
            <span className="text-white font-semibold">{totalItems.toLocaleString()}</span> visual assets
          </div>

          <div className="flex items-center gap-2">
            <button
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page <= 1}
              className="flex items-center gap-1 px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 disabled:opacity-40 transition"
            >
              <ChevronLeft className="w-3.5 h-3.5" />
              <span>Previous</span>
            </button>

            <span className="px-3 py-1.5 rounded-lg bg-slate-950 border border-slate-800 text-slate-300">
              Page {page} of {totalPages}
            </span>

            <button
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page >= totalPages}
              className="flex items-center gap-1 px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 disabled:opacity-40 transition"
            >
              <span>Next</span>
              <ChevronRight className="w-3.5 h-3.5" />
            </button>
          </div>
        </div>
      )}

      {/* 5. Cinema Player & Photo Lightbox Modal */}
      {isCinemaOpen && activeMedia && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-6 bg-slate-950/90 backdrop-blur-md animate-in fade-in duration-200">
          <div className="relative w-full max-w-6xl max-h-[92vh] flex flex-col bg-slate-900 border border-slate-800 rounded-2xl shadow-2xl overflow-hidden">
            {/* Modal Header */}
            <div className="flex items-center justify-between p-4 border-b border-slate-800 bg-slate-950/70">
              <div className="flex items-center space-x-3 truncate">
                <span className={`p-1.5 rounded-lg text-xs font-mono flex items-center gap-1 ${
                  activeMedia.category === "video" ? "bg-purple-950 text-purple-300 border border-purple-700/50" : "bg-amber-950 text-amber-300 border border-amber-700/50"
                }`}>
                  {activeMedia.category === "video" ? <Film className="w-3.5 h-3.5" /> : <ImageIcon className="w-3.5 h-3.5" />}
                  <span className="capitalize">{activeMedia.category}</span>
                </span>
                <span className="font-semibold text-white truncate text-sm">
                  {activeMedia.title || activeMedia.file_name}
                </span>
                {activeMedia.dasha_mahadasha && (
                  <span className="hidden sm:inline-flex text-[11px] font-mono px-2 py-0.5 rounded-full bg-slate-800 text-cyan-300 border border-slate-700">
                    {activeMedia.dasha_mahadasha} Mahadasha ({activeMedia.sacred_metal || "Sacred Metal"})
                  </span>
                )}
              </div>

              <div className="flex items-center space-x-2 shrink-0">
                <button
                  onClick={() => navigateMedia("prev")}
                  className="p-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 transition"
                  title="Previous (Arrow Left)"
                >
                  <ChevronLeft className="w-4 h-4" />
                </button>
                <button
                  onClick={() => navigateMedia("next")}
                  className="p-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 transition"
                  title="Next (Arrow Right)"
                >
                  <ChevronRight className="w-4 h-4" />
                </button>
                <button
                  onClick={closeMedia}
                  className="p-1.5 rounded-lg bg-slate-800 hover:bg-red-950/80 hover:text-red-400 text-slate-300 transition"
                  title="Close (Escape)"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>
            </div>

            {/* Modal Body: Media Display & Details Panel */}
            <div className="flex-1 overflow-y-auto grid grid-cols-1 lg:grid-cols-3 gap-0">
              {/* Left / Center: Visual Display Canvas */}
              <div className="lg:col-span-2 bg-black flex flex-col items-center justify-center p-2 sm:p-4 min-h-[380px] relative">
                {activeMedia.category === "video" ? (
                  <div className="w-full h-full flex flex-col items-center justify-center">
                    <video
                      ref={videoRef}
                      src={activeMedia.stream_url}
                      controls
                      autoPlay
                      playsInline
                      className="max-h-[60vh] max-w-full rounded-lg shadow-lg border border-slate-800"
                    />

                    {/* Enhanced Playback Tool Ribbon */}
                    <div className="flex flex-wrap items-center justify-center gap-2 mt-3 text-xs font-mono">
                      <button
                        onClick={() => handleSeekDelta(-10)}
                        className="px-2.5 py-1 rounded bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 transition flex items-center gap-1"
                        title="Seek -10 seconds"
                      >
                        <RotateCcw className="w-3 h-3 text-cyan-400" />
                        <span>-10s</span>
                      </button>

                      <button
                        onClick={() => handleSeekDelta(10)}
                        className="px-2.5 py-1 rounded bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 transition flex items-center gap-1"
                        title="Seek +10 seconds"
                      >
                        <RotateCw className="w-3 h-3 text-cyan-400" />
                        <span>+10s</span>
                      </button>

                      <button
                        onClick={() => handleStepFrame(-1)}
                        className="px-2 py-1 rounded bg-slate-800 hover:bg-slate-700 text-slate-300 border border-slate-700"
                        title="Step backward 1 frame"
                      >
                        &lt; Frame
                      </button>

                      <button
                        onClick={() => handleStepFrame(1)}
                        className="px-2 py-1 rounded bg-slate-800 hover:bg-slate-700 text-slate-300 border border-slate-700"
                        title="Step forward 1 frame"
                      >
                        Frame &gt;
                      </button>

                      {/* Speed Multipliers */}
                      <div className="flex items-center rounded bg-slate-950 border border-slate-800 p-0.5">
                        {[0.5, 1.0, 1.25, 1.5, 2.0].map((rate) => (
                          <button
                            key={rate}
                            onClick={() => handlePlaybackRateChange(rate)}
                            className={`px-1.5 py-0.5 rounded text-[10px] ${
                              playbackRate === rate ? "bg-cyan-950 text-cyan-300 font-bold" : "text-slate-400 hover:text-white"
                            }`}
                          >
                            {rate}x
                          </button>
                        ))}
                      </div>
                    </div>
                  </div>
                ) : (
                  <div className="w-full h-full flex items-center justify-center">
                    <img
                      src={activeMedia.stream_url}
                      alt={activeMedia.file_name}
                      className="max-h-[65vh] max-w-full object-contain rounded-lg shadow-xl"
                    />
                  </div>
                )}
              </div>

              {/* Right: Technical Inspector & Astrological Matrix */}
              <div className="p-4 sm:p-5 bg-slate-900/90 border-t lg:border-t-0 lg:border-l border-slate-800 space-y-4 font-mono text-xs overflow-y-auto">
                {/* Astrological Alignment Card */}
                {activeMedia.dasha_mahadasha && (
                  <div className="p-3.5 rounded-xl bg-gradient-to-br from-slate-950 to-slate-900 border border-cyan-800/40 space-y-2">
                    <div className="flex items-center justify-between text-cyan-400 font-semibold">
                      <span className="flex items-center gap-1.5">
                        <Compass className="w-3.5 h-3.5" />
                        <span>Astrological Resonance</span>
                      </span>
                      <span className="text-[10px] px-1.5 py-0.5 rounded bg-cyan-950 text-cyan-300 border border-cyan-700">
                        Vimshottari
                      </span>
                    </div>

                    <div className="space-y-1 text-slate-300 text-xs">
                      <div>Mahadasha: <span className="text-white font-bold">{activeMedia.dasha_mahadasha}</span></div>
                      {activeMedia.dasha_antardasha && (
                        <div>Antardasha: <span className="text-cyan-300">{activeMedia.dasha_antardasha}</span></div>
                      )}
                      {activeMedia.sacred_metal && (
                        <div>Sacred Metal: <span className="text-amber-300">{activeMedia.sacred_metal}</span></div>
                      )}
                      {activeMedia.hermetic_axiom && (
                        <p className="text-[11px] text-slate-400 italic pt-1 border-t border-slate-800/80">
                          &ldquo;{activeMedia.hermetic_axiom}&rdquo;
                        </p>
                      )}
                    </div>
                  </div>
                )}

                {/* Technical Metadata */}
                <div className="space-y-2">
                  <h4 className="text-slate-400 font-semibold flex items-center gap-1.5 uppercase tracking-wider text-[11px]">
                    <Info className="w-3.5 h-3.5 text-slate-500" />
                    <span>Technical Telemetry</span>
                  </h4>

                  <div className="space-y-1.5 bg-slate-950 p-3 rounded-xl border border-slate-800/80 text-slate-300">
                    <div className="flex justify-between">
                      <span className="text-slate-500">File Name:</span>
                      <span className="text-white truncate max-w-[180px]" title={activeMedia.file_name}>{activeMedia.file_name}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-slate-500">Source:</span>
                      <span className="text-cyan-300 capitalize">{activeMedia.source.replace("_", " ")}</span>
                    </div>
                    {activeMedia.device_model && (
                      <div className="flex justify-between">
                        <span className="text-slate-500">Device Model:</span>
                        <span className="text-white">{activeMedia.device_model}</span>
                      </div>
                    )}
                    <div className="flex justify-between">
                      <span className="text-slate-500">Capture Date:</span>
                      <span className="text-white">{activeMedia.media_date || activeMedia.mod_time}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-slate-500">Size:</span>
                      <span className="text-white">{formatBytes(activeMedia.size_bytes)}</span>
                    </div>
                    {activeMedia.resolution && (
                      <div className="flex justify-between">
                        <span className="text-slate-500">Resolution:</span>
                        <span className="text-white">{activeMedia.resolution}</span>
                      </div>
                    )}
                    {activeMedia.duration_str && (
                      <div className="flex justify-between">
                        <span className="text-slate-500">Duration:</span>
                        <span className="text-purple-300 font-semibold">{activeMedia.duration_str}</span>
                      </div>
                    )}
                    {activeMedia.album && (
                      <div className="flex justify-between">
                        <span className="text-slate-500">Album / Vault:</span>
                        <span className="text-white">{activeMedia.album}</span>
                      </div>
                    )}
                  </div>
                </div>

                {/* Canonical File Path */}
                <div className="space-y-1.5">
                  <div className="flex items-center justify-between text-slate-400">
                    <span className="text-[11px] font-semibold uppercase tracking-wider flex items-center gap-1">
                      <HardDrive className="w-3.5 h-3.5 text-slate-500" />
                      <span>Canonical Substrate Path</span>
                    </span>
                    <button
                      onClick={() => copyToClipboard(activeMedia.full_path, "modal-path")}
                      className="text-cyan-400 hover:text-cyan-300 transition flex items-center gap-1"
                    >
                      {copiedId === "modal-path" ? <Check className="w-3 h-3 text-emerald-400" /> : <Copy className="w-3 h-3" />}
                      <span>{copiedId === "modal-path" ? "Copied" : "Copy"}</span>
                    </button>
                  </div>
                  <div className="p-2 rounded-lg bg-slate-950 border border-slate-800 text-[11px] text-slate-400 break-all select-all font-mono">
                    {activeMedia.full_path}
                  </div>
                </div>

                {/* Quick Actions */}
                <div className="pt-2 flex flex-col gap-2">
                  <a
                    href={activeMedia.stream_url}
                    download={activeMedia.file_name}
                    className="w-full flex items-center justify-center space-x-2 px-3 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-white font-medium border border-slate-700 transition shadow-sm"
                  >
                    <Download className="w-3.5 h-3.5 text-cyan-400" />
                    <span>Download Original Media</span>
                  </a>

                  <a
                    href={activeMedia.stream_url}
                    target="_blank"
                    rel="noreferrer"
                    className="w-full flex items-center justify-center space-x-2 px-3 py-2 rounded-xl bg-slate-950 hover:bg-slate-800 text-slate-300 font-medium border border-slate-800 transition"
                  >
                    <ExternalLink className="w-3.5 h-3.5 text-slate-400" />
                    <span>Open Raw HTTP 206 Stream</span>
                  </a>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

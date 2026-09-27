"use client";

import React, { useState, useEffect, useCallback } from "react";
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
  AlertCircle
} from "lucide-react";

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

export default function DropboxOmniExplorer() {
  const [query, setQuery] = useState("");
  const [selectedCategory, setSelectedCategory] = useState<string>("all");
  const [selectedExt, setSelectedExt] = useState<string>("");
  const [results, setResults] = useState<IndexEntry[]>([]);
  const [totalMatches, setTotalMatches] = useState(0);
  const [loading, setLoading] = useState(false);
  const [status, setStatus] = useState<CrawlStatus | null>(null);
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [activePreview, setActivePreview] = useState<IndexEntry | null>(null);

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

  const searchFiles = useCallback(async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams();
      if (query.trim()) params.set("q", query.trim());
      if (selectedCategory && selectedCategory !== "all") params.set("category", selectedCategory);
      if (selectedExt) params.set("ext", selectedExt);
      params.set("limit", "60");

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
    const interval = setInterval(fetchStatus, 4000);
    return () => clearInterval(interval);
  }, [fetchStatus]);

  useEffect(() => {
    const delayDebounce = setTimeout(() => {
      searchFiles();
    }, 250);
    return () => clearTimeout(delayDebounce);
  }, [searchFiles]);

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

  const copyPath = (path: string, id: string) => {
    navigator.clipboard.writeText(path);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB"];
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

  return (
    <div className="space-y-6">
      {/* Top Banner & Control Bar */}
      <div className="p-6 rounded-2xl bg-zinc-900/80 border border-zinc-800 shadow-xl backdrop-blur-md">
        <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-6">
          <div>
            <div className="flex items-center gap-3">
              <div className="p-2.5 rounded-xl bg-amber-500/10 text-amber-400 border border-amber-500/20">
                <FolderTree className="w-6 h-6" />
              </div>
              <div>
                <h2 className="text-xl font-bold text-zinc-100 flex items-center gap-2">
                  Dropbox Omni-Storehouse Indexer
                  <span className="text-xs px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/30">
                    Local POSIX Engine
                  </span>
                </h2>
                <p className="text-sm text-zinc-400 mt-0.5">
                  Direct filesystem access to <code className="text-amber-300 font-mono text-xs">/home/justin/Dropbox</code> with BoltDB indexed metadata.
                </p>
              </div>
            </div>
          </div>

          {/* Action Trigger Buttons */}
          <div className="flex flex-wrap items-center gap-3">
            <button
              onClick={() => triggerScan("all")}
              disabled={status?.is_scanning}
              className={`flex items-center gap-2 px-4 py-2 rounded-xl text-sm font-medium transition-all ${
                status?.is_scanning
                  ? "bg-amber-500/20 text-amber-300 border border-amber-500/30 cursor-not-allowed animate-pulse"
                  : "bg-amber-500 text-zinc-950 hover:bg-amber-400 font-semibold shadow-lg shadow-amber-500/20"
              }`}
            >
              <RefreshCw className={`w-4 h-4 ${status?.is_scanning ? "animate-spin" : ""}`} />
              {status?.is_scanning ? `Indexing (${status.current_category || "..."})...` : "Crawl & Index All"}
            </button>

            <button
              onClick={() => triggerScan("books")}
              disabled={status?.is_scanning}
              className="flex items-center gap-1.5 px-3 py-2 rounded-xl text-xs font-medium bg-zinc-800 hover:bg-zinc-700 text-zinc-200 border border-zinc-700 transition-colors"
            >
              <BookOpen className="w-3.5 h-3.5 text-purple-400" />
              Index Books (35k)
            </button>
            <button
              onClick={() => triggerScan("code")}
              disabled={status?.is_scanning}
              className="flex items-center gap-1.5 px-3 py-2 rounded-xl text-xs font-medium bg-zinc-800 hover:bg-zinc-700 text-zinc-200 border border-zinc-700 transition-colors"
            >
              <Code2 className="w-3.5 h-3.5 text-emerald-400" />
              Index Code
            </button>
          </div>
        </div>

        {/* Live Indexing Status Pill */}
        {status?.is_scanning && (
          <div className="mt-4 p-3 rounded-xl bg-amber-500/10 border border-amber-500/30 flex items-center justify-between text-xs text-amber-200">
            <div className="flex items-center gap-2">
              <RefreshCw className="w-4 h-4 animate-spin text-amber-400" />
              <span>Scanning category <strong>{status.current_category}</strong>... Found {status.total_seen} files, indexed {status.total_indexed}.</span>
            </div>
          </div>
        )}

        {/* Category Stats Grid */}
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3 mt-6 pt-6 border-t border-zinc-800/80">
          <div 
            onClick={() => setSelectedCategory("all")} 
            className={`p-3 rounded-xl border transition-all cursor-pointer ${
              selectedCategory === "all" ? "bg-amber-500/10 border-amber-500/40" : "bg-zinc-950/50 border-zinc-800/60 hover:border-zinc-700"
            }`}
          >
            <div className="text-xs text-zinc-400 font-medium">Total Indexed</div>
            <div className="text-lg font-bold text-zinc-100 mt-1">{(counts.total || 0).toLocaleString()}</div>
          </div>

          <div 
            onClick={() => setSelectedCategory("books")} 
            className={`p-3 rounded-xl border transition-all cursor-pointer ${
              selectedCategory === "books" ? "bg-purple-500/10 border-purple-500/40" : "bg-zinc-950/50 border-zinc-800/60 hover:border-zinc-700"
            }`}
          >
            <div className="text-xs text-purple-400 font-medium flex items-center gap-1"><BookOpen className="w-3 h-3" /> Books</div>
            <div className="text-lg font-bold text-zinc-100 mt-1">{(counts.books || 0).toLocaleString()}</div>
          </div>

          <div 
            onClick={() => setSelectedCategory("text")} 
            className={`p-3 rounded-xl border transition-all cursor-pointer ${
              selectedCategory === "text" ? "bg-cyan-500/10 border-cyan-500/40" : "bg-zinc-950/50 border-zinc-800/60 hover:border-zinc-700"
            }`}
          >
            <div className="text-xs text-cyan-400 font-medium flex items-center gap-1"><FileText className="w-3 h-3" /> Text</div>
            <div className="text-lg font-bold text-zinc-100 mt-1">{(counts.text || 0).toLocaleString()}</div>
          </div>

          <div 
            onClick={() => setSelectedCategory("audio")} 
            className={`p-3 rounded-xl border transition-all cursor-pointer ${
              selectedCategory === "audio" ? "bg-amber-500/10 border-amber-500/40" : "bg-zinc-950/50 border-zinc-800/60 hover:border-zinc-700"
            }`}
          >
            <div className="text-xs text-amber-400 font-medium flex items-center gap-1"><Music className="w-3 h-3" /> Audio</div>
            <div className="text-lg font-bold text-zinc-100 mt-1">{(counts.audio || 0).toLocaleString()}</div>
          </div>

          <div 
            onClick={() => setSelectedCategory("video")} 
            className={`p-3 rounded-xl border transition-all cursor-pointer ${
              selectedCategory === "video" ? "bg-rose-500/10 border-rose-500/40" : "bg-zinc-950/50 border-zinc-800/60 hover:border-zinc-700"
            }`}
          >
            <div className="text-xs text-rose-400 font-medium flex items-center gap-1"><Video className="w-3 h-3" /> Video</div>
            <div className="text-lg font-bold text-zinc-100 mt-1">{(counts.video || 0).toLocaleString()}</div>
          </div>

          <div 
            onClick={() => setSelectedCategory("code")} 
            className={`p-3 rounded-xl border transition-all cursor-pointer ${
              selectedCategory === "code" ? "bg-emerald-500/10 border-emerald-500/40" : "bg-zinc-950/50 border-zinc-800/60 hover:border-zinc-700"
            }`}
          >
            <div className="text-xs text-emerald-400 font-medium flex items-center gap-1"><Code2 className="w-3 h-3" /> Code</div>
            <div className="text-lg font-bold text-zinc-100 mt-1">{(counts.code || 0).toLocaleString()}</div>
          </div>
        </div>
      </div>

      {/* Search & Filtering Bar */}
      <div className="p-4 rounded-2xl bg-zinc-900/60 border border-zinc-800 space-y-4">
        <div className="relative">
          <Search className="w-5 h-5 absolute left-3.5 top-1/2 -translate-y-1/2 text-zinc-400" />
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search across entire Dropbox library (titles, authors, notes, audio, code)..."
            className="w-full bg-zinc-950 border border-zinc-800 rounded-xl pl-11 pr-4 py-3 text-sm text-zinc-100 placeholder-zinc-500 focus:outline-none focus:border-amber-500/60 focus:ring-1 focus:ring-amber-500/60 transition-all"
          />
        </div>

        {/* Filter Quick Pills */}
        <div className="flex flex-wrap items-center justify-between gap-3 text-xs">
          <div className="flex flex-wrap items-center gap-1.5">
            <span className="text-zinc-500 font-medium mr-1">Extension:</span>
            {["", ".pdf", ".txt", ".mp4", ".wav", ".py", ".go", ".json"].map((ext) => (
              <button
                key={ext || "all"}
                onClick={() => setSelectedExt(ext)}
                className={`px-2.5 py-1 rounded-lg border font-mono transition-colors ${
                  selectedExt === ext
                    ? "bg-amber-500/20 text-amber-300 border-amber-500/40 font-semibold"
                    : "bg-zinc-950 text-zinc-400 border-zinc-800 hover:border-zinc-700"
                }`}
              >
                {ext || "all"}
              </button>
            ))}
          </div>

          <div className="text-zinc-400">
            Found <strong className="text-zinc-100">{totalMatches}</strong> matching files
          </div>
        </div>
      </div>

      {/* Results List */}
      <div className="space-y-3">
        {loading ? (
          <div className="p-12 text-center text-zinc-400 space-y-3">
            <RefreshCw className="w-6 h-6 animate-spin mx-auto text-amber-400" />
            <p className="text-sm">Searching storehouse index...</p>
          </div>
        ) : results.length === 0 ? (
          <div className="p-12 text-center rounded-2xl bg-zinc-900/40 border border-zinc-800/80 space-y-3">
            <AlertCircle className="w-8 h-8 text-zinc-500 mx-auto" />
            <h3 className="text-base font-semibold text-zinc-200">No indexed files found</h3>
            <p className="text-xs text-zinc-400 max-w-md mx-auto">
              {counts.total === 0 
                ? "Your local index is currently empty. Click 'Crawl & Index All' above to build the BoltDB index from /home/justin/Dropbox." 
                : "No items matched your current search filters. Try broadening your query."}
            </p>
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-3">
            {results.map((entry) => (
              <div
                key={entry.id}
                className="p-4 rounded-xl bg-zinc-900/70 border border-zinc-800/90 hover:border-zinc-700 transition-all flex flex-col md:flex-row md:items-center justify-between gap-4 group"
              >
                <div className="space-y-1.5 min-w-0 flex-1">
                  <div className="flex flex-wrap items-center gap-2">
                    {getCategoryBadge(entry.category)}
                    <span className="text-sm font-semibold text-zinc-100 truncate group-hover:text-amber-300 transition-colors">
                      {entry.file_name}
                    </span>
                    <span className="text-xs font-mono text-zinc-500">
                      {formatBytes(entry.size_bytes)}
                    </span>
                  </div>

                  <div className="text-xs text-zinc-400 truncate font-mono">
                    {entry.path}
                  </div>

                  {entry.snippet && (
                    <p className="text-xs text-zinc-400 line-clamp-2 bg-zinc-950/60 p-2 rounded-lg border border-zinc-800/60 font-mono text-xs">
                      {entry.snippet}
                    </p>
                  )}

                  {/* Tags / Metadata attributes */}
                  {entry.tags && entry.tags.length > 0 && (
                    <div className="flex flex-wrap gap-1 mt-1">
                      {entry.tags.slice(0, 5).map((tag, idx) => (
                        <span key={idx} className="text-[10px] px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-400 border border-zinc-700/60">
                          #{tag}
                        </span>
                      ))}
                    </div>
                  )}
                </div>

                {/* Action Buttons */}
                <div className="flex items-center gap-2 self-end md:self-center shrink-0">
                  <button
                    onClick={() => setActivePreview(entry)}
                    className="p-2 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-zinc-200 border border-zinc-700 transition-colors flex items-center gap-1.5 text-xs"
                    title="Open Stream / Preview"
                  >
                    <Play className="w-3.5 h-3.5 text-amber-400" />
                    Preview
                  </button>

                  <button
                    onClick={() => copyPath(entry.full_path || entry.path, entry.id)}
                    className="p-2 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-zinc-300 border border-zinc-700 transition-colors"
                    title="Copy Path"
                  >
                    {copiedId === entry.id ? <Check className="w-4 h-4 text-emerald-400" /> : <Copy className="w-4 h-4" />}
                  </button>

                  <a
                    href={`/api/v1/index/content?path=${encodeURIComponent(entry.path)}`}
                    target="_blank"
                    rel="noreferrer"
                    className="p-2 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-zinc-300 border border-zinc-700 transition-colors"
                    title="Direct Stream Link"
                  >
                    <ExternalLink className="w-4 h-4" />
                  </a>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Media Stream & Preview Modal */}
      {activePreview && (
        <div className="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-zinc-900 border border-zinc-800 rounded-2xl max-w-3xl w-full max-h-[85vh] overflow-hidden flex flex-col shadow-2xl">
            <div className="p-4 border-b border-zinc-800 flex items-center justify-between bg-zinc-950/50">
              <div className="flex items-center gap-2 min-w-0">
                {getCategoryBadge(activePreview.category)}
                <h4 className="text-sm font-semibold text-zinc-100 truncate">{activePreview.file_name}</h4>
              </div>
              <button 
                onClick={() => setActivePreview(null)}
                className="text-zinc-400 hover:text-zinc-100 p-1.5 rounded-lg hover:bg-zinc-800 transition-colors"
              >
                ✕
              </button>
            </div>

            <div className="p-6 overflow-y-auto space-y-4 flex-1">
              {/* Category-Specific Player/Viewer */}
              {activePreview.category === "audio" && (
                <div className="p-6 rounded-xl bg-zinc-950 border border-zinc-800 text-center space-y-4">
                  <Music className="w-12 h-12 text-amber-400 mx-auto animate-bounce" />
                  <audio 
                    controls 
                    autoPlay 
                    src={`/api/v1/index/content?path=${encodeURIComponent(activePreview.path)}`}
                    className="w-full"
                  />
                  <p className="text-xs text-zinc-500 font-mono">Direct local streaming with HTTP Range seeking support</p>
                </div>
              )}

              {activePreview.category === "video" && (
                <div className="rounded-xl overflow-hidden bg-black border border-zinc-800">
                  <video 
                    controls 
                    autoPlay 
                    src={`/api/v1/index/content?path=${encodeURIComponent(activePreview.path)}`}
                    className="w-full max-h-[400px]"
                  />
                </div>
              )}

              {(activePreview.category === "books" || activePreview.extension === ".pdf") && (
                <div className="p-6 rounded-xl bg-zinc-950 border border-zinc-800 text-center space-y-4">
                  <BookOpen className="w-12 h-12 text-purple-400 mx-auto" />
                  <div>
                    <h5 className="text-sm font-medium text-zinc-200">{activePreview.file_name}</h5>
                    <p className="text-xs text-zinc-500 mt-1">Size: {formatBytes(activePreview.size_bytes)}</p>
                  </div>
                  <a
                    href={`/api/v1/index/content?path=${encodeURIComponent(activePreview.path)}`}
                    target="_blank"
                    rel="noreferrer"
                    className="inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-purple-600 hover:bg-purple-500 text-white text-xs font-semibold shadow-lg shadow-purple-600/30 transition-colors"
                  >
                    <ExternalLink className="w-4 h-4" />
                    Open PDF in Dedicated Browser Viewer
                  </a>
                </div>
              )}

              {(activePreview.category === "text" || activePreview.category === "code") && (
                <div className="space-y-2">
                  <div className="flex items-center justify-between text-xs text-zinc-400">
                    <span>File Preview</span>
                    <span className="font-mono">{activePreview.path}</span>
                  </div>
                  <pre className="p-4 rounded-xl bg-zinc-950 border border-zinc-800 text-xs font-mono text-zinc-300 overflow-x-auto whitespace-pre-wrap max-h-[400px]">
                    {activePreview.snippet || "No text snippet available."}
                  </pre>
                </div>
              )}

              {/* Detailed Metadata Fields */}
              <div className="pt-4 border-t border-zinc-800 grid grid-cols-2 gap-3 text-xs">
                <div>
                  <span className="text-zinc-500">Path:</span>
                  <p className="font-mono text-zinc-300 truncate">{activePreview.path}</p>
                </div>
                <div>
                  <span className="text-zinc-500">Size:</span>
                  <p className="font-mono text-zinc-300">{formatBytes(activePreview.size_bytes)}</p>
                </div>
                <div>
                  <span className="text-zinc-500">Indexed At:</span>
                  <p className="font-mono text-zinc-300">{new Date(activePreview.indexed_at).toLocaleString()}</p>
                </div>
                <div>
                  <span className="text-zinc-500">Full Local Path:</span>
                  <p className="font-mono text-zinc-300 truncate">{activePreview.full_path}</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

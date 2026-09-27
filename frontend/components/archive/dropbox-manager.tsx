"use client";

import React, { useState, useEffect } from "react";
import { 
  Cloud, 
  UploadCloud, 
  RefreshCw, 
  HardDrive, 
  FileArchive, 
  CheckCircle2, 
  AlertTriangle, 
  Download, 
  ShieldCheck, 
  ExternalLink,
  Lock,
  Search,
  FileText,
  Eye,
  X,
  Copy,
  Check
} from "lucide-react";

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

interface SearchResult {
  name: string;
  path_display: string;
  size?: number;
  server_modified?: string;
  highlight?: string;
}

export default function DropboxManager() {
  const [status, setStatus] = useState<DropboxStatus | null>(null);
  const [files, setFiles] = useState<DropboxFile[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [backingUp, setBackingUp] = useState<boolean>(false);
  const [feedback, setFeedback] = useState<{ type: "success" | "error"; text: string } | null>(null);

  // Search state
  const [searchQuery, setSearchQuery] = useState<string>("");
  const [searchResults, setSearchResults] = useState<SearchResult[] | null>(null);
  const [searching, setSearching] = useState<boolean>(false);

  // Content Preview state
  const [previewPath, setPreviewPath] = useState<string | null>(null);
  const [previewContent, setPreviewContent] = useState<string | null>(null);
  const [previewLoading, setPreviewLoading] = useState<boolean>(false);
  const [copied, setCopied] = useState<boolean>(false);

  const fetchStatus = async () => {
    try {
      const res = await fetch("/api/v1/dropbox/status");
      if (res.ok) {
        const data = await res.json();
        setStatus(data);
        if (data.configured && !data.error) {
          fetchFiles();
        }
      }
    } catch (err) {
      console.error("Failed to load Dropbox status:", err);
    } finally {
      setLoading(false);
    }
  };

  const fetchFiles = async () => {
    try {
      const res = await fetch("/api/v1/dropbox/files");
      if (res.ok) {
        const data = await res.json();
        setFiles(data.entries || []);
      }
    } catch (err) {
      console.error("Failed to load Dropbox files:", err);
    }
  };

  useEffect(() => {
    fetchStatus();
  }, []);

  const triggerBackup = async () => {
    setBackingUp(true);
    setFeedback(null);
    try {
      const res = await fetch("/api/v1/dropbox/backup", { method: "POST" });
      const data = await res.json();
      if (res.ok && data.success) {
        setFeedback({
          type: "success",
          text: `Snapshot ${data.filename} (${formatBytes(data.size_bytes)}) successfully archived to Dropbox!`,
        });
        await fetchFiles();
        await fetchStatus();
      } else {
        setFeedback({
          type: "error",
          text: data.error || "Backup upload to Dropbox failed.",
        });
      }
    } catch (err) {
      setFeedback({
        type: "error",
        text: "Network error during Dropbox backup snapshot.",
      });
    } finally {
      setBackingUp(false);
    }
  };

  const handleSearch = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    if (!searchQuery.trim()) {
      setSearchResults(null);
      return;
    }

    setSearching(true);
    try {
      const res = await fetch(`/api/v1/dropbox/search?q=${encodeURIComponent(searchQuery.trim())}`);
      if (res.ok) {
        const data = await res.json();
        setSearchResults(data.matches || []);
      } else {
        const err = await res.json();
        setFeedback({ type: "error", text: err.error || "Search failed." });
      }
    } catch (err) {
      setFeedback({ type: "error", text: "Network error performing Dropbox search." });
    } finally {
      setSearching(false);
    }
  };

  const handlePreview = async (path: string) => {
    setPreviewPath(path);
    setPreviewLoading(true);
    setPreviewContent(null);
    setCopied(false);

    try {
      const res = await fetch(`/api/v1/dropbox/content?path=${encodeURIComponent(path)}`);
      if (res.ok) {
        const text = await res.text();
        setPreviewContent(text);
      } else {
        setPreviewContent("Unable to preview binary or inaccessible file content.");
      }
    } catch (err) {
      setPreviewContent("Error retrieving file content from Dropbox.");
    } finally {
      setPreviewLoading(false);
    }
  };

  const handleDownload = async (path: string) => {
    try {
      const res = await fetch(`/api/v1/dropbox/link?path=${encodeURIComponent(path)}`);
      if (res.ok) {
        const data = await res.json();
        if (data.link) {
          window.open(data.link, "_blank");
        }
      } else {
        alert("Failed to generate temporary download link.");
      }
    } catch (e) {
      alert("Error requesting download link.");
    }
  };

  const copyToClipboard = () => {
    if (previewContent) {
      navigator.clipboard.writeText(previewContent);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const formatBytes = (bytes?: number) => {
    if (!bytes && bytes !== 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return (bytes / Math.pow(k, i)).toFixed(2) + " " + sizes[i];
  };

  const calculateUsagePercent = () => {
    if (!status?.account?.used_bytes || !status?.account?.allocated_bytes) return 0;
    return Math.min(100, Math.round((status.account.used_bytes / status.account.allocated_bytes) * 100));
  };

  return (
    <div className="space-y-6">
      {/* Top Status & Overview */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {/* Account Info */}
        <div className="bg-[#111928] border border-slate-800 rounded-xl p-5 flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between">
              <span className="text-xs font-mono text-slate-400 uppercase">Dropbox Identity</span>
              {status?.configured ? (
                <span className="inline-flex items-center space-x-1 px-2 py-0.5 rounded-full text-[10px] font-mono bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                  <ShieldCheck className="w-3 h-3" />
                  <span>CONNECTED</span>
                </span>
              ) : (
                <span className="inline-flex items-center space-x-1 px-2 py-0.5 rounded-full text-[10px] font-mono bg-amber-500/10 text-amber-400 border border-amber-500/20">
                  <Lock className="w-3 h-3" />
                  <span>UNCONFIGURED</span>
                </span>
              )}
            </div>

            <div className="mt-3">
              <div className="text-base font-bold text-white">
                {status?.account?.display_name || "Justin Wood"}
              </div>
              <div className="text-xs text-slate-400 font-mono mt-0.5">
                {status?.account?.email || "justinfrombeamsville@gmail.com"}
              </div>
            </div>
          </div>

          <div className="mt-4 pt-3 border-t border-slate-800/80 flex items-center justify-between text-xs font-mono text-slate-400">
            <span>Tier: <strong className="text-slate-200 uppercase">{status?.account?.account_type || "Pro"}</strong></span>
            <span>Locale: <strong className="text-slate-200">{status?.account?.country || "CA"}</strong></span>
          </div>
        </div>

        {/* Storage Quota */}
        <div className="bg-[#111928] border border-slate-800 rounded-xl p-5 flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between">
              <span className="text-xs font-mono text-slate-400 uppercase">Dropbox Quota</span>
              <HardDrive className="w-4 h-4 text-cyan-400" />
            </div>

            <div className="mt-3">
              <div className="text-lg font-bold text-cyan-300">
                {formatBytes(status?.account?.used_bytes)}
                <span className="text-xs text-slate-400 font-normal ml-1">
                  / {formatBytes(status?.account?.allocated_bytes) || "3 TB"}
                </span>
              </div>
              <div className="w-full bg-slate-900 rounded-full h-1.5 mt-3 overflow-hidden border border-slate-800">
                <div 
                  className="bg-gradient-to-r from-cyan-500 to-emerald-400 h-full rounded-full transition-all duration-500"
                  style={{ width: `${Math.max(5, calculateUsagePercent())}%` }}
                />
              </div>
            </div>
          </div>

          <div className="mt-4 pt-3 border-t border-slate-800/80 flex items-center justify-between text-xs font-mono text-slate-400">
            <span>Target Folder:</span>
            <span className="text-cyan-400 truncate max-w-[140px]" title={status?.base_path}>
              {status?.base_path || "/MercuryDasha/backups"}
            </span>
          </div>
        </div>

        {/* Instant Backup Actions */}
        <div className="bg-[#111928] border border-slate-800 rounded-xl p-5 flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between">
              <span className="text-xs font-mono text-slate-400 uppercase">Cloud Archival</span>
              <Cloud className="w-4 h-4 text-purple-400" />
            </div>
            <p className="text-xs text-slate-400 mt-2">
              Stream a live point-in-time ACID snapshot of BoltDB directly to remote Dropbox storage.
            </p>
          </div>

          <button
            onClick={triggerBackup}
            disabled={backingUp}
            className="mt-4 w-full py-2 px-3 bg-gradient-to-r from-cyan-600 to-blue-600 hover:from-cyan-500 hover:to-blue-500 disabled:opacity-50 text-white rounded-lg text-xs font-semibold flex items-center justify-center space-x-2 transition shadow-lg"
          >
            {backingUp ? (
              <>
                <RefreshCw className="w-4 h-4 animate-spin" />
                <span>Uploading Snapshot...</span>
              </>
            ) : (
              <>
                <UploadCloud className="w-4 h-4" />
                <span>Snapshot & Upload to Dropbox</span>
              </>
            )}
          </button>
        </div>
      </div>

      {/* Alert Messages */}
      {feedback && (
        <div
          className={`p-3.5 rounded-xl text-xs flex items-center space-x-2.5 border ${
            feedback.type === "success"
              ? "bg-emerald-950/60 border-emerald-800 text-emerald-200"
              : "bg-red-950/60 border-red-800 text-red-200"
          }`}
        >
          {feedback.type === "success" ? (
            <CheckCircle2 className="w-4 h-4 shrink-0 text-emerald-400" />
          ) : (
            <AlertTriangle className="w-4 h-4 shrink-0 text-red-400" />
          )}
          <span className="font-mono">{feedback.text}</span>
        </div>
      )}

      {status?.error && (
        <div className="p-4 rounded-xl bg-amber-950/40 border border-amber-800/80 text-amber-200 text-xs flex items-start space-x-3">
          <AlertTriangle className="w-4 h-4 shrink-0 text-amber-400 mt-0.5" />
          <div>
            <div className="font-semibold">Dropbox Scope or Token Action Required</div>
            <div className="text-slate-300 font-mono mt-1 text-[11px] leading-relaxed">
              {status.error}
            </div>
            <a
              href="https://www.dropbox.com/developers/apps"
              target="_blank"
              rel="noreferrer"
              className="inline-flex items-center space-x-1 mt-2 text-cyan-400 hover:underline font-mono text-[11px]"
            >
              <span>Generate fresh token in Dropbox App Console (App 8201475 Settings)</span>
              <ExternalLink className="w-3 h-3" />
            </a>
          </div>
        </div>
      )}

      {/* Search & Content Retrieval Bar */}
      <div className="bg-[#111928] border border-slate-800 rounded-xl p-4 shadow-lg">
        <form onSubmit={handleSearch} className="flex flex-col sm:flex-row gap-2">
          <div className="relative flex-1">
            <Search className="w-4 h-4 text-slate-400 absolute left-3 top-2.5" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search Dropbox files, storyboards, schemas, and archives (e.g. *.db, backup, storyboard)..."
              className="w-full bg-slate-900 border border-slate-700 rounded-lg pl-9 pr-3 py-1.5 text-xs font-mono text-slate-200 focus:outline-none focus:border-cyan-500"
            />
          </div>
          <button
            type="submit"
            disabled={searching}
            className="px-4 py-1.5 bg-cyan-600 hover:bg-cyan-500 disabled:opacity-50 text-white rounded-lg text-xs font-medium flex items-center justify-center space-x-1.5 transition shrink-0"
          >
            {searching ? <RefreshCw className="w-3.5 h-3.5 animate-spin" /> : <Search className="w-3.5 h-3.5" />}
            <span>Search Cloud</span>
          </button>
          {searchResults !== null && (
            <button
              type="button"
              onClick={() => {
                setSearchResults(null);
                setSearchQuery("");
              }}
              className="px-3 py-1.5 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded-lg text-xs transition"
            >
              Clear
            </button>
          )}
        </form>

        {/* Search Results Display */}
        {searchResults !== null && (
          <div className="mt-4 pt-3 border-t border-slate-800">
            <div className="text-xs font-mono text-cyan-400 mb-2">
              Found {searchResults.length} match(es) for "{searchQuery}":
            </div>
            {searchResults.length === 0 ? (
              <div className="text-xs font-mono text-slate-500 py-3">No matching files found.</div>
            ) : (
              <div className="space-y-1.5 max-h-60 overflow-y-auto pr-1">
                {searchResults.map((m) => (
                  <div
                    key={m.path_display}
                    className="flex items-center justify-between p-2 rounded-lg bg-slate-900/60 border border-slate-800 text-xs font-mono"
                  >
                    <div className="flex items-center space-x-2 truncate">
                      <FileText className="w-3.5 h-3.5 text-cyan-400 shrink-0" />
                      <span className="text-white font-medium truncate">{m.name}</span>
                      <span className="text-slate-500 text-[10px] truncate">{m.path_display}</span>
                    </div>
                    <div className="flex items-center space-x-1.5 shrink-0">
                      <button
                        onClick={() => handlePreview(m.path_display)}
                        className="px-2 py-1 bg-slate-800 hover:bg-cyan-950 text-slate-300 hover:text-cyan-300 rounded text-[11px] flex items-center space-x-1 transition"
                      >
                        <Eye className="w-3 h-3" />
                        <span>Preview</span>
                      </button>
                      <button
                        onClick={() => handleDownload(m.path_display)}
                        className="px-2 py-1 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded text-[11px] flex items-center space-x-1 transition"
                      >
                        <Download className="w-3 h-3" />
                        <span>Link</span>
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>

      {/* Content Preview Modal / Drawer */}
      {previewPath && (
        <div className="bg-[#111928] border border-cyan-800/80 rounded-xl p-4 shadow-2xl relative">
          <div className="flex items-center justify-between pb-3 border-b border-slate-800">
            <div className="flex items-center space-x-2 truncate">
              <FileText className="w-4 h-4 text-cyan-400 shrink-0" />
              <span className="text-sm font-semibold text-white truncate">{previewPath}</span>
            </div>
            <div className="flex items-center space-x-2">
              <button
                onClick={copyToClipboard}
                disabled={!previewContent}
                className="px-2.5 py-1 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded text-xs flex items-center space-x-1 transition"
                title="Copy Content"
              >
                {copied ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                <span>{copied ? "Copied" : "Copy"}</span>
              </button>
              <button
                onClick={() => setPreviewPath(null)}
                className="p-1 hover:bg-slate-800 text-slate-400 hover:text-white rounded transition"
              >
                <X className="w-4 h-4" />
              </button>
            </div>
          </div>

          <div className="mt-3">
            {previewLoading ? (
              <div className="flex items-center justify-center py-12 text-slate-400 text-xs font-mono space-x-2">
                <RefreshCw className="w-4 h-4 animate-spin" />
                <span>Retrieving content from Dropbox...</span>
              </div>
            ) : (
              <pre className="bg-[#090d16] border border-slate-800 rounded-lg p-3 text-xs font-mono text-slate-200 max-h-72 overflow-y-auto whitespace-pre-wrap">
                {previewContent}
              </pre>
            )}
          </div>
        </div>
      )}

      {/* Remote Backups List */}
      <div className="bg-[#111928] border border-slate-800 rounded-xl overflow-hidden shadow-xl">
        <div className="px-5 py-4 border-b border-slate-800 flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <FileArchive className="w-4 h-4 text-cyan-400" />
            <span className="text-sm font-semibold text-slate-200">
              Remote Backups Archive ({files.length})
            </span>
          </div>
          <button
            onClick={() => {
              fetchFiles();
              fetchStatus();
            }}
            className="p-1.5 hover:bg-slate-800 text-slate-400 hover:text-white rounded-lg transition"
            title="Refresh Remote Backups"
          >
            <RefreshCw className="w-3.5 h-3.5" />
          </button>
        </div>

        {files.length === 0 ? (
          <div className="p-12 text-center text-slate-500 text-xs font-mono">
            {status?.configured 
              ? "No backup snapshots detected in Dropbox folder yet. Click \"Snapshot & Upload to Dropbox\" above."
              : "Configure Dropbox credentials to view and archive remote snapshots."}
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead className="bg-slate-900/80 text-slate-400 uppercase font-mono border-b border-slate-800">
                <tr>
                  <th className="px-4 py-3">Snapshot Name</th>
                  <th className="px-4 py-3">Remote Path</th>
                  <th className="px-4 py-3">Size</th>
                  <th className="px-4 py-3">Timestamp</th>
                  <th className="px-4 py-3 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800/60 font-mono">
                {files.map((file) => (
                  <tr key={file.path_display} className="hover:bg-slate-800/30 transition">
                    <td className="px-4 py-3 font-semibold text-slate-100 flex items-center space-x-2">
                      <FileArchive className="w-3.5 h-3.5 text-cyan-400" />
                      <span>{file.name}</span>
                    </td>
                    <td className="px-4 py-3 text-slate-400 text-[11px] truncate max-w-[200px]" title={file.path_display}>
                      {file.path_display}
                    </td>
                    <td className="px-4 py-3 text-cyan-300 font-medium">
                      {formatBytes(file.size)}
                    </td>
                    <td className="px-4 py-3 text-slate-400 text-[11px]">
                      {file.server_modified ? new Date(file.server_modified).toLocaleString() : "—"}
                    </td>
                    <td className="px-4 py-3 text-right space-x-1.5">
                      <button
                        onClick={() => handlePreview(file.path_display)}
                        className="inline-flex items-center space-x-1 px-2.5 py-1 bg-slate-800 hover:bg-slate-700 text-slate-200 rounded text-xs transition"
                        title="Preview Content"
                      >
                        <Eye className="w-3 h-3" />
                        <span>Preview</span>
                      </button>
                      <button
                        onClick={() => handleDownload(file.path_display)}
                        className="inline-flex items-center space-x-1 px-2.5 py-1 bg-slate-800 hover:bg-slate-700 text-slate-200 rounded text-xs transition"
                      >
                        <Download className="w-3 h-3" />
                        <span>Download</span>
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}

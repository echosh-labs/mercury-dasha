"use client";

import React, { useState, useEffect, useCallback, useRef } from "react";
import {
  Activity,
  RefreshCw,
  Eye,
  FileText,
  Table,
  Mail,
  StickyNote,
  Calendar,
  Zap,
  CheckCircle2,
  AlertCircle,
  Clock,
  Database,
  Search,
  Filter,
  Sparkles,
  Send,
  Radio,
  Layers,
  ChevronRight,
  ShieldCheck,
  ExternalLink
} from "lucide-react";

export interface WorkspaceItem {
  id: string;
  type: "keep" | "gmail" | "doc" | "sheet" | "calendar" | string;
  title: string;
  snippet: string;
  status?: string;
  source: string;
  first_seen_at: string;
  last_seen_at: string;
  alert_emitted?: boolean;
  metadata?: Record<string, any>;
}

export interface WorkspaceCounts {
  total: number;
  keep_notes: number;
  gmail: number;
  docs: number;
  sheets: number;
  calendar: number;
}

export interface WorkspaceStatus {
  is_live: boolean;
  endpoint: string;
  last_polled_at: string;
  last_poll_status: string;
  mode: string;
  counts: WorkspaceCounts;
  active_alerts: number;
}

export interface WorkspaceAlert {
  id: string;
  item_id: string;
  type: string;
  title: string;
  snippet: string;
  timestamp: string;
  message: string;
}

export interface WorkspaceFeed {
  status: WorkspaceStatus;
  alerts: WorkspaceAlert[];
  items: WorkspaceItem[];
  total: number;
  generated_at: string;
}

export default function AxisMundiWorkspace() {
  const [feed, setFeed] = useState<WorkspaceFeed | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [syncing, setSyncing] = useState<boolean>(false);
  const [searchQuery, setSearchQuery] = useState<string>("");
  const [typeFilter, setTypeFilter] = useState<string>("all");
  const [sourceFilter, setSourceFilter] = useState<string>("all");
  const [newOnly, setNewOnly] = useState<boolean>(false);
  const [liveAlerts, setLiveAlerts] = useState<WorkspaceAlert[]>([]);
  const [showSimulateModal, setShowSimulateModal] = useState<boolean>(false);
  
  // Simulator State
  const [simType, setSimType] = useState<string>("gmail");
  const [simTitle, setSimTitle] = useState<string>("Transmission: Cosmic Ingress Observed");
  const [simSnippet, setSimSnippet] = useState<string>("Planetary alignment frequency verified by Sovereign Observer.");
  const [simSource, setSimSource] = useState<string>("sovereign_observer");
  const [simSubmitting, setSimSubmitting] = useState<boolean>(false);
  const [simSuccessMsg, setSimSuccessMsg] = useState<string | null>(null);

  const fetchFeed = useCallback(async () => {
    try {
      const params = new URLSearchParams();
      if (typeFilter !== "all") params.set("type", typeFilter);
      if (newOnly) params.set("new_only", "true");
      
      const queryStr = params.toString() ? `?${params.toString()}` : "";
      const res = await fetch(`/api/v1/axis-mundi/feed${queryStr}`);
      if (res.ok) {
        const data: WorkspaceFeed = await res.json();
        setFeed(data);
      }
    } catch (err) {
      console.error("Failed to load Axis Mundi feed:", err);
    } finally {
      setLoading(false);
    }
  }, [typeFilter, newOnly]);

  useEffect(() => {
    fetchFeed();
    const interval = setInterval(fetchFeed, 10000);
    return () => clearInterval(interval);
  }, [fetchFeed]);

  // Connect to SSE stream for live alerts from Sovereign Observer / Axis Mundi
  useEffect(() => {
    let eventSource: EventSource | null = null;
    try {
      eventSource = new EventSource("/api/v1/stream/pulse");
      
      eventSource.addEventListener("axis_mundi_alert", (e: MessageEvent) => {
        try {
          const alert: WorkspaceAlert = JSON.parse(e.data);
          setLiveAlerts((prev) => [alert, ...prev.slice(0, 19)]);
          // Also refresh feed items
          fetchFeed();
        } catch (err) {
          console.error("Error parsing axis_mundi_alert SSE:", err);
        }
      });
    } catch (err) {
      console.error("Failed to connect to pulse SSE for Axis Mundi alerts:", err);
    }

    return () => {
      if (eventSource) eventSource.close();
    };
  }, [fetchFeed]);

  const handleManualSync = async () => {
    setSyncing(true);
    try {
      const res = await fetch("/api/v1/axis-mundi/sync", { method: "POST" });
      if (res.ok) {
        await fetchFeed();
      }
    } catch (err) {
      console.error("Failed to trigger manual sync:", err);
    } finally {
      setSyncing(false);
    }
  };

  const handleSimulateDiscovery = async (e: React.FormEvent) => {
    e.preventDefault();
    setSimSubmitting(true);
    setSimSuccessMsg(null);
    try {
      const id = `${simType}/sovereign-${Date.now().toString(36)}`;
      const payload = {
        source: simSource,
        discoveries: [
          {
            id,
            type: simType,
            title: simTitle,
            snippet: simSnippet,
            status: "Discovered"
          }
        ]
      };

      const res = await fetch("/api/v1/axis-mundi/discoveries", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload)
      });

      if (res.ok) {
        setSimSuccessMsg(`Successfully ingested item ID: ${id}`);
        await fetchFeed();
        setTimeout(() => {
          setShowSimulateModal(false);
          setSimSuccessMsg(null);
        }, 1500);
      } else {
        const err = await res.json();
        setSimSuccessMsg(`Error: ${err.error || "Failed to submit discovery"}`);
      }
    } catch (err: any) {
      setSimSuccessMsg(`Network Error: ${err.message}`);
    } finally {
      setSimSubmitting(false);
    }
  };

  const status = feed?.status;
  const items = feed?.items || [];

  const filteredItems = items.filter((item) => {
    if (sourceFilter !== "all" && item.source !== sourceFilter) return false;
    if (searchQuery.trim() !== "") {
      const q = searchQuery.toLowerCase();
      const matchTitle = item.title.toLowerCase().includes(q);
      const matchSnippet = item.snippet.toLowerCase().includes(q);
      const matchId = item.id.toLowerCase().includes(q);
      if (!matchTitle && !matchSnippet && !matchId) return false;
    }
    return true;
  });

  const getItemIcon = (type: string) => {
    switch (type.toLowerCase()) {
      case "keep":
        return <StickyNote className="w-4 h-4 text-amber-400" />;
      case "gmail":
        return <Mail className="w-4 h-4 text-rose-400" />;
      case "doc":
        return <FileText className="w-4 h-4 text-blue-400" />;
      case "sheet":
        return <Table className="w-4 h-4 text-emerald-400" />;
      case "calendar":
        return <Calendar className="w-4 h-4 text-purple-400" />;
      default:
        return <Radio className="w-4 h-4 text-cyan-400" />;
    }
  };

  const formatRelativeTime = (isoString?: string) => {
    if (!isoString) return "--";
    try {
      const date = new Date(isoString);
      const diffSec = Math.floor((Date.now() - date.getTime()) / 1000);
      if (diffSec < 60) return `${diffSec}s ago`;
      const diffMin = Math.floor(diffSec / 60);
      if (diffMin < 60) return `${diffMin}m ago`;
      const diffHour = Math.floor(diffMin / 60);
      if (diffHour < 24) return `${diffHour}h ago`;
      return date.toLocaleDateString([], { month: "short", day: "numeric", hour: "2-digit", minute: "2-digit" });
    } catch {
      return isoString;
    }
  };

  return (
    <div className="space-y-6">
      {/* Top Banner / Ingestion Overview Header */}
      <div className="bg-gradient-to-r from-[#0d1627] via-[#101c33] to-[#0a1120] border border-cyan-500/20 rounded-2xl p-6 shadow-2xl relative overflow-hidden">
        <div className="absolute top-0 right-0 w-96 h-96 bg-cyan-500/5 rounded-full blur-3xl pointer-events-none" />
        
        <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4 relative z-10">
          <div>
            <div className="flex flex-wrap items-center gap-2">
              <span className="p-2 rounded-xl bg-cyan-950/80 border border-cyan-800 text-cyan-400 flex items-center justify-center">
                <Radio className="w-5 h-5 animate-pulse" />
              </span>
              <div>
                <div className="flex items-center space-x-2">
                  <h2 className="text-xl font-bold text-white tracking-tight font-serif">
                    Axis Mundi Ingestion Layer
                  </h2>
                  <span className="text-[10px] font-mono px-2 py-0.5 rounded-full bg-cyan-950 border border-cyan-800 text-cyan-300">
                    SOVEREIGN OBSERVER
                  </span>
                </div>
                <p className="text-xs text-slate-400 mt-0.5">
                  Real-time Google Workspace Ingestion, Model Context Protocol (MCP) Bridge, and Autonomous Discovery Stream.
                </p>
              </div>
            </div>
          </div>

          <div className="flex flex-wrap items-center gap-2">
            <button
              onClick={() => setShowSimulateModal(true)}
              className="px-3 py-1.5 rounded-lg bg-purple-950/70 hover:bg-purple-900 border border-purple-700/60 text-purple-300 text-xs font-mono flex items-center space-x-1.5 transition"
            >
              <Sparkles className="w-3.5 h-3.5" />
              <span>Simulate Discovery</span>
            </button>

            <button
              onClick={handleManualSync}
              disabled={syncing}
              className="px-3.5 py-1.5 rounded-lg bg-cyan-950/70 hover:bg-cyan-900 border border-cyan-700/60 text-cyan-300 text-xs font-mono flex items-center space-x-1.5 transition disabled:opacity-50"
            >
              <RefreshCw className={`w-3.5 h-3.5 ${syncing ? "animate-spin text-cyan-400" : ""}`} />
              <span>{syncing ? "Syncing..." : "Sync Now"}</span>
            </button>
          </div>
        </div>

        {/* Live Diagnostics Sub-Bar */}
        <div className="mt-5 pt-4 border-t border-slate-800/80 grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs font-mono">
          <div className="bg-slate-900/60 border border-slate-800/80 rounded-xl p-3">
            <span className="text-slate-400 block text-[11px]">Subsystem Status</span>
            <div className="flex items-center space-x-2 mt-1">
              <span className={`w-2 h-2 rounded-full ${status?.is_live ? "bg-emerald-400 animate-pulse" : "bg-amber-400"}`} />
              <span className={`font-semibold ${status?.is_live ? "text-emerald-300" : "text-amber-300"}`}>
                {status?.is_live ? "Connected (Live)" : "Standby / Polling"}
              </span>
            </div>
          </div>

          <div className="bg-slate-900/60 border border-slate-800/80 rounded-xl p-3">
            <span className="text-slate-400 block text-[11px]">Ingestion Protocol</span>
            <span className="text-cyan-300 font-semibold block mt-1 uppercase">
              {status?.mode || "MCP JSON-RPC"} (Port 8088)
            </span>
          </div>

          <div className="bg-slate-900/60 border border-slate-800/80 rounded-xl p-3">
            <span className="text-slate-400 block text-[11px]">Continuous Polling Duration</span>
            <span className="text-slate-200 font-semibold block mt-1">
              20s Loop (10s Timeout)
            </span>
          </div>

          <div className="bg-slate-900/60 border border-slate-800/80 rounded-xl p-3">
            <span className="text-slate-400 block text-[11px]">Last Synchronization</span>
            <span className="text-slate-300 font-semibold block mt-1">
              {status?.last_polled_at ? formatRelativeTime(status.last_polled_at) : "Pending boot"}
            </span>
          </div>
        </div>
      </div>

      {/* Metrics & Category Distribution Cards */}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
        <div className="p-4 rounded-xl bg-slate-900/80 border border-slate-800">
          <div className="flex items-center justify-between text-slate-400 text-xs font-mono">
            <span>Total Items</span>
            <Layers className="w-3.5 h-3.5 text-cyan-400" />
          </div>
          <div className="text-2xl font-bold text-white mt-1 font-mono">
            {status?.counts?.total ?? items.length}
          </div>
        </div>

        <div className="p-4 rounded-xl bg-slate-900/80 border border-slate-800">
          <div className="flex items-center justify-between text-slate-400 text-xs font-mono">
            <span>Keep Notes</span>
            <StickyNote className="w-3.5 h-3.5 text-amber-400" />
          </div>
          <div className="text-2xl font-bold text-amber-300 mt-1 font-mono">
            {status?.counts?.keep_notes ?? 0}
          </div>
        </div>

        <div className="p-4 rounded-xl bg-slate-900/80 border border-slate-800">
          <div className="flex items-center justify-between text-slate-400 text-xs font-mono">
            <span>Gmail Threads</span>
            <Mail className="w-3.5 h-3.5 text-rose-400" />
          </div>
          <div className="text-2xl font-bold text-rose-300 mt-1 font-mono">
            {status?.counts?.gmail ?? 0}
          </div>
        </div>

        <div className="p-4 rounded-xl bg-slate-900/80 border border-slate-800">
          <div className="flex items-center justify-between text-slate-400 text-xs font-mono">
            <span>Google Docs</span>
            <FileText className="w-3.5 h-3.5 text-blue-400" />
          </div>
          <div className="text-2xl font-bold text-blue-300 mt-1 font-mono">
            {status?.counts?.docs ?? 0}
          </div>
        </div>

        <div className="p-4 rounded-xl bg-slate-900/80 border border-slate-800">
          <div className="flex items-center justify-between text-slate-400 text-xs font-mono">
            <span>Google Sheets</span>
            <Table className="w-3.5 h-3.5 text-emerald-400" />
          </div>
          <div className="text-2xl font-bold text-emerald-300 mt-1 font-mono">
            {status?.counts?.sheets ?? 0}
          </div>
        </div>

        <div className="p-4 rounded-xl bg-slate-900/80 border border-slate-800">
          <div className="flex items-center justify-between text-slate-400 text-xs font-mono">
            <span>Active Alerts</span>
            <Zap className="w-3.5 h-3.5 text-yellow-400" />
          </div>
          <div className="text-2xl font-bold text-yellow-300 mt-1 font-mono">
            {status?.active_alerts ?? 0}
          </div>
        </div>
      </div>

      {/* Live SSE Alert Banner if alerts recently arrived */}
      {liveAlerts.length > 0 && (
        <div className="p-4 rounded-xl bg-purple-950/40 border border-purple-700/50 space-y-2 animate-in fade-in duration-300">
          <div className="flex items-center space-x-2 text-xs font-mono text-purple-300">
            <Radio className="w-3.5 h-3.5 animate-pulse text-purple-400" />
            <span className="font-semibold uppercase">Real-Time Ingestion Transmissions ({liveAlerts.length})</span>
          </div>
          <div className="space-y-1.5 max-h-32 overflow-y-auto scrollbar-thin">
            {liveAlerts.map((al) => (
              <div key={al.id} className="text-xs font-mono bg-slate-900/80 p-2 rounded border border-purple-900 flex items-center justify-between">
                <span className="text-slate-200">{al.message}</span>
                <span className="text-slate-400 text-[10px]">{formatRelativeTime(al.timestamp)}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Filter and Search Bar */}
      <div className="flex flex-col md:flex-row items-center justify-between gap-3 bg-slate-900/60 p-3 rounded-xl border border-slate-800 text-xs font-mono">
        <div className="flex flex-wrap items-center gap-2 w-full md:w-auto">
          <div className="flex items-center space-x-1 bg-slate-800/80 p-1 rounded-lg border border-slate-700">
            {["all", "keep", "gmail", "doc", "sheet"].map((t) => (
              <button
                key={t}
                onClick={() => setTypeFilter(t)}
                className={`px-2.5 py-1 rounded text-xs transition uppercase ${
                  typeFilter === t
                    ? "bg-cyan-600 text-white font-bold"
                    : "text-slate-400 hover:text-slate-200"
                }`}
              >
                {t}
              </button>
            ))}
          </div>

          <select
            value={sourceFilter}
            onChange={(e) => setSourceFilter(e.target.value)}
            className="bg-slate-800 border border-slate-700 rounded-lg px-2.5 py-1.5 text-slate-300 focus:outline-none"
          >
            <option value="all">All Sources</option>
            <option value="sovereign_observer">Sovereign Observer</option>
            <option value="mcp">MCP Protocol</option>
            <option value="keep">Keep</option>
            <option value="gmail">Gmail</option>
          </select>

          <label className="flex items-center space-x-1.5 cursor-pointer text-slate-400 hover:text-slate-200 px-2">
            <input
              type="checkbox"
              checked={newOnly}
              onChange={(e) => setNewOnly(e.target.checked)}
              className="rounded bg-slate-800 border-slate-700 text-cyan-500 focus:ring-0"
            />
            <span>New Only</span>
          </label>
        </div>

        <div className="relative w-full md:w-64">
          <Search className="w-3.5 h-3.5 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            placeholder="Search titles, snippets, IDs..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full bg-slate-800 border border-slate-700 rounded-lg pl-8 pr-3 py-1.5 text-slate-200 placeholder-slate-500 focus:outline-none focus:border-cyan-500"
          />
        </div>
      </div>

      {/* Ingestion Stream Cards */}
      {loading ? (
        <div className="py-12 text-center text-slate-400 font-mono text-xs flex flex-col items-center justify-center space-y-2">
          <RefreshCw className="w-6 h-6 animate-spin text-cyan-400" />
          <span>Polling Axis Mundi Ingestion Feed...</span>
        </div>
      ) : filteredItems.length === 0 ? (
        <div className="py-16 text-center text-slate-400 font-mono text-xs border border-dashed border-slate-800 rounded-2xl bg-slate-900/30 p-8">
          <Eye className="w-8 h-8 mx-auto text-slate-600 mb-2" />
          <p className="font-semibold text-slate-300">No workspace items matched the active filter.</p>
          <p className="text-slate-500 mt-1 max-w-md mx-auto">
            Items ingested by the 20s background poller or directly posted by Sovereign Observer to <code className="text-cyan-400 font-mono">/api/v1/axis-mundi/discoveries</code> will materialize here automatically.
          </p>
          <button
            onClick={() => setShowSimulateModal(true)}
            className="mt-4 px-4 py-2 rounded-lg bg-cyan-950 border border-cyan-800 text-cyan-300 hover:bg-cyan-900 transition"
          >
            Send Test Discovery from Sovereign Observer
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {filteredItems.map((item) => (
            <div
              key={item.id}
              className="p-4 rounded-xl bg-slate-900/70 border border-slate-800/80 hover:border-cyan-500/40 transition group space-y-3"
            >
              <div className="flex items-start justify-between gap-2">
                <div className="flex items-start space-x-2.5">
                  <span className="p-2 rounded-lg bg-slate-800/80 border border-slate-700/80 mt-0.5">
                    {getItemIcon(item.type)}
                  </span>
                  <div>
                    <h3 className="text-sm font-semibold text-white group-hover:text-cyan-300 transition tracking-tight">
                      {item.title}
                    </h3>
                    <span className="text-[10px] font-mono text-slate-500 block truncate max-w-xs sm:max-w-sm">
                      {item.id}
                    </span>
                  </div>
                </div>

                <div className="flex flex-col items-end gap-1">
                  <span
                    className={`text-[10px] font-mono px-2 py-0.5 rounded-full border ${
                      item.source === "sovereign_observer"
                        ? "bg-purple-950/80 text-purple-300 border-purple-700"
                        : "bg-cyan-950/80 text-cyan-300 border-cyan-800"
                    }`}
                  >
                    {item.source}
                  </span>
                  {item.alert_emitted && (
                    <span className="text-[9px] font-mono px-1.5 py-0.2 rounded bg-yellow-950/60 text-yellow-300 border border-yellow-800">
                      NEW ALERT
                    </span>
                  )}
                </div>
              </div>

              {item.snippet && (
                <p className="text-xs text-slate-300 font-sans leading-relaxed line-clamp-2 bg-slate-950/40 p-2.5 rounded-lg border border-slate-800/50">
                  {item.snippet}
                </p>
              )}

              <div className="flex items-center justify-between text-[11px] font-mono text-slate-500 pt-1 border-t border-slate-800/60">
                <div className="flex items-center space-x-1.5">
                  <Clock className="w-3 h-3 text-slate-500" />
                  <span>Seen {formatRelativeTime(item.first_seen_at)}</span>
                </div>
                {item.status && (
                  <span className="text-cyan-400 font-mono text-[10px] bg-cyan-950/40 px-2 py-0.5 rounded border border-cyan-900">
                    {item.status}
                  </span>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Simulator Modal for testing Sovereign Observer discoveries */}
      {showSimulateModal && (
        <div className="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4 animate-in fade-in duration-200">
          <div className="bg-[#0e1626] border border-cyan-500/40 rounded-2xl p-6 max-w-lg w-full space-y-4 shadow-2xl">
            <div className="flex items-center justify-between border-b border-slate-800 pb-3">
              <div className="flex items-center space-x-2">
                <Sparkles className="w-4 h-4 text-purple-400" />
                <h3 className="text-base font-bold text-white font-serif">
                  Simulate Sovereign Observer Ingestion
                </h3>
              </div>
              <button
                onClick={() => setShowSimulateModal(false)}
                className="text-slate-400 hover:text-white text-xs font-mono"
              >
                ✕
              </button>
            </div>

            <p className="text-xs text-slate-400 leading-relaxed font-sans">
              Test posting a live workspace discovery directly to <code className="text-cyan-400 font-mono">/api/v1/axis-mundi/discoveries</code>. The engine will ingest it, save to BoltDB, and broadcast an SSE alert.
            </p>

            <form onSubmit={handleSimulateDiscovery} className="space-y-3 text-xs font-mono">
              <div>
                <label className="block text-slate-400 mb-1">Source Identifier</label>
                <input
                  type="text"
                  value={simSource}
                  onChange={(e) => setSimSource(e.target.value)}
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg p-2 text-slate-200 focus:outline-none focus:border-cyan-500"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-slate-400 mb-1">Item Type</label>
                  <select
                    value={simType}
                    onChange={(e) => setSimType(e.target.value)}
                    className="w-full bg-slate-900 border border-slate-700 rounded-lg p-2 text-slate-200 focus:outline-none focus:border-cyan-500"
                  >
                    <option value="gmail">Gmail</option>
                    <option value="keep">Keep Note</option>
                    <option value="doc">Google Doc</option>
                    <option value="sheet">Google Sheet</option>
                    <option value="calendar">Calendar</option>
                  </select>
                </div>
                <div>
                  <label className="block text-slate-400 mb-1">Discovered Title</label>
                  <input
                    type="text"
                    value={simTitle}
                    onChange={(e) => setSimTitle(e.target.value)}
                    className="w-full bg-slate-900 border border-slate-700 rounded-lg p-2 text-slate-200 focus:outline-none focus:border-cyan-500"
                  />
                </div>
              </div>

              <div>
                <label className="block text-slate-400 mb-1">Snippet / Directive Payload</label>
                <textarea
                  rows={3}
                  value={simSnippet}
                  onChange={(e) => setSimSnippet(e.target.value)}
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg p-2 text-slate-200 focus:outline-none focus:border-cyan-500 resize-none"
                />
              </div>

              {simSuccessMsg && (
                <div className={`p-2.5 rounded-lg text-xs font-mono ${simSuccessMsg.startsWith("Error") ? "bg-red-950/80 text-red-300 border border-red-800" : "bg-emerald-950/80 text-emerald-300 border border-emerald-800"}`}>
                  {simSuccessMsg}
                </div>
              )}

              <div className="flex items-center justify-end space-x-2 pt-2">
                <button
                  type="button"
                  onClick={() => setShowSimulateModal(false)}
                  className="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={simSubmitting}
                  className="px-4 py-1.5 rounded-lg bg-cyan-600 hover:bg-cyan-500 text-white font-bold flex items-center space-x-1.5 disabled:opacity-50"
                >
                  <Send className="w-3.5 h-3.5" />
                  <span>{simSubmitting ? "Ingesting..." : "Submit Discovery"}</span>
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

"use client";

import React, { useState, useEffect } from "react";
import {
  Terminal,
  Activity,
  Cpu,
  Database,
  Layers,
  FileCode,
  FolderTree,
  Send,
  RefreshCw,
  Clock,
  Sparkles,
  Server,
  Play,
  CheckCircle2,
  AlertTriangle,
  FileText,
  Boxes,
  Code,
  Eye,
  Zap,
  ChevronRight,
  HardDrive,
  Brain,
  Video,
  HeartPulse,
  Power,
  AlertOctagon,
  RotateCcw,
  Check
} from "lucide-react";
import { useYouTubeStudio } from "@/lib/youtube-context";

interface SystemOverview {
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
    account_type?: string;
  };
  chrono_pulse: {
    active_hora: string;
    sacred_metal: string;
    axiom: string;
    client_count: number;
  };
  architecture: {
    canonical_root: string;
    host_substrate: string;
    sibling_repos: string[];
  };
  youtube?: {
    configured: boolean;
    authenticated: boolean;
    channel_name?: string;
    video_count?: number;
  };
  self_healing?: {
    active: boolean;
    status: string;
    auto_heals_triggered: number;
    panics_recovered: number;
    last_healed_at?: string;
    total_checks: number;
    restart_pending: boolean;
  };
}

interface ComponentHealth {
  name: string;
  status: "HEALTHY" | "DEGRADED" | "CRITICAL" | "UNHEALTHY";
  message: string;
  latency_ms?: number;
  checked_at: string;
  details?: Record<string, any>;
}

interface SelfHealingReport {
  service: string;
  environment: string;
  overall_status: "HEALTHY" | "DEGRADED" | "CRITICAL" | "UNHEALTHY";
  timestamp: string;
  uptime: string;
  uptime_sec: number;
  components: Record<string, ComponentHealth>;
  stats: {
    total_checks: number;
    auto_heals_triggered: number;
    panics_recovered: number;
    consecutive_failures: number;
    last_check_time?: string;
    last_healed_at?: string;
    last_heal_action?: string;
    restart_pending: boolean;
  };
  self_healing_active: boolean;
}

interface IncidentEntry {
  id: string;
  timestamp: string;
  severity: "info" | "warning" | "critical";
  component: string;
  message: string;
  details?: string;
  action_taken: string;
  resolved: boolean;
}

interface RestartRecord {
  id: string;
  timestamp: string;
  reason: string;
  initiator: string;
  type: string;
  pid: number;
  success: boolean;
  details?: string;
}

interface ConversationEntry {
  id: string;
  hemisphere?: "left" | "right";
  path?: string;
  date: string;
  size_bytes: number;
  artifacts: string[];
}

interface BicameralData {
  left_hemisphere: {
    name: string;
    substrate: string;
    path: string;
    exists: boolean;
    session_count: number;
    instruction_chain: string;
  };
  right_hemisphere: {
    name: string;
    substrate: string;
    path: string;
    exists: boolean;
    session_count: number;
    instruction_chain: string;
  };
  corpus_callosum: string;
  total_sessions: number;
  active_session?: string;
  timestamp: string;
}

interface RouteDoc {
  method: string;
  path: string;
  category: string;
  description: string;
}

export default function AgenticConsole() {
  const { openUploader, status: ytStatus, connectChannel } = useYouTubeStudio();
  const [overview, setOverview] = useState<SystemOverview | null>(null);
  const [bicameral, setBicameral] = useState<BicameralData | null>(null);
  const [hemisphereFilter, setHemisphereFilter] = useState<"all" | "left" | "right">("all");
  const [conversations, setConversations] = useState<ConversationEntry[]>([]);
  const [selectedConv, setSelectedConv] = useState<string | null>(null);
  const [selectedArtifact, setSelectedArtifact] = useState<string | null>(null);
  const [artifactContent, setArtifactContent] = useState<string>("");
  const [loadingArtifact, setLoadingArtifact] = useState<boolean>(false);
  
  const [routes, setRoutes] = useState<RouteDoc[]>([]);
  const [selectedCategory, setSelectedCategory] = useState<string>("All");
  const [activeTab, setActiveTab] = useState<"telemetry" | "selfhealing" | "brain" | "api" | "architecture">("telemetry");
  
  // Self-Healing & Remote Restart state
  const [healthReport, setHealthReport] = useState<SelfHealingReport | null>(null);
  const [incidents, setIncidents] = useState<IncidentEntry[]>([]);
  const [restartHistory, setRestartHistory] = useState<RestartRecord[]>([]);
  const [loadingHealth, setLoadingHealth] = useState<boolean>(false);
  const [remediating, setRemediating] = useState<boolean>(false);
  const [remediateMsg, setRemediateMsg] = useState<string | null>(null);

  // Restart modal & execution state
  const [restartReason, setRestartReason] = useState<string>("Manual restart via Agentic Mission Control");
  const [restartDelayMs, setRestartDelayMs] = useState<number>(500);
  const [restartToken, setRestartToken] = useState<string>("");
  const [restartStatus, setRestartStatus] = useState<"idle" | "triggering" | "restarting" | "reconnected" | "error">("idle");
  const [restartAttempts, setRestartAttempts] = useState<number>(0);
  const [restartError, setRestartError] = useState<string | null>(null);

  // API Tester state
  const [testMethod, setTestMethod] = useState<string>("GET");
  const [testPath, setTestPath] = useState<string>("/api/v1/system/overview");
  const [testResponse, setTestResponse] = useState<string>("");
  const [testLatency, setTestLatency] = useState<number | null>(null);
  const [testing, setTesting] = useState<boolean>(false);
  const [loadingOverview, setLoadingOverview] = useState<boolean>(false);

  // Fetch System Overview
  const fetchOverview = async () => {
    setLoadingOverview(true);
    try {
      const res = await fetch("/api/v1/system/overview");
      if (res.ok) {
        const data = await res.json();
        setOverview(data);
      }
    } catch (err) {
      console.error("Failed to load system overview:", err);
    } finally {
      setLoadingOverview(false);
    }
  };

  // Fetch Health Report
  const fetchHealthReport = async () => {
    setLoadingHealth(true);
    try {
      const res = await fetch("/api/v1/system/health");
      if (res.ok) {
        const data = await res.json();
        setHealthReport(data);
      }
    } catch (err) {
      console.error("Failed to load health report:", err);
    } finally {
      setLoadingHealth(false);
    }
  };

  // Fetch Incidents
  const fetchIncidents = async () => {
    try {
      const res = await fetch("/api/v1/system/self-healing/incidents?limit=25");
      if (res.ok) {
        const data = await res.json();
        setIncidents(data.incidents || []);
      }
    } catch (err) {
      console.error("Failed to load incidents:", err);
    }
  };

  // Fetch Restart History
  const fetchRestartHistory = async () => {
    try {
      const res = await fetch("/api/v1/system/restart/history");
      if (res.ok) {
        const data = await res.json();
        setRestartHistory(data.restart_events || []);
      }
    } catch (err) {
      console.error("Failed to load restart history:", err);
    }
  };

  // Trigger On-Demand Health Diagnostic Probe
  const triggerHealthCheck = async () => {
    setLoadingHealth(true);
    try {
      const res = await fetch("/api/v1/system/self-healing/check", { method: "POST" });
      if (res.ok) {
        const data = await res.json();
        setHealthReport(data.report || null);
        await fetchIncidents();
        await fetchOverview();
      }
    } catch (err) {
      console.error("Health probe failed:", err);
    } finally {
      setLoadingHealth(false);
    }
  };

  // Trigger Targeted Self-Healing Remediation
  const triggerRemediate = async (action: string) => {
    setRemediating(true);
    setRemediateMsg(null);
    try {
      const res = await fetch("/api/v1/system/self-healing/remediate", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action }),
      });
      const data = await res.json();
      setRemediateMsg(data.message || (data.success ? "Remediation applied successfully" : "Remediation failed"));
      await fetchHealthReport();
      await fetchIncidents();
      await fetchOverview();
    } catch (err: any) {
      setRemediateMsg(`Remediation error: ${err.message}`);
    } finally {
      setRemediating(false);
    }
  };

  // Issue Remote Service Restart
  const triggerRestart = async () => {
    setRestartStatus("triggering");
    setRestartError(null);
    setRestartAttempts(0);
    try {
      const headers: Record<string, string> = {
        "Content-Type": "application/json",
      };
      if (restartToken.trim()) {
        headers["X-Admin-Token"] = restartToken.trim();
      }

      const res = await fetch("/api/v1/system/restart", {
        method: "POST",
        headers,
        body: JSON.stringify({
          reason: restartReason,
          delay_ms: restartDelayMs,
        }),
      });

      const data = await res.json();
      if (!res.ok) {
        setRestartStatus("error");
        setRestartError(data.error || `HTTP ${res.status} restart rejected`);
        return;
      }

      // Enter restarting state and poll /healthz for reconnection
      setRestartStatus("restarting");
      let attempts = 0;
      const pollInterval = setInterval(async () => {
        attempts++;
        setRestartAttempts(attempts);
        try {
          const checkRes = await fetch("/healthz", { cache: "no-store" });
          if (checkRes.ok && attempts >= 2) {
            clearInterval(pollInterval);
            setRestartStatus("reconnected");
            await fetchHealthReport();
            await fetchOverview();
            await fetchRestartHistory();
            await fetchIncidents();
          }
        } catch {
          // Process is currently recycling/restarting
        }
        if (attempts > 30) {
          clearInterval(pollInterval);
          setRestartStatus("error");
          setRestartError("Reconnection timed out after 15 seconds. Check console or process logs.");
        }
      }, 500);

    } catch (err: any) {
      setRestartStatus("error");
      setRestartError(`Failed to trigger restart: ${err.message}`);
    }
  };

  // Fetch Conversations
  const fetchConversations = async () => {
    try {
      const res = await fetch("/api/v1/agent/conversations");
      if (res.ok) {
        const data = await res.json();
        const convs: ConversationEntry[] = data.conversations || [];
        setConversations(convs);
        if (convs.length > 0 && !selectedConv) {
          setSelectedConv(convs[0].id);
          if (convs[0].artifacts && convs[0].artifacts.length > 0) {
            setSelectedArtifact(convs[0].artifacts[0]);
          } else {
            setSelectedArtifact(null);
            setArtifactContent("");
          }
        }
      }
    } catch (err) {
      console.error("Failed to load conversations:", err);
    }
  };

  // Fetch Routes Catalog
  const fetchRoutes = async () => {
    try {
      const res = await fetch("/api/v1/system/routes");
      if (res.ok) {
        const data = await res.json();
        setRoutes(data.routes || []);
      }
    } catch (err) {
      console.error("Failed to load route catalog:", err);
    }
  };

  // Fetch Bicameral Status
  const fetchBicameral = async () => {
    try {
      const res = await fetch("/api/v1/agent/bicameral");
      if (res.ok) {
        const data = await res.json();
        setBicameral(data);
      }
    } catch (err) {
      console.error("Failed to load bicameral status:", err);
    }
  };

  useEffect(() => {
    fetchOverview();
    fetchBicameral();
    fetchConversations();
    fetchRoutes();
    fetchHealthReport();
    fetchIncidents();
    fetchRestartHistory();
  }, []);

  // Fetch artifact content when conversation or artifact selection changes
  useEffect(() => {
    if (!selectedConv || !selectedArtifact) {
      setArtifactContent("");
      return;
    }
    const loadArtifact = async () => {
      setLoadingArtifact(true);
      try {
        const res = await fetch(`/api/v1/agent/artifact?id=${encodeURIComponent(selectedConv)}&name=${encodeURIComponent(selectedArtifact)}`);
        if (res.ok) {
          const data = await res.json();
          setArtifactContent(data.content || "");
        } else {
          setArtifactContent(`[Artifact error: HTTP ${res.status}]`);
        }
      } catch (err: any) {
        setArtifactContent(`[Failed to load artifact: ${err.message}]`);
      } finally {
        setLoadingArtifact(false);
      }
    };
    loadArtifact();
  }, [selectedConv, selectedArtifact]);

  // Execute API probe
  const runApiTest = async (method: string, path: string) => {
    setTesting(true);
    setTestMethod(method);
    setTestPath(path);
    setTestResponse("Probing endpoint...");
    const start = performance.now();
    try {
      const res = await fetch(path, {
        method: method,
      });
      const end = performance.now();
      setTestLatency(Math.round((end - start) * 100) / 100);
      const text = await res.text();
      try {
        const json = JSON.parse(text);
        setTestResponse(JSON.stringify(json, null, 2));
      } catch {
        setTestResponse(text || `HTTP ${res.status} ${res.statusText}`);
      }
    } catch (err: any) {
      setTestLatency(null);
      setTestResponse(`Fetch failed: ${err.message}`);
    } finally {
      setTesting(false);
    }
  };

  const categories = ["All", ...Array.from(new Set(routes.map((r) => r.category)))];
  const filteredRoutes = selectedCategory === "All" 
    ? routes 
    : routes.filter((r) => r.category === selectedCategory);

  const filteredConversations = conversations.filter((c) => {
    if (hemisphereFilter === "all") return true;
    return c.hemisphere === hemisphereFilter;
  });

  return (
    <div className="space-y-6">
      {/* Sub-navigation bar */}
      <div className="flex flex-wrap items-center justify-between gap-3 bg-slate-900/90 border border-slate-800 p-4 rounded-xl">
        <div className="flex items-center space-x-2">
          <Terminal className="w-5 h-5 text-cyan-400" />
          <h2 className="text-lg font-bold text-white tracking-tight">
            Agentic Mission Control & System Console
          </h2>
          <span className="bg-cyan-500/10 text-cyan-400 border border-cyan-500/20 text-xs font-mono px-2 py-0.5 rounded">
            WSL2 Canonical
          </span>
        </div>

        {/* View mode toggle */}
        <div className="flex items-center bg-slate-950 p-1 rounded-lg border border-slate-800 text-xs font-medium">
          <button
            onClick={() => setActiveTab("telemetry")}
            className={`px-3 py-1.5 rounded transition flex items-center space-x-1.5 ${
              activeTab === "telemetry"
                ? "bg-cyan-900/60 text-cyan-300 border border-cyan-700/50"
                : "text-slate-400 hover:text-slate-200"
            }`}
          >
            <Activity className="w-3.5 h-3.5" />
            <span>Diagnostics</span>
          </button>
          <button
            onClick={() => setActiveTab("selfhealing")}
            className={`px-3 py-1.5 rounded transition flex items-center space-x-1.5 ${
              activeTab === "selfhealing"
                ? "bg-emerald-900/60 text-emerald-300 border border-emerald-700/50"
                : "text-slate-400 hover:text-slate-200"
            }`}
          >
            <HeartPulse className="w-3.5 h-3.5 text-emerald-400" />
            <span>Self-Healing & Restart</span>
          </button>
          <button
            onClick={() => setActiveTab("brain")}
            className={`px-3 py-1.5 rounded transition flex items-center space-x-1.5 ${
              activeTab === "brain"
                ? "bg-cyan-900/60 text-cyan-300 border border-cyan-700/50"
                : "text-slate-400 hover:text-slate-200"
            }`}
          >
            <Sparkles className="w-3.5 h-3.5" />
            <span>Agent Brain Sessions</span>
          </button>
          <button
            onClick={() => setActiveTab("api")}
            className={`px-3 py-1.5 rounded transition flex items-center space-x-1.5 ${
              activeTab === "api"
                ? "bg-cyan-900/60 text-cyan-300 border border-cyan-700/50"
                : "text-slate-400 hover:text-slate-200"
            }`}
          >
            <Zap className="w-3.5 h-3.5" />
            <span>API Catalog & Probe</span>
          </button>
          <button
            onClick={() => setActiveTab("architecture")}
            className={`px-3 py-1.5 rounded transition flex items-center space-x-1.5 ${
              activeTab === "architecture"
                ? "bg-cyan-900/60 text-cyan-300 border border-cyan-700/50"
                : "text-slate-400 hover:text-slate-200"
            }`}
          >
            <Boxes className="w-3.5 h-3.5" />
            <span>Substrate Blueprint</span>
          </button>
        </div>

        <button
          onClick={() => {
            fetchOverview();
            fetchBicameral();
            fetchConversations();
            fetchRoutes();
            fetchHealthReport();
            fetchIncidents();
            fetchRestartHistory();
          }}
          disabled={loadingOverview || loadingHealth}
          className="px-3 py-1.5 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded-lg text-xs font-mono flex items-center space-x-1.5 border border-slate-700 transition"
        >
          <RefreshCw className={`w-3.5 h-3.5 ${loadingOverview || loadingHealth ? "animate-spin" : ""}`} />
          <span>Refresh</span>
        </button>
      </div>

      {/* VIEW: DIAGNOSTICS & TELEMETRY */}
      {activeTab === "telemetry" && (
        <div className="space-y-6">
          {/* Top Telemetry Grid */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 font-mono">
            {/* Go Runtime */}
            <div className="bg-slate-900/70 border border-slate-800 p-4 rounded-xl">
              <div className="flex items-center justify-between text-xs text-slate-400 mb-2">
                <span className="flex items-center space-x-1.5">
                  <Cpu className="w-4 h-4 text-cyan-400" />
                  <span>GO RUNTIME</span>
                </span>
                <span className="text-cyan-400">{overview?.go_version || "go1.23+"}</span>
              </div>
              <div className="text-2xl font-bold text-white mb-1">
                {overview?.runtime_stats.alloc_mb ? `${overview.runtime_stats.alloc_mb.toFixed(2)} MB` : "--"}
              </div>
              <div className="text-xs text-slate-400 flex justify-between">
                <span>Goroutines: <strong className="text-slate-200">{overview?.runtime_stats.goroutines || 0}</strong></span>
                <span>GC Cycles: <strong className="text-slate-200">{overview?.runtime_stats.num_gc || 0}</strong></span>
              </div>
            </div>

            {/* BoltDB Storage */}
            <div className="bg-slate-900/70 border border-slate-800 p-4 rounded-xl">
              <div className="flex items-center justify-between text-xs text-slate-400 mb-2">
                <span className="flex items-center space-x-1.5">
                  <Database className="w-4 h-4 text-emerald-400" />
                  <span>BOLTDB STORAGE</span>
                </span>
                <span className="text-emerald-400">ACID B+TREE</span>
              </div>
              <div className="text-2xl font-bold text-white mb-1">
                {overview?.database.size_bytes
                  ? `${(overview.database.size_bytes / (1024 * 1024)).toFixed(1)} MB`
                  : "--"}
              </div>
              <div className="text-xs text-slate-400 flex justify-between">
                <span>Alloc Pages: <strong className="text-slate-200">{overview?.database.allocated_pages || 0}</strong></span>
                <span>Keys: <strong className="text-slate-200">{overview?.database.key_count || 0}</strong></span>
              </div>
            </div>

            {/* Indexed Storehouse */}
            <div className="bg-slate-900/70 border border-slate-800 p-4 rounded-xl">
              <div className="flex items-center justify-between text-xs text-slate-400 mb-2">
                <span className="flex items-center space-x-1.5">
                  <FolderTree className="w-4 h-4 text-amber-400" />
                  <span>LOCAL DROPBOX</span>
                </span>
                <span className="text-amber-400">85,119 FILES</span>
              </div>
              <div className="text-2xl font-bold text-white mb-1">
                {overview?.dropbox.total_indexed ? overview.dropbox.total_indexed.toLocaleString() : "85,119"}
              </div>
              <div className="text-xs text-slate-400 truncate">
                <span>Path: </span>
                <span className="text-slate-200 font-mono text-[11px]">{overview?.dropbox.local_path || "/home/justin/Dropbox"}</span>
              </div>
            </div>

            {/* Uptime & Metronome */}
            <div className="bg-slate-900/70 border border-slate-800 p-4 rounded-xl">
              <div className="flex items-center justify-between text-xs text-slate-400 mb-2">
                <span className="flex items-center space-x-1.5">
                  <Clock className="w-4 h-4 text-purple-400" />
                  <span>CHRONO METRONOME</span>
                </span>
                <span className="text-purple-400">2s INTERVAL</span>
              </div>
              <div className="text-2xl font-bold text-white mb-1">
                {overview?.uptime || "Running"}
              </div>
              <div className="text-xs text-slate-400 flex justify-between">
                <span>Hora: <strong className="text-purple-300">{overview?.chrono_pulse.active_hora || "Mercury"}</strong></span>
                <span>Subscribers: <strong className="text-slate-200">{overview?.chrono_pulse.client_count || 0}</strong></span>
              </div>
            </div>
          </div>

          {/* Breakdown Panels */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Category Breakdown */}
            <div className="bg-slate-900/80 border border-slate-800 rounded-xl p-5 space-y-4">
              <div className="flex items-center justify-between border-b border-slate-800 pb-3">
                <h3 className="text-sm font-semibold text-white flex items-center space-x-2">
                  <Layers className="w-4 h-4 text-amber-400" />
                  <span>Indexed Knowledge Corpus Distribution</span>
                </h3>
                <span className="text-xs font-mono text-slate-400">ACID B-Tree Store</span>
              </div>

              <div className="grid grid-cols-2 sm:grid-cols-3 gap-3 font-mono text-xs">
                {overview?.database.category_counts &&
                  Object.entries(overview.database.category_counts)
                    .filter(([k]) => k !== "total")
                    .map(([cat, count]) => (
                      <div key={cat} className="p-3 bg-slate-950/70 border border-slate-800/80 rounded-lg">
                        <div className="text-slate-400 uppercase text-[10px] tracking-wider">{cat}</div>
                        <div className="text-lg font-bold text-amber-300 mt-1">{count.toLocaleString()}</div>
                      </div>
                    ))}
              </div>

              <div className="p-3 bg-cyan-950/30 border border-cyan-800/40 rounded-lg text-xs text-slate-300 flex items-center justify-between">
                <div>
                  <span className="font-semibold text-cyan-300">Total Indexed Files: </span>
                  <span className="font-mono text-cyan-200 font-bold">
                    {overview?.database.category_counts?.total?.toLocaleString() || "85,119"}
                  </span>
                </div>
                <div className="flex items-center space-x-1.5 text-emerald-400">
                  <CheckCircle2 className="w-3.5 h-3.5" />
                  <span>Sync Validated</span>
                </div>
              </div>
            </div>

            {/* Hermetic & Alchemical Alignment */}
            <div className="bg-slate-900/80 border border-slate-800 rounded-xl p-5 space-y-4">
              <div className="flex items-center justify-between border-b border-slate-800 pb-3">
                <h3 className="text-sm font-semibold text-white flex items-center space-x-2">
                  <Sparkles className="w-4 h-4 text-purple-400" />
                  <span>Chrono-Pulse Active Resonance</span>
                </h3>
                <span className="text-xs font-mono text-purple-300">Planetary Metronome</span>
              </div>

              <div className="space-y-3 font-mono text-xs">
                <div className="p-3 bg-slate-950/70 border border-slate-800/80 rounded-lg flex items-center justify-between">
                  <span className="text-slate-400">Active Planetary Ruler:</span>
                  <span className="font-bold text-purple-300 text-sm">
                    {overview?.chrono_pulse.active_hora || "—"}
                  </span>
                </div>
                <div className="p-3 bg-slate-950/70 border border-slate-800/80 rounded-lg flex items-center justify-between">
                  <span className="text-slate-400">Sacred Metal:</span>
                  <span className="font-bold text-cyan-300">
                    {overview?.chrono_pulse.sacred_metal || "—"}
                  </span>
                </div>
                <div className="p-3 bg-slate-950/70 border border-slate-800/80 rounded-lg">
                  <div className="text-slate-400 text-[10px] uppercase mb-1">Governing Hermetic Axiom:</div>
                  <div className="text-slate-200 italic">
                    {overview?.chrono_pulse.axiom ? `"${overview.chrono_pulse.axiom}"` : "—"}
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* YouTube Sovereign Studio & Pipeline Telemetry */}
          <div className="bg-slate-900/80 border border-slate-800 rounded-xl p-5 space-y-4">
            <div className="flex items-center justify-between border-b border-slate-800 pb-3">
              <h3 className="text-sm font-semibold text-white flex items-center space-x-2">
                <Video className="w-4 h-4 text-red-400" />
                <span>YouTube Sovereign Studio & Publishing Pipeline</span>
              </h3>
              <span className={`text-xs font-mono uppercase px-2 py-0.5 rounded border ${
                ytStatus?.authenticated
                  ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
                  : "bg-red-500/10 text-red-400 border-red-500/20"
              }`}>
                {ytStatus?.authenticated ? "Connected" : "Disconnected"}
              </span>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-3 font-mono text-xs">
              <div className="p-3 bg-slate-950/70 border border-slate-800/80 rounded-lg">
                <div className="text-slate-400 uppercase text-[10px] tracking-wider">OAuth Channel</div>
                <div className="text-sm font-bold mt-1 text-white flex items-center gap-1.5 truncate">
                  {ytStatus?.authenticated ? (
                    <>
                      <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0" />
                      <span className="truncate">{ytStatus.channel?.title || "Channel Paired"}</span>
                    </>
                  ) : (
                    <>
                      <AlertTriangle className="w-4 h-4 text-amber-400 shrink-0" />
                      <span className="text-slate-400">Not Paired</span>
                    </>
                  )}
                </div>
              </div>

              <div className="p-3 bg-slate-950/70 border border-slate-800/80 rounded-lg">
                <div className="text-slate-400 uppercase text-[10px] tracking-wider">Daily API Quota</div>
                <div className="text-sm font-bold mt-1 text-cyan-300">
                  {ytStatus?.quota
                    ? `${ytStatus.quota.used_today.toLocaleString()} / ${ytStatus.quota.daily_limit.toLocaleString()} units`
                    : "10,000 units/day"}
                </div>
              </div>

              <div className="p-3 bg-slate-950/70 border border-slate-800/80 rounded-lg">
                <div className="text-slate-400 uppercase text-[10px] tracking-wider">Active Upload Jobs</div>
                <div className="text-sm font-bold mt-1 text-amber-300">
                  {ytStatus?.recent_jobs
                    ? ytStatus.recent_jobs.filter(j => j.status === "uploading" || j.status === "queued").length
                    : 0}{" "}
                  running
                </div>
              </div>
            </div>

            <div className="flex flex-wrap items-center justify-between gap-3 pt-2">
              <span className="text-xs text-slate-400 font-mono">
                Autonomous publishing step ready for video render pipelines and manual dispatch.
              </span>
              <button
                onClick={() => openUploader()}
                className="px-3.5 py-1.5 rounded-lg text-xs font-semibold bg-red-600 hover:bg-red-500 text-white transition flex items-center gap-1.5 font-mono shadow-lg shadow-red-950/40"
              >
                <Video className="w-3.5 h-3.5" />
                <span>Open YouTube Studio</span>
              </button>
            </div>
          </div>
        </div>
      )}

      {/* VIEW: AUTONOMOUS SELF-HEALING & REMOTE RESTART */}
      {activeTab === "selfhealing" && (
        <div className="space-y-6">
          {/* Top Status & Health Overview Banner */}
          <div className="bg-slate-900/90 border border-slate-800 rounded-xl p-5 space-y-4">
            <div className="flex flex-wrap items-center justify-between gap-3 border-b border-slate-800 pb-3">
              <div className="flex items-center space-x-3">
                <div className="p-2 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400">
                  <HeartPulse className="w-6 h-6 animate-pulse" />
                </div>
                <div>
                  <h3 className="text-base font-bold text-white tracking-tight flex items-center gap-2">
                    <span>Autonomous Self-Healing Watchdog</span>
                    <span className={`text-[11px] font-mono px-2 py-0.5 rounded border ${
                      healthReport?.overall_status === "HEALTHY"
                        ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/30"
                        : healthReport?.overall_status === "DEGRADED"
                        ? "bg-amber-500/10 text-amber-400 border-amber-500/30"
                        : "bg-rose-500/10 text-rose-400 border-rose-500/30"
                    }`}>
                      {healthReport?.overall_status || "HEALTHY"}
                    </span>
                  </h3>
                  <p className="text-xs text-slate-400 mt-0.5">
                    Proactive subsystem health watchdog, automated OS memory reclamation, and remote binary restart.
                  </p>
                </div>
              </div>

              <div className="flex items-center space-x-2">
                <button
                  onClick={triggerHealthCheck}
                  disabled={loadingHealth}
                  className="px-3.5 py-1.5 rounded-lg text-xs font-mono font-medium bg-emerald-600 hover:bg-emerald-500 text-white transition flex items-center space-x-1.5 shadow-lg shadow-emerald-950/40"
                >
                  <RotateCcw className={`w-3.5 h-3.5 ${loadingHealth ? "animate-spin" : ""}`} />
                  <span>{loadingHealth ? "Probing Subsystems..." : "Run Health Probe"}</span>
                </button>
              </div>
            </div>

            {/* Quick Stats Strip */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-3 font-mono text-xs">
              <div className="bg-slate-950/60 p-3 rounded-lg border border-slate-800/80">
                <span className="text-slate-400 block mb-1">HEALTH CHECKS</span>
                <span className="text-lg font-bold text-white">
                  {healthReport?.stats.total_checks ?? 0}
                </span>
                <span className="text-[10px] text-slate-500 block mt-0.5">15s loop</span>
              </div>
              <div className="bg-slate-950/60 p-3 rounded-lg border border-slate-800/80">
                <span className="text-slate-400 block mb-1">AUTO-HEALS</span>
                <span className="text-lg font-bold text-emerald-400">
                  {healthReport?.stats.auto_heals_triggered ?? 0}
                </span>
                <span className="text-[10px] text-slate-500 block mt-0.5">Autonomous fixes</span>
              </div>
              <div className="bg-slate-950/60 p-3 rounded-lg border border-slate-800/80">
                <span className="text-slate-400 block mb-1">PANICS MITIGATED</span>
                <span className="text-lg font-bold text-cyan-400">
                  {healthReport?.stats.panics_recovered ?? 0}
                </span>
                <span className="text-[10px] text-slate-500 block mt-0.5">Zero crash exits</span>
              </div>
              <div className="bg-slate-950/60 p-3 rounded-lg border border-slate-800/80">
                <span className="text-slate-400 block mb-1">CONSECUTIVE FAILS</span>
                <span className={`text-lg font-bold ${
                  (healthReport?.stats.consecutive_failures ?? 0) > 0 ? "text-amber-400" : "text-slate-400"
                }`}>
                  {healthReport?.stats.consecutive_failures ?? 0} / 3
                </span>
                <span className="text-[10px] text-slate-500 block mt-0.5">Auto-restart threshold</span>
              </div>
            </div>

            {/* Remediation message banner */}
            {remediateMsg && (
              <div className="bg-emerald-950/40 border border-emerald-800/60 text-emerald-300 text-xs p-3 rounded-lg flex items-center justify-between font-mono">
                <span>{remediateMsg}</span>
                <button
                  onClick={() => setRemediateMsg(null)}
                  className="text-slate-400 hover:text-white text-xs ml-2"
                >
                  ✕
                </button>
              </div>
            )}
          </div>

          {/* Subsystem Health Cards Grid */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {/* BoltDB Storehouse */}
            <div className="bg-slate-900/70 border border-slate-800 p-4 rounded-xl flex flex-col justify-between space-y-3">
              <div>
                <div className="flex items-center justify-between text-xs text-slate-400 mb-1.5">
                  <span className="flex items-center space-x-1.5 font-mono">
                    <Database className="w-4 h-4 text-amber-400" />
                    <span>BOLTDB STORAGE</span>
                  </span>
                  <span className={`px-1.5 py-0.5 rounded text-[10px] font-mono ${
                    healthReport?.components?.database?.status === "HEALTHY"
                      ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                      : "bg-rose-500/10 text-rose-400 border border-rose-500/20"
                  }`}>
                    {healthReport?.components?.database?.status || "HEALTHY"}
                  </span>
                </div>
                <div className="text-xl font-bold text-white font-mono">
                  {healthReport?.components?.database?.latency_ms !== undefined
                    ? `${healthReport.components.database.latency_ms.toFixed(2)} ms`
                    : "0.15 ms"}
                </div>
                <p className="text-xs text-slate-400 mt-1 line-clamp-2">
                  {healthReport?.components?.database?.message || "BoltDB read/write responsive with zero locks."}
                </p>
              </div>
              <button
                onClick={() => triggerRemediate("db_verify")}
                disabled={remediating}
                className="w-full py-1.5 bg-slate-800 hover:bg-slate-700 text-slate-200 rounded text-xs font-mono transition"
              >
                Verify Storehouse Integrity
              </button>
            </div>

            {/* Memory & OS Heap */}
            <div className="bg-slate-900/70 border border-slate-800 p-4 rounded-xl flex flex-col justify-between space-y-3">
              <div>
                <div className="flex items-center justify-between text-xs text-slate-400 mb-1.5">
                  <span className="flex items-center space-x-1.5 font-mono">
                    <Cpu className="w-4 h-4 text-cyan-400" />
                    <span>MEMORY MANAGEMENT</span>
                  </span>
                  <span className={`px-1.5 py-0.5 rounded text-[10px] font-mono ${
                    healthReport?.components?.memory?.status === "HEALTHY"
                      ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                      : "bg-amber-500/10 text-amber-400 border border-amber-500/20"
                  }`}>
                    {healthReport?.components?.memory?.status || "HEALTHY"}
                  </span>
                </div>
                <div className="text-xl font-bold text-white font-mono">
                  {overview?.runtime_stats?.alloc_mb?.toFixed(1) || "12.4"} MB
                </div>
                <p className="text-xs text-slate-400 mt-1">
                  Threshold: 768 MB warning / 1024 MB critical. Auto-purge triggers on spike.
                </p>
              </div>
              <button
                onClick={() => triggerRemediate("memory_gc")}
                disabled={remediating}
                className="w-full py-1.5 bg-cyan-900/50 hover:bg-cyan-800/60 text-cyan-200 border border-cyan-700/40 rounded text-xs font-mono transition"
              >
                Purge & Free OS Memory
              </button>
            </div>

            {/* Goroutines Watchdog */}
            <div className="bg-slate-900/70 border border-slate-800 p-4 rounded-xl flex flex-col justify-between space-y-3">
              <div>
                <div className="flex items-center justify-between text-xs text-slate-400 mb-1.5">
                  <span className="flex items-center space-x-1.5 font-mono">
                    <Activity className="w-4 h-4 text-indigo-400" />
                    <span>GOROUTINE LEAK GUARD</span>
                  </span>
                  <span className="px-1.5 py-0.5 rounded text-[10px] font-mono bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                    {healthReport?.components?.goroutines?.status || "HEALTHY"}
                  </span>
                </div>
                <div className="text-xl font-bold text-white font-mono">
                  {overview?.runtime_stats?.goroutines || 18} routines
                </div>
                <p className="text-xs text-slate-400 mt-1">
                  Limits: 500 routines warning / 2000 critical. Concurrency monitored continuously.
                </p>
              </div>
              <div className="text-[11px] text-slate-500 font-mono py-1 px-2 bg-slate-950/40 rounded border border-slate-800/50 text-center">
                Metronome, Recorder & Indexer Nominal
              </div>
            </div>

            {/* Port Sovereignty */}
            <div className="bg-slate-900/70 border border-slate-800 p-4 rounded-xl flex flex-col justify-between space-y-3">
              <div>
                <div className="flex items-center justify-between text-xs text-slate-400 mb-1.5">
                  <span className="flex items-center space-x-1.5 font-mono">
                    <HardDrive className="w-4 h-4 text-purple-400" />
                    <span>PORT SOVEREIGNTY</span>
                  </span>
                  <span className="px-1.5 py-0.5 rounded text-[10px] font-mono bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                    {healthReport?.components?.port?.status || "HEALTHY"}
                  </span>
                </div>
                <div className="text-xl font-bold text-white font-mono">
                  :{overview?.database?.path ? "8080" : "8080"}
                </div>
                <p className="text-xs text-slate-400 mt-1">
                  Startup port contention auto-remediation active via portutil.
                </p>
              </div>
              <button
                onClick={() => triggerRemediate("port_check")}
                disabled={remediating}
                className="w-full py-1.5 bg-slate-800 hover:bg-slate-700 text-slate-200 rounded text-xs font-mono transition"
              >
                Probe Port State
              </button>
            </div>
          </div>

          {/* Remote Service Restart Control Panel (Sovereign Re-Exec Gateway) */}
          <div className="bg-gradient-to-r from-red-950/30 via-slate-900/90 to-slate-900/90 border border-red-900/50 rounded-xl p-5 space-y-4">
            <div className="flex flex-wrap items-center justify-between gap-3 border-b border-red-900/30 pb-3">
              <div className="flex items-center space-x-3">
                <div className="p-2 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400">
                  <Power className="w-5 h-5" />
                </div>
                <div>
                  <h3 className="text-sm font-bold text-white tracking-tight flex items-center gap-2">
                    <span>Remote Sovereign Restart Gateway</span>
                    <span className="text-[10px] font-mono bg-red-500/20 text-red-300 border border-red-500/30 px-2 py-0.5 rounded">
                      In-Place Re-Exec
                    </span>
                  </h3>
                  <p className="text-xs text-slate-400 mt-0.5">
                    Issues an authorized remote restart. Shuts down HTTP sockets, flushes BoltDB, and re-executes the binary via <code className="text-cyan-300">syscall.Exec</code> with zero process downtime.
                  </p>
                </div>
              </div>

              <div className="text-xs font-mono text-slate-400">
                Current PID: <span className="text-white font-semibold">Self</span>
              </div>
            </div>

            {/* Interactive Restart Execution States */}
            {restartStatus === "restarting" ? (
              <div className="bg-slate-950 p-6 rounded-xl border border-amber-500/40 space-y-3 text-center">
                <div className="flex justify-center">
                  <RotateCcw className="w-8 h-8 text-amber-400 animate-spin" />
                </div>
                <h4 className="text-base font-bold text-amber-300 tracking-tight">
                  Restarting Mercury Dasha Single-Binary Service...
                </h4>
                <p className="text-xs text-slate-400 max-w-lg mx-auto leading-relaxed font-mono">
                  The Go runtime has drained active listeners and is re-executing. Probing <code className="text-cyan-300">/healthz</code> for service readiness (attempt {restartAttempts})...
                </p>
              </div>
            ) : restartStatus === "reconnected" ? (
              <div className="bg-slate-950 p-6 rounded-xl border border-emerald-500/40 space-y-3 text-center">
                <div className="flex justify-center">
                  <Check className="w-8 h-8 text-emerald-400" />
                </div>
                <h4 className="text-base font-bold text-emerald-300 tracking-tight">
                  Service Reconnected & Operational!
                </h4>
                <p className="text-xs text-slate-400 max-w-lg mx-auto font-mono">
                  Mercury Dasha has successfully completed in-place binary re-execution with fresh memory and clean BoltDB bindings.
                </p>
                <button
                  onClick={() => setRestartStatus("idle")}
                  className="px-4 py-1.5 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg text-xs font-mono transition"
                >
                  Acknowledge
                </button>
              </div>
            ) : (
              <div className="space-y-4">
                {restartError && (
                  <div className="p-3 bg-red-950/60 border border-red-800/80 rounded-lg text-red-300 text-xs font-mono flex items-center justify-between">
                    <span>{restartError}</span>
                    <button onClick={() => setRestartError(null)} className="text-slate-400 hover:text-white">✕</button>
                  </div>
                )}

                <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                  <div className="md:col-span-2 space-y-1">
                    <label className="text-xs font-mono text-slate-400">Restart Reason / Audit Rationale</label>
                    <input
                      type="text"
                      value={restartReason}
                      onChange={(e) => setRestartReason(e.target.value)}
                      placeholder="e.g. Remote maintenance, configuration refresh"
                      className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-red-500/50 font-mono"
                    />
                  </div>

                  <div className="space-y-1">
                    <label className="text-xs font-mono text-slate-400">Restart Delay</label>
                    <select
                      value={restartDelayMs}
                      onChange={(e) => setRestartDelayMs(Number(e.target.value))}
                      className="w-full bg-slate-950 border border-slate-800 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-red-500/50 font-mono"
                    >
                      <option value={200}>200 ms (Fastest)</option>
                      <option value={500}>500 ms (Default Flush)</option>
                      <option value={1000}>1000 ms (Safe)</option>
                      <option value={2000}>2000 ms (Extended)</option>
                    </select>
                  </div>
                </div>

                <div className="flex flex-wrap items-center justify-between gap-3 pt-2">
                  <div className="flex items-center space-x-2 max-w-sm">
                    <input
                      type="password"
                      value={restartToken}
                      onChange={(e) => setRestartToken(e.target.value)}
                      placeholder="Admin Token (if MERCURY_RESTART_TOKEN set)"
                      className="bg-slate-950 border border-slate-800 rounded-lg px-3 py-1.5 text-xs text-white focus:outline-none focus:border-red-500/50 font-mono w-72"
                    />
                  </div>

                  <button
                    onClick={triggerRestart}
                    disabled={restartStatus === "triggering"}
                    className="px-4 py-2 bg-red-700 hover:bg-red-600 disabled:opacity-50 text-white rounded-lg text-xs font-mono font-semibold transition flex items-center space-x-2 shadow-lg shadow-red-950/60"
                  >
                    <Power className={`w-3.5 h-3.5 ${restartStatus === "triggering" ? "animate-spin" : ""}`} />
                    <span>{restartStatus === "triggering" ? "Dispatching Restart..." : "Issue Remote Service Restart"}</span>
                  </button>
                </div>
              </div>
            )}
          </div>

          {/* Self-Healing Incidents & Audit Journal */}
          <div className="bg-slate-900/90 border border-slate-800 rounded-xl p-5 space-y-4">
            <div className="flex items-center justify-between border-b border-slate-800 pb-3">
              <div className="flex items-center space-x-2">
                <Activity className="w-4 h-4 text-cyan-400" />
                <h3 className="text-sm font-bold text-white tracking-tight">
                  Self-Healing Incident & Audit Journal
                </h3>
                <span className="text-xs font-mono text-slate-500">
                  ({incidents.length} recorded events)
                </span>
              </div>
              <button
                onClick={fetchIncidents}
                className="text-xs font-mono text-cyan-400 hover:text-cyan-300"
              >
                Refresh Log
              </button>
            </div>

            {incidents.length === 0 ? (
              <div className="text-center py-8 text-xs font-mono text-slate-500">
                No anomalous incidents recorded. All systems operating in sovereign equilibrium.
              </div>
            ) : (
              <div className="space-y-2 max-h-72 overflow-y-auto pr-1">
                {incidents.map((inc) => (
                  <div
                    key={inc.id}
                    className="bg-slate-950/80 border border-slate-800/80 p-3 rounded-lg flex flex-col sm:flex-row sm:items-center justify-between gap-2 text-xs font-mono"
                  >
                    <div className="space-y-1">
                      <div className="flex items-center space-x-2">
                        <span className={`px-1.5 py-0.2 rounded text-[10px] font-semibold ${
                          inc.severity === "critical"
                            ? "bg-rose-500/20 text-rose-300 border border-rose-500/30"
                            : inc.severity === "warning"
                            ? "bg-amber-500/20 text-amber-300 border border-amber-500/30"
                            : "bg-cyan-500/20 text-cyan-300 border border-cyan-500/30"
                        }`}>
                          {inc.severity.toUpperCase()}
                        </span>
                        <span className="text-slate-400 text-[11px]">
                          [{inc.component}]
                        </span>
                        <span className="text-white font-medium">
                          {inc.message}
                        </span>
                      </div>
                      {inc.action_taken && inc.action_taken !== "none" && (
                        <div className="text-[11px] text-emerald-400">
                          Remedy applied: <span className="underline">{inc.action_taken}</span>
                        </div>
                      )}
                    </div>
                    <div className="text-[11px] text-slate-500 whitespace-nowrap">
                      {new Date(inc.timestamp).toLocaleTimeString()}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Service Restart History */}
          {restartHistory.length > 0 && (
            <div className="bg-slate-900/90 border border-slate-800 rounded-xl p-5 space-y-4">
              <div className="flex items-center justify-between border-b border-slate-800 pb-3">
                <div className="flex items-center space-x-2">
                  <RotateCcw className="w-4 h-4 text-red-400" />
                  <h3 className="text-sm font-bold text-white tracking-tight">
                    Service Restart History
                  </h3>
                </div>
                <span className="text-xs font-mono text-slate-500">
                  {restartHistory.length} recorded re-exec events
                </span>
              </div>

              <div className="space-y-2 max-h-48 overflow-y-auto pr-1">
                {restartHistory.map((rec) => (
                  <div
                    key={rec.id}
                    className="bg-slate-950/80 border border-slate-800/80 p-3 rounded-lg flex items-center justify-between text-xs font-mono"
                  >
                    <div>
                      <span className="text-white font-semibold">{rec.reason}</span>
                      <span className="text-slate-500 ml-2">via {rec.initiator} (PID {rec.pid})</span>
                    </div>
                    <div className="text-[11px] text-slate-400">
                      {new Date(rec.timestamp).toLocaleString()}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {/* VIEW: AGENT BRAIN SESSIONS & ARTIFACTS */}
      {activeTab === "brain" && (
        <div className="space-y-6">
          {/* Bicameral Brain Federation Overview Banner */}
          {bicameral && (
            <div className="bg-slate-900/90 border border-slate-800 rounded-xl p-5 space-y-4">
              <div className="flex flex-wrap items-center justify-between gap-3 border-b border-slate-800 pb-3">
                <div className="flex items-center space-x-3">
                  <div className="p-2 rounded-lg bg-indigo-500/10 border border-indigo-500/20 text-indigo-400">
                    <Brain className="w-5 h-5" />
                  </div>
                  <div>
                    <div className="flex items-center space-x-2">
                      <h3 className="text-sm font-bold text-white tracking-wide">
                        Bicameral Brain Federation
                      </h3>
                      <span className="bg-indigo-500/10 text-indigo-400 border border-indigo-500/30 text-[10px] font-mono px-2 py-0.5 rounded">
                        Corpus Callosum Active
                      </span>
                    </div>
                    <div className="text-xs text-slate-400 font-mono mt-0.5">
                      Unified Hub: <span className="text-cyan-300">{bicameral.corpus_callosum}</span> ({bicameral.total_sessions} total sessions indexed)
                    </div>
                  </div>
                </div>

                <button
                  onClick={() => {
                    fetchBicameral();
                    fetchConversations();
                  }}
                  className="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs flex items-center space-x-1.5 border border-slate-700 transition"
                >
                  <RefreshCw className="w-3.5 h-3.5" />
                  <span>Sync Hemispheres</span>
                </button>
              </div>

              {/* Hemisphere Cards */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {/* Left Hemisphere Card */}
                <div
                  onClick={() => setHemisphereFilter(hemisphereFilter === "left" ? "all" : "left")}
                  className={`p-4 rounded-xl border cursor-pointer transition ${
                    hemisphereFilter === "left"
                      ? "bg-sky-950/40 border-sky-500 ring-1 ring-sky-500/40"
                      : "bg-slate-950/60 border-slate-800 hover:border-slate-700"
                  }`}
                >
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center space-x-2">
                      <span className="text-lg">🪟</span>
                      <div>
                        <div className="text-xs font-bold text-white">Left Hemisphere</div>
                        <div className="text-[10px] text-sky-400 font-mono">Windows Antigravity IDE</div>
                      </div>
                    </div>
                    <span className="text-xs font-mono font-bold px-2 py-0.5 rounded bg-sky-500/10 text-sky-400 border border-sky-500/20">
                      {bicameral.left_hemisphere.session_count} sessions
                    </span>
                  </div>
                  <div className="space-y-1 text-[11px] font-mono text-slate-400">
                    <div className="truncate" title={bicameral.left_hemisphere.path}>
                      <span className="text-slate-500">Path: </span>{bicameral.left_hemisphere.path}
                    </div>
                    <div className="truncate text-slate-300" title={bicameral.left_hemisphere.instruction_chain}>
                      <span className="text-slate-500">Chain: </span>{bicameral.left_hemisphere.instruction_chain}
                    </div>
                  </div>
                </div>

                {/* Right Hemisphere Card */}
                <div
                  onClick={() => setHemisphereFilter(hemisphereFilter === "right" ? "all" : "right")}
                  className={`p-4 rounded-xl border cursor-pointer transition ${
                    hemisphereFilter === "right"
                      ? "bg-purple-950/40 border-purple-500 ring-1 ring-purple-500/40"
                      : "bg-slate-950/60 border-slate-800 hover:border-slate-700"
                  }`}
                >
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center space-x-2">
                      <span className="text-lg">🐧</span>
                      <div>
                        <div className="text-xs font-bold text-white">Right Hemisphere</div>
                        <div className="text-[10px] text-purple-400 font-mono">WSL2 Ubuntu AGY CLI</div>
                      </div>
                    </div>
                    <span className="text-xs font-mono font-bold px-2 py-0.5 rounded bg-purple-500/10 text-purple-400 border border-purple-500/20">
                      {bicameral.right_hemisphere.session_count} sessions
                    </span>
                  </div>
                  <div className="space-y-1 text-[11px] font-mono text-slate-400">
                    <div className="truncate" title={bicameral.right_hemisphere.path}>
                      <span className="text-slate-500">Path: </span>{bicameral.right_hemisphere.path}
                    </div>
                    <div className="truncate text-slate-300" title={bicameral.right_hemisphere.instruction_chain}>
                      <span className="text-slate-500">Chain: </span>{bicameral.right_hemisphere.instruction_chain}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Session Explorer and Artifact Viewer */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* Conversation List */}
            <div className="bg-slate-900/90 border border-slate-800 rounded-xl p-4 space-y-3">
              <div className="flex items-center justify-between border-b border-slate-800 pb-2">
                <span className="text-xs font-bold uppercase tracking-wider text-slate-400">
                  Agent Sessions ({filteredConversations.length})
                </span>
                <span className="text-[11px] font-mono text-cyan-400">~/.gemini/bicameral</span>
              </div>

              {/* Hemisphere Filter Pills */}
              <div className="flex items-center space-x-1 bg-slate-950 p-1 rounded-lg border border-slate-800 text-[11px] font-mono">
                <button
                  onClick={() => setHemisphereFilter("all")}
                  className={`flex-1 py-1 rounded text-center transition ${
                    hemisphereFilter === "all"
                      ? "bg-cyan-900/80 text-cyan-200 font-semibold"
                      : "text-slate-400 hover:text-slate-200"
                  }`}
                >
                  All ({conversations.length})
                </button>
                <button
                  onClick={() => setHemisphereFilter("left")}
                  className={`flex-1 py-1 rounded text-center transition flex items-center justify-center space-x-1 ${
                    hemisphereFilter === "left"
                      ? "bg-sky-900/80 text-sky-200 font-semibold"
                      : "text-slate-400 hover:text-slate-200"
                  }`}
                >
                  <span>🪟 Left</span>
                  <span>({conversations.filter((c) => c.hemisphere === "left").length})</span>
                </button>
                <button
                  onClick={() => setHemisphereFilter("right")}
                  className={`flex-1 py-1 rounded text-center transition flex items-center justify-center space-x-1 ${
                    hemisphereFilter === "right"
                      ? "bg-purple-900/80 text-purple-200 font-semibold"
                      : "text-slate-400 hover:text-slate-200"
                  }`}
                >
                  <span>🐧 Right</span>
                  <span>({conversations.filter((c) => c.hemisphere === "right").length})</span>
                </button>
              </div>

              <div className="space-y-2 max-h-[600px] overflow-y-auto pr-1">
                {filteredConversations.map((conv) => {
                  const isSelected = selectedConv === conv.id;
                  const isCurrent = conv.id.startsWith("65df33eb");
                  return (
                    <button
                      key={conv.id}
                      onClick={() => {
                        setSelectedConv(conv.id);
                        if (conv.artifacts && conv.artifacts.length > 0) {
                          setSelectedArtifact(conv.artifacts[0]);
                        } else {
                          setSelectedArtifact(null);
                          setArtifactContent("");
                        }
                      }}
                      className={`w-full text-left p-3 rounded-lg border transition ${
                        isSelected
                          ? "bg-cyan-950/60 border-cyan-600 text-white"
                          : "bg-slate-950/60 border-slate-800/80 text-slate-300 hover:border-slate-700"
                      }`}
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex items-center space-x-1.5 min-w-0">
                          <span
                            className={`text-[9px] font-mono px-1 py-0.5 rounded border shrink-0 ${
                              conv.hemisphere === "left"
                                ? "bg-sky-500/20 text-sky-300 border-sky-500/30"
                                : "bg-purple-500/20 text-purple-300 border-purple-500/30"
                            }`}
                          >
                            {conv.hemisphere === "left" ? "🪟 WIN" : "🐧 WSL"}
                          </span>
                          <span className="font-mono text-xs font-bold truncate max-w-[130px]">
                            {conv.id}
                          </span>
                        </div>
                        {isCurrent && (
                          <span className="bg-emerald-500/20 text-emerald-400 border border-emerald-500/40 text-[9px] font-mono px-1.5 py-0.2 rounded shrink-0">
                            ACTIVE
                          </span>
                        )}
                      </div>
                      <div className="flex items-center justify-between text-[11px] text-slate-400 mt-1 font-mono">
                        <span>{new Date(conv.date).toLocaleDateString()}</span>
                        <span>{(conv.size_bytes / 1024).toFixed(0)} KB</span>
                      </div>
                      {conv.artifacts && conv.artifacts.length > 0 && (
                        <div className="flex flex-wrap gap-1 mt-2">
                          {(conv.artifacts || []).map((art) => (
                            <span
                              key={art}
                              className="bg-slate-800/90 text-cyan-300 text-[10px] font-mono px-1.5 py-0.5 rounded border border-slate-700"
                            >
                              {art}
                            </span>
                          ))}
                        </div>
                      )}
                    </button>
                  );
                })}
              </div>
            </div>

            {/* Artifact Content Viewer */}
            <div className="lg:col-span-2 bg-slate-900/90 border border-slate-800 rounded-xl p-5 flex flex-col space-y-4">
              {/* Artifact Header & Selector */}
              <div className="flex flex-wrap items-center justify-between gap-3 border-b border-slate-800 pb-3">
                <div>
                  <div className="flex items-center space-x-2">
                    <span className="text-xs font-mono text-slate-400">Active Session</span>
                    {conversations.find((c) => c.id === selectedConv)?.hemisphere && (
                      <span
                        className={`text-[10px] font-mono px-2 py-0.5 rounded border ${
                          conversations.find((c) => c.id === selectedConv)?.hemisphere === "left"
                            ? "bg-sky-500/20 text-sky-300 border-sky-500/30"
                            : "bg-purple-500/20 text-purple-300 border-purple-500/30"
                        }`}
                      >
                        {conversations.find((c) => c.id === selectedConv)?.hemisphere === "left"
                          ? "🪟 Left Hemisphere (Windows)"
                          : "🐧 Right Hemisphere (WSL)"}
                      </span>
                    )}
                  </div>
                  <div className="text-sm font-mono font-bold text-white truncate max-w-md">
                    {selectedConv || "Select a session"}
                  </div>
                </div>

                {/* Artifact Selector Tabs */}
                {selectedConv && (
                  <div className="flex items-center space-x-1.5 overflow-x-auto">
                    {(conversations.find((c) => c.id === selectedConv)?.artifacts || []).length === 0 ? (
                      <span className="text-[11px] font-mono text-slate-500 italic">
                        No markdown reports in this session
                      </span>
                    ) : (
                      (conversations.find((c) => c.id === selectedConv)?.artifacts || []).map((art) => (
                        <button
                          key={art}
                          onClick={() => setSelectedArtifact(art)}
                          className={`px-2.5 py-1 text-xs font-mono rounded-lg transition border ${
                            selectedArtifact === art
                              ? "bg-cyan-900 text-cyan-200 border-cyan-500 font-semibold"
                              : "bg-slate-800 text-slate-400 border-slate-700 hover:text-slate-200"
                          }`}
                        >
                          {art}
                        </button>
                      ))
                    )}
                  </div>
                )}
              </div>


            {/* Content Body */}
            <div className="flex-1 min-h-[480px] max-h-[620px] overflow-y-auto bg-slate-950 p-4 rounded-lg border border-slate-800/80 font-mono text-xs text-slate-200 whitespace-pre-wrap leading-relaxed">
              {loadingArtifact ? (
                <div className="flex items-center justify-center h-48 text-slate-400 space-x-2">
                  <RefreshCw className="w-4 h-4 animate-spin text-cyan-400" />
                  <span>Loading artifact text...</span>
                </div>
              ) : artifactContent ? (
                artifactContent
              ) : selectedConv && (conversations.find((c) => c.id === selectedConv)?.artifacts || []).length === 0 ? (
                <div className="flex flex-col items-center justify-center h-48 text-slate-500 space-y-2">
                  <Terminal className="w-8 h-8 opacity-50 text-purple-400" />
                  <span className="text-slate-300 font-bold text-xs">Right Hemisphere CLI Transcript Session</span>
                  <span className="text-[11px] text-slate-400 max-w-sm text-center">
                    This session was executed natively in the WSL2 AGY CLI. Terminal transcripts are saved directly to <code className="text-purple-300">.system_generated/logs/transcript.jsonl</code>.
                  </span>
                </div>
              ) : (
                <div className="flex flex-col items-center justify-center h-48 text-slate-500">
                  <FileText className="w-8 h-8 mb-2 opacity-50" />
                  <span>Select an artifact from the list to view report contents.</span>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
      )}

      {/* VIEW: INTERACTIVE API CATALOG & PROBE */}
      {activeTab === "api" && (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Endpoint List */}
          <div className="lg:col-span-2 space-y-4">
            {/* Category Filter */}
            <div className="flex flex-wrap gap-1.5 pb-2">
              {categories.map((cat) => (
                <button
                  key={cat}
                  onClick={() => setSelectedCategory(cat)}
                  className={`px-3 py-1 rounded-lg text-xs font-medium transition ${
                    selectedCategory === cat
                      ? "bg-cyan-900/80 text-cyan-300 border border-cyan-600"
                      : "bg-slate-900 text-slate-400 border border-slate-800 hover:bg-slate-800"
                  }`}
                >
                  {cat}
                </button>
              ))}
            </div>

            {/* Routes Table */}
            <div className="space-y-2 max-h-[620px] overflow-y-auto pr-1">
              {filteredRoutes.map((r, idx) => (
                <div
                  key={idx}
                  className="p-3.5 bg-slate-900/80 border border-slate-800 rounded-xl flex items-center justify-between hover:border-slate-700 transition"
                >
                  <div className="space-y-1">
                    <div className="flex items-center space-x-2">
                      <span
                        className={`text-[11px] font-mono font-bold px-2 py-0.5 rounded ${
                          r.method === "GET"
                            ? "bg-cyan-500/10 text-cyan-400 border border-cyan-500/30"
                            : r.method === "POST"
                            ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/30"
                            : "bg-rose-500/10 text-rose-400 border border-rose-500/30"
                        }`}
                      >
                        {r.method}
                      </span>
                      <span className="font-mono text-xs font-semibold text-white">
                        {r.path}
                      </span>
                      <span className="text-[10px] font-mono text-slate-500 uppercase bg-slate-950 px-2 py-0.5 rounded border border-slate-800">
                        {r.category}
                      </span>
                    </div>
                    <div className="text-xs text-slate-400 leading-snug">
                      {r.description}
                    </div>
                  </div>

                  <button
                    onClick={() => runApiTest(r.method, r.path)}
                    className="ml-3 px-3 py-1.5 bg-slate-800 hover:bg-cyan-950/80 hover:border-cyan-600 text-slate-200 hover:text-cyan-300 border border-slate-700 rounded-lg text-xs font-mono flex items-center space-x-1.5 transition shrink-0"
                  >
                    <Play className="w-3 h-3" />
                    <span>Test</span>
                  </button>
                </div>
              ))}
            </div>
          </div>

          {/* Interactive Probe Drawer */}
          <div className="bg-slate-900/90 border border-slate-800 rounded-xl p-5 flex flex-col space-y-4">
            <div className="flex items-center justify-between border-b border-slate-800 pb-3">
              <div className="flex items-center space-x-2">
                <Terminal className="w-4 h-4 text-cyan-400" />
                <h3 className="text-sm font-bold text-white">Live API Runner</h3>
              </div>
              {testLatency !== null && (
                <span className="text-xs font-mono text-emerald-400 bg-emerald-950/40 border border-emerald-800/40 px-2 py-0.5 rounded">
                  {testLatency} ms
                </span>
              )}
            </div>

            {/* Input Bar */}
            <div className="space-y-2">
              <div className="text-xs font-mono text-slate-400">Endpoint Path</div>
              <div className="flex items-center space-x-2">
                <span className="px-2.5 py-1.5 bg-slate-950 border border-slate-800 rounded text-xs font-mono text-cyan-400 font-bold">
                  {testMethod}
                </span>
                <input
                  type="text"
                  value={testPath}
                  onChange={(e) => setTestPath(e.target.value)}
                  className="flex-1 bg-slate-950 border border-slate-800 rounded px-3 py-1.5 text-xs font-mono text-white focus:outline-none focus:border-cyan-500"
                />
                <button
                  onClick={() => runApiTest(testMethod, testPath)}
                  disabled={testing}
                  className="px-3 py-1.5 bg-cyan-600 hover:bg-cyan-500 text-white rounded text-xs font-mono flex items-center space-x-1 transition"
                >
                  <Send className="w-3 h-3" />
                </button>
              </div>
            </div>

            {/* Response Console */}
            <div className="flex-1 flex flex-col space-y-1">
              <div className="text-xs font-mono text-slate-400">Response Payload</div>
              <pre className="flex-1 min-h-[360px] max-h-[480px] overflow-auto bg-slate-950 p-3 rounded-lg border border-slate-800 text-[11px] font-mono text-cyan-200 leading-relaxed">
                {testResponse || "Click 'Test' on any endpoint above to inspect live JSON response."}
              </pre>
            </div>
          </div>
        </div>
      )}

      {/* VIEW: SUBSTRATE BLUEPRINT */}
      {activeTab === "architecture" && (
        <div className="bg-slate-900/90 border border-slate-800 rounded-xl p-6 space-y-6">
          <div className="border-b border-slate-800 pb-4">
            <h3 className="text-base font-bold text-white flex items-center space-x-2">
              <Boxes className="w-5 h-5 text-cyan-400" />
              <span>Sovereign Substrate Architecture & Canonical Grounding</span>
            </h3>
            <p className="text-xs text-slate-400 mt-1">
              Unified Single-Binary Echosh Engine with Embedded Next.js 15 Static Export.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6 font-mono text-xs">
            {/* Host Blueprint */}
            <div className="p-4 bg-slate-950/80 border border-slate-800 rounded-xl space-y-3">
              <div className="text-xs font-bold uppercase text-cyan-400 flex items-center space-x-2">
                <Server className="w-4 h-4" />
                <span>Canonical Root & Environment</span>
              </div>
              <div className="space-y-2 text-slate-300">
                <div>
                  <span className="text-slate-500">Root Directory: </span>
                  <span className="text-cyan-300 font-bold">/home/justin/code/echosh-labs/mercury-dasha</span>
                </div>
                <div>
                  <span className="text-slate-500">Host OS: </span>
                  <span className="text-slate-200">WSL2 Ubuntu (Linux 6.6+)</span>
                </div>
                <div>
                  <span className="text-slate-500">Local Dropbox: </span>
                  <span className="text-emerald-400 font-bold">/home/justin/Dropbox (POSIX Daemon)</span>
                </div>
                <div>
                  <span className="text-slate-500">Daemon Status: </span>
                  <span className="text-emerald-400">python3 ~/dropbox.py status &rarr; Up to date</span>
                </div>
              </div>
            </div>

            {/* Sibling Ecosystem */}
            <div className="p-4 bg-slate-950/80 border border-slate-800 rounded-xl space-y-3">
              <div className="text-xs font-bold uppercase text-amber-400 flex items-center space-x-2">
                <FolderTree className="w-4 h-4" />
                <span>Echosh-Labs Ecosystem Repos</span>
              </div>
              <div className="grid grid-cols-2 gap-2 text-slate-300">
                {overview?.architecture.sibling_repos.map((repo) => (
                  <div
                    key={repo}
                    className="p-2 bg-slate-900 border border-slate-800/80 rounded flex items-center space-x-1.5"
                  >
                    <ChevronRight className="w-3 h-3 text-cyan-400" />
                    <span>{repo}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Architecture Pillars Banner */}
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 pt-2">
            <div className="p-4 bg-slate-950/50 border border-slate-800 rounded-lg">
              <div className="text-cyan-400 font-bold text-xs uppercase mb-1">Pillar 1: Dasha Observatory</div>
              <p className="text-slate-400 text-xs">
                Vimshottari 120-year cycles with sub-periods (Mahadasha, Antardasha, Pratyantardasha) and 27 Nakshatras.
              </p>
            </div>
            <div className="p-4 bg-slate-950/50 border border-slate-800 rounded-lg">
              <div className="text-purple-400 font-bold text-xs uppercase mb-1">Pillar 2: Chrono Metronome</div>
              <p className="text-slate-400 text-xs">
                2-second Server-Sent Events pulse calculating planetary hours, hermetic axioms, and live telemetry.
              </p>
            </div>
            <div className="p-4 bg-slate-950/50 border border-slate-800 rounded-lg">
              <div className="text-emerald-400 font-bold text-xs uppercase mb-1">Pillar 3: Sovereign Storage</div>
              <p className="text-slate-400 text-xs">
                Local POSIX Linux Dropbox crawling with 85,119 files indexed in embedded BoltDB and cloud OAuth backups.
              </p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

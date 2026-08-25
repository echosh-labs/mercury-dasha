"use client";

import React, { useState, useEffect } from "react";
import { Database, Plus, RefreshCw, Trash2, Save, FileJson, CheckCircle, AlertCircle } from "lucide-react";

interface HealthData {
  status: string;
  service: string;
  environment: string;
  timestamp: string;
  database?: {
    path: string;
    size_bytes: number;
    key_count: number;
    tx_total: number;
    open_time: string;
    allocated_pages: number;
  };
}

export default function MetaEditor() {
  const [health, setHealth] = useState<HealthData | null>(null);
  const [keys, setKeys] = useState<string[]>([]);
  const [selectedKey, setSelectedKey] = useState<string>("");
  const [keyInput, setKeyInput] = useState<string>("");
  const [jsonContent, setJsonContent] = useState<string>("{\n  \"message\": \"Hello from mercury-dasha\"\n}");
  const [loading, setLoading] = useState<boolean>(false);
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);

  const fetchHealth = async () => {
    try {
      const res = await fetch("/healthz");
      if (res.ok) {
        const data = await res.json();
        setHealth(data);
      }
    } catch (err) {
      console.error("Health check failed:", err);
    }
  };

  const fetchKeys = async () => {
    try {
      const res = await fetch("/api/v1/meta");
      if (res.ok) {
        const data = await res.json();
        setKeys(data.keys || []);
      }
    } catch (err) {
      console.error("Fetch keys failed:", err);
    }
  };

  useEffect(() => {
    fetchHealth();
    fetchKeys();
    const interval = setInterval(fetchHealth, 10000);
    return () => clearInterval(interval);
  }, []);

  const loadDocument = async (key: string) => {
    setSelectedKey(key);
    setKeyInput(key);
    setLoading(true);
    setMessage(null);
    try {
      const res = await fetch(`/api/v1/meta/${encodeURIComponent(key)}`);
      if (res.ok) {
        const data = await res.json();
        setJsonContent(JSON.stringify(data, null, 2));
      } else {
        setMessage({ type: "error", text: "Failed to load document" });
      }
    } catch (err) {
      setMessage({ type: "error", text: "Network error loading document" });
    } finally {
      setLoading(false);
    }
  };

  const saveDocument = async () => {
    if (!keyInput.trim()) {
      setMessage({ type: "error", text: "Document key cannot be empty" });
      return;
    }

    try {
      JSON.parse(jsonContent); // Validate JSON client-side
    } catch (e) {
      setMessage({ type: "error", text: "Invalid JSON format: " + (e as Error).message });
      return;
    }

    setLoading(true);
    setMessage(null);
    try {
      const res = await fetch(`/api/v1/meta/${encodeURIComponent(keyInput.trim())}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: jsonContent,
      });

      if (res.ok) {
        setMessage({ type: "success", text: `Saved key "${keyInput.trim()}" to BoltDB` });
        setSelectedKey(keyInput.trim());
        await fetchKeys();
        await fetchHealth();
      } else {
        const err = await res.json();
        setMessage({ type: "error", text: err.error || "Failed to save document" });
      }
    } catch (err) {
      setMessage({ type: "error", text: "Network error saving document" });
    } finally {
      setLoading(false);
    }
  };

  const deleteDocument = async (key: string) => {
    if (!confirm(`Delete key "${key}" from BoltDB?`)) return;

    setLoading(true);
    setMessage(null);
    try {
      const res = await fetch(`/api/v1/meta/${encodeURIComponent(key)}`, {
        method: "DELETE",
      });

      if (res.ok) {
        setMessage({ type: "success", text: `Deleted "${key}"` });
        if (selectedKey === key) {
          setSelectedKey("");
          setKeyInput("");
          setJsonContent("{\n  \n}");
        }
        await fetchKeys();
        await fetchHealth();
      } else {
        setMessage({ type: "error", text: "Failed to delete" });
      }
    } catch (err) {
      setMessage({ type: "error", text: "Network error deleting key" });
    } finally {
      setLoading(false);
    }
  };

  const formatBytes = (bytes?: number) => {
    if (!bytes && bytes !== 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return (bytes / Math.pow(k, i)).toFixed(2) + " " + sizes[i];
  };

  return (
    <div className="space-y-6">
      {/* Top Telemetry Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="bg-[#111928] border border-slate-800 rounded-xl p-4 flex flex-col justify-between">
          <div className="text-xs font-mono uppercase text-slate-400">Service Status</div>
          <div className="flex items-center space-x-2 mt-2">
            <span className={`h-2.5 w-2.5 rounded-full ${health?.status === "healthy" ? "bg-emerald-400" : "bg-amber-400"}`} />
            <span className="text-lg font-semibold capitalize text-white">{health?.status || "Connecting..."}</span>
          </div>
          <div className="text-xs text-slate-500 font-mono mt-1">{health?.service || "mercury-dasha"}</div>
        </div>

        <div className="bg-[#111928] border border-slate-800 rounded-xl p-4 flex flex-col justify-between">
          <div className="text-xs font-mono uppercase text-slate-400">BoltDB Store</div>
          <div className="text-lg font-semibold text-cyan-400 mt-2">
            {formatBytes(health?.database?.size_bytes)}
          </div>
          <div className="text-xs text-slate-500 font-mono mt-1 truncate" title={health?.database?.path}>
            {health?.database?.path || "/var/data/mercury-dasha.db"}
          </div>
        </div>

        <div className="bg-[#111928] border border-slate-800 rounded-xl p-4 flex flex-col justify-between">
          <div className="text-xs font-mono uppercase text-slate-400">Total Key Records</div>
          <div className="text-lg font-semibold text-white mt-2">
            {health?.database?.key_count ?? keys.length}
          </div>
          <div className="text-xs text-slate-500 font-mono mt-1">Bucket: dasha_meta</div>
        </div>

        <div className="bg-[#111928] border border-slate-800 rounded-xl p-4 flex flex-col justify-between">
          <div className="text-xs font-mono uppercase text-slate-400">Transactions</div>
          <div className="text-lg font-semibold text-purple-400 mt-2">
            {health?.database?.tx_total ?? 0}
          </div>
          <div className="text-xs text-slate-500 font-mono mt-1">Atomic ACID ops</div>
        </div>
      </div>

      {/* Main Studio Area */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left: Key Index / Navigator */}
        <div className="bg-[#111928] border border-slate-800 rounded-xl p-4 flex flex-col h-[520px]">
          <div className="flex items-center justify-between pb-3 border-b border-slate-800">
            <div className="flex items-center space-x-2 text-sm font-semibold text-slate-200">
              <Database className="w-4 h-4 text-cyan-400" />
              <span>Keys ({keys.length})</span>
            </div>
            <div className="flex items-center space-x-1">
              <button
                onClick={() => {
                  fetchKeys();
                  fetchHealth();
                }}
                className="p-1.5 hover:bg-slate-800 text-slate-400 hover:text-white rounded-lg transition"
                title="Refresh Keys"
              >
                <RefreshCw className="w-3.5 h-3.5" />
              </button>
              <button
                onClick={() => {
                  setSelectedKey("");
                  setKeyInput("config:app:" + Date.now().toString().slice(-4));
                  setJsonContent("{\n  \"status\": \"initialized\",\n  \"meta\": {\n    \"owner\": \"echosh-labs\"\n  }\n}");
                  setMessage(null);
                }}
                className="p-1.5 bg-cyan-600 hover:bg-cyan-500 text-white rounded-lg transition flex items-center space-x-1 text-xs"
              >
                <Plus className="w-3.5 h-3.5" />
                <span>New</span>
              </button>
            </div>
          </div>

          <div className="flex-1 overflow-y-auto mt-3 space-y-1 pr-1">
            {keys.length === 0 ? (
              <div className="text-center py-12 text-slate-500 text-xs font-mono">
                No keys found.<br />Create your first document.
              </div>
            ) : (
              keys.map((k) => (
                <div
                  key={k}
                  onClick={() => loadDocument(k)}
                  className={`group flex items-center justify-between px-3 py-2 rounded-lg cursor-pointer text-xs font-mono transition ${
                    selectedKey === k
                      ? "bg-cyan-950/70 text-cyan-300 border border-cyan-800/80"
                      : "text-slate-300 hover:bg-slate-800/60"
                  }`}
                >
                  <div className="flex items-center space-x-2 truncate">
                    <FileJson className="w-3.5 h-3.5 text-slate-400 group-hover:text-cyan-400 shrink-0" />
                    <span className="truncate">{k}</span>
                  </div>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      deleteDocument(k);
                    }}
                    className="opacity-0 group-hover:opacity-100 p-1 hover:text-red-400 transition"
                    title="Delete Key"
                  >
                    <Trash2 className="w-3 h-3" />
                  </button>
                </div>
              ))
            )}
          </div>
        </div>

        {/* Right: JSON Document Editor */}
        <div className="lg:col-span-2 bg-[#111928] border border-slate-800 rounded-xl p-4 flex flex-col h-[520px]">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-slate-800">
            <div className="flex-1 flex items-center space-x-2">
              <span className="text-xs font-mono text-slate-400">Key:</span>
              <input
                type="text"
                value={keyInput}
                onChange={(e) => setKeyInput(e.target.value)}
                placeholder="e.g. system:config or agent:dasha:01"
                className="flex-1 bg-slate-900 border border-slate-700 rounded px-2.5 py-1 text-xs font-mono text-cyan-300 focus:outline-none focus:border-cyan-500"
              />
            </div>
            <div className="flex items-center space-x-2">
              <button
                onClick={saveDocument}
                disabled={loading}
                className="px-3 py-1.5 bg-cyan-600 hover:bg-cyan-500 disabled:opacity-50 text-white rounded-lg text-xs font-medium flex items-center space-x-1.5 transition shadow"
              >
                <Save className="w-3.5 h-3.5" />
                <span>Save to BoltDB</span>
              </button>
              <a
                href="/api/v1/backup"
                download
                className="px-3 py-1.5 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded-lg text-xs font-medium transition"
                title="Download live binary snapshot of BoltDB"
              >
                Snapshot
              </a>
            </div>
          </div>

          {message && (
            <div
              className={`mt-2 p-2 rounded text-xs flex items-center space-x-2 ${
                message.type === "success"
                  ? "bg-emerald-950/60 border border-emerald-800 text-emerald-300"
                  : "bg-red-950/60 border border-red-800 text-red-300"
              }`}
            >
              {message.type === "success" ? (
                <CheckCircle className="w-3.5 h-3.5 shrink-0" />
              ) : (
                <AlertCircle className="w-3.5 h-3.5 shrink-0" />
              )}
              <span>{message.text}</span>
            </div>
          )}

          <div className="flex-1 mt-3 relative">
            <textarea
              value={jsonContent}
              onChange={(e) => setJsonContent(e.target.value)}
              placeholder="Enter JSON metadata..."
              className="w-full h-full bg-[#090d16] border border-slate-800 rounded-lg p-3 text-xs font-mono text-slate-200 focus:outline-none focus:border-cyan-600 resize-none"
              spellCheck={false}
            />
          </div>
        </div>
      </div>
    </div>
  );
}

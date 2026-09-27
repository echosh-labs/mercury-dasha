"use client";

import React, { useState, useEffect, useCallback } from "react";
import {
  Radio,
  X,
  ExternalLink,
  StickyNote,
  Mail,
  FileText,
  Table,
  Calendar,
  Zap,
  Sparkles,
  Compass,
  CheckCircle2,
  Clock,
  RefreshCw
} from "lucide-react";
import { savePrecursorClassification, type PrecursorClassification } from "./axis-mundi-toast-listener";

interface WorkspaceItem {
  id: string;
  type: string;
  title: string;
  snippet: string;
  source: string;
  first_seen_at: string;
  status?: string;
}

interface WorkspaceFeed {
  status: {
    is_live: boolean;
    endpoint: string;
    mode: string;
    counts: {
      total: number;
      keep_notes: number;
      gmail: number;
      docs: number;
      sheets: number;
    };
  };
  items: WorkspaceItem[];
}

interface AxisMundiTriageDrawerProps {
  isOpen: boolean;
  onClose: () => void;
}

export default function AxisMundiTriageDrawer({ isOpen, onClose }: AxisMundiTriageDrawerProps) {
  const [feed, setFeed] = useState<WorkspaceFeed | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [classifications, setClassifications] = useState<Record<string, { classification: PrecursorClassification; classified_at: string }>>({});

  const loadClassifications = useCallback(() => {
    if (typeof window === "undefined") return;
    try {
      const raw = localStorage.getItem("mercury_axis_classifications") || "{}";
      setClassifications(JSON.parse(raw));
    } catch {
      // non-blocking
    }
  }, []);

  const fetchFeed = useCallback(async () => {
    try {
      const res = await fetch("/api/v1/axis-mundi/feed");
      if (res.ok) {
        const data: WorkspaceFeed = await res.json();
        setFeed(data);
      }
    } catch {
      // non-blocking
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (isOpen) {
      loadClassifications();
      fetchFeed();
    }
  }, [isOpen, loadClassifications, fetchFeed]);

  useEffect(() => {
    const handleClassified = () => {
      loadClassifications();
    };
    if (typeof window !== "undefined") {
      window.addEventListener("mercury_axis_classified", handleClassified);
      return () => window.removeEventListener("mercury_axis_classified", handleClassified);
    }
  }, [loadClassifications]);

  if (!isOpen) return null;

  const items = feed?.items || [];
  const status = feed?.status;

  const getItemIcon = (type: string) => {
    switch (type.toLowerCase()) {
      case "keep":
        return <StickyNote className="w-3.5 h-3.5 text-amber-400" />;
      case "gmail":
        return <Mail className="w-3.5 h-3.5 text-rose-400" />;
      case "doc":
        return <FileText className="w-3.5 h-3.5 text-blue-400" />;
      case "sheet":
        return <Table className="w-3.5 h-3.5 text-emerald-400" />;
      case "calendar":
        return <Calendar className="w-3.5 h-3.5 text-purple-400" />;
      default:
        return <Radio className="w-3.5 h-3.5 text-cyan-400" />;
    }
  };

  const getClassificationBadge = (classification?: PrecursorClassification) => {
    switch (classification) {
      case "ACTIONABLE_DIRECTIVE":
        return (
          <span className="px-2 py-0.5 rounded text-[9px] font-mono bg-amber-950/80 text-amber-300 border border-amber-800 flex items-center space-x-1">
            <Zap className="w-2.5 h-2.5" />
            <span>Action</span>
          </span>
        );
      case "CREATIVE_INSPIRATION":
        return (
          <span className="px-2 py-0.5 rounded text-[9px] font-mono bg-purple-950/80 text-purple-300 border border-purple-800 flex items-center space-x-1">
            <Sparkles className="w-2.5 h-2.5" />
            <span>Creative</span>
          </span>
        );
      case "ESOTERIC_ALIGNMENT":
        return (
          <span className="px-2 py-0.5 rounded text-[9px] font-mono bg-cyan-950/80 text-cyan-300 border border-cyan-800 flex items-center space-x-1">
            <Compass className="w-2.5 h-2.5" />
            <span>Esoteric</span>
          </span>
        );
      default:
        return null;
    }
  };

  return (
    <div className="fixed inset-0 z-50 overflow-hidden bg-black/60 backdrop-blur-sm flex justify-end animate-in fade-in duration-200">
      <div className="w-full max-w-md bg-[#0a0f1d] border-l border-slate-800 h-full flex flex-col shadow-2xl animate-in slide-in-from-right duration-200">
        {/* Drawer Header */}
        <div className="p-4 border-b border-slate-800 flex items-center justify-between bg-[#0e1628]/80">
          <div className="flex items-center space-x-2">
            <Radio className={`w-4 h-4 ${status?.is_live ? "text-emerald-400 animate-pulse" : "text-cyan-400"}`} />
            <div>
              <h3 className="text-sm font-bold text-white tracking-tight font-serif">
                Axis Mundi Inbound Triage
              </h3>
              <span className="text-[10px] text-slate-400 font-mono block">
                {status?.counts?.total ?? items.length} Monitored Items • Port 8088
              </span>
            </div>
          </div>

          <div className="flex items-center space-x-2">
            <a
              href="http://localhost:8088"
              target="_blank"
              rel="noreferrer"
              className="p-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-cyan-300 text-xs font-mono flex items-center space-x-1 transition"
              title="Open full Axis Mundi Command Center in new tab"
            >
              <span className="text-[10px]">Open TUI</span>
              <ExternalLink className="w-3 h-3" />
            </a>

            <button
              onClick={onClose}
              className="p-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-400 hover:text-white transition"
              title="Close drawer"
            >
              <X className="w-4 h-4" />
            </button>
          </div>
        </div>

        {/* Inbound Items List */}
        <div className="flex-1 overflow-y-auto p-4 space-y-3 scrollbar-thin">
          {loading ? (
            <div className="py-12 flex flex-col items-center justify-center space-y-2 text-slate-500 font-mono text-xs">
              <RefreshCw className="w-5 h-5 animate-spin text-cyan-400" />
              <span>Checking Axis Mundi inbound stream...</span>
            </div>
          ) : items.length === 0 ? (
            <div className="py-12 text-center text-slate-500 font-mono text-xs p-6 border border-dashed border-slate-800 rounded-xl">
              No recent inbounds recorded.
            </div>
          ) : (
            items.map((item) => {
              const currentClass = classifications[item.id]?.classification;
              return (
                <div
                  key={item.id}
                  className="p-3 rounded-xl bg-slate-900/80 border border-slate-800/80 hover:border-slate-700 transition space-y-2 text-xs font-mono"
                >
                  <div className="flex items-start justify-between gap-2">
                    <div className="flex items-center space-x-2 min-w-0">
                      <span className="p-1.5 rounded bg-slate-800 shrink-0">
                        {getItemIcon(item.type)}
                      </span>
                      <h4 className="text-slate-200 font-semibold truncate text-[11px]">
                        {item.title}
                      </h4>
                    </div>

                    {getClassificationBadge(currentClass)}
                  </div>

                  {item.snippet && (
                    <p className="text-slate-400 font-sans text-[11px] leading-relaxed line-clamp-2 bg-slate-950/40 p-2 rounded border border-slate-800/60">
                      {item.snippet}
                    </p>
                  )}

                  {/* Precursor Triage Bar */}
                  <div className="pt-1.5 border-t border-slate-800/60 flex items-center justify-between">
                    <span className="text-[9px] text-slate-500">
                      {item.source}
                    </span>

                    <div className="flex items-center space-x-1">
                      <button
                        onClick={() => savePrecursorClassification(item.id, "ACTIONABLE_DIRECTIVE")}
                        className={`p-1 rounded text-[9px] font-semibold transition ${
                          currentClass === "ACTIONABLE_DIRECTIVE"
                            ? "bg-amber-600 text-white"
                            : "bg-slate-800 text-amber-400 hover:bg-slate-700"
                        }`}
                        title="Mark as Actionable Directive"
                      >
                        ⚡ Action
                      </button>

                      <button
                        onClick={() => savePrecursorClassification(item.id, "CREATIVE_INSPIRATION")}
                        className={`p-1 rounded text-[9px] font-semibold transition ${
                          currentClass === "CREATIVE_INSPIRATION"
                            ? "bg-purple-600 text-white"
                            : "bg-slate-800 text-purple-400 hover:bg-slate-700"
                        }`}
                        title="Mark as Creative Inspiration"
                      >
                        🎨 Story
                      </button>

                      <button
                        onClick={() => savePrecursorClassification(item.id, "ESOTERIC_ALIGNMENT")}
                        className={`p-1 rounded text-[9px] font-semibold transition ${
                          currentClass === "ESOTERIC_ALIGNMENT"
                            ? "bg-cyan-600 text-white"
                            : "bg-slate-800 text-cyan-400 hover:bg-slate-700"
                        }`}
                        title="Mark as Esoteric Planetary Resonance"
                      >
                        🔮 Esoteric
                      </button>
                    </div>
                  </div>
                </div>
              );
            })
          )}
        </div>

        {/* Footer Link to Authoritative Hub */}
        <div className="p-3 border-t border-slate-800 bg-[#0c1322] flex items-center justify-between text-xs font-mono text-slate-400">
          <span>Authoritative Hub: Port 8088</span>
          <a
            href="http://localhost:8088"
            target="_blank"
            rel="noreferrer"
            className="text-cyan-400 hover:underline flex items-center space-x-1"
          >
            <span>Launch Axis Mundi</span>
            <ExternalLink className="w-3 h-3" />
          </a>
        </div>
      </div>
    </div>
  );
}

"use client";

import React, { useEffect, useCallback } from "react";
import { toast } from "sonner";
import {
  StickyNote,
  Mail,
  FileText,
  Table,
  Calendar,
  Radio,
  ExternalLink,
  Zap,
  Sparkles,
  Compass,
  Check,
  X
} from "lucide-react";

export type PrecursorClassification =
  | "ACTIONABLE_DIRECTIVE"
  | "CREATIVE_INSPIRATION"
  | "ESOTERIC_ALIGNMENT"
  | "PASSIVE_REFERENCE";

export interface WorkspaceAlert {
  id: string;
  item_id: string;
  type: "keep" | "gmail" | "doc" | "sheet" | "calendar" | string;
  title: string;
  snippet: string;
  timestamp: string;
  message: string;
}

// Gentle ambient harmonic chime (528 Hz Solfeggio / Love frequency)
function playInboundChime() {
  if (typeof window === "undefined") return;
  try {
    const AudioCtx = window.AudioContext || (window as any).webkitAudioContext;
    if (!AudioCtx) return;
    const ctx = new AudioCtx();
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();

    osc.type = "sine";
    osc.frequency.setValueAtTime(528, ctx.currentTime); // 528Hz
    osc.frequency.exponentialRampToValueAtTime(792, ctx.currentTime + 0.15); // Harmonic fifth

    gain.gain.setValueAtTime(0.04, ctx.currentTime);
    gain.gain.exponentialRampToValueAtTime(0.0001, ctx.currentTime + 0.5);

    osc.connect(gain);
    gain.connect(ctx.destination);

    osc.start();
    osc.stop(ctx.currentTime + 0.5);
  } catch {
    // Non-blocking if audio context is restricted
  }
}

export function savePrecursorClassification(itemId: string, classification: PrecursorClassification) {
  if (typeof window === "undefined") return;
  try {
    const raw = localStorage.getItem("mercury_axis_classifications") || "{}";
    const data = JSON.parse(raw);
    data[itemId] = {
      classification,
      classified_at: new Date().toISOString(),
    };
    localStorage.setItem("mercury_axis_classifications", JSON.stringify(data));
    window.dispatchEvent(new CustomEvent("mercury_axis_classified", { detail: { itemId, classification } }));
  } catch {
    // Non-blocking
  }
}

export default function AxisMundiToastListener() {
  const handleAlert = useCallback((alert: WorkspaceAlert) => {
    playInboundChime();

    const channelDetails = getChannelMeta(alert.type);
    const Icon = channelDetails.icon;

    toast.custom(
      (t) => (
        <div className="bg-[#0b1220] border border-slate-700/80 rounded-2xl p-4 shadow-2xl backdrop-blur-xl w-full max-w-md text-xs font-mono space-y-3 pointer-events-auto">
          {/* Header & Channel Badge */}
          <div className="flex items-center justify-between gap-2 border-b border-slate-800/80 pb-2.5">
            <div className="flex items-center space-x-2">
              <span className={`p-1.5 rounded-lg border ${channelDetails.badgeStyle}`}>
                <Icon className={`w-3.5 h-3.5 ${channelDetails.textColor}`} />
              </span>
              <div>
                <span className={`font-bold tracking-wider uppercase text-[10px] ${channelDetails.textColor}`}>
                  {channelDetails.label}
                </span>
                <span className="text-slate-500 text-[9px] block">
                  Axis Mundi // Port 8088
                </span>
              </div>
            </div>

            <button
              onClick={() => toast.dismiss(t)}
              className="text-slate-500 hover:text-slate-200 transition p-1 rounded-md"
              title="Dismiss"
            >
              <X className="w-3.5 h-3.5" />
            </button>
          </div>

          {/* Title & Snippet Payload */}
          <div className="space-y-1">
            <h4 className="text-white font-semibold text-xs leading-snug tracking-tight">
              {alert.title || "Inbound Transmission"}
            </h4>
            {alert.snippet && (
              <p className="text-slate-300 font-sans text-[11px] leading-relaxed line-clamp-2 bg-slate-900/80 p-2 rounded-lg border border-slate-800/70">
                {alert.snippet}
              </p>
            )}
          </div>

          {/* Precursor Classification Actions */}
          <div className="pt-2 border-t border-slate-800/80 space-y-1.5">
            <div className="text-[10px] text-slate-400 uppercase tracking-wider flex items-center justify-between">
              <span>Quick Precursor Classification:</span>
              <a
                href="http://localhost:8088"
                target="_blank"
                rel="noreferrer"
                className="text-cyan-400 hover:text-cyan-300 transition flex items-center space-x-1"
                title="Inspect in Axis Mundi Command Center"
              >
                <span>Triage in Axis Mundi</span>
                <ExternalLink className="w-2.5 h-2.5" />
              </a>
            </div>

            <div className="grid grid-cols-3 gap-1.5 pt-0.5">
              <button
                onClick={() => {
                  savePrecursorClassification(alert.item_id || alert.id, "ACTIONABLE_DIRECTIVE");
                  toast.success("Classified as Actionable Directive", { duration: 1500 });
                  toast.dismiss(t);
                }}
                className="px-2 py-1.5 rounded-lg bg-amber-950/60 hover:bg-amber-900/70 border border-amber-800/70 text-amber-300 hover:text-amber-200 text-[10px] font-semibold flex items-center justify-center space-x-1 transition"
              >
                <Zap className="w-3 h-3 text-amber-400" />
                <span>Directive</span>
              </button>

              <button
                onClick={() => {
                  savePrecursorClassification(alert.item_id || alert.id, "CREATIVE_INSPIRATION");
                  toast.success("Classified as Creative Story Inspiration", { duration: 1500 });
                  toast.dismiss(t);
                }}
                className="px-2 py-1.5 rounded-lg bg-purple-950/60 hover:bg-purple-900/70 border border-purple-800/70 text-purple-300 hover:text-purple-200 text-[10px] font-semibold flex items-center justify-center space-x-1 transition"
              >
                <Sparkles className="w-3 h-3 text-purple-400" />
                <span>Creative</span>
              </button>

              <button
                onClick={() => {
                  savePrecursorClassification(alert.item_id || alert.id, "ESOTERIC_ALIGNMENT");
                  toast.success("Classified as Esoteric Planetary Resonance", { duration: 1500 });
                  toast.dismiss(t);
                }}
                className="px-2 py-1.5 rounded-lg bg-cyan-950/60 hover:bg-cyan-900/70 border border-cyan-800/70 text-cyan-300 hover:text-cyan-200 text-[10px] font-semibold flex items-center justify-center space-x-1 transition"
              >
                <Compass className="w-3 h-3 text-cyan-400" />
                <span>Esoteric</span>
              </button>
            </div>
          </div>
        </div>
      ),
      { duration: 8000 }
    );
  }, []);

  useEffect(() => {
    let eventSource: EventSource | null = null;
    try {
      eventSource = new EventSource("/api/v1/stream/pulse");

      eventSource.addEventListener("axis_mundi_alert", (e: MessageEvent) => {
        try {
          const alert: WorkspaceAlert = JSON.parse(e.data);
          handleAlert(alert);
        } catch (err) {
          console.error("Error parsing axis_mundi_alert SSE:", err);
        }
      });
    } catch (err) {
      console.error("Failed to connect to SSE stream for Axis Mundi alerts:", err);
    }

    return () => {
      if (eventSource) eventSource.close();
    };
  }, [handleAlert]);

  return null;
}

function getChannelMeta(type: string) {
  switch (type.toLowerCase()) {
    case "keep":
    case "note":
      return {
        label: "Keep Directive",
        icon: StickyNote,
        textColor: "text-amber-400",
        badgeStyle: "bg-amber-950/80 border-amber-800/80",
      };
    case "gmail":
    case "email":
      return {
        label: "Gmail Transmission",
        icon: Mail,
        textColor: "text-rose-400",
        badgeStyle: "bg-rose-950/80 border-rose-800/80",
      };
    case "doc":
    case "docs":
      return {
        label: "Manuscript Doc",
        icon: FileText,
        textColor: "text-blue-400",
        badgeStyle: "bg-blue-950/80 border-blue-800/80",
      };
    case "sheet":
    case "sheets":
      return {
        label: "Ledger Sheet",
        icon: Table,
        textColor: "text-emerald-400",
        badgeStyle: "bg-emerald-950/80 border-emerald-800/80",
      };
    case "calendar":
    case "event":
      return {
        label: "Temporal Event",
        icon: Calendar,
        textColor: "text-purple-400",
        badgeStyle: "bg-purple-950/80 border-purple-800/80",
      };
    default:
      return {
        label: "Workspace Inbound",
        icon: Radio,
        textColor: "text-cyan-400",
        badgeStyle: "bg-cyan-950/80 border-cyan-800/80",
      };
  }
}

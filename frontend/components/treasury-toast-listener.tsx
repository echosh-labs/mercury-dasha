"use client";

import React, { useEffect, useCallback, useRef } from "react";
import { toast } from "sonner";
import {
  Trophy,
  Users,
  Youtube,
  Coins,
  Sparkles,
  TrendingUp,
  ExternalLink,
  CheckCircle2,
  X
} from "lucide-react";

export type TreasurySuccessType =
  | "youtube_view"
  | "youtube_subscriber"
  | "youtube_milestone"
  | "amra_subscriber"
  | "treasury_tx"
  | "revenue_yield";

export interface TreasurySuccessAlert {
  id: string;
  type: TreasurySuccessType;
  title: string;
  message: string;
  metric_label?: string;
  metric_value?: string | number;
  delta?: string | number;
  channel_title?: string;
  video_title?: string;
  video_url?: string;
  timestamp: string;
}

// Lush celebratory harmonic arpeggio (Solfeggio 528Hz -> 660Hz -> 792Hz -> 1056Hz octave)
function playTreasuryTriumphChime() {
  if (typeof window === "undefined") return;
  try {
    const AudioCtx = window.AudioContext || (window as any).webkitAudioContext;
    if (!AudioCtx) return;
    const ctx = new AudioCtx();
    const chord = [528, 660, 792, 1056]; // Harmonics of transformation & triumph

    chord.forEach((freq, idx) => {
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();

      const startTime = ctx.currentTime + idx * 0.08;
      const stopTime = startTime + 0.55;

      osc.type = "sine";
      osc.frequency.setValueAtTime(freq, startTime);

      gain.gain.setValueAtTime(0.0001, startTime);
      gain.gain.exponentialRampToValueAtTime(0.045, startTime + 0.04);
      gain.gain.exponentialRampToValueAtTime(0.0001, stopTime);

      osc.connect(gain);
      gain.connect(ctx.destination);

      osc.start(startTime);
      osc.stop(stopTime);
    });
  } catch {
    // Non-blocking if audio context is restricted by user interaction policy
  }
}

export default function TreasuryToastListener() {
  const baselineInitialized = useRef<boolean>(false);
  const lastState = useRef<{
    views: number;
    subscribers: number;
    amraSubscribers: number;
  }>({ views: 0, subscribers: 0, amraSubscribers: 0 });

  const handleAlert = useCallback((alert: TreasurySuccessAlert) => {
    playTreasuryTriumphChime();

    const meta = getSuccessMeta(alert.type);
    const Icon = meta.icon;

    // Dispatch custom top-right toast with distinct radiant alchemical emerald & gold styling
    toast.custom(
      (t) => (
        <div className="pointer-events-auto w-full max-w-md bg-gradient-to-br from-[#041a13]/95 via-[#081814]/95 to-[#020b08]/95 border border-emerald-500/40 hover:border-amber-400/60 rounded-2xl p-4 shadow-[0_10px_35px_rgba(16,185,129,0.25)] backdrop-blur-2xl text-xs font-mono space-y-3 relative overflow-hidden group">
          {/* Subtle Ambient Golden Shimmer Aura */}
          <div className="absolute -top-10 -right-10 w-28 h-28 bg-emerald-500/15 rounded-full blur-2xl pointer-events-none group-hover:bg-amber-400/20 transition-all duration-500" />

          {/* Header & Badging */}
          <div className="flex items-center justify-between gap-2 border-b border-emerald-500/20 pb-2.5">
            <div className="flex items-center space-x-2">
              <span className={`p-1.5 rounded-lg border shadow-sm ${meta.badgeBg}`}>
                <Icon className={`w-4 h-4 ${meta.textColor}`} />
              </span>
              <div>
                <span className={`font-bold tracking-wider uppercase text-[10px] ${meta.textColor}`}>
                  {meta.label}
                </span>
                <span className="text-emerald-400/70 text-[9px] block">
                  AMRA Treasury // Port 8050
                </span>
              </div>
            </div>

            <button
              onClick={() => toast.dismiss(t)}
              className="text-slate-400 hover:text-white transition p-1 rounded-md"
              title="Acknowledge & Dismiss"
            >
              <X className="w-3.5 h-3.5" />
            </button>
          </div>

          {/* Title & Metric Highlight Banner */}
          <div className="space-y-2">
            <h4 className="text-white font-bold text-sm tracking-tight flex items-center space-x-1.5">
              <span>{alert.title}</span>
            </h4>

            {/* Radiant Metric Delta Banner */}
            {alert.metric_label && (
              <div className="bg-emerald-950/70 border border-emerald-500/30 rounded-xl p-2.5 flex items-center justify-between shadow-inner">
                <div className="flex items-center space-x-2">
                  <span className="w-2 h-2 rounded-full bg-emerald-400 animate-ping" />
                  <span className="text-emerald-200 text-xs font-semibold">
                    {alert.metric_label}
                  </span>
                </div>
                {alert.metric_value !== undefined && (
                  <span className="text-amber-300 font-bold text-xs bg-amber-500/10 border border-amber-500/20 px-2 py-0.5 rounded-md font-mono">
                    {alert.metric_value}
                  </span>
                )}
              </div>
            )}

            {alert.message && (
              <p className="text-slate-300 font-sans text-xs leading-relaxed">
                {alert.message}
              </p>
            )}

            {alert.video_title && (
              <div className="text-[11px] text-emerald-300/90 bg-emerald-950/40 p-2 rounded-lg border border-emerald-900/50 truncate font-sans">
                <span className="text-slate-400 font-mono text-[10px] block">Associated Media:</span>
                &quot;{alert.video_title}&quot;
              </div>
            )}
          </div>

          {/* Action Row */}
          <div className="pt-2 border-t border-emerald-500/20 flex items-center justify-between text-[11px]">
            <a
              href="http://localhost:8050"
              target="_blank"
              rel="noreferrer"
              className="text-emerald-300 hover:text-emerald-200 font-semibold flex items-center space-x-1 transition"
            >
              <span>Inspect in Treasury (:8050)</span>
              <ExternalLink className="w-3 h-3 text-emerald-400" />
            </a>

            {alert.video_url ? (
              <a
                href={alert.video_url}
                target="_blank"
                rel="noreferrer"
                className="text-amber-300 hover:text-amber-200 flex items-center space-x-1 transition"
              >
                <span>Watch Video</span>
                <ExternalLink className="w-3 h-3 text-amber-400" />
              </a>
            ) : (
              <button
                onClick={() => toast.dismiss(t)}
                className="px-2.5 py-1 rounded-md bg-emerald-900/60 hover:bg-emerald-800/80 text-emerald-200 border border-emerald-700/60 text-[10px] transition"
              >
                Acknowledge
              </button>
            )}
          </div>
        </div>
      ),
      {
        position: "top-right",
        duration: 9000,
      }
    );
  }, []);

  // 1. SSE Stream Listener for push events from backend
  useEffect(() => {
    let eventSource: EventSource | null = null;
    try {
      eventSource = new EventSource("/api/v1/stream/pulse");

      eventSource.addEventListener("treasury_success_alert", (e: MessageEvent) => {
        try {
          const alert: TreasurySuccessAlert = JSON.parse(e.data);
          handleAlert(alert);
        } catch (err) {
          console.error("Error parsing treasury_success_alert SSE:", err);
        }
      });
    } catch (err) {
      console.error("Failed to connect to SSE stream for Treasury alerts:", err);
    }

    return () => {
      if (eventSource) eventSource.close();
    };
  }, [handleAlert]);

  // 2. Client-side Success Poller & Delta Evaluator (detects new views, subscribers, ledger updates)
  useEffect(() => {
    const pollTreasuryMetrics = async () => {
      try {
        // Fetch YouTube channel profile & analytics
        const [ytStatusRes, ytAnalyticsRes] = await Promise.allSettled([
          fetch("/api/v1/youtube/status"),
          fetch("/api/v1/youtube/analytics"),
        ]);

        let currentViews = lastState.current.views;
        let currentSubs = lastState.current.subscribers;
        let channelTitle = "Justin Andrew Wood";

        if (ytStatusRes.status === "fulfilled" && ytStatusRes.value.ok) {
          const statusData = await ytStatusRes.value.json();
          if (statusData.channel) {
            currentSubs = Number(statusData.channel.subscriber_count || 0);
            channelTitle = statusData.channel.title || channelTitle;
          }
        }

        if (ytAnalyticsRes.status === "fulfilled" && ytAnalyticsRes.value.ok) {
          const analyticsData = await ytAnalyticsRes.value.json();
          if (analyticsData.total_views !== undefined) {
            currentViews = Number(analyticsData.total_views);
          }
        }

        // Initialize baseline on first poll
        if (!baselineInitialized.current) {
          baselineInitialized.current = true;
          lastState.current = {
            views: currentViews,
            subscribers: currentSubs,
            amraSubscribers: 0,
          };
          return;
        }

        // Check for YouTube view increase
        if (currentViews > lastState.current.views) {
          const delta = currentViews - lastState.current.views;
          handleAlert({
            id: `yt-view-${Date.now()}`,
            type: "youtube_view",
            title: "🎉 New YouTube Audience Expansion",
            message: `Recorded ${delta} new view${delta > 1 ? "s" : ""} on channel ${channelTitle}. Active viewership resonance confirmed.`,
            metric_label: `+${delta} New View${delta > 1 ? "s" : ""}`,
            metric_value: `Total: ${currentViews}`,
            delta,
            channel_title: channelTitle,
            timestamp: new Date().toISOString(),
          });
          lastState.current.views = currentViews;
        }

        // Check for YouTube subscriber increase
        if (currentSubs > lastState.current.subscribers) {
          const delta = currentSubs - lastState.current.subscribers;
          handleAlert({
            id: `yt-sub-${Date.now()}`,
            type: "youtube_subscriber",
            title: "👑 New YouTube Subscriber Enlisted",
            message: `A new supporter has subscribed to ${channelTitle}! Total channel subscribers updated to ${currentSubs}.`,
            metric_label: `+${delta} New Subscriber${delta > 1 ? "s" : ""}`,
            metric_value: `Subscribers: ${currentSubs}`,
            delta,
            channel_title: channelTitle,
            timestamp: new Date().toISOString(),
          });
          lastState.current.subscribers = currentSubs;
        }
      } catch {
        // Non-blocking
      }
    };

    pollTreasuryMetrics();
    const interval = setInterval(pollTreasuryMetrics, 30000); // Check every 30s

    return () => clearInterval(interval);
  }, [handleAlert]);

  // 3. Listen to ambient window events for manual / test trigger
  useEffect(() => {
    const handleCustomTrigger = (e: Event) => {
      const customEvent = e as CustomEvent<TreasurySuccessAlert>;
      if (customEvent.detail) {
        handleAlert(customEvent.detail);
      }
    };
    window.addEventListener("mercury_treasury_alert", handleCustomTrigger);
    return () => window.removeEventListener("mercury_treasury_alert", handleCustomTrigger);
  }, [handleAlert]);

  return null;
}

function getSuccessMeta(type: TreasurySuccessType) {
  switch (type) {
    case "youtube_subscriber":
    case "amra_subscriber":
      return {
        label: "Subscriber Milestone",
        icon: Users,
        textColor: "text-amber-300",
        badgeBg: "bg-amber-950/80 border-amber-600/60 shadow-amber-950/40",
      };
    case "youtube_view":
      return {
        label: "Audience Expansion",
        icon: Youtube,
        textColor: "text-emerald-300",
        badgeBg: "bg-emerald-950/80 border-emerald-600/60 shadow-emerald-950/40",
      };
    case "revenue_yield":
    case "treasury_tx":
      return {
        label: "Treasury Inflow",
        icon: Coins,
        textColor: "text-yellow-400",
        badgeBg: "bg-yellow-950/80 border-yellow-600/60 shadow-yellow-950/40",
      };
    case "youtube_milestone":
    default:
      return {
        label: "Treasury Victory",
        icon: Trophy,
        textColor: "text-emerald-400",
        badgeBg: "bg-emerald-950/80 border-emerald-600/60 shadow-emerald-950/40",
      };
  }
}

"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { 
  Settings, 
  Globe, 
  Compass, 
  Coins, 
  Terminal, 
  ArrowUpRight, 
  ShieldCheck, 
  Cloud,
  ExternalLink,
  Video,
  Users,
  Sun
} from "lucide-react";
import { DetectClientTemporalEnvironment, DetectedLocation } from "@/lib/timezone";

export default function GlobalFooter() {
  const [activeTz, setActiveTz] = useState<string>("America/Toronto");
  const [activeOffset, setActiveOffset] = useState<number>(-4.0);

  useEffect(() => {
    let mounted = true;

    // 1. Fetch current server settings
    const loadServerSettings = async () => {
      try {
        const res = await fetch("/api/v1/settings");
        if (res.ok) {
          const s: DetectedLocation = await res.json();
          if (mounted && s && s.timezone) {
            setActiveTz(s.timezone);
            setActiveOffset(s.utc_offset_hours);
            return;
          }
        }
        
        // 2. Only auto-detect and sync if server explicitly reports no settings configured (404)
        if (res.status === 404) {
          const detected = await DetectClientTemporalEnvironment();
          if (!mounted) return;
          setActiveTz(detected.timezone);
          setActiveOffset(detected.utc_offset_hours);

          await fetch("/api/v1/settings", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(detected),
          });
        }
      } catch (err) {
        console.warn("[GlobalFooter] Could not fetch server settings:", err);
      }
    };

    loadServerSettings();

    const handleSettingsUpdated = () => {
      loadServerSettings();
    };

    window.addEventListener("mercury_settings_updated", handleSettingsUpdated);
    return () => {
      mounted = false;
      window.removeEventListener("mercury_settings_updated", handleSettingsUpdated);
    };
  }, []);

  const handleSettingsSaved = (newSettings: DetectedLocation) => {
    setActiveTz(newSettings.timezone);
    setActiveOffset(newSettings.utc_offset_hours);
  };

  const offsetStr = activeOffset >= 0 ? `+${activeOffset}` : `${activeOffset}`;

  return (
    <>
      <footer className="border-t border-slate-800/80 bg-[#070b12] text-xs font-mono">
        {/* Dedicated Route Navigation Section */}
        <div className="max-w-7xl mx-auto px-4 sm:px-8 pt-6 pb-5 border-b border-slate-800/60">
          <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-4">
            <div className="flex items-center space-x-2">
              <span className="h-2 w-2 rounded-full bg-emerald-400" />
              <span className="text-[11px] uppercase tracking-wider text-slate-400 font-semibold">
                Sovereign Route Navigation Hub
              </span>
            </div>
            <div className="flex items-center space-x-3 text-[11px] text-slate-500">
              <span className="flex items-center space-x-1">
                <Cloud className="w-3 h-3 text-blue-400" />
                <span>GCP Cloud Run gen2</span>
              </span>
              <span>•</span>
              <span className="flex items-center space-x-1">
                <ShieldCheck className="w-3 h-3 text-emerald-400" />
                <span>Single Binary Go 1.23 + BoltDB</span>
              </span>
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-3">
            {/* 1. Main Observatory Hub Route */}
            <Link
              href="/"
              className="p-3.5 rounded-xl bg-slate-900/60 hover:bg-slate-900 border border-slate-800/80 hover:border-cyan-500/40 transition flex items-start space-x-3 group"
            >
              <div className="p-2 rounded-lg bg-cyan-950/60 border border-cyan-800/60 text-cyan-400 group-hover:scale-105 transition-transform shrink-0">
                <Compass className="w-4 h-4" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center justify-between">
                  <span className="font-semibold text-slate-200 group-hover:text-cyan-300 transition text-xs">
                    Dasha Observatory Hub
                  </span>
                  <span className="text-[10px] text-slate-500 bg-slate-800 px-1.5 py-0.2 rounded font-sans">
                    /
                  </span>
                </div>
                <p className="text-[11px] text-slate-400 mt-1 font-sans leading-relaxed line-clamp-2">
                  Sidereal Moon ephemeris, 27 Nakshatras, Vimshottari timelines, and alchemical lab.
                </p>
              </div>
            </Link>

            {/* 2. Dedicated AMRA Treasury Route */}
            <Link
              href="/treasury/"
              className="p-3.5 rounded-xl bg-slate-900/60 hover:bg-slate-900 border border-emerald-500/20 hover:border-emerald-500/50 transition flex items-start space-x-3 group shadow-sm shadow-emerald-950/20"
            >
              <div className="p-2 rounded-lg bg-emerald-950/60 border border-emerald-800/60 text-emerald-400 group-hover:scale-105 transition-transform shrink-0">
                <Coins className="w-4 h-4" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-1.5">
                    <span className="font-semibold text-slate-200 group-hover:text-emerald-300 transition text-xs">
                      AMRA Sovereign Treasury
                    </span>
                    <span className="text-[9px] bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 px-1 py-0.2 rounded">
                      ROUTE
                    </span>
                  </div>
                  <ArrowUpRight className="w-3.5 h-3.5 text-emerald-400 group-hover:translate-x-0.5 group-hover:-translate-y-0.5 transition-transform" />
                </div>
                <p className="text-[11px] text-slate-400 mt-1 font-sans leading-relaxed line-clamp-2">
                  Immutable BoltDB ledger, Google Cloud billing, and alchemical shadow transmutation.
                </p>
              </div>
            </Link>

            {/* 3. Dedicated Foundations Studio Route */}
            <Link
              href="/foundations/"
              className="p-3.5 rounded-xl bg-slate-900/60 hover:bg-slate-900 border border-red-500/20 hover:border-red-500/50 transition flex items-start space-x-3 group shadow-sm shadow-red-950/20"
            >
              <div className="p-2 rounded-lg bg-red-950/60 border border-red-800/60 text-red-400 group-hover:scale-105 transition-transform shrink-0">
                <Video className="w-4 h-4" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-1.5">
                    <span className="font-semibold text-slate-200 group-hover:text-red-300 transition text-xs">
                      Foundations Studio
                    </span>
                    <span className="text-[9px] bg-red-500/10 text-red-400 border border-red-500/20 px-1 py-0.2 rounded">
                      ROUTE
                    </span>
                  </div>
                  <ArrowUpRight className="w-3.5 h-3.5 text-red-400 group-hover:translate-x-0.5 group-hover:-translate-y-0.5 transition-transform" />
                </div>
                <p className="text-[11px] text-slate-400 mt-1 font-sans leading-relaxed line-clamp-2">
                  Video pipeline, Sonic Chronicle rendering, channel presence, and audience telemetry.
                </p>
              </div>
            </Link>

            {/* 4. Dedicated Temporal Alignment & Ephemeris Route */}
            <Link
              href="/alignment/"
              className="p-3.5 rounded-xl bg-slate-900/60 hover:bg-slate-900 border border-purple-500/20 hover:border-purple-500/50 transition flex items-start space-x-3 group shadow-sm shadow-purple-950/20"
            >
              <div className="p-2 rounded-lg bg-purple-950/60 border border-purple-800/60 text-purple-400 group-hover:scale-105 transition-transform shrink-0">
                <Settings className="w-4 h-4 group-hover:rotate-90 transition-transform duration-300" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-1.5">
                    <span className="font-semibold text-slate-200 group-hover:text-purple-300 transition text-xs">
                      Temporal Alignment
                    </span>
                    <span className="text-[9px] bg-purple-500/10 text-purple-400 border border-purple-500/20 px-1 py-0.2 rounded">
                      ROUTE
                    </span>
                  </div>
                  <ArrowUpRight className="w-3.5 h-3.5 text-purple-400 group-hover:translate-x-0.5 group-hover:-translate-y-0.5 transition-transform" />
                </div>
                <p className="text-[11px] text-slate-400 mt-1 font-sans leading-relaxed line-clamp-2">
                  24-Hour sub-horas, Lagna arc traversal, and topocentric observer calibration.
                </p>
              </div>
            </Link>

            {/* 5. Character Sanctuary Route */}
            <Link
              href="/characters/"
              className="p-3.5 rounded-xl bg-slate-900/60 hover:bg-slate-900 border border-cyan-500/20 hover:border-cyan-500/50 transition flex items-start space-x-3 group shadow-sm shadow-cyan-950/20"
            >
              <div className="p-2 rounded-lg bg-cyan-950/60 border border-cyan-800/60 text-cyan-400 group-hover:scale-105 transition-transform shrink-0">
                <Users className="w-4 h-4 group-hover:scale-110 transition-transform" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-1.5">
                    <span className="font-semibold text-slate-200 group-hover:text-cyan-300 transition text-xs">
                      Character Sanctuary
                    </span>
                    <span className="text-[9px] bg-cyan-500/10 text-cyan-400 border border-cyan-500/20 px-1 py-0.2 rounded">
                      ROUTE
                    </span>
                  </div>
                  <ArrowUpRight className="w-3.5 h-3.5 text-cyan-400 group-hover:translate-x-0.5 group-hover:-translate-y-0.5 transition-transform" />
                </div>
                <p className="text-[11px] text-slate-400 mt-1 font-sans leading-relaxed line-clamp-2">
                  Create, switch &amp; inspect sovereign profiles, Nava Graha dignities, and Bhavas.
                </p>
              </div>
            </Link>

            {/* 6. Dedicated Daily Planetary Ephemeris Route */}
            <Link
              href="/ephemeris/"
              className="p-3.5 rounded-xl bg-slate-900/60 hover:bg-slate-900 border border-amber-500/20 hover:border-amber-500/50 transition flex items-start space-x-3 group shadow-sm shadow-amber-950/20"
            >
              <div className="p-2 rounded-lg bg-amber-950/60 border border-amber-800/60 text-amber-400 group-hover:scale-105 transition-transform shrink-0">
                <Sun className="w-4 h-4 group-hover:rotate-45 transition-transform duration-300" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-1.5">
                    <span className="font-semibold text-slate-200 group-hover:text-amber-300 transition text-xs">
                      Daily Ephemeris
                    </span>
                    <span className="text-[9px] bg-amber-500/10 text-amber-400 border border-amber-500/20 px-1 py-0.2 rounded">
                      ROUTE
                    </span>
                  </div>
                  <ArrowUpRight className="w-3.5 h-3.5 text-amber-400 group-hover:translate-x-0.5 group-hover:-translate-y-0.5 transition-transform" />
                </div>
                <p className="text-[11px] text-slate-400 mt-1 font-sans leading-relaxed line-clamp-2">
                  Complete 24-hour planetary ephemeris, day lord, solar anchors, and hourly metals.
                </p>
              </div>
            </Link>
          </div>
        </div>

        {/* Bottom Bar: Stack & Ecosystem Telemetry */}
        <div className="max-w-7xl mx-auto py-3.5 px-4 sm:px-8 flex flex-col sm:flex-row items-center justify-between gap-3 text-slate-500">
          <div className="flex items-center space-x-2">
            <span className="text-slate-400 font-semibold">Mercury Stack</span>
            <span>•</span>
            <span>Go 1.23 + bbolt &amp; Next.js 15 Static Export</span>
            <span>•</span>
            <span className="text-cyan-400">echosh-labs</span>
          </div>

          <div className="flex items-center space-x-4">
            <a
              href="https://echosh-labs.com"
              target="_blank"
              rel="noreferrer"
              className="text-slate-400 hover:text-slate-200 transition flex items-center space-x-1"
            >
              <span>echosh-labs.com</span>
              <ExternalLink className="w-3 h-3" />
            </a>
          </div>
        </div>
      </footer>


    </>
  );
}

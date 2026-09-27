"use client";

import React, { useState } from "react";
import Link from "next/link";
import { 
  Flame, 
  ExternalLink,
  ShieldCheck, 
  Cloud, 
  FolderTree, 
  Compass, 
  Terminal, 
  Mic, 
  Coins,
  ArrowUpRight,
  Video,
  Film,
  Camera
} from "lucide-react";
import DropboxSovereignStorehouse from "@/components/dropbox-sovereign-storehouse";
import ChronoPulse from "@/components/chrono-pulse";
import DashaCalculator from "@/components/dasha-calculator";
import AlchemicalLab from "@/components/alchemical-lab";
import AgenticConsole from "@/components/agentic-console";
import AudioStudio from "@/components/audio-studio";
import UnifiedMediaPortal from "@/components/unified-media-portal";
import AxisMundiWorkspace from "@/components/axis-mundi-workspace";
import { Radio } from "lucide-react";

interface TelemetryData {
  service: string;
  uptime: string;
  alloc_mb: number;
  sys_mb: number;
  goroutines: number;
  timestamp: string;
}

export default function Home() {
  const [activeTab, setActiveTab] = useState<"observatory" | "alchemy" | "dropbox" | "audio" | "media" | "axis-mundi" | "console">("observatory");
  const [telemetry, setTelemetry] = useState<TelemetryData | null>(null);

  // Read URL query parameter for tab selection (e.g., /?tab=axis-mundi)
  React.useEffect(() => {
    if (typeof window !== "undefined") {
      const params = new URLSearchParams(window.location.search);
      const tabParam = params.get("tab");
      if (tabParam && ["observatory", "alchemy", "dropbox", "audio", "media", "axis-mundi", "console"].includes(tabParam)) {
        setActiveTab(tabParam as any);
      }
    }
  }, []);

  return (
    <div className="space-y-6">
      {/* Hero Header */}
      <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4 bg-gradient-to-r from-slate-900 via-[#111928] to-slate-900 p-6 rounded-2xl border border-slate-800 shadow-xl">
        <div>
          <div className="flex items-center space-x-2">
            <h1 className="text-2xl font-bold text-white tracking-tight">
              Mercury Dasha
            </h1>
            <span className="bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 text-xs font-mono px-2.5 py-0.5 rounded-full flex items-center space-x-1">
              <ShieldCheck className="w-3 h-3" />
              <span>OPERATIONAL</span>
            </span>
          </div>
          <p className="text-xs text-slate-400 mt-1 max-w-2xl leading-relaxed">
            High-Throughput Astrological Engine & Telemetry Service for <span className="text-cyan-300 font-medium">echosh-labs</span>.
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-2 text-xs font-mono">
          <ChronoPulse onTelemetryUpdate={(tel) => setTelemetry((prev) => prev ? { ...prev, ...tel } : null)} />
          <a
            href="https://echosh-labs.com/compendium"
            target="_blank"
            rel="noreferrer"
            className="px-3 py-1.5 rounded-lg bg-cyan-950/70 border border-cyan-800 text-cyan-300 hover:bg-cyan-900/60 transition flex items-center space-x-1"
          >
            <span>Compendium</span>
            <ExternalLink className="w-3 h-3" />
          </a>
        </div>
      </div>

      {/* Navigation Tabs */}
      <div className="flex border-b border-slate-800 space-x-2 text-sm font-medium overflow-x-auto pb-1">
        <button
          onClick={() => setActiveTab("observatory")}
          className={`flex items-center space-x-2 px-4 py-2.5 rounded-lg transition ${
            activeTab === "observatory"
              ? "bg-slate-800 text-cyan-400 border border-slate-700 font-semibold"
              : "text-slate-400 hover:text-slate-200 hover:bg-slate-800/40"
          }`}
        >
          <Compass className="w-4 h-4 text-cyan-400" />
          <span>Dasha Observatory</span>
        </button>

        <button
          onClick={() => setActiveTab("alchemy")}
          className={`flex items-center space-x-2 px-4 py-2.5 rounded-lg transition ${
            activeTab === "alchemy"
              ? "bg-slate-800 text-amber-400 border border-slate-700 font-semibold"
              : "text-slate-400 hover:text-slate-200 hover:bg-slate-800/40"
          }`}
        >
          <Flame className="w-4 h-4 text-amber-400" />
          <span>Alchemical Laboratory</span>
        </button>

        <button
          onClick={() => setActiveTab("dropbox")}
          className={`flex items-center space-x-2 px-4 py-2.5 rounded-lg transition ${
            activeTab === "dropbox"
              ? "bg-slate-800 text-amber-400 border border-slate-700 font-semibold"
              : "text-slate-400 hover:text-slate-200 hover:bg-slate-800/40"
          }`}
        >
          <FolderTree className="w-4 h-4 text-amber-400" />
          <span>Dropbox Sovereign Storehouse</span>
        </button>

        <button
          onClick={() => setActiveTab("audio")}
          className={`flex items-center space-x-2 px-4 py-2.5 rounded-lg transition ${
            activeTab === "audio"
              ? "bg-slate-800 text-rose-400 border border-slate-700 font-semibold"
              : "text-slate-400 hover:text-slate-200 hover:bg-slate-800/40"
          }`}
        >
          <Mic className="w-4 h-4 text-rose-400" />
          <span>Sonic Chronicle</span>
        </button>

        <button
          onClick={() => setActiveTab("media")}
          className={`flex items-center space-x-2 px-4 py-2.5 rounded-lg transition ${
            activeTab === "media"
              ? "bg-slate-800 text-cyan-400 border border-slate-700 font-semibold"
              : "text-slate-400 hover:text-slate-200 hover:bg-slate-800/40"
          }`}
        >
          <Camera className="w-4 h-4 text-cyan-400" />
          <span>Visual Chronicle</span>
        </button>

        <button
          onClick={() => setActiveTab("axis-mundi")}
          className={`flex items-center space-x-2 px-4 py-2.5 rounded-lg transition ${
            activeTab === "axis-mundi"
              ? "bg-slate-800 text-cyan-400 border border-slate-700 font-semibold"
              : "text-slate-400 hover:text-slate-200 hover:bg-slate-800/40"
          }`}
        >
          <Radio className="w-4 h-4 text-cyan-400" />
          <span>Axis Mundi Ingestion</span>
        </button>

        <button
          onClick={() => setActiveTab("console")}
          className={`flex items-center space-x-2 px-4 py-2.5 rounded-lg transition ${
            activeTab === "console"
              ? "bg-slate-800 text-purple-400 border border-slate-700 font-semibold"
              : "text-slate-400 hover:text-slate-200 hover:bg-slate-800/40"
          }`}
        >
          <Terminal className="w-4 h-4 text-purple-400" />
          <span>Agentic Mission Control</span>
        </button>

        <Link
          href="/treasury/"
          className="flex items-center space-x-2 px-4 py-2.5 rounded-lg transition text-slate-400 hover:text-emerald-300 hover:bg-emerald-950/40 border border-transparent hover:border-emerald-500/30 group shrink-0"
          title="Open Dedicated AMRA Sovereign Treasury Route"
        >
          <Coins className="w-4 h-4 text-emerald-400 group-hover:scale-110 transition-transform" />
          <span>AMRA Treasury &amp; Studio</span>
          <ArrowUpRight className="w-3.5 h-3.5 text-emerald-400/70 group-hover:text-emerald-300 transition" />
        </Link>
      </div>

      {/* Tab Contents */}
      {activeTab === "observatory" && <DashaCalculator />}
      {activeTab === "alchemy" && <AlchemicalLab />}
      {activeTab === "dropbox" && <DropboxSovereignStorehouse />}
      {activeTab === "audio" && <AudioStudio />}
      {activeTab === "media" && <UnifiedMediaPortal />}
      {activeTab === "axis-mundi" && <AxisMundiWorkspace />}
      {activeTab === "console" && <AgenticConsole />}
    </div>
  );
}

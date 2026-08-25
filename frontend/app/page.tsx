"use client";

import React, { useState, useEffect } from "react";
import { 
  Activity, 
  Database, 
  Sparkles, 
  Orbit, 
  Flame, 
  Layers, 
  RefreshCw, 
  ExternalLink,
  ShieldCheck,
  Cpu
} from "lucide-react";
import MetaEditor from "@/components/meta-editor";

interface TelemetryData {
  service: string;
  uptime: string;
  alloc_mb: number;
  sys_mb: number;
  goroutines: number;
  timestamp: string;
}

interface Planet {
  name: string;
  sanskrit_name: string;
  duration_years: number;
  frequency_hz: number;
  element: string;
  chakra: string;
  color_hex: string;
}

interface DashaOverview {
  mahadasha_lord: string;
  total_years: number;
  root_frequency_hz: number;
  planets: Planet[];
}

interface Nakshatra {
  id: string;
  name: string;
  range: string;
  frequency: number;
  deity: string;
  symbol: string;
  quality: string;
}

interface Axiom {
  title: string;
  text: string;
}

export default function Home() {
  const [activeTab, setActiveTab] = useState<"overview" | "nakshatras" | "alchemy" | "database">("overview");
  const [telemetry, setTelemetry] = useState<TelemetryData | null>(null);
  const [overview, setOverview] = useState<DashaOverview | null>(null);
  const [nakshatras, setNakshatras] = useState<Nakshatra[]>([]);
  const [axioms, setAxioms] = useState<Axiom[]>([]);
  const [loading, setLoading] = useState<boolean>(true);

  const fetchData = async () => {
    try {
      const [tRes, oRes, nRes, aRes] = await Promise.all([
        fetch("/api/telemetry"),
        fetch("/api/dasha/overview"),
        fetch("/api/dasha/nakshatras"),
        fetch("/api/dasha/alchemy")
      ]);

      if (tRes.ok) setTelemetry(await tRes.json());
      if (oRes.ok) setOverview(await oRes.json());
      if (nRes.ok) {
        const nData = await nRes.json();
        setNakshatras(nData.nakshatras || []);
      }
      if (aRes.ok) {
        const aData = await aRes.json();
        setAxioms(aData.axioms || []);
      }
    } catch (e) {
      console.error("Failed to load domain data:", e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 10000);
    return () => clearInterval(interval);
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
          <div className="px-3 py-1.5 rounded-lg bg-slate-800/90 border border-slate-700 text-slate-300 flex items-center space-x-1.5">
            <Cpu className="w-3.5 h-3.5 text-cyan-400" />
            <span>RAM: {telemetry ? `${telemetry.alloc_mb.toFixed(2)} MB` : "..."}</span>
          </div>
          <div className="px-3 py-1.5 rounded-lg bg-slate-800/90 border border-slate-700 text-slate-300 flex items-center space-x-1.5">
            <Activity className="w-3.5 h-3.5 text-emerald-400" />
            <span>Uptime: {telemetry?.uptime || "..."}</span>
          </div>
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
          onClick={() => setActiveTab("overview")}
          className={`flex items-center space-x-2 px-4 py-2.5 rounded-lg transition ${
            activeTab === "overview"
              ? "bg-slate-800 text-cyan-400 border border-slate-700 font-semibold"
              : "text-slate-400 hover:text-slate-200 hover:bg-slate-800/40"
          }`}
        >
          <Orbit className="w-4 h-4" />
          <span>Planetary Cycles</span>
        </button>

        <button
          onClick={() => setActiveTab("nakshatras")}
          className={`flex items-center space-x-2 px-4 py-2.5 rounded-lg transition ${
            activeTab === "nakshatras"
              ? "bg-slate-800 text-cyan-400 border border-slate-700 font-semibold"
              : "text-slate-400 hover:text-slate-200 hover:bg-slate-800/40"
          }`}
        >
          <Sparkles className="w-4 h-4" />
          <span>Nakshatras</span>
        </button>

        <button
          onClick={() => setActiveTab("alchemy")}
          className={`flex items-center space-x-2 px-4 py-2.5 rounded-lg transition ${
            activeTab === "alchemy"
              ? "bg-slate-800 text-cyan-400 border border-slate-700 font-semibold"
              : "text-slate-400 hover:text-slate-200 hover:bg-slate-800/40"
          }`}
        >
          <Flame className="w-4 h-4" />
          <span>Hermetic Alchemy</span>
        </button>

        <button
          onClick={() => setActiveTab("database")}
          className={`flex items-center space-x-2 px-4 py-2.5 rounded-lg transition ${
            activeTab === "database"
              ? "bg-slate-800 text-cyan-400 border border-slate-700 font-semibold"
              : "text-slate-400 hover:text-slate-200 hover:bg-slate-800/40"
          }`}
        >
          <Database className="w-4 h-4" />
          <span>BoltDB Key/Value</span>
        </button>
      </div>

      {/* Tab Contents */}
      {activeTab === "overview" && (
        <div className="space-y-6">
          {/* Key Metrics */}
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div className="bg-[#111928] border border-slate-800 rounded-xl p-4">
              <div className="text-xs font-mono text-slate-400 uppercase">Mahadasha Lord</div>
              <div className="text-lg font-bold text-emerald-400 mt-1">
                {overview?.mahadasha_lord || "Mercury (17-Year Cycle)"}
              </div>
              <div className="text-xs text-slate-500 font-mono mt-1">Primary Ruler</div>
            </div>
            <div className="bg-[#111928] border border-slate-800 rounded-xl p-4">
              <div className="text-xs font-mono text-slate-400 uppercase">Root Frequency</div>
              <div className="text-lg font-bold text-cyan-400 mt-1">
                {overview?.root_frequency_hz || 141.27} Hz
              </div>
              <div className="text-xs text-slate-500 font-mono mt-1">Mercury Resonant Tone</div>
            </div>
            <div className="bg-[#111928] border border-slate-800 rounded-xl p-4">
              <div className="text-xs font-mono text-slate-400 uppercase">Total Cycle Span</div>
              <div className="text-lg font-bold text-purple-400 mt-1">
                {overview?.total_years || 120} Solar Years
              </div>
              <div className="text-xs text-slate-500 font-mono mt-1">Vimshottari Dasha System</div>
            </div>
          </div>

          {/* Planetary Table */}
          <div className="bg-[#111928] border border-slate-800 rounded-xl overflow-hidden shadow">
            <div className="px-5 py-4 border-b border-slate-800 flex items-center justify-between">
              <div className="text-sm font-semibold text-slate-200">Planetary Rulers & Resonances</div>
              <span className="text-xs font-mono text-slate-400">9 Planetary Archetypes</span>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead className="bg-slate-900/80 text-slate-400 uppercase font-mono border-b border-slate-800">
                  <tr>
                    <th className="px-4 py-3">Planet (Graha)</th>
                    <th className="px-4 py-3">Duration</th>
                    <th className="px-4 py-3">Frequency</th>
                    <th className="px-4 py-3">Element</th>
                    <th className="px-4 py-3">Chakra Center</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-800/60 font-mono">
                  {overview?.planets.map((p) => (
                    <tr key={p.name} className="hover:bg-slate-800/30 transition">
                      <td className="px-4 py-3 font-semibold text-slate-100 flex items-center space-x-2">
                        <span className="w-2.5 h-2.5 rounded-full" style={{ backgroundColor: p.color_hex }} />
                        <span>{p.name}</span>
                        <span className="text-slate-500 text-[10px]">({p.sanskrit_name})</span>
                      </td>
                      <td className="px-4 py-3 text-slate-300">{p.duration_years} Years</td>
                      <td className="px-4 py-3 text-cyan-400 font-bold">{p.frequency_hz} Hz</td>
                      <td className="px-4 py-3 text-slate-400">{p.element}</td>
                      <td className="px-4 py-3 text-slate-300">{p.chakra}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {activeTab === "nakshatras" && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {nakshatras.map((n) => (
            <div key={n.id} className="bg-[#111928] border border-slate-800 rounded-xl p-5 flex flex-col justify-between hover:border-cyan-800/80 transition shadow-lg">
              <div>
                <div className="flex items-center justify-between">
                  <span className="font-mono text-xs text-cyan-400 font-bold uppercase">{n.id}</span>
                  <span className="px-2 py-0.5 rounded bg-cyan-950 text-cyan-300 text-xs font-mono border border-cyan-800">
                    {n.frequency} Hz
                  </span>
                </div>
                <h3 className="text-base font-bold text-white mt-2">{n.name}</h3>
                <div className="text-xs font-mono text-slate-400 mt-1">{n.range}</div>

                <div className="mt-4 space-y-2 text-xs">
                  <div className="bg-slate-900/90 p-2.5 rounded border border-slate-800">
                    <span className="text-slate-500 font-mono block text-[10px] uppercase">Symbol</span>
                    <span className="text-slate-200 font-medium">{n.symbol}</span>
                  </div>
                  <div className="bg-slate-900/90 p-2.5 rounded border border-slate-800">
                    <span className="text-slate-500 font-mono block text-[10px] uppercase">Presiding Deity</span>
                    <span className="text-slate-200 font-medium">{n.deity}</span>
                  </div>
                </div>
              </div>

              <div className="mt-4 pt-3 border-t border-slate-800/80 text-xs text-slate-400 leading-relaxed italic">
                "{n.quality}"
              </div>
            </div>
          ))}
        </div>
      )}

      {activeTab === "alchemy" && (
        <div className="space-y-6">
          <div className="bg-gradient-to-r from-cyan-950/40 via-[#111928] to-slate-900 p-5 rounded-xl border border-cyan-900/50 flex flex-col sm:flex-row items-center justify-between gap-4">
            <div>
              <div className="text-xs font-mono text-cyan-400 uppercase tracking-widest">Elemental Correspondence</div>
              <h2 className="text-lg font-bold text-white mt-1">Quicksilver (Liquid Mercury)</h2>
              <p className="text-xs text-slate-400 mt-1">The bridge between volatile spirit and fixed matter.</p>
            </div>
            <span className="px-3 py-1.5 rounded-lg bg-cyan-900/40 border border-cyan-700 text-cyan-300 font-mono text-xs">
              Hermetic & Alchemical Philosophy
            </span>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {axioms.map((ax, idx) => (
              <div key={ax.title} className="bg-[#111928] border border-slate-800 rounded-xl p-4 hover:border-slate-700 transition">
                <div className="flex items-center space-x-2 text-xs font-mono text-cyan-400">
                  <span className="bg-slate-800 text-slate-400 w-5 h-5 rounded flex items-center justify-center text-[10px]">
                    {idx + 1}
                  </span>
                  <span className="font-semibold">{ax.title}</span>
                </div>
                <div className="text-xs text-slate-300 mt-2 font-serif italic leading-relaxed pl-7">
                  "{ax.text}"
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {activeTab === "database" && (
        <MetaEditor />
      )}
    </div>
  );
}

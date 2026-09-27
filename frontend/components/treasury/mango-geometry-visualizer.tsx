"use client";

import React from "react";
import { Sparkles, RefreshCw } from "lucide-react";
import { AmraGeometryData } from "@/lib/types/treasury";
import EsotericShadowAlchemy from "./esoteric-shadow-alchemy";

interface MangoGeometryVisualizerProps {
  geoBelly: number;
  geoHook: number;
  geoShadow: number;
  geoHeat: number;
  geoData: AmraGeometryData | null;
  loadingGeo: boolean;
  onBellyChange: (val: number) => void;
  onHookChange: (val: number) => void;
  onShadowChange: (val: number) => void;
  onHeatChange: (val: number) => void;
  onPreset: (belly: number, hook: number, shadow: number, heat: number) => void;
  onReload: () => void;
}

export default function MangoGeometryVisualizer({
  geoBelly,
  geoHook,
  geoShadow,
  geoHeat,
  geoData,
  loadingGeo,
  onBellyChange,
  onHookChange,
  onShadowChange,
  onHeatChange,
  onPreset,
  onReload,
}: MangoGeometryVisualizerProps) {
  return (
    <div className="space-y-6 animate-in fade-in duration-300">
      {/* Header Bar */}
      <div className="p-4 rounded-xl bg-slate-900/80 border border-slate-800 flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <div className="flex items-center space-x-2">
            <span className="p-1.5 rounded-lg bg-amber-500/10 border border-amber-500/30 text-amber-400">
              <Sparkles className="w-5 h-5" />
            </span>
            <div>
              <h3 className="text-base font-bold text-white tracking-tight flex items-center gap-2">
                <span>Sacred Geometry of Āmra &amp; Shadow Alchemy</span>
                <span className="text-xs font-mono px-2 py-0.5 rounded bg-amber-500/10 text-amber-300 border border-amber-500/20 font-normal">
                  Jnana-Phala • Kairi Calculus
                </span>
              </h3>
              <p className="text-xs text-slate-400 mt-0.5">
                {geoData?.philosophy?.summary || "Parametric Kairi curve, indestructible Bīja seed, and shadow transmutation."}
              </p>
            </div>
          </div>
        </div>

        <div className="flex items-center space-x-2">
          <button
            onClick={() => onPreset(125, 35, 0.25, 0.85)}
            className="px-2.5 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-mono border border-slate-700 transition"
          >
            Canonical Core
          </button>
          <button
            onClick={() => onPreset(140, 45, 0.85, 0.45)}
            className="px-2.5 py-1.5 rounded-lg bg-indigo-950/60 hover:bg-indigo-900/60 text-indigo-300 text-xs font-mono border border-indigo-700/50 transition"
          >
            Manthan (Churning)
          </button>
          <button
            onClick={() => onPreset(130, 40, 0.40, 1.0)}
            className="px-2.5 py-1.5 rounded-lg bg-emerald-950/60 hover:bg-emerald-900/60 text-emerald-300 text-xs font-mono border border-emerald-700/50 transition"
          >
            Pakva (Ripe Nectar)
          </button>
          <button
            onClick={onReload}
            className="p-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-400 hover:text-white border border-slate-700 transition"
            title="Reload from API"
          >
            <RefreshCw className={`w-4 h-4 ${loadingGeo ? "animate-spin" : ""}`} />
          </button>
        </div>
      </div>

      {/* Canvas & Controls Layout */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
        {/* Left: SVG Canvas (7 cols) */}
        <div className="lg:col-span-7 bg-slate-900/90 border border-slate-800 rounded-2xl p-6 shadow-2xl flex flex-col items-center relative overflow-hidden">
          <div className="w-full flex justify-between items-center mb-4 text-xs font-mono text-slate-400">
            <span className="flex items-center gap-1.5">
              <span className="w-2 h-2 rounded-full bg-amber-400 animate-ping"></span>
              <span>Parametric Engine [/api/v1/amra/geometry]</span>
            </span>
            <span className="text-slate-500">
              Arc Length: {geoData?.body?.arc_length || "calculating..."}px
            </span>
          </div>

          {/* Responsive SVG Viewer */}
          <div className="w-full aspect-square max-w-[460px] relative flex items-center justify-center">
            <svg viewBox="-250 -250 500 500" className="w-full h-full drop-shadow-2xl">
              <defs>
                <radialGradient id="amraBodyGlow" cx="45%" cy="55%" r="65%">
                  <stop offset="0%" stopColor="#fef08a" stopOpacity="0.95" />
                  <stop offset="45%" stopColor="#f59e0b" stopOpacity="0.90" />
                  <stop offset="80%" stopColor="#b45309" stopOpacity="0.80" />
                  <stop offset="100%" stopColor="#451a03" stopOpacity="0.95" />
                </radialGradient>

                <radialGradient id="amraShadowVortex" cx="50%" cy="50%" r="50%">
                  <stop offset="0%" stopColor="#818cf8" stopOpacity="0.75" />
                  <stop offset="60%" stopColor="#3730a3" stopOpacity="0.5" />
                  <stop offset="100%" stopColor="#0f172a" stopOpacity="0" />
                </radialGradient>

                <radialGradient id="amraBijaGlow" cx="50%" cy="50%" r="50%">
                  <stop offset="0%" stopColor="#ffffff" stopOpacity="0.98" />
                  <stop offset="40%" stopColor="#fef3c7" stopOpacity="0.85" />
                  <stop offset="80%" stopColor="#d97706" stopOpacity="0.6" />
                  <stop offset="100%" stopColor="#78350f" stopOpacity="0" />
                </radialGradient>
              </defs>

              {/* Sacred Mandala Concentric Rings & Crosshairs */}
              <circle cx="0" cy="0" r="220" fill="none" stroke="#334155" strokeDasharray="4 6" opacity="0.4" />
              <circle cx="0" cy="0" r="160" fill="none" stroke="#334155" strokeDasharray="2 4" opacity="0.3" />
              <circle cx="0" cy="0" r="80" fill="none" stroke="#334155" strokeDasharray="2 2" opacity="0.3" />
              <line x1="-230" y1="0" x2="230" y2="0" stroke="#334155" opacity="0.25" strokeDasharray="2 4" />
              <line x1="0" y1="-230" x2="0" y2="230" stroke="#334155" opacity="0.25" strokeDasharray="2 4" />

              {/* Asuric Shadow Field & Demons */}
              <g opacity={Math.min(1.0, geoShadow * 1.2)}>
                <circle cx="-60" cy="70" r="90" fill="url(#amraShadowVortex)" />
                <circle cx="80" cy="-30" r="80" fill="url(#amraShadowVortex)" />

                {/* 6 Arishadvarga Vectors */}
                {geoData?.transmutation?.demons?.map((demon) => (
                  <g key={demon.index}>
                    <line
                      x1="0"
                      y1="20"
                      x2={demon.coord.x}
                      y2={demon.coord.y}
                      stroke="#818cf8"
                      strokeWidth="1"
                      strokeDasharray="3 3"
                      opacity="0.6"
                    />
                    <circle
                      cx={demon.coord.x}
                      cy={demon.coord.y}
                      r="3.5"
                      fill="#818cf8"
                      stroke="#c7d2fe"
                      strokeWidth="1"
                    />
                  </g>
                ))}
              </g>

              {/* 5 Mango Leaves (Āmra-Pallava) */}
              {geoData?.leaves?.map((leaf) => (
                <path
                  key={leaf.index}
                  d={leaf.svg_path}
                  fill="#065f46"
                  stroke="#34d399"
                  strokeWidth="1.3"
                  opacity="0.8"
                />
              ))}

              {/* The Mango Body (Āmra Rūpa) */}
              {geoData?.body?.svg_path && (
                <path
                  d={geoData.body.svg_path}
                  fill="url(#amraBodyGlow)"
                  stroke="#f59e0b"
                  strokeWidth="2.5"
                />
              )}

              {/* The Deathless Seed (Amara Bīja) */}
              {geoData?.bija?.svg_path && (
                <path
                  d={geoData.bija.svg_path}
                  fill="url(#amraBijaGlow)"
                  stroke="#fbbf24"
                  strokeWidth="1.5"
                  strokeDasharray="3 2"
                />
              )}

              {/* Apex Bindu */}
              <circle
                cx="0"
                cy={-175 * 0.82 + 30}
                r="4.5"
                fill="#fef08a"
                stroke="#d97706"
                strokeWidth="2"
              />
            </svg>
          </div>

          {/* Status Bar */}
          <div className="w-full mt-4 pt-4 border-t border-slate-800 grid grid-cols-3 gap-3 text-center text-xs">
            <div className="bg-slate-950/80 p-2.5 rounded-lg border border-slate-800">
              <span className="block text-slate-500 text-[10px] font-mono">GOLDEN RATIO Φ</span>
              <span className="font-mono text-amber-300 font-semibold">
                {geoData?.body?.golden_ratio_phi?.toFixed(6) || "1.618033"}
              </span>
            </div>
            <div className="bg-slate-950/80 p-2.5 rounded-lg border border-slate-800">
              <span className="block text-slate-500 text-[10px] font-mono">CHURNING BALANCE</span>
              <span className="font-mono text-indigo-300 font-semibold">
                {geoData?.transmutation?.devic_ratio?.toFixed(0) || "50"}% Light / {geoData?.transmutation?.asuric_ratio?.toFixed(0) || "50"}% Shadow
              </span>
            </div>
            <div className="bg-slate-950/80 p-2.5 rounded-lg border border-slate-800">
              <span className="block text-slate-500 text-[10px] font-mono">ALCHEMICAL STATE</span>
              <span
                className={`font-mono font-semibold ${
                  geoData?.transmutation?.state === "pakva"
                    ? "text-emerald-300"
                    : geoData?.transmutation?.state === "manthan"
                    ? "text-amber-300"
                    : "text-rose-300"
                }`}
              >
                {geoData?.transmutation?.state === "pakva"
                  ? "Pakva (Ripe Nectar)"
                  : geoData?.transmutation?.state === "manthan"
                  ? "Manthan (Churning)"
                  : "Āma (Raw Acid)"}
              </span>
            </div>
          </div>
        </div>

        {/* Right: Parametric Sliders (5 cols) */}
        <div className="lg:col-span-5 space-y-6">
          <div className="bg-slate-900/90 border border-slate-800 rounded-2xl p-6 shadow-xl space-y-5">
            <h4 className="text-sm font-semibold text-white flex items-center justify-between border-b border-slate-800 pb-3">
              <span>Parametric Calculus</span>
              <span className="text-xs font-mono text-amber-400">Real-Time Synthesis</span>
            </h4>

            {/* Slider 1: Belly Fullness */}
            <div className="space-y-1.5">
              <div className="flex justify-between text-xs">
                <label className="text-slate-300 font-medium">Belly Fullness (Garbha • Receptivity)</label>
                <span className="font-mono text-amber-400">{geoBelly}</span>
              </div>
              <input
                type="range"
                min="80"
                max="170"
                value={geoBelly}
                onChange={(e) => onBellyChange(parseFloat(e.target.value))}
                className="w-full accent-amber-500 bg-slate-800 rounded-lg cursor-pointer h-1.5"
              />
              <p className="text-[11px] text-slate-500">Governs lateral capacity to store digested experience and creative harvest.</p>
            </div>

            {/* Slider 2: Apex Crest Hook */}
            <div className="space-y-1.5">
              <div className="flex justify-between text-xs">
                <label className="text-slate-300 font-medium">Crest Hook Deflection (Pratyāhāra • Inward Bow)</label>
                <span className="font-mono text-amber-400">{geoHook}</span>
              </div>
              <input
                type="range"
                min="10"
                max="65"
                step="1"
                value={geoHook}
                onChange={(e) => onHookChange(parseFloat(e.target.value))}
                className="w-full accent-amber-500 bg-slate-800 rounded-lg cursor-pointer h-1.5"
              />
              <p className="text-[11px] text-slate-500">The sacred inflection: bowing inward towards the source rather than egoic expansion.</p>
            </div>

            {/* Slider 3: Shadow Churning */}
            <div className="space-y-1.5">
              <div className="flex justify-between text-xs">
                <label className="text-slate-300 font-medium">Inner Shadow Integration (Asuric Tension)</label>
                <span className="font-mono text-indigo-400">{geoShadow.toFixed(2)}</span>
              </div>
              <input
                type="range"
                min="0"
                max="1"
                step="0.05"
                value={geoShadow}
                onChange={(e) => onShadowChange(parseFloat(e.target.value))}
                className="w-full accent-indigo-500 bg-slate-800 rounded-lg cursor-pointer h-1.5"
              />
              <p className="text-[11px] text-slate-500">The counter-weight torque of unresolved shadow required to churn the ocean.</p>
            </div>

            {/* Slider 4: Ripening Heat */}
            <div className="space-y-1.5">
              <div className="flex justify-between text-xs">
                <label className="text-slate-300 font-medium">Ripening Fire (Surya Agni • Solar Awareness)</label>
                <span className="font-mono text-emerald-400">{geoHeat.toFixed(2)}</span>
              </div>
              <input
                type="range"
                min="0.1"
                max="1.0"
                step="0.05"
                value={geoHeat}
                onChange={(e) => onHeatChange(parseFloat(e.target.value))}
                className="w-full accent-emerald-500 bg-slate-800 rounded-lg cursor-pointer h-1.5"
              />
              <p className="text-[11px] text-slate-500">Transmutes caustic acids into sweet golden nectar without destroying the essence.</p>
            </div>
          </div>

          {/* Arishadvarga Subcomponent */}
          <EsotericShadowAlchemy
            demons={geoData?.transmutation?.demons}
            philosophy={geoData?.philosophy || null}
          />
        </div>
      </div>

      {/* Mathematical Equations Card */}
      {geoData?.mathematical_formulas && Object.keys(geoData.mathematical_formulas).length > 0 && (
        <div className="bg-slate-900/90 border border-slate-800 rounded-2xl p-6 shadow-xl space-y-3">
          <h4 className="text-sm font-semibold text-amber-300 font-mono flex items-center gap-2">
            <Sparkles className="w-4 h-4 text-amber-400" />
            <span>Mathematical Formulations of the Esoteric (Storehouse Engine)</span>
          </h4>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 text-xs font-mono">
            {Object.entries(geoData.mathematical_formulas).map(([formulaKey, formulaVal]) => {
              const label = formulaKey.replace(/_/g, " ").toUpperCase();
              return (
                <div key={formulaKey} className="bg-slate-950 p-3 rounded-xl border border-slate-800 space-y-1">
                  <span className="text-amber-400 text-[11px] font-semibold block">{label}</span>
                  <div className="text-slate-300 break-all text-[11px]">
                    {formulaVal}
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}

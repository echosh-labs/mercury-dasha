"use client";

import React, { useState, useEffect } from "react";
import { 
  Flame, 
  Sparkles, 
  Layers, 
  Orbit, 
  Compass, 
  ShieldCheck, 
  Zap, 
  Radio, 
  Feather, 
  BookOpen 
} from "lucide-react";
import {
  fetchAlchemyLazy,
  fetchActiveProfile,
  type AlchemyData,
  type SacredMetal,
  type HermeticAxiom,
  type MagnumOpusStage,
  type LiveAlignment,
  type SovereignProfile,
  type SymbioticResonance
} from "@/lib/storehouse";

export default function AlchemicalLab() {
  const [alchemy, setAlchemy] = useState<AlchemyData | null>(null);
  const [profile, setProfile] = useState<SovereignProfile | null>(null);
  const [selectedMetal, setSelectedMetal] = useState<SacredMetal | null>(null);
  const [selectedAxiom, setSelectedAxiom] = useState<HermeticAxiom | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [transmuting, setTransmuting] = useState<boolean>(false);

  useEffect(() => {
    let mounted = true;

    const loadData = async (force = false) => {
      try {
        const [alchemyData, profileData] = await Promise.all([
          fetchAlchemyLazy(force),
          fetchActiveProfile(undefined, force)
        ]);

        if (!mounted) return;
        if (alchemyData) {
          setAlchemy(alchemyData);
          if (alchemyData.sacred_metals && alchemyData.sacred_metals.length > 0) {
            setSelectedMetal((prev) => prev || alchemyData.sacred_metals[0]);
          }
          if (alchemyData.hermetic_axioms && alchemyData.hermetic_axioms.length > 0) {
            setSelectedAxiom((prev) => prev || alchemyData.hermetic_axioms[0]);
          }
        }
        if (profileData) {
          setProfile(profileData);
        }
      } catch (err) {
        console.error("Failed to load alchemical domain data:", err);
      } finally {
        if (mounted) setLoading(false);
      }
    };

    loadData(false);

    const handleHoraUpdated = () => {
      if (!mounted) return;
      setTransmuting(true);
      loadData(true).then(() => {
        setTimeout(() => {
          if (mounted) setTransmuting(false);
        }, 1200);
      });
    };

    window.addEventListener("mercury_hora_updated", handleHoraUpdated);
    window.addEventListener("mercury_settings_updated", () => loadData(true));
    return () => {
      mounted = false;
      window.removeEventListener("mercury_hora_updated", handleHoraUpdated);
      window.removeEventListener("mercury_settings_updated", () => loadData(true));
    };
  }, []);

  const live = alchemy?.live_alignment;
  const natalAlchemy = profile?.alchemy || alchemy?.natal_alchemy;
  const activeAlchemy = profile?.active_alchemy || alchemy?.active_alchemy;
  const resonance: SymbioticResonance | undefined = profile?.symbiotic_resonance || alchemy?.symbiotic_resonance;

  return (
    <div className="space-y-8">
      {/* Alchemical Laboratory Header */}
      <div className="bg-gradient-to-r from-[#111928] via-[#0d1624] to-[#111928] border border-cyan-900/50 p-6 rounded-2xl shadow-xl space-y-4">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-slate-800 pb-4">
          <div>
            <div className="flex items-center space-x-2">
              <Flame className="w-5 h-5 text-emerald-400" />
              <h2 className="text-xl font-bold text-white tracking-tight">
                Hermetic Alchemy Laboratory & Quicksilver Engine
              </h2>
              <span className="px-2 py-0.5 rounded text-[10px] font-mono bg-emerald-950/70 border border-emerald-800 text-emerald-300">
                Transmutation Core
              </span>
            </div>
            <p className="text-xs text-slate-400 mt-1">
              The sacred qualitative dimension of the Mercury Stack: 7 metals, 7 cosmic axioms, and the Magnum Opus stages.
            </p>
          </div>

          {live && (
            <div className={`flex items-center space-x-2 text-xs font-mono px-3 py-1.5 rounded-lg bg-slate-900 border transition-all duration-500 shadow ${
              transmuting ? "border-emerald-400 ring-2 ring-emerald-400/40 bg-emerald-950/30" : "border-emerald-800/60"
            }`}>
              <span className={`h-2 w-2 rounded-full ${transmuting ? "bg-cyan-400 animate-spin" : "bg-emerald-400 animate-ping"}`} />
              <span className="text-slate-400">Live Vessel State:</span>
              <span className="font-bold text-emerald-300">
                {live.active_metal.symbol} {live.active_metal.name}
              </span>
              <span className="text-slate-500">•</span>
              <span className="text-cyan-400">{live.magnum_opus_stage.name}</span>
              {transmuting && (
                <span className="ml-1 text-[10px] text-cyan-300 animate-pulse font-semibold">
                  [TRANSMUTING]
                </span>
              )}
            </div>
          )}
        </div>

        {/* Dual-Vessel Solve et Coagula Transmutation Crucible */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 pt-2">
          {/* Vessel A: Natal Fixed Anchor */}
          <div className="p-4 rounded-xl bg-slate-900/90 border border-cyan-900/60 flex flex-col justify-between space-y-3 font-mono text-xs shadow-inner">
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-[10px] uppercase tracking-wider text-cyan-400 font-bold">
                  Natal Anchor (Fixed)
                </span>
                <span className="px-2 py-0.5 rounded text-[10px] bg-slate-800 border border-slate-700 text-slate-300">
                  {profile?.name || "Sovereign Genesis"}
                </span>
              </div>

              {natalAlchemy ? (
                <div className="space-y-2 pt-1">
                  <div className="flex items-center space-x-3">
                    <span
                      className="text-2xl font-bold px-2.5 py-0.5 rounded-lg bg-slate-950 border border-slate-800"
                      style={{ color: natalAlchemy.sacred_metal?.color_hex || "#10b981" }}
                    >
                      {natalAlchemy.sacred_metal?.symbol}
                    </span>
                    <div>
                      <div className="font-bold text-slate-100 text-sm">
                        {natalAlchemy.sacred_metal?.name}
                      </div>
                      <div className="text-[10px] text-slate-400">
                        Governing Graha: {natalAlchemy.sacred_metal?.governing_graha}
                      </div>
                    </div>
                  </div>

                  <div className="text-[11px] text-slate-300 font-sans italic border-l-2 border-cyan-500/50 pl-2">
                    &ldquo;{natalAlchemy.governing_axiom?.title}&rdquo;
                  </div>

                  <div className="grid grid-cols-2 gap-2 pt-1 text-[10px] text-slate-400">
                    <div className="bg-slate-950/60 p-1.5 rounded border border-slate-800">
                      <span className="text-slate-500 block">Tattva / Dosha:</span>
                      <span className="text-emerald-300 font-semibold">
                        {profile?.astrology?.elemental_tattva || "Earth"}
                        {profile?.astrology?.ayurveda?.dosha ? ` • ${profile.astrology.ayurveda.dosha}` : ""}
                      </span>
                    </div>
                    <div className="bg-slate-950/60 p-1.5 rounded border border-slate-800">
                      <span className="text-slate-500 block">Vara / Chakra:</span>
                      <span className="text-cyan-300 font-semibold">
                        {profile?.astrology?.panchanga?.vara ? `${profile.astrology.panchanga.vara.split(" ")[0]}` : natalAlchemy.chakra_anchor?.split(" ")[0]}
                      </span>
                    </div>
                  </div>

                  <p className="text-[10px] text-slate-400 font-sans leading-tight pt-1">
                    &ldquo;{natalAlchemy.alchemical_motto}&rdquo;
                  </p>
                </div>
              ) : (
                <div className="text-slate-500 italic text-[11px] py-4">
                  Calculating natal alchemical anchor...
                </div>
              )}
            </div>
          </div>

          {/* Resonance Bridge: Solve et Coagula */}
          {(() => {
            const isResonant = !!(resonance && (resonance.is_janma_resonant || resonance.is_mahadasha_resonant || resonance.is_antardasha_resonant));
            return (
              <div className={`p-4 rounded-xl border flex flex-col justify-between space-y-3 font-mono text-xs transition-all ${
                isResonant
                  ? "bg-gradient-to-b from-amber-950/40 via-slate-900 to-slate-900 border-amber-500/60 shadow-lg"
                  : "bg-slate-900/60 border-slate-800"
              }`}>
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-[10px] uppercase tracking-wider text-amber-400 font-bold flex items-center space-x-1">
                      <Sparkles className="w-3 h-3 text-amber-400" />
                      <span>Solve et Coagula</span>
                    </span>
                    {isResonant ? (
                      <span className="px-2 py-0.5 rounded text-[10px] bg-amber-500/20 text-amber-300 border border-amber-500/40 font-bold animate-pulse">
                        RESONANT
                      </span>
                    ) : (
                      <span className="px-2 py-0.5 rounded text-[10px] bg-slate-800 text-slate-400 border border-slate-700">
                        TRANSIT
                      </span>
                    )}
                  </div>

                  {resonance ? (
                    <div className="space-y-2 pt-1">
                      <div className="text-center py-1 bg-slate-950/70 rounded-lg border border-slate-800/80">
                        <div className="text-[10px] text-slate-400">Resonance Tier</div>
                        <div className="font-bold text-amber-300 text-sm">{resonance.resonance_tier?.replace(/_/g, " ")}</div>
                        <div className="text-[10px] text-slate-500 font-mono mt-0.5">
                          Hora Lord: <strong className="text-cyan-300">{resonance.transit_hora_planet_name}</strong>
                        </div>
                      </div>

                      <div className="text-[11px] text-slate-300 font-sans leading-snug">
                        {resonance.transmutation_guidance}
                      </div>

                      <div className="text-[10px] text-slate-400 font-sans italic border-t border-slate-800/80 pt-1">
                        Transit Vessel: {resonance.transit_metal?.symbol} {resonance.transit_metal?.name} • &ldquo;{resonance.transit_axiom?.title}&rdquo;
                      </div>
                    </div>
                  ) : (
                    <div className="text-slate-500 italic text-[11px] py-4 text-center">
                      Synthesizing symbiotic bridge...
                    </div>
                  )}
                </div>
              </div>
            );
          })()}

          {/* Vessel B: Live Sky Transmutation (Volatile) */}
          <div className="p-4 rounded-xl bg-slate-900/90 border border-emerald-900/60 flex flex-col justify-between space-y-3 font-mono text-xs shadow-inner">
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-[10px] uppercase tracking-wider text-emerald-400 font-bold">
                  Sky Vessel (Volatile)
                </span>
                <span className="px-2 py-0.5 rounded text-[10px] bg-emerald-950/60 border border-emerald-800 text-emerald-300">
                  Live Hora
                </span>
              </div>

              {live ? (
                <div className="space-y-2 pt-1">
                  <div className="flex items-center space-x-3">
                    <span
                      className="text-2xl font-bold px-2.5 py-0.5 rounded-lg bg-slate-950 border border-slate-800"
                      style={{ color: live.active_metal.color_hex }}
                    >
                      {live.active_metal.symbol}
                    </span>
                    <div>
                      <div className="font-bold text-slate-100 text-sm">
                        {live.active_metal.name}
                      </div>
                      <div className="text-[10px] text-slate-400">
                        Hora Lord: {live.active_metal.governing_planet} ({live.active_metal.governing_graha})
                      </div>
                    </div>
                  </div>

                  <div className="text-[11px] text-slate-300 font-sans italic border-l-2 border-emerald-500/50 pl-2">
                    &ldquo;{live.governing_axiom.title}&rdquo;
                  </div>

                  <div className="grid grid-cols-2 gap-2 pt-1 text-[10px] text-slate-400">
                    <div className="bg-slate-950/60 p-1.5 rounded border border-slate-800">
                      <span className="text-slate-500 block">Stage:</span>
                      <span className="text-emerald-300 font-semibold">{live.magnum_opus_stage.name}</span>
                    </div>
                    <div className="bg-slate-950/60 p-1.5 rounded border border-slate-800">
                      <span className="text-slate-500 block">Chakra:</span>
                      <span className="text-cyan-300 font-semibold">{live.active_metal.chakra_center.split(" ")[0]}</span>
                    </div>
                  </div>

                  <p className="text-[10px] text-slate-400 font-sans italic leading-tight pt-1">
                    &ldquo;{live.alchemical_motto}&rdquo;
                  </p>
                </div>
              ) : (
                <div className="text-slate-500 italic text-[11px] py-4">
                  Observing celestial hora...
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Magnum Opus Transmutation Sequence */}
      {alchemy?.magnum_opus_stages && (
        <div className="bg-[#111928] border border-slate-800 rounded-2xl p-6 shadow-xl space-y-4">
          <div className="flex items-center justify-between border-b border-slate-800 pb-3">
            <div className="flex items-center space-x-2">
              <Layers className="w-5 h-5 text-cyan-400" />
              <h3 className="text-base font-bold text-white">
                The Seven Stages of the Magnum Opus
              </h3>
            </div>
            <span className="text-xs font-mono text-slate-400">Great Alchemical Work</span>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 lg:grid-cols-7 gap-2 font-mono text-xs">
            {alchemy.magnum_opus_stages.map((stage) => {
              const isCurrent = live?.magnum_opus_stage.name === stage.name;
              return (
                <div
                  key={stage.index}
                  className={`p-3 rounded-xl border transition flex flex-col justify-between ${
                    isCurrent
                      ? "bg-emerald-950/40 border-emerald-500 shadow-md scale-102"
                      : "bg-slate-900/60 border-slate-800 hover:border-slate-700"
                  }`}
                >
                  <div>
                    <div className="flex items-center justify-between text-[10px] text-slate-500 mb-1">
                      <span>0{stage.index}.</span>
                      <span className="text-cyan-400 uppercase">{stage.element}</span>
                    </div>
                    <div className="font-bold text-slate-200 text-sm">{stage.name}</div>
                    <div className="text-[10px] text-slate-400 italic mb-2">{stage.latin_name}</div>
                    <p className="text-[11px] text-slate-400 font-sans leading-tight">
                      {stage.description}
                    </p>
                  </div>
                  {isCurrent && (
                    <div className="mt-2 pt-1 border-t border-emerald-500/30 text-[9px] text-emerald-400 font-bold uppercase text-center">
                      Active Stage
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* The 7 Sacred Metals & Quicksilver Core */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Metals Selector List */}
        <div className="bg-[#111928] border border-slate-800 rounded-2xl p-6 shadow-xl space-y-4">
          <div className="flex items-center justify-between border-b border-slate-800 pb-3">
            <h3 className="text-base font-bold text-white flex items-center space-x-2">
              <Zap className="w-4 h-4 text-amber-400" />
              <span>The 7 Sacred Metals</span>
            </h3>
            <span className="text-xs font-mono text-slate-400">Planetary Vessels</span>
          </div>

          <div className="space-y-2 font-mono text-xs">
            {alchemy?.sacred_metals.map((metal) => {
              const isSelected = selectedMetal?.id === metal.id;
              const isLive = live?.active_metal.id === metal.id;

              return (
                <div
                  key={metal.id}
                  onClick={() => setSelectedMetal(metal)}
                  className={`p-3 rounded-xl border cursor-pointer transition flex items-center justify-between ${
                    isSelected
                      ? "bg-slate-800/90 border-cyan-500 text-white shadow-sm"
                      : "bg-slate-900/50 border-slate-800/80 text-slate-400 hover:bg-slate-800/40"
                  }`}
                >
                  <div className="flex items-center space-x-3">
                    <span
                      className="text-lg font-bold px-2 py-0.5 rounded bg-slate-950 border border-slate-800"
                      style={{ color: metal.color_hex }}
                    >
                      {metal.symbol}
                    </span>
                    <div>
                      <div className="font-semibold text-slate-200">{metal.name}</div>
                      <div className="text-[10px] text-slate-500">
                        {metal.governing_graha} • {metal.chakra_center.split(" ")[0]}
                      </div>
                    </div>
                  </div>

                  {isLive && (
                    <span className="px-2 py-0.5 rounded text-[9px] font-bold bg-emerald-500/20 text-emerald-400 border border-emerald-500/30">
                      LIVE HORA
                    </span>
                  )}
                </div>
              );
            })}
          </div>
        </div>

        {/* Selected Metal Hologram Card */}
        {selectedMetal && (
          <div className="lg:col-span-2 bg-[#111928] border border-slate-800 rounded-2xl p-6 shadow-xl space-y-5">
            <div className="flex items-center justify-between border-b border-slate-800 pb-3">
              <div className="flex items-center space-x-3">
                <span
                  className="text-3xl font-bold px-3 py-1 rounded-xl bg-slate-900 border border-slate-700"
                  style={{ color: selectedMetal.color_hex }}
                >
                  {selectedMetal.symbol}
                </span>
                <div>
                  <h4 className="text-xl font-bold text-white tracking-tight">{selectedMetal.name}</h4>
                  <div className="text-xs font-mono text-slate-400">
                    Governed by {selectedMetal.governing_planet} ({selectedMetal.governing_graha})
                  </div>
                </div>
              </div>
              <span
                className="px-2.5 py-1 rounded-full text-xs font-mono font-bold border"
                style={{
                  color: selectedMetal.color_hex,
                  borderColor: `${selectedMetal.color_hex}66`,
                  backgroundColor: `${selectedMetal.color_hex}15`,
                }}
              >
                {selectedMetal.conductivity_rating}
              </span>
            </div>

            <div className="space-y-4 font-mono text-xs">
              <div className="p-3 bg-slate-900/80 border border-slate-800 rounded-xl space-y-1">
                <div className="text-slate-500 uppercase text-[10px]">Alchemical Essence</div>
                <div className="text-slate-200 text-sm font-sans">{selectedMetal.alchemical_essence}</div>
              </div>

              <div className="p-3 bg-slate-900/80 border border-slate-800 rounded-xl space-y-1">
                <div className="text-slate-500 uppercase text-[10px]">Transmutation Role in Great Work</div>
                <div className="text-emerald-300 font-sans">{selectedMetal.transmutation_role}</div>
              </div>

              <div className="p-3 bg-slate-900/80 border border-slate-800 rounded-xl space-y-1">
                <div className="text-slate-500 uppercase text-[10px]">Quicksilver (Mercury) Affinity</div>
                <div className="text-cyan-300 font-sans italic">{selectedMetal.quicksilver_affinity}</div>
              </div>

              <div className="flex items-center justify-between pt-2 border-t border-slate-800 text-slate-400">
                <span>Chakra Center: <strong className="text-slate-200">{selectedMetal.chakra_center}</strong></span>
                <span>Color Frequency: <strong style={{ color: selectedMetal.color_hex }}>{selectedMetal.color_hex}</strong></span>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* The 7 Hermetic Axioms & Foundations Narrative Backbone */}
      <div className="bg-[#111928] border border-slate-800 rounded-2xl p-6 shadow-xl space-y-4">
        <div className="flex items-center justify-between border-b border-slate-800 pb-3">
          <div className="flex items-center space-x-2">
            <BookOpen className="w-5 h-5 text-emerald-400" />
            <h3 className="text-base font-bold text-white">
              The Seven Hermetic Axioms & Foundations Storytelling Matrix
            </h3>
          </div>
          <span className="text-xs font-mono text-slate-400">Universal Laws</span>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 font-mono text-xs">
          {alchemy?.hermetic_axioms.map((axiom) => {
            const isLiveAxiom = live?.governing_axiom.number === axiom.number;

            return (
              <div
                key={axiom.number}
                className={`p-4 rounded-xl border space-y-3 transition flex flex-col justify-between ${
                  isLiveAxiom
                    ? "bg-cyan-950/30 border-cyan-500 shadow"
                    : "bg-slate-900/50 border-slate-800/80 hover:border-slate-700"
                }`}
              >
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-slate-500 font-bold">Principle 0{axiom.number}</span>
                    <span className="px-2 py-0.5 rounded text-[10px] bg-slate-800 text-cyan-300 border border-slate-700">
                      {axiom.governing_planet}
                    </span>
                  </div>

                  <h5 className="text-sm font-bold text-slate-200">{axiom.title}</h5>

                  <p className="text-xs text-slate-300 font-sans italic border-l-2 border-cyan-500/60 pl-2 py-0.5">
                    &ldquo;{axiom.canonical_text}&rdquo;
                  </p>
                </div>

                <div className="pt-2 border-t border-slate-800 space-y-1.5 text-[11px]">
                  <div className="text-slate-400 font-sans">
                    <strong className="text-slate-300">Esoteric Law: </strong>
                    {axiom.esoteric_meaning}
                  </div>
                  <div className="text-emerald-400/90 font-sans">
                    <strong className="text-emerald-300">Foundations Story Theme: </strong>
                    {axiom.foundations_theme}
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}

"use client";

import React from "react";
import { Sparkles } from "lucide-react";
import { ArishadvargaVector, EsotericPhilosophy } from "@/lib/types/treasury";

interface EsotericShadowAlchemyProps {
  demons: ArishadvargaVector[] | undefined;
  philosophy: EsotericPhilosophy | null;
}

export default function EsotericShadowAlchemy({
  demons,
  philosophy,
}: EsotericShadowAlchemyProps) {
  return (
    <div className="space-y-6">
      {/* Subtle Vedic Philosophical Harmony Card */}
      <div className="p-4 rounded-xl bg-gradient-to-r from-amber-950/20 via-slate-900/80 to-emerald-950/20 border border-amber-500/20 text-xs">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-amber-500/10 pb-2">
          <div className="flex items-center space-x-2">
            <Sparkles className="w-4 h-4 text-amber-400 shrink-0" />
            <span className="font-serif text-amber-300 font-semibold text-sm">
              {philosophy?.title || "Vedic Philosophy of Āmra (आम्र)"}
              {philosophy?.subtitle ? ` • ${philosophy.subtitle}` : ""}
            </span>
          </div>
          <span className="font-mono text-[10px] text-amber-400/80 bg-amber-500/10 px-2 py-0.5 rounded border border-amber-500/20 self-start sm:self-auto">
            Sovereign Fruition
          </span>
        </div>
        <div className="text-slate-400 text-xs leading-relaxed mt-2.5 font-sans">
          {philosophy?.karma_phala ? (
            <p className="space-y-1.5">
              <span className="text-slate-300 font-medium block">{philosophy.karma_phala}</span>
              {philosophy.purna_kumbha && (
                <span className="text-emerald-300/90 block">{philosophy.purna_kumbha}</span>
              )}
              {philosophy.summary && !philosophy.purna_kumbha && (
                <span className="text-emerald-300/90 block">{philosophy.summary}</span>
              )}
            </p>
          ) : (
            <span className="text-slate-500 italic">
              Hydrating sovereign Vedic philosophy from storehouse...
            </span>
          )}
        </div>
      </div>

      {/* The 6 Arishadvarga Demons & Transmutation Matrix */}
      <div className="bg-slate-900/90 border border-slate-800 rounded-2xl p-5 shadow-xl space-y-3">
        <h4 className="text-xs font-mono uppercase tracking-wider text-indigo-300 flex items-center justify-between">
          <span>Arishadvarga Transmutation</span>
          <span className="text-[10px] text-slate-500">6 Inner Adversaries &amp; Divine Virtues</span>
        </h4>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 text-xs">
          {demons && demons.map((d) => (
            <div
              key={d.index}
              className="p-3 rounded-xl bg-slate-950/70 border border-slate-800/80 hover:border-indigo-500/30 transition flex flex-col justify-between"
            >
              <div>
                <div className="font-semibold text-slate-200 text-xs flex items-center justify-between">
                  <span>{d.name}</span>
                  <span className="text-[9px] font-mono text-slate-500 bg-slate-900 px-1.5 py-0.5 rounded">
                    #{d.index}
                  </span>
                </div>
                {d.shadow_distortion && (
                  <div className="text-[10px] text-rose-400/80 font-mono mt-1">
                    Shadow: {d.shadow_distortion}
                  </div>
                )}
                {d.teaching && (
                  <p className="text-[11px] text-slate-400 mt-1.5 leading-relaxed font-sans">
                    {d.teaching}
                  </p>
                )}
              </div>
              <div className="text-[11px] text-emerald-400 font-mono font-medium mt-2 pt-2 border-t border-slate-800/60 flex items-center space-x-1.5">
                <span className="text-amber-400">➔</span>
                <span>{d.transmuted_as}</span>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

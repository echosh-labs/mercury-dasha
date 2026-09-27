"use client";

import React from "react";
import Link from "next/link";
import { ArrowLeft, Sparkles, ShieldCheck, ExternalLink } from "lucide-react";
import AlchemicalLab from "@/components/alchemical-lab";
import ChronoPulse from "@/components/chrono-pulse";

export default function AlchemyPage() {
  return (
    <div className="space-y-6">
      {/* Route Navigation & Hero Banner */}
      <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4 bg-gradient-to-r from-amber-950/40 via-slate-900 to-slate-900 p-6 rounded-2xl border border-amber-500/20 shadow-2xl">
        <div>
          <div className="flex items-center space-x-3">
            <Link
              href="/"
              className="px-3 py-1.5 rounded-lg bg-slate-800/80 hover:bg-slate-700 text-slate-300 hover:text-white border border-slate-700 text-xs font-mono flex items-center space-x-1.5 transition"
            >
              <ArrowLeft className="w-3.5 h-3.5" />
              <span>Observatory Hub</span>
            </Link>
            <span className="text-slate-600">/</span>
            <div className="flex items-center space-x-2">
              <Sparkles className="w-5 h-5 text-amber-400" />
              <h1 className="text-2xl font-bold text-white tracking-tight font-serif">
                Alchemical Laboratory
              </h1>
            </div>
            <span className="bg-amber-500/10 text-amber-400 border border-amber-500/20 text-xs font-mono px-2.5 py-0.5 rounded-full flex items-center space-x-1">
              <ShieldCheck className="w-3 h-3" />
              <span>HERMETIC TRANSMUTATION</span>
            </span>
          </div>
          <p className="text-xs text-slate-400 mt-2 max-w-3xl leading-relaxed">
            The 7 Sacred Metals, Hermetic Axioms of Thoth/Hermes, Quicksilver Metronome, and Alchemical Vessel Transmutation matrix for <span className="text-cyan-300 font-medium font-mono">echosh-labs</span>.
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-2 text-xs font-mono">
          <ChronoPulse />
          <a
            href="https://echosh-labs.com/compendium"
            target="_blank"
            rel="noreferrer"
            className="px-3 py-1.5 rounded-lg bg-amber-950/60 border border-amber-800 text-amber-300 hover:bg-amber-900/60 transition flex items-center space-x-1"
          >
            <span>Compendium</span>
            <ExternalLink className="w-3 h-3" />
          </a>
        </div>
      </div>

      {/* Alchemical Laboratory Workspace */}
      <AlchemicalLab />
    </div>
  );
}

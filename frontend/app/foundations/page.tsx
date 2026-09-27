"use client";

import React from "react";
import Link from "next/link";
import { ArrowLeft, Video, ShieldCheck, ExternalLink, Coins } from "lucide-react";
import FoundationsCreatorHub from "@/components/foundations/foundations-creator-hub";
import ChronoPulse from "@/components/chrono-pulse";

export default function FoundationsPage() {
  return (
    <div className="space-y-6">
      {/* Route Navigation & Hero Banner */}
      <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4 bg-gradient-to-r from-red-950/40 via-slate-900 to-slate-900 p-6 rounded-2xl border border-red-500/20 shadow-2xl">
        <div>
          <div className="flex flex-wrap items-center gap-2">
            <Link
              href="/"
              className="px-3 py-1.5 rounded-lg bg-slate-800/80 hover:bg-slate-700 text-slate-300 hover:text-white border border-slate-700 text-xs font-mono flex items-center space-x-1.5 transition"
            >
              <ArrowLeft className="w-3.5 h-3.5" />
              <span>Observatory Hub</span>
            </Link>

            <Link
              href="/treasury/"
              className="px-3 py-1.5 rounded-lg bg-emerald-950/40 hover:bg-emerald-900/50 text-emerald-300 hover:text-emerald-200 border border-emerald-500/30 text-xs font-mono flex items-center space-x-1.5 transition"
            >
              <Coins className="w-3.5 h-3.5 text-emerald-400" />
              <span>AMRA Treasury</span>
            </Link>

            <span className="text-slate-600 hidden sm:inline">/</span>

            <div className="flex items-center space-x-2">
              <Video className="w-5 h-5 text-red-400" />
              <h1 className="text-2xl font-bold text-white tracking-tight font-serif">
                Foundations Studio
              </h1>
            </div>

            <span className="bg-red-500/10 text-red-400 border border-red-500/20 text-xs font-mono px-2.5 py-0.5 rounded-full flex items-center space-x-1">
              <ShieldCheck className="w-3 h-3" />
              <span>SOVEREIGN BROADCASTER</span>
            </span>
          </div>

          <p className="text-xs text-slate-400 mt-2 max-w-3xl leading-relaxed">
            Autonomous YouTube Video Pipeline, Multi-Format Timeline Compiler, Channel Identity Telemetry, and Audience Feedback Loops for <span className="text-cyan-300 font-medium font-mono">echosh-labs</span>.
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-2 text-xs font-mono">
          <ChronoPulse />
          <a
            href="https://echosh-labs.com/compendium"
            target="_blank"
            rel="noreferrer"
            className="px-3 py-1.5 rounded-lg bg-red-950/60 border border-red-800 text-red-300 hover:bg-red-900/60 transition flex items-center space-x-1"
          >
            <span>Compendium</span>
            <ExternalLink className="w-3 h-3" />
          </a>
        </div>
      </div>

      {/* Sovereign Foundations Creator Hub */}
      <FoundationsCreatorHub />
    </div>
  );
}

"use client";

import React from "react";
import Link from "next/link";
import {
  DollarSign,
  TrendingUp,
  BarChart3,
  RefreshCw,
  Video,
  ArrowUpRight,
  ShieldCheck,
  CreditCard,
} from "lucide-react";
import { FinancialReport } from "@/lib/types/treasury";

interface YouTubeRevenueCardProps {
  finance: FinancialReport | null;
  syncingAMRA: boolean;
  authenticated: boolean;
  onSyncAMRA: () => void;
}

export default function YouTubeRevenueCard({
  finance,
  syncingAMRA,
  authenticated,
  onSyncAMRA,
}: YouTubeRevenueCardProps) {
  return (
    <div className="space-y-6">
      {/* Financial Accrual HUD */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Total Estimated Revenue */}
        <div className="p-4 rounded-xl bg-slate-900/90 border border-slate-800">
          <div className="text-[11px] font-mono uppercase text-red-400 flex items-center justify-between">
            <span>Estimated Total Revenue</span>
            <DollarSign className="w-4 h-4 text-red-400" />
          </div>
          <div className="text-2xl font-bold text-white mt-1">
            ${(finance?.total_estimated_revenue || 0).toLocaleString("en-US", { minimumFractionDigits: 2 })}
          </div>
          <div className="text-[10px] font-mono text-slate-500 mt-1">
            Currency: {finance?.currency || "USD"} • 30d Accrual
          </div>
        </div>

        {/* Ad Partner Revenue */}
        <div className="p-4 rounded-xl bg-slate-900/90 border border-slate-800">
          <div className="text-[11px] font-mono uppercase text-emerald-400 flex items-center justify-between">
            <span>Estimated Ad Revenue</span>
            <TrendingUp className="w-4 h-4 text-emerald-400" />
          </div>
          <div className="text-2xl font-bold text-emerald-400 mt-1">
            ${(finance?.total_estimated_ad_revenue || 0).toLocaleString("en-US", { minimumFractionDigits: 2 })}
          </div>
          <div className="text-[10px] font-mono text-slate-500 mt-1">
            Auction &amp; Reserved Placements
          </div>
        </div>

        {/* YouTube Red / Premium Partner Revenue */}
        <div className="p-4 rounded-xl bg-slate-900/90 border border-slate-800">
          <div className="text-[11px] font-mono uppercase text-cyan-400 flex items-center justify-between">
            <span>YouTube Premium Revenue</span>
            <CreditCard className="w-4 h-4 text-cyan-400" />
          </div>
          <div className="text-2xl font-bold text-cyan-300 mt-1">
            ${(finance?.total_estimated_red_revenue || 0).toLocaleString("en-US", { minimumFractionDigits: 2 })}
          </div>
          <div className="text-[10px] font-mono text-slate-500 mt-1">
            Subscriber Watch Time Allocation
          </div>
        </div>

        {/* Average CPM & Playback CPM */}
        <div className="p-4 rounded-xl bg-slate-900/90 border border-slate-800">
          <div className="text-[11px] font-mono uppercase text-amber-400 flex items-center justify-between">
            <span>Effective CPM</span>
            <BarChart3 className="w-4 h-4 text-amber-400" />
          </div>
          <div className="text-2xl font-bold text-amber-300 mt-1">
            ${(finance?.average_cpm || 0).toFixed(2)}
          </div>
          <div className="text-[10px] font-mono text-slate-500 mt-1">
            Playback CPM: ${(finance?.average_playback_based_cpm || 0).toFixed(2)}
          </div>
        </div>
      </div>

      {/* Ingestion Action & Cross-Route Link Banner */}
      <div className="p-4 rounded-xl bg-slate-900/80 border border-slate-800 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="flex items-center space-x-3">
          <span className="p-2 rounded-lg bg-red-500/10 text-red-400 border border-red-500/30">
            <Video className="w-5 h-5" />
          </span>
          <div>
            <h4 className="text-xs font-mono font-bold text-slate-200">
              Foundations Production &amp; Channel Telemetry
            </h4>
            <p className="text-[11px] text-slate-400">
              Video generation, upload pipelines, channel identity, and audience analytics have moved to their dedicated route.
            </p>
          </div>
        </div>

        <div className="flex items-center space-x-2">
          <Link
            href="/foundations/"
            className="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-mono flex items-center space-x-1.5 border border-slate-700 transition"
          >
            <span>Open Foundations Studio</span>
            <ArrowUpRight className="w-3.5 h-3.5 text-red-400" />
          </Link>

          <button
            onClick={onSyncAMRA}
            disabled={syncingAMRA || !authenticated}
            className="px-3.5 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white font-semibold text-xs transition flex items-center space-x-1.5 shadow disabled:opacity-50"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${syncingAMRA ? "animate-spin" : ""}`} />
            <span>Ingest Revenue into AMRA Ledger</span>
          </button>
        </div>
      </div>

      {/* Daily Revenue Feed Table */}
      <div className="p-5 rounded-xl bg-slate-900/80 border border-slate-800 space-y-4">
        <div className="flex items-center justify-between border-b border-slate-800 pb-3">
          <div>
            <h3 className="text-sm font-semibold text-white flex items-center space-x-2">
              <BarChart3 className="w-4 h-4 text-cyan-400" />
              <span>Daily Monetization &amp; Accrual Feed</span>
            </h3>
            <p className="text-xs text-slate-400">
              Audit-grade daily revenue time-series from YouTube Reporting API v2.
            </p>
          </div>
          <span className="text-xs font-mono text-emerald-400 bg-emerald-500/10 px-2 py-0.5 rounded border border-emerald-500/20">
            {finance?.monetized ? "Monetization Active" : "Unmonetized"}
          </span>
        </div>

        {finance?.status_message && (
          <div className="p-3 rounded-lg bg-amber-500/10 border border-amber-500/20 text-xs font-mono text-amber-300">
            ℹ {finance.status_message}
          </div>
        )}

        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs font-mono">
            <thead className="border-b border-slate-800 text-slate-400 uppercase text-[10px] bg-slate-950/40">
              <tr>
                <th className="py-2.5 px-3">Date</th>
                <th className="py-2.5 px-3">Gross Revenue</th>
                <th className="py-2.5 px-3">Ad Revenue</th>
                <th className="py-2.5 px-3">Red Revenue</th>
                <th className="py-2.5 px-3">CPM</th>
                <th className="py-2.5 px-3">Monetized Playbacks</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/60 text-slate-300">
              {finance?.daily_rows && finance.daily_rows.length > 0 ? (
                finance.daily_rows.map((row) => (
                  <tr key={row.day} className="hover:bg-slate-800/30 transition">
                    <td className="py-2.5 px-3 text-white font-medium">{row.day}</td>
                    <td className="py-2.5 px-3 text-emerald-400 font-bold">
                      ${row.gross_revenue.toFixed(2)}
                    </td>
                    <td className="py-2.5 px-3 text-slate-300">
                      ${row.estimated_ad_revenue.toFixed(2)}
                    </td>
                    <td className="py-2.5 px-3 text-cyan-300">
                      ${row.estimated_red_partner_revenue.toFixed(2)}
                    </td>
                    <td className="py-2.5 px-3 text-amber-300">
                      ${row.cpm.toFixed(2)}
                    </td>
                    <td className="py-2.5 px-3 text-slate-400">
                      {row.monetized_playbacks.toLocaleString()}
                    </td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan={6} className="py-8 text-center text-slate-500">
                    No monetization records returned. Connect your channel in Foundations Studio to authorize financial reporting scopes.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

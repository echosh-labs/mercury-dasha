"use client";

import React from "react";
import { TrendingUp, CreditCard, Video, History } from "lucide-react";
import { FinancialMetrics, FinancialReport, LedgerTransaction } from "@/lib/types/treasury";

interface TreasuryMetricCardsProps {
  metrics: FinancialMetrics | null;
  finance: FinancialReport | null;
  ledger: LedgerTransaction[];
}

export default function TreasuryMetricCards({
  metrics,
  finance,
  ledger,
}: TreasuryMetricCardsProps) {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mt-6">
      {/* Total Gross Ecosystem */}
      <div className="p-4 rounded-xl bg-slate-900/90 border border-emerald-500/30 shadow-inner">
        <div className="text-[11px] font-mono uppercase text-emerald-400 font-semibold flex items-center justify-between">
          <span>Total Ecosystem Gross</span>
          <TrendingUp className="w-4 h-4 text-emerald-400" />
        </div>
        <div className="text-2xl font-bold text-white mt-1">
          ${(metrics?.total_gross_ecosystem || 0).toLocaleString("en-US", { minimumFractionDigits: 2 })}
        </div>
        <div className="text-[10px] font-mono text-slate-400 mt-1">
          Karma-Phala: SaaS MRR + 30d Media Accrual
        </div>
      </div>

      {/* SaaS Subscriptions MRR */}
      <div className="p-4 rounded-xl bg-slate-900/90 border border-slate-800">
        <div className="text-[11px] font-mono uppercase text-cyan-400 font-semibold flex items-center justify-between">
          <span>SaaS Recurring (MRR)</span>
          <CreditCard className="w-4 h-4 text-cyan-400" />
        </div>
        <div className="text-2xl font-bold text-white mt-1">
          ${(metrics?.saas_mrr || 0).toLocaleString("en-US", { minimumFractionDigits: 2 })}
        </div>
        <div className="text-[10px] font-mono text-slate-400 mt-1">
          {metrics?.active_subscribers || 0} active subscribers
        </div>
      </div>

      {/* YouTube Ad & Partner Revenue */}
      <div className="p-4 rounded-xl bg-slate-900/90 border border-slate-800">
        <div className="text-[11px] font-mono uppercase text-red-400 font-semibold flex items-center justify-between">
          <span>YouTube Partner Ad Revenue</span>
          <Video className="w-4 h-4 text-red-400" />
        </div>
        <div className="text-2xl font-bold text-white mt-1">
          ${(metrics?.youtube_accrued_30d || finance?.total_estimated_revenue || 0).toLocaleString("en-US", { minimumFractionDigits: 2 })}
        </div>
        <div className="text-[10px] font-mono text-slate-400 mt-1">
          Avg CPM: ${(finance?.average_cpm || 0).toFixed(2)}
        </div>
      </div>

      {/* BoltDB Immutable Ledger Entries */}
      <div className="p-4 rounded-xl bg-slate-900/90 border border-slate-800">
        <div className="text-[11px] font-mono uppercase text-purple-400 font-semibold flex items-center justify-between">
          <span>Ledger Transactions</span>
          <History className="w-4 h-4 text-purple-400" />
        </div>
        <div className="text-2xl font-bold text-white mt-1">
          {(metrics?.ledger_total_transactions || ledger.length).toLocaleString()}
        </div>
        <div className="text-[10px] font-mono text-slate-400 mt-1">
          BoltDB bucket: amra_ledger
        </div>
      </div>
    </div>
  );
}

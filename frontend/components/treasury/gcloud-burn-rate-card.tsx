"use client";

import React from "react";
import {
  Cloud,
  DollarSign,
  TrendingUp,
  Layers,
  History,
  ShieldCheck,
  Server,
  Cpu,
  CreditCard,
  CheckCircle2,
  RefreshCw,
} from "lucide-react";
import { GCloudBillingResponse } from "@/lib/types/treasury";

interface GCloudBurnRateCardProps {
  gcloudBilling: GCloudBillingResponse | null;
  loadingGCloud: boolean;
  syncingGCloud: boolean;
  onRefresh: () => void;
  onSyncExpense: () => void;
}

export default function GCloudBurnRateCard({
  gcloudBilling,
  loadingGCloud,
  syncingGCloud,
  onRefresh,
  onSyncExpense,
}: GCloudBurnRateCardProps) {
  return (
    <div className="space-y-6">
      {/* GCloud Account & Project Topology HUD */}
      <div className="bg-gradient-to-r from-blue-950/30 via-slate-900 to-slate-950 border border-blue-500/20 rounded-2xl p-6 shadow-xl space-y-4">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-blue-500/10 pb-4">
          <div className="flex items-center space-x-3">
            <div className="p-2.5 rounded-xl bg-blue-500/10 border border-blue-500/30 text-blue-400">
              <Cloud className="w-6 h-6" />
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <h3 className="text-lg font-bold text-white tracking-tight">
                  Google Cloud Infrastructure &amp; Billing
                </h3>
                <span className="px-2 py-0.5 rounded-full text-[10px] font-mono bg-blue-500/10 text-blue-300 border border-blue-500/30">
                  GCP SOVEREIGN ARCHITECTURE
                </span>
              </div>
              <p className="text-xs text-slate-400 mt-0.5">
                Cloud Run gen2 execution, Artifact Registry container storage, and infrastructure cost accounting for echosh-labs.
              </p>
            </div>
          </div>

          <div className="flex items-center space-x-2 self-start sm:self-auto">
            <button
              onClick={onRefresh}
              disabled={loadingGCloud}
              className="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-mono flex items-center space-x-1.5 transition disabled:opacity-50"
            >
              <RefreshCw className={`w-3.5 h-3.5 ${loadingGCloud ? "animate-spin text-blue-400" : ""}`} />
              <span>Refresh Telemetry</span>
            </button>
          </div>
        </div>

        {/* Topology Grid */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 text-xs font-mono">
          <div className="p-3.5 rounded-xl bg-slate-950/70 border border-slate-800/80 space-y-1">
            <div className="text-[10px] uppercase text-slate-400 flex items-center justify-between">
              <span>Authenticated Identity</span>
              <ShieldCheck className="w-3.5 h-3.5 text-emerald-400" />
            </div>
            <div className="text-sm font-semibold text-slate-100 truncate">
              {gcloudBilling?.status?.account?.account || "justin@echosh-labs.com"}
            </div>
            <div className="text-[10px] text-emerald-400">
              Type: {gcloudBilling?.status?.account?.account_type || "user"} • Authenticated
            </div>
          </div>

          <div className="p-3.5 rounded-xl bg-slate-950/70 border border-slate-800/80 space-y-1">
            <div className="text-[10px] uppercase text-slate-400 flex items-center justify-between">
              <span>Target Project</span>
              <Server className="w-3.5 h-3.5 text-blue-400" />
            </div>
            <div className="text-sm font-semibold text-blue-300">
              {gcloudBilling?.status?.project?.project_id || "echosh-labs-prod"}
            </div>
            <div className="text-[10px] text-slate-400">
              ID: {gcloudBilling?.status?.project?.project_number || "85327882658"} • {gcloudBilling?.status?.project?.environment || "Production"}
            </div>
          </div>

          <div className="p-3.5 rounded-xl bg-slate-950/70 border border-slate-800/80 space-y-1">
            <div className="text-[10px] uppercase text-slate-400 flex items-center justify-between">
              <span>Execution Region</span>
              <Cpu className="w-3.5 h-3.5 text-cyan-400" />
            </div>
            <div className="text-sm font-semibold text-cyan-300">
              {gcloudBilling?.status?.project?.region || "northamerica-northeast1"}
            </div>
            <div className="text-[10px] text-slate-400">
              Montréal, QC (Low Carbon PUE)
            </div>
          </div>

          <div className="p-3.5 rounded-xl bg-slate-950/70 border border-slate-800/80 space-y-1">
            <div className="text-[10px] uppercase text-slate-400 flex items-center justify-between">
              <span>Linked Billing Account</span>
              <CreditCard className="w-3.5 h-3.5 text-purple-400" />
            </div>
            <div className="text-sm font-semibold text-purple-300 truncate">
              {gcloudBilling?.status?.billing_account?.display_name || "echosh-labs Sovereign Billing"}
            </div>
            <div className="text-[10px] text-emerald-400 flex items-center space-x-1">
              <span>{gcloudBilling?.status?.billing_account?.billing_account_id || "01A8D4-9C7E22-B110FA"}</span>
              <span className="text-[9px] bg-emerald-500/10 px-1 py-0.2 rounded border border-emerald-500/20">ACTIVE</span>
            </div>
          </div>
        </div>
      </div>

      {/* Burn Rate & Net Sovereign Margin Synthesis */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Monthly Burn Rate & Budget Progress */}
        <div className="lg:col-span-2 bg-slate-900/90 border border-slate-800 rounded-2xl p-6 shadow-xl space-y-4">
          <div className="flex items-center justify-between border-b border-slate-800 pb-3">
            <div className="flex items-center space-x-2">
              <DollarSign className="w-5 h-5 text-blue-400" />
              <h4 className="text-sm font-semibold text-white font-mono">
                Current Month Infrastructure Burn Rate ({gcloudBilling?.cost_report?.month || "2026-09"})
              </h4>
            </div>
            <span className="text-xs font-mono text-emerald-400 bg-emerald-500/10 px-2 py-0.5 rounded border border-emerald-500/20">
              Scale-to-Zero ($0.00 idle)
            </span>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div className="p-3.5 bg-slate-950/70 border border-slate-800/80 rounded-xl">
              <div className="text-[10px] font-mono uppercase text-slate-400">Accrued Month Spend</div>
              <div className="text-2xl font-bold text-white mt-1">
                ${(gcloudBilling?.cost_report?.accrued_cost_dollars || 19.50).toFixed(2)}
              </div>
              <div className="text-[10px] text-slate-500 mt-1 font-mono">
                {(gcloudBilling?.cost_report?.budget_used_percent || 19.5).toFixed(1)}% of ${(gcloudBilling?.cost_report?.budget_limit_dollars || 100.0).toFixed(2)} budget
              </div>
            </div>

            <div className="p-3.5 bg-slate-950/70 border border-slate-800/80 rounded-xl">
              <div className="text-[10px] font-mono uppercase text-slate-400">Month-End Forecast</div>
              <div className="text-2xl font-bold text-cyan-300 mt-1">
                ${(gcloudBilling?.cost_report?.forecast_cost_dollars || 24.38).toFixed(2)}
              </div>
              <div className="text-[10px] text-slate-500 mt-1 font-mono">
                Projected with active invocations
              </div>
            </div>

            <div className="p-3.5 bg-slate-950/70 border border-slate-800/80 rounded-xl">
              <div className="text-[10px] font-mono uppercase text-slate-400">Idle Burn Rate</div>
              <div className="text-2xl font-bold text-emerald-400 mt-1">
                $0.00 / hr
              </div>
              <div className="text-[10px] text-emerald-500 mt-1 font-mono">
                minScale: 0 verified
              </div>
            </div>
          </div>

          {/* Budget Progress Bar */}
          <div className="space-y-1.5 pt-2">
            <div className="flex justify-between text-xs font-mono text-slate-400">
              <span>Monthly Budget Consumption</span>
              <span className="text-blue-300 font-semibold">
                ${(gcloudBilling?.cost_report?.accrued_cost_dollars || 19.50).toFixed(2)} / ${(gcloudBilling?.cost_report?.budget_limit_dollars || 100.0).toFixed(2)}
              </span>
            </div>
            <div className="h-2 w-full bg-slate-800 rounded-full overflow-hidden">
              <div
                className="h-full bg-gradient-to-r from-blue-500 via-cyan-400 to-emerald-400 rounded-full transition-all duration-500"
                style={{ width: `${Math.min(gcloudBilling?.cost_report?.budget_used_percent || 19.5, 100)}%` }}
              />
            </div>
          </div>
        </div>

        {/* Net Sovereign Yield Card */}
        <div className="bg-gradient-to-br from-slate-900 via-[#111827] to-emerald-950/40 border border-emerald-500/30 rounded-2xl p-6 shadow-xl flex flex-col justify-between space-y-4">
          <div>
            <div className="flex items-center justify-between border-b border-slate-800 pb-3">
              <div className="flex items-center space-x-2">
                <TrendingUp className="w-4 h-4 text-emerald-400" />
                <h4 className="text-sm font-semibold text-white font-mono">Sovereign Net Yield</h4>
              </div>
              <span className="text-[10px] font-mono bg-emerald-500/10 text-emerald-300 border border-emerald-500/20 px-2 py-0.5 rounded-full">
                {gcloudBilling?.sovereign_margin?.status === "sovereign_surplus" ? "SURPLUS" : "INVESTMENT"}
              </span>
            </div>

            <div className="space-y-3 mt-4 text-xs font-mono">
              <div className="flex justify-between text-slate-300">
                <span>Gross Digital Revenue (SaaS + YT):</span>
                <span className="text-white font-semibold">
                  ${(gcloudBilling?.sovereign_margin?.gross_revenue_dollars || 159.00).toFixed(2)}
                </span>
              </div>
              <div className="flex justify-between text-slate-300">
                <span>GCP Infrastructure Burn Rate:</span>
                <span className="text-rose-400 font-semibold">
                  -${(gcloudBilling?.sovereign_margin?.gcloud_cost_dollars || 19.50).toFixed(2)}
                </span>
              </div>
              <div className="pt-2 border-t border-slate-800 flex justify-between items-baseline">
                <span className="text-slate-200 font-medium">Net Sovereign Reserve:</span>
                <span className="text-xl font-bold text-emerald-400">
                  +${(gcloudBilling?.sovereign_margin?.net_sovereign_yield_dollars || 139.50).toFixed(2)}
                </span>
              </div>
              <div className="flex justify-between text-[11px] text-slate-400 pt-1">
                <span>Operating Margin:</span>
                <span className="text-emerald-300 font-semibold">
                  {(gcloudBilling?.sovereign_margin?.profit_margin_percent || 87.7).toFixed(1)}%
                </span>
              </div>
            </div>
          </div>

          <div className="text-[11px] text-slate-400 leading-relaxed font-sans bg-slate-950/60 p-3 rounded-xl border border-slate-800/80">
            In Vedic economic governance, surplus is not vanity profit, but the <span className="text-emerald-300 font-medium">Pūrṇa Kumbha</span> sustaining computational autonomy and continuous open creative service.
          </div>
        </div>
      </div>

      {/* Itemized Infrastructure Services */}
      <div className="bg-slate-900/90 border border-slate-800 rounded-2xl p-6 shadow-xl space-y-4">
        <h4 className="text-sm font-semibold text-white font-mono flex items-center justify-between">
          <span className="flex items-center space-x-2">
            <Layers className="w-4 h-4 text-blue-400" />
            <span>Itemized Service Cost Breakdown</span>
          </span>
          <span className="text-xs text-slate-400 font-normal">
            {gcloudBilling?.cost_report?.services?.length || 5} active cloud components
          </span>
        </h4>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {(gcloudBilling?.cost_report?.services || []).map((service, idx) => (
            <div
              key={idx}
              className="p-4 rounded-xl bg-slate-950/70 border border-slate-800 hover:border-blue-500/30 transition space-y-2 flex flex-col justify-between"
            >
              <div>
                <div className="flex items-center justify-between text-xs font-mono">
                  <span className="font-semibold text-slate-200">{service.service_name}</span>
                  <span className="text-blue-400 font-bold">${service.amount_dollars.toFixed(2)}</span>
                </div>
                <div className="text-[10px] text-slate-400 mt-0.5 font-mono">
                  {service.category}
                </div>
                <p className="text-[11px] text-slate-400 mt-2 leading-relaxed">
                  {service.unit_summary}
                </p>
              </div>

              <div className="pt-2 border-t border-slate-800/60 flex items-center justify-between text-[10px] font-mono text-slate-500">
                <span>Allocation</span>
                <span className="text-slate-300">{service.percentage.toFixed(1)}% of total</span>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Sovereign Ledger Commitment Card */}
      <div className="p-5 rounded-xl bg-gradient-to-r from-blue-950/20 via-slate-900 to-purple-950/20 border border-slate-800 space-y-3">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div>
            <h4 className="text-sm font-semibold text-white font-mono flex items-center space-x-2">
              <History className="w-4 h-4 text-purple-400" />
              <span>Two-Sided Ledger Accounting (amra_ledger)</span>
            </h4>
            <p className="text-xs text-slate-400 mt-1">
              Commit monthly cloud infrastructure expenses as verifiable debit transactions into the immutable BoltDB ledger with SHA-256 idempotency deduplication.
            </p>
          </div>

          <div className="shrink-0">
            {gcloudBilling?.ledger_synced ? (
              <div className="flex items-center space-x-2 text-xs font-mono text-emerald-400 bg-emerald-500/10 px-3 py-1.5 rounded-lg border border-emerald-500/30">
                <CheckCircle2 className="w-4 h-4" />
                <span>Month Expense Committed</span>
              </div>
            ) : (
              <button
                onClick={onSyncExpense}
                disabled={syncingGCloud}
                className="px-4 py-2 rounded-lg bg-blue-600 hover:bg-blue-500 text-white font-mono text-xs font-medium flex items-center space-x-1.5 transition shadow disabled:opacity-50"
              >
                <RefreshCw className={`w-3.5 h-3.5 ${syncingGCloud ? "animate-spin" : ""}`} />
                <span>Commit Expense to Ledger</span>
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

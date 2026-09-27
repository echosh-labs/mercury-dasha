"use client";

import React, { useState } from "react";
import { History, RefreshCw, Search } from "lucide-react";
import { LedgerTransaction } from "@/lib/types/treasury";

interface AMRALedgerTableProps {
  ledger: LedgerTransaction[];
  loadingLedger: boolean;
  onRefresh: () => void;
}

export default function AMRALedgerTable({
  ledger,
  loadingLedger,
  onRefresh,
}: AMRALedgerTableProps) {
  const [filterQuery, setFilterQuery] = useState("");
  const [filterProvider, setFilterProvider] = useState<string>("all");

  const filtered = ledger.filter((tx) => {
    const matchesProvider =
      filterProvider === "all" || tx.provider.toLowerCase().includes(filterProvider.toLowerCase());
    const matchesQuery =
      filterQuery === "" ||
      tx.id.toLowerCase().includes(filterQuery.toLowerCase()) ||
      tx.description.toLowerCase().includes(filterQuery.toLowerCase()) ||
      tx.customer_id.toLowerCase().includes(filterQuery.toLowerCase());
    return matchesProvider && matchesQuery;
  });

  return (
    <div className="p-5 rounded-xl bg-slate-900/80 border border-slate-800 space-y-4">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-slate-800 pb-3">
        <div>
          <h3 className="text-sm font-semibold text-white flex items-center space-x-2">
            <History className="w-4 h-4 text-purple-400" />
            <div className="flex items-center space-x-2">
              <span>AMRA Immutable Financial Ledger (`amra_ledger`)</span>
              <span className="text-[10px] font-serif bg-amber-500/10 text-amber-300 border border-amber-500/20 px-2 py-0.5 rounded-full">
                Pūrṇa Kumbha
              </span>
            </div>
          </h3>
          <p className="text-xs text-slate-400 mt-0.5">
            Pūrṇa Kumbha ACID audit registry: Permanent, unencumbered ledger of SaaS subscriptions, YouTube revenue, and infrastructure debits.
          </p>
        </div>

        <div className="flex items-center space-x-2">
          {/* Provider Filter */}
          <select
            value={filterProvider}
            onChange={(e) => setFilterProvider(e.target.value)}
            className="bg-slate-800 border border-slate-700 text-xs font-mono text-slate-300 rounded-lg px-2.5 py-1.5 focus:outline-none focus:border-purple-500/50"
          >
            <option value="all">All Providers</option>
            <option value="youtube">YouTube Partner</option>
            <option value="stripe">Stripe SaaS</option>
            <option value="gcloud">Google Cloud</option>
          </select>

          <button
            onClick={onRefresh}
            disabled={loadingLedger}
            className="p-1.5 rounded-lg bg-slate-800 text-slate-400 hover:text-white border border-slate-700 transition"
            title="Refresh Ledger"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${loadingLedger ? "animate-spin text-purple-400" : ""}`} />
          </button>
        </div>
      </div>

      {/* Filter Bar */}
      <div className="relative">
        <Search className="w-3.5 h-3.5 absolute left-3 top-2.5 text-slate-500" />
        <input
          type="text"
          value={filterQuery}
          onChange={(e) => setFilterQuery(e.target.value)}
          placeholder="Filter by transaction ID, description, or customer..."
          className="w-full bg-slate-950/80 border border-slate-800/80 text-xs font-mono text-slate-300 rounded-lg pl-9 pr-3 py-2 focus:outline-none focus:border-purple-500/40 placeholder:text-slate-600"
        />
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-left text-xs font-mono">
          <thead className="border-b border-slate-800 text-slate-400 uppercase text-[10px] bg-slate-950/40">
            <tr>
              <th className="py-2.5 px-3">Transaction ID</th>
              <th className="py-2.5 px-3">Provider</th>
              <th className="py-2.5 px-3">Amount</th>
              <th className="py-2.5 px-3">Status</th>
              <th className="py-2.5 px-3">Description</th>
              <th className="py-2.5 px-3">Timestamp</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800/60 text-slate-300">
            {filtered.length > 0 ? (
              filtered.map((tx) => {
                const isDebit = tx.amount_cents < 0;
                return (
                  <tr key={tx.id} className="hover:bg-slate-800/30 transition">
                    <td className="py-2.5 px-3 text-cyan-300 font-medium truncate max-w-[150px]">
                      {tx.id}
                    </td>
                    <td className="py-2.5 px-3">
                      <span
                        className={`px-2 py-0.5 rounded text-[10px] uppercase ${
                          tx.provider === "youtube_partner"
                            ? "bg-red-500/10 text-red-400 border border-red-500/20"
                            : tx.provider === "gcloud_billing"
                            ? "bg-blue-500/10 text-blue-400 border border-blue-500/20"
                            : "bg-cyan-500/10 text-cyan-400 border border-cyan-500/20"
                        }`}
                      >
                        {tx.provider}
                      </span>
                    </td>
                    <td className={`py-2.5 px-3 font-bold ${isDebit ? "text-rose-400" : "text-emerald-400"}`}>
                      {isDebit ? "-" : ""}${Math.abs(tx.amount_cents / 100).toFixed(2)} {tx.currency}
                    </td>
                    <td className="py-2.5 px-3">
                      <span className="capitalize text-emerald-400 font-semibold">{tx.status}</span>
                    </td>
                    <td className="py-2.5 px-3 text-slate-300 truncate max-w-[300px]" title={tx.description}>
                      {tx.description}
                    </td>
                    <td className="py-2.5 px-3 text-slate-400">
                      {new Date(tx.created_at).toLocaleString()}
                    </td>
                  </tr>
                );
              })
            ) : (
              <tr>
                <td colSpan={6} className="py-8 text-center text-slate-500">
                  {ledger.length === 0 ? "No ledger transactions recorded yet." : "No matching transactions found."}
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

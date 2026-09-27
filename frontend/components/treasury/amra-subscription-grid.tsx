"use client";

import React from "react";
import { CheckCircle2 } from "lucide-react";
import { AMRAPlan } from "@/lib/types/treasury";

interface AMRASubscriptionGridProps {
  plans: AMRAPlan[];
  onCheckout: (planId: string) => Promise<void>;
}

export default function AMRASubscriptionGrid({
  plans,
  onCheckout,
}: AMRASubscriptionGridProps) {
  return (
    <div className="space-y-4">
      <div className="text-xs text-slate-400 font-mono">
        Sovereign AMRA subscription tiers for echosh-labs platform capabilities.
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {plans.map((p) => (
          <div
            key={p.id}
            className="p-5 rounded-xl bg-slate-900/90 border border-slate-800 flex flex-col justify-between hover:border-cyan-500/40 transition"
          >
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-xs font-mono uppercase px-2 py-0.5 rounded bg-cyan-500/10 text-cyan-400 border border-cyan-500/20">
                  {p.tier}
                </span>
                <span className="text-xs font-mono text-slate-400 capitalize">{p.interval}</span>
              </div>

              <div>
                <h4 className="text-lg font-bold text-white">{p.name}</h4>
                <div className="text-2xl font-extrabold text-cyan-300 mt-1">
                  ${(p.price_cents / 100).toFixed(2)}
                  <span className="text-xs font-normal text-slate-400"> / {p.interval}</span>
                </div>
              </div>

              <div className="pt-2 border-t border-slate-800/80 space-y-1.5">
                <div className="text-[11px] font-mono text-slate-400 uppercase tracking-wider">Entitlements:</div>
                {p.entitlements.map((ent) => (
                  <div key={ent} className="text-xs text-slate-300 flex items-center space-x-1.5">
                    <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                    <span className="font-mono text-[11px]">{ent}</span>
                  </div>
                ))}
              </div>
            </div>

            <div className="pt-6">
              <button
                onClick={() => onCheckout(p.id)}
                className="w-full py-2 rounded-lg bg-cyan-600 hover:bg-cyan-500 text-white font-semibold text-xs transition"
              >
                Simulate Entitlement Checkout
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

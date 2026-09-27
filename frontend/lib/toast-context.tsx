"use client";

import React, { createContext, useContext, useState, useCallback, ReactNode } from "react";
import { CheckCircle2, AlertCircle, Info, X } from "lucide-react";

export type ToastType = "success" | "info" | "warning" | "error";

export interface ToastItem {
  id: string;
  message: string;
  type: ToastType;
  durationMs: number;
}

interface ToastContextType {
  toast: (message: string, type?: ToastType, durationMs?: number) => void;
  success: (message: string, durationMs?: number) => void;
  error: (message: string, durationMs?: number) => void;
  info: (message: string, durationMs?: number) => void;
  warning: (message: string, durationMs?: number) => void;
}

const ToastContext = createContext<ToastContextType | undefined>(undefined);

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<ToastItem[]>([]);

  const removeToast = useCallback((id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  const toast = useCallback(
    (message: string, type: ToastType = "info", durationMs: number = 2000) => {
      const id = `${Date.now()}_${Math.random().toString(36).substring(2, 7)}`;
      const newToast: ToastItem = { id, message, type, durationMs };

      setToasts((prev) => [...prev.slice(-4), newToast]); // keep at most 5 toasts

      if (durationMs > 0) {
        setTimeout(() => {
          removeToast(id);
        }, durationMs);
      }
    },
    [removeToast]
  );

  const success = useCallback((msg: string, d?: number) => toast(msg, "success", d ?? 2000), [toast]);
  const error = useCallback((msg: string, d?: number) => toast(msg, "error", d ?? 3000), [toast]);
  const info = useCallback((msg: string, d?: number) => toast(msg, "info", d ?? 2000), [toast]);
  const warning = useCallback((msg: string, d?: number) => toast(msg, "warning", d ?? 2500), [toast]);

  return (
    <ToastContext.Provider value={{ toast, success, error, info, warning }}>
      {children}
      {/* Floating Toast Notification Container */}
      <div
        className="fixed bottom-5 right-5 z-[9999] flex flex-col space-y-2 pointer-events-none max-w-sm w-full sm:w-auto"
        aria-live="polite"
      >
        {toasts.map((t) => (
          <div
            key={t.id}
            className={`pointer-events-auto flex items-center justify-between space-x-2.5 px-4 py-2.5 rounded-xl border shadow-2xl backdrop-blur-md text-xs font-mono transition-all duration-200 animate-in fade-in slide-in-from-bottom-2 ${
              t.type === "success"
                ? "bg-slate-900/95 border-emerald-500/40 text-emerald-300 shadow-emerald-950/40"
                : t.type === "error"
                ? "bg-slate-900/95 border-rose-500/40 text-rose-300 shadow-rose-950/40"
                : t.type === "warning"
                ? "bg-slate-900/95 border-amber-500/40 text-amber-300 shadow-amber-950/40"
                : "bg-slate-900/95 border-cyan-500/40 text-cyan-300 shadow-cyan-950/40"
            }`}
          >
            <div className="flex items-center space-x-2 min-w-0">
              {t.type === "success" && <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0" />}
              {t.type === "error" && <AlertCircle className="w-4 h-4 text-rose-400 shrink-0" />}
              {t.type === "warning" && <AlertCircle className="w-4 h-4 text-amber-400 shrink-0" />}
              {t.type === "info" && <Info className="w-4 h-4 text-cyan-400 shrink-0" />}
              <span className="truncate">{t.message}</span>
            </div>
            <button
              onClick={() => removeToast(t.id)}
              className="text-slate-400 hover:text-white p-0.5 rounded transition shrink-0"
              title="Dismiss"
            >
              <X className="w-3.5 h-3.5" />
            </button>
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  );
}

export function useToast(): ToastContextType {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error("useToast must be used within a ToastProvider");
  }
  return context;
}

import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "mercury-dasha | echosh-labs",
  description: "Unified Go Engine, BoltDB, and Static Next.js on Cloud Run",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className="antialiased min-h-screen bg-[#090d16] text-slate-100 flex flex-col">
        <header className="border-b border-slate-800 bg-[#0c1220]/80 backdrop-blur sticky top-0 z-50">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <span className="h-3 w-3 rounded-full bg-cyan-400 animate-pulse" />
              <span className="font-mono text-xs uppercase tracking-widest text-cyan-400 bg-cyan-950/60 px-2 py-0.5 rounded border border-cyan-800">
                echosh-labs
              </span>
              <span className="font-bold tracking-tight text-white">mercury-dasha</span>
              <span className="text-xs text-slate-400 font-mono">v1.0.0</span>
            </div>
            <div className="flex items-center space-x-4 text-xs font-mono text-slate-400">
              <span className="hidden sm:inline-block">Runtime: Go 1.23 + bbolt</span>
              <span className="px-2 py-1 rounded bg-slate-800 text-slate-300 border border-slate-700">
                Cloud Run gen2
              </span>
            </div>
          </div>
        </header>
        <main className="flex-1 max-w-7xl w-full mx-auto p-4 sm:p-6 lg:p-8">
          {children}
        </main>
        <footer className="border-t border-slate-800/80 py-4 text-center text-xs text-slate-500">
          Mercury Stack • Built with Go, bbolt & Next.js for echosh-labs
        </footer>
      </body>
    </html>
  );
}

import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "mercury-dasha | echosh-labs",
  description: "Unified Go Engine, BoltDB, and Static Next.js on Cloud Run",
};

import GlobalFooter from "@/components/global-footer";
import GlobalNav from "@/components/global-nav";
import { YouTubeStudioProvider } from "@/lib/youtube-context";
import { ToastProvider } from "@/lib/toast-context";
import { TemporalProvider } from "@/lib/temporal-context";
import YouTubeStudioDock from "@/components/youtube-studio-dock";

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className="antialiased min-h-screen bg-[#090d16] text-slate-100 flex flex-col" suppressHydrationWarning>
        <ToastProvider>
          <TemporalProvider>
            <YouTubeStudioProvider>
            <GlobalNav />
            <main className="flex-1 max-w-7xl w-full mx-auto p-4 sm:p-6 lg:p-8">
              {children}
            </main>
            <YouTubeStudioDock />
            <GlobalFooter />
          </YouTubeStudioProvider>
        </TemporalProvider>
      </ToastProvider>
      </body>
    </html>
  );
}

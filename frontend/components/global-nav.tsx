"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  Compass,
  Clock,
  Sun,
  Users,
  Video,
  Film,
  Coins,
  Radio,
  ExternalLink,
  Menu,
  X,
  ChevronDown,
  Sparkles,
  Terminal,
  FolderTree,
  Flame,
  Layers
} from "lucide-react";

interface NavItem {
  name: string;
  href: string;
  icon: React.ElementType;
  description?: string;
  badge?: string;
}

const OBSERVATORY_LINKS: NavItem[] = [
  { name: "Observatory Hub", href: "/", icon: Compass, description: "Unified Vimshottari, Hora & Ephemeris portal" },
  { name: "24h Planetary Alignment", href: "/alignment/", icon: Clock, description: "Topocentric Chaldean matrix & hourly metals" },
  { name: "Daily Ephemeris Chart", href: "/ephemeris/", icon: Sun, description: "Day lord, solar anchors & 24h planetary hours" },
  { name: "Character Sanctuary", href: "/characters/", icon: Users, description: "Living protagonist blueprints & character forge" },
];

const STUDIO_LINKS: NavItem[] = [
  { name: "Foundations Studio", href: "/foundations/", icon: Video, description: "Autonomous YouTube storyline pipeline" },
  { name: "AV Studio 1", href: "/studio/", icon: Film, description: "Vocal narration & cinematic timeline engine" },
  { name: "Sonic Chronicle", href: "/foundations/", icon: Sparkles, description: "Audiovisual syntheses & narration chronicle" },
];

export default function GlobalNav() {
  const pathname = usePathname() || "/";
  const [mobileMenuOpen, setMobileMenuOpen] = useState<boolean>(false);
  const [observatoryDropdown, setObservatoryDropdown] = useState<boolean>(false);
  const [studiosDropdown, setStudiosDropdown] = useState<boolean>(false);
  const [axisStatus, setAxisStatus] = useState<{ isLive: boolean; count: number }>({ isLive: false, count: 0 });

  // Close menus on route change
  useEffect(() => {
    setMobileMenuOpen(false);
    setObservatoryDropdown(false);
    setStudiosDropdown(false);
  }, [pathname]);

  // Fetch quick status for Axis Mundi badge
  useEffect(() => {
    const checkAxisStatus = async () => {
      try {
        const res = await fetch("/api/v1/axis-mundi/status");
        if (res.ok) {
          const data = await res.json();
          setAxisStatus({
            isLive: !!data.is_live,
            count: data.counts?.total ?? 0,
          });
        }
      } catch {
        // non-blocking
      }
    };
    checkAxisStatus();
    const interval = setInterval(checkAxisStatus, 20000);
    return () => clearInterval(interval);
  }, []);

  const isObservatoryActive = pathname === "/" || pathname.startsWith("/alignment") || pathname.startsWith("/ephemeris") || pathname.startsWith("/characters");
  const isAlchemyActive = pathname.startsWith("/alchemy");
  const isStorehouseActive = pathname.startsWith("/storehouse");
  const isStudiosActive = pathname.startsWith("/foundations") || pathname.startsWith("/studio");
  const isTreasuryActive = pathname.startsWith("/treasury");
  const isConsoleActive = pathname.startsWith("/console");

  return (
    <>
      <header className="border-b border-slate-800 bg-[#0c1220]/90 backdrop-blur-md sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between gap-4">
          {/* Brand & Identity */}
          <div className="flex items-center space-x-3">
            <Link href="/" className="flex items-center space-x-2.5 group">
              <span className="h-3 w-3 rounded-full bg-cyan-400 animate-pulse" />
              <span className="font-mono text-xs uppercase tracking-widest text-cyan-400 bg-cyan-950/60 px-2 py-0.5 rounded border border-cyan-800 group-hover:border-cyan-500 transition">
                echosh-labs
              </span>
              <span className="font-bold tracking-tight text-white group-hover:text-cyan-300 transition text-sm sm:text-base">
                mercury-dasha
              </span>
            </Link>
            <span className="text-[10px] text-slate-500 font-mono hidden md:inline-block">v1.1.0</span>
          </div>

          {/* Unified Desktop Navigation Links */}
          <nav className="hidden lg:flex items-center space-x-1 text-xs font-mono">
            {/* 1. Dasha Observatory Dropdown */}
            <div className="relative">
              <button
                onClick={() => {
                  setObservatoryDropdown(!observatoryDropdown);
                  setStudiosDropdown(false);
                }}
                onMouseEnter={() => setObservatoryDropdown(true)}
                className={`flex items-center space-x-1.5 px-3 py-2 rounded-lg transition ${
                  isObservatoryActive
                    ? "bg-slate-800 text-cyan-400 font-semibold border border-slate-700"
                    : "text-slate-300 hover:text-white hover:bg-slate-800/50"
                }`}
              >
                <Compass className="w-3.5 h-3.5 text-cyan-400" />
                <span>Dasha Observatory</span>
                <ChevronDown className="w-3 h-3 text-slate-500" />
              </button>

              {observatoryDropdown && (
                <div
                  onMouseLeave={() => setObservatoryDropdown(false)}
                  className="absolute left-0 mt-1 w-64 rounded-xl bg-[#0e1626] border border-slate-800 shadow-2xl p-2 space-y-1 animate-in fade-in zoom-in-95 duration-150 z-50"
                >
                  {OBSERVATORY_LINKS.map((item) => {
                    const Icon = item.icon;
                    const isActive = pathname === item.href || (item.href !== "/" && pathname.startsWith(item.href));
                    return (
                      <Link
                        key={item.href}
                        href={item.href}
                        className={`flex items-start space-x-2.5 p-2 rounded-lg transition ${
                          isActive
                            ? "bg-cyan-950/60 text-cyan-300 border border-cyan-800/60"
                            : "text-slate-300 hover:bg-slate-800/70 hover:text-white"
                        }`}
                      >
                        <Icon className="w-4 h-4 text-cyan-400 mt-0.5 shrink-0" />
                        <div>
                          <div className="font-semibold">{item.name}</div>
                          {item.description && (
                            <div className="text-[10px] text-slate-400 font-sans mt-0.5 leading-snug">
                              {item.description}
                            </div>
                          )}
                        </div>
                      </Link>
                    );
                  })}
                </div>
              )}
            </div>

            {/* 2. Alchemical Laboratory */}
            <Link
              href="/alchemy/"
              className={`flex items-center space-x-1.5 px-3 py-2 rounded-lg transition ${
                isAlchemyActive
                  ? "bg-amber-950/60 text-amber-300 font-semibold border border-amber-800/70"
                  : "text-slate-300 hover:text-amber-300 hover:bg-slate-800/50"
              }`}
            >
              <Sparkles className="w-3.5 h-3.5 text-amber-400" />
              <span>Alchemical Laboratory</span>
            </Link>

            {/* 3. Dropbox Sovereign Storehouse */}
            <Link
              href="/storehouse/"
              className={`flex items-center space-x-1.5 px-3 py-2 rounded-lg transition ${
                isStorehouseActive
                  ? "bg-blue-950/60 text-blue-300 font-semibold border border-blue-800/70"
                  : "text-slate-300 hover:text-blue-300 hover:bg-slate-800/50"
              }`}
            >
              <FolderTree className="w-3.5 h-3.5 text-blue-400" />
              <span>Dropbox Sovereign Storehouse</span>
            </Link>

            {/* 4. Production Studios Dropdown */}
            <div className="relative">
              <button
                onClick={() => {
                  setStudiosDropdown(!studiosDropdown);
                  setObservatoryDropdown(false);
                }}
                onMouseEnter={() => setStudiosDropdown(true)}
                className={`flex items-center space-x-1.5 px-3 py-2 rounded-lg transition ${
                  isStudiosActive
                    ? "bg-slate-800 text-purple-400 font-semibold border border-slate-700"
                    : "text-slate-300 hover:text-white hover:bg-slate-800/50"
                }`}
              >
                <Video className="w-3.5 h-3.5 text-purple-400" />
                <span>Sonic Chronicle</span>
                <ChevronDown className="w-3 h-3 text-slate-500" />
              </button>

              {studiosDropdown && (
                <div
                  onMouseLeave={() => setStudiosDropdown(false)}
                  className="absolute left-0 mt-1 w-64 rounded-xl bg-[#0e1626] border border-slate-800 shadow-2xl p-2 space-y-1 animate-in fade-in zoom-in-95 duration-150 z-50"
                >
                  {STUDIO_LINKS.map((item) => {
                    const Icon = item.icon;
                    const isActive = pathname.startsWith(item.href);
                    return (
                      <Link
                        key={item.href}
                        href={item.href}
                        className={`flex items-start space-x-2.5 p-2 rounded-lg transition ${
                          isActive
                            ? "bg-purple-950/60 text-purple-300 border border-purple-800/60"
                            : "text-slate-300 hover:bg-slate-800/70 hover:text-white"
                        }`}
                      >
                        <Icon className="w-4 h-4 text-purple-400 mt-0.5 shrink-0" />
                        <div>
                          <div className="font-semibold">{item.name}</div>
                          {item.description && (
                            <div className="text-[10px] text-slate-400 font-sans mt-0.5 leading-snug">
                              {item.description}
                            </div>
                          )}
                        </div>
                      </Link>
                    );
                  })}
                </div>
              )}
            </div>

            {/* 5. AMRA Treasury & Studio */}
            <Link
              href="/treasury/"
              className={`flex items-center space-x-1.5 px-3 py-2 rounded-lg transition ${
                isTreasuryActive
                  ? "bg-emerald-950/60 text-emerald-300 font-semibold border border-emerald-800/70"
                  : "text-slate-300 hover:text-emerald-300 hover:bg-slate-800/50"
              }`}
            >
              <Coins className="w-3.5 h-3.5 text-emerald-400" />
              <span>AMRA Treasury &amp; Studio</span>
            </Link>

            {/* 6. Agentic Mission Control */}
            <Link
              href="/console/"
              className={`flex items-center space-x-1.5 px-3 py-2 rounded-lg transition ${
                isConsoleActive
                  ? "bg-purple-950/60 text-purple-300 font-semibold border border-purple-800/70"
                  : "text-slate-300 hover:text-purple-300 hover:bg-slate-800/50"
              }`}
            >
              <Terminal className="w-3.5 h-3.5 text-purple-400" />
              <span>Agentic Mission Control</span>
            </Link>
          </nav>

          {/* Right Status Utilities & Controls */}
          <div className="flex items-center space-x-2 sm:space-x-3 text-xs font-mono">
            {/* Direct Axis Mundi Engine Link (:8088) */}
            <a
              href="http://localhost:8088"
              target="_blank"
              rel="noreferrer"
              title="Open Axis Mundi Engine on :8088"
              className="flex items-center space-x-1.5 px-2.5 py-1 rounded-lg border transition group bg-slate-900 border-slate-800 hover:border-cyan-500/50 cursor-pointer text-slate-400 hover:text-cyan-300"
            >
              <Radio className={`w-3 h-3 ${axisStatus.isLive ? "text-emerald-400 animate-pulse" : "text-cyan-400"}`} />
              <span className="hidden sm:inline">Axis Mundi:</span>
              <span className="text-cyan-300 font-semibold">{axisStatus.count} items</span>
              <ExternalLink className="w-2.5 h-2.5 text-slate-500 group-hover:text-cyan-400 transition" />
            </a>

            {/* Runtime Target Pill */}
            <span className="hidden xl:inline-block px-2.5 py-1 rounded-lg bg-slate-900 text-slate-400 border border-slate-800">
              Internal // :8080
            </span>

            {/* External Compendium Link */}
            <a
              href="https://echosh-labs.com"
              target="_blank"
              rel="noreferrer"
              className="hidden sm:flex items-center space-x-1 px-2.5 py-1 rounded-lg bg-slate-900 hover:bg-slate-800 border border-slate-800 text-slate-400 hover:text-cyan-300 transition"
            >
              <span>Dossier</span>
              <ExternalLink className="w-3 h-3" />
            </a>

            {/* Mobile Menu Hamburger */}
            <button
              onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
              className="lg:hidden p-2 rounded-lg bg-slate-900 border border-slate-800 text-slate-300 hover:text-white"
              aria-label="Toggle navigation menu"
            >
              {mobileMenuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
            </button>
          </div>
        </div>

        {/* Mobile Drawer Menu */}
        {mobileMenuOpen && (
          <div className="lg:hidden border-t border-slate-800 bg-[#0c1220] p-4 space-y-4 animate-in slide-in-from-top duration-200 max-h-[80vh] overflow-y-auto">
            {/* 1. Observatory */}
            <div className="space-y-1">
              <span className="text-[11px] font-mono uppercase text-cyan-400 tracking-wider block px-2 mb-1">
                Dasha Observatory
              </span>
              {OBSERVATORY_LINKS.map((item) => {
                const Icon = item.icon;
                const isActive = pathname === item.href || (item.href !== "/" && pathname.startsWith(item.href));
                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    className={`flex items-center space-x-2 px-3 py-2 rounded-lg text-xs font-mono transition ${
                      isActive
                        ? "bg-cyan-950/80 text-cyan-300 font-bold border border-cyan-800"
                        : "text-slate-300 hover:bg-slate-800 hover:text-white"
                    }`}
                  >
                    <Icon className="w-4 h-4 text-cyan-400" />
                    <span>{item.name}</span>
                  </Link>
                );
              })}
            </div>

            {/* 2. Alchemy */}
            <div className="space-y-1 pt-2 border-t border-slate-800/80">
              <Link
                href="/alchemy/"
                className={`flex items-center space-x-2 px-3 py-2 rounded-lg text-xs font-mono transition ${
                  isAlchemyActive
                    ? "bg-amber-950/80 text-amber-300 font-bold border border-amber-800"
                    : "text-slate-300 hover:bg-slate-800 hover:text-white"
                }`}
              >
                <Sparkles className="w-4 h-4 text-amber-400" />
                <span>Alchemical Laboratory</span>
              </Link>
            </div>

            {/* 3. Storehouse */}
            <div className="space-y-1 pt-2 border-t border-slate-800/80">
              <Link
                href="/storehouse/"
                className={`flex items-center space-x-2 px-3 py-2 rounded-lg text-xs font-mono transition ${
                  isStorehouseActive
                    ? "bg-blue-950/80 text-blue-300 font-bold border border-blue-800"
                    : "text-slate-300 hover:bg-slate-800 hover:text-white"
                }`}
              >
                <FolderTree className="w-4 h-4 text-blue-400" />
                <span>Dropbox Sovereign Storehouse</span>
              </Link>
            </div>

            {/* 4. Studios */}
            <div className="space-y-1 pt-2 border-t border-slate-800/80">
              <span className="text-[11px] font-mono uppercase text-purple-400 tracking-wider block px-2 mb-1">
                Sonic Chronicle &amp; Studios
              </span>
              {STUDIO_LINKS.map((item) => {
                const Icon = item.icon;
                const isActive = pathname.startsWith(item.href);
                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    className={`flex items-center space-x-2 px-3 py-2 rounded-lg text-xs font-mono transition ${
                      isActive
                        ? "bg-purple-950/80 text-purple-300 font-bold border border-purple-800"
                        : "text-slate-300 hover:bg-slate-800 hover:text-white"
                    }`}
                  >
                    <Icon className="w-4 h-4 text-purple-400" />
                    <span>{item.name}</span>
                  </Link>
                );
              })}
            </div>

            {/* 5. Treasury */}
            <div className="space-y-1 pt-2 border-t border-slate-800/80">
              <Link
                href="/treasury/"
                className={`flex items-center space-x-2 px-3 py-2 rounded-lg text-xs font-mono transition ${
                  isTreasuryActive
                    ? "bg-emerald-950/80 text-emerald-300 font-bold border border-emerald-800"
                    : "text-slate-300 hover:bg-slate-800 hover:text-white"
                }`}
              >
                <Coins className="w-4 h-4 text-emerald-400" />
                <span>AMRA Treasury &amp; Studio</span>
              </Link>
            </div>

            {/* 6. Mission Control */}
            <div className="space-y-1 pt-2 border-t border-slate-800/80">
              <Link
                href="/console/"
                className={`flex items-center space-x-2 px-3 py-2 rounded-lg text-xs font-mono transition ${
                  isConsoleActive
                    ? "bg-purple-950/80 text-purple-300 font-bold border border-purple-800"
                    : "text-slate-300 hover:bg-slate-800 hover:text-white"
                }`}
              >
                <Terminal className="w-4 h-4 text-purple-400" />
                <span>Agentic Mission Control</span>
              </Link>
            </div>

            {/* 7. Axis Mundi Engine Link */}
            <div className="space-y-1 pt-2 border-t border-slate-800/80">
              <a
                href="http://localhost:8088"
                target="_blank"
                rel="noreferrer"
                onClick={() => setMobileMenuOpen(false)}
                className="w-full flex items-center justify-between px-3 py-2 rounded-lg text-xs font-mono transition text-slate-300 hover:bg-slate-800 hover:text-white"
              >
                <div className="flex items-center space-x-2">
                  <Radio className={`w-4 h-4 ${axisStatus.isLive ? "text-emerald-400 animate-pulse" : "text-cyan-400"}`} />
                  <span>Axis Mundi Engine (:8088)</span>
                </div>
                <div className="flex items-center space-x-1.5">
                  <span className="text-[10px] bg-cyan-950 text-cyan-400 border border-cyan-800 px-2 py-0.5 rounded-full font-semibold">
                    {axisStatus.count} items
                  </span>
                  <ExternalLink className="w-3 h-3 text-slate-400" />
                </div>
              </a>
            </div>

            <div className="pt-2 border-t border-slate-800/80 flex items-center justify-between text-xs font-mono text-slate-400 px-2">
              <span>Runtime: Go 1.23 + bbolt</span>
              <a
                href="https://echosh-labs.com"
                target="_blank"
                rel="noreferrer"
                className="text-cyan-400 hover:underline flex items-center space-x-1"
              >
                <span>Ecosystem Dossier</span>
                <ExternalLink className="w-3 h-3" />
              </a>
            </div>
          </div>
        )}
      </header>
    </>
  );
}

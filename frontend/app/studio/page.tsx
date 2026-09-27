"use client";

import React from "react";
import Link from "next/link";
import { ArrowLeft, Film } from "lucide-react";
import VocalNarrationStudio from "@/components/vocal-narration-studio";

export default function StudioPage() {
  return (
    <div className="space-y-6 max-w-7xl mx-auto py-4 px-2">
      <div className="flex items-center justify-between border-b border-slate-800 pb-4">
        <Link
          href="/"
          className="flex items-center space-x-2 text-xs font-mono text-slate-400 hover:text-cyan-400 transition"
        >
          <ArrowLeft className="w-3.5 h-3.5" />
          <span>Return to Observatory</span>
        </Link>
        <div className="text-xs font-mono text-amber-400 flex items-center space-x-1.5">
          <Film className="w-3.5 h-3.5" />
          <span>AV STUDIO 1 // CINEMATIC ENGINE</span>
        </div>
      </div>

      <VocalNarrationStudio slideshowId="temple-of-illumination" />
    </div>
  );
}

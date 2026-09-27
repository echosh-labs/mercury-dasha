"use client";

import React from "react";
import UnifiedAudioPortal from "./unified-audio-portal";

/**
 * AudioStudio serves as the sovereign entrypoint for Sonic Chronicle.
 * Refactored into a single-page UnifiedAudioPortal that hosts, manages,
 * and plays both transcript-driven audio files and transcriptless vault files.
 */
export default function AudioStudio() {
  return <UnifiedAudioPortal />;
}

export interface SceneBeat {
  act: number;
  title: string;
  premise: string;
  emotional_tone: string;
  prompt_directive: string;
  visual_keywords?: string[];
}

export interface VisualDirectives {
  color_palette: string[];
  atmosphere: string;
  lighting: string;
  suggested_motion: string;
}

export interface CosmicAlignment {
  mahadasha: string;
  mahadasha_sanskrit: string;
  antardasha: string;
  antardasha_sanskrit: string;
  sacred_metal: string;
  hermetic_axiom: string;
  magnum_opus_stage: string;
  active_archetype: string;
  resonant_frequency_hz: number;
  chakra_focus?: string;
}

export interface StorySeed {
  id: string;
  title: string;
  source_path: string;
  source_type: "written_note" | "spoken_transcript";
  has_audio: boolean;
  audio_url?: string;
  transcript_url?: string;
  sanctuary_id: string;
  sanctuary_label: string;
  sub_sanctuary?: string;
  note_date: string;
  year: number;
  word_count: number;
  narrative_hook: string;
  dialogue_anchor: string;
  protagonist_role: string;
  cosmic_alignment: CosmicAlignment;
  scene_beats: SceneBeat[];
  visual_directives: VisualDirectives;
  tags?: string[];
  created_at: string;
}

export interface ResonantExcerpt {
  catalog_id: string;
  title: string;
  path: string;
  source_type: "written_note" | "spoken_transcript";
  sanctuary: string;
  sanctuary_label: string;
  note_date: string;
  year: number;
  has_audio: boolean;
  audio_url?: string;
  resonance_reason: string;
  matching_attribute: string;
  snippet: string;
  tags?: string[];
}

export interface ResonantTransitResponse {
  current_transit: {
    mahadasha: string;
    antardasha: string;
    sacred_metal: string;
    hermetic_axiom: string;
    archetype: string;
  };
  count: number;
  excerpts: ResonantExcerpt[];
}

export interface SeedFromTextRequest {
  path?: string;
  id?: string;
  content?: string;
  title?: string;
  note_date?: string;
  sanctuary?: string;
  protagonist?: string;
  save?: boolean;
}

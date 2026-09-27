"use client";

// Sovereign client service for BoltDB-backed Esoteric Documents (/api/v1/esoteric/*).
// Provides typed hydration and single-flight in-memory deduplication.

import {
  EsotericPhilosophy,
  ArishadvargaDemonDoc,
  TransmutationTriadDoc,
  MathematicalFoundationsDoc,
} from "@/lib/types/treasury";

export interface EsotericCatalogResponse {
  keys: string[];
  count: number;
  prefix: string;
  source: string;
}

export interface ArishadvargaResponse {
  demons: ArishadvargaDemonDoc[];
  count: number;
  source: string;
}

// In-memory single-flight promise cache
const docPromises = new Map<string, Promise<any>>();
const docCache = new Map<string, any>();

/**
 * Lists available esoteric document keys in BoltDB with an optional prefix.
 */
export async function fetchEsotericCatalog(prefix = "esoteric:"): Promise<string[]> {
  try {
    const url = `/api/v1/esoteric?prefix=${encodeURIComponent(prefix)}`;
    const res = await fetch(url);
    if (!res.ok) {
      throw new Error(`Failed to list esoteric catalog: ${res.statusText}`);
    }
    const data: EsotericCatalogResponse = await res.json();
    return data.keys || [];
  } catch (err) {
    console.error("[Esoteric Client] Catalog fetch error:", err);
    return [];
  }
}

/**
 * Retrieves an arbitrary esoteric document by its key from BoltDB.
 */
export async function fetchEsotericDoc<T = any>(key: string, forceRefresh = false): Promise<T | null> {
  const cleanKey = key.startsWith("esoteric:") ? key.replace("esoteric:", "") : key;

  if (!forceRefresh && docCache.has(cleanKey)) {
    return docCache.get(cleanKey);
  }

  if (!forceRefresh && docPromises.has(cleanKey)) {
    return docPromises.get(cleanKey);
  }

  const promise = (async () => {
    try {
      const res = await fetch(`/api/v1/esoteric/${encodeURIComponent(cleanKey)}`);
      if (!res.ok) {
        throw new Error(`Failed to load esoteric document ${cleanKey}: ${res.statusText}`);
      }
      const data: T = await res.json();
      docCache.set(cleanKey, data);
      return data;
    } catch (err) {
      console.error(`[Esoteric Client] Document fetch error for ${cleanKey}:`, err);
      docPromises.delete(cleanKey);
      return null;
    }
  })();

  docPromises.set(cleanKey, promise);
  return promise;
}

/**
 * Loads the 6 classical Arishadvarga inner adversaries and transmutations from BoltDB.
 */
export async function fetchArishadvargaDemons(forceRefresh = false): Promise<ArishadvargaDemonDoc[]> {
  const cacheKey = "arishadvarga";
  if (!forceRefresh && docCache.has(cacheKey)) {
    return docCache.get(cacheKey);
  }

  try {
    const res = await fetch("/api/v1/esoteric/arishadvarga");
    if (!res.ok) {
      throw new Error(`Failed to load Arishadvarga demons: ${res.statusText}`);
    }
    const data: ArishadvargaResponse = await res.json();
    const demons = data.demons || [];
    docCache.set(cacheKey, demons);
    return demons;
  } catch (err) {
    console.error("[Esoteric Client] Arishadvarga fetch error:", err);
    return [];
  }
}

/**
 * Loads the authoritative Vedic Philosophy of Āmra document from BoltDB.
 */
export async function fetchAmraPhilosophy(forceRefresh = false): Promise<EsotericPhilosophy | null> {
  return fetchEsotericDoc<EsotericPhilosophy>("amra_philosophy", forceRefresh);
}

/**
 * Loads the 3-stage alchemical transmutation triad (Āma -> Āmra -> Amara) from BoltDB.
 */
export async function fetchTransmutationTriad(forceRefresh = false): Promise<TransmutationTriadDoc | null> {
  return fetchEsotericDoc<TransmutationTriadDoc>("transmutation_triad", forceRefresh);
}

/**
 * Loads the sacred geometry formulas and mathematical foundations from BoltDB.
 */
export async function fetchMathematicalFoundations(forceRefresh = false): Promise<MathematicalFoundationsDoc | null> {
  return fetchEsotericDoc<MathematicalFoundationsDoc>("mathematical_foundations", forceRefresh);
}

/**
 * Flushes all in-memory cached esoteric documents.
 */
export function invalidateEsotericCache(): void {
  docPromises.clear();
  docCache.clear();
}

// Timezone and Geolocation Auto-Detection Substrate

export interface DetectedLocation {
  timezone: string;
  utc_offset_hours: number;
  latitude: number;
  longitude: number;
  auto_detect: boolean;
  location_name: string;
}

// Built-in IANA timezone centroids for zero-friction fallback
export const IANACentroids: Record<string, { latitude: number; longitude: number; name: string }> = {
  "america/new_york": { latitude: 40.7128, longitude: -74.0060, name: "New York, USA" },
  "america/chicago": { latitude: 41.8781, longitude: -87.6298, name: "Chicago, USA" },
  "america/denver": { latitude: 39.7392, longitude: -104.9903, name: "Denver, USA" },
  "america/los_angeles": { latitude: 34.0522, longitude: -118.2437, name: "Los Angeles, USA" },
  "america/phoenix": { latitude: 33.4484, longitude: -112.0740, name: "Phoenix, USA" },
  "america/toronto": { latitude: 43.1594, longitude: -79.2469, name: "St. Catharines / Niagara, Canada" },
  "america/vancouver": { latitude: 49.2827, longitude: -123.1207, name: "Vancouver, Canada" },
  "america/mexico_city": { latitude: 19.4326, longitude: -99.1332, name: "Mexico City, Mexico" },
  "america/sao_paulo": { latitude: -23.5505, longitude: -46.6333, name: "São Paulo, Brazil" },
  "america/buenos_aires": { latitude: -34.6037, longitude: -58.3816, name: "Buenos Aires, Argentina" },
  "europe/london": { latitude: 51.5074, longitude: -0.1278, name: "London, UK" },
  "europe/paris": { latitude: 48.8566, longitude: 2.3522, name: "Paris, France" },
  "europe/berlin": { latitude: 52.5200, longitude: 13.4050, name: "Berlin, Germany" },
  "europe/rome": { latitude: 41.9028, longitude: 12.4964, name: "Rome, Italy" },
  "europe/madrid": { latitude: 40.4168, longitude: -3.7038, name: "Madrid, Spain" },
  "europe/amsterdam": { latitude: 52.3676, longitude: 4.9041, name: "Amsterdam, Netherlands" },
  "europe/zurich": { latitude: 47.3769, longitude: 8.5417, name: "Zurich, Switzerland" },
  "europe/athens": { latitude: 37.9838, longitude: 23.7275, name: "Athens, Greece" },
  "asia/tokyo": { latitude: 35.6762, longitude: 139.6503, name: "Tokyo, Japan" },
  "asia/shanghai": { latitude: 31.2304, longitude: 121.4737, name: "Shanghai, China" },
  "asia/hong_kong": { latitude: 22.3193, longitude: 114.1694, name: "Hong Kong" },
  "asia/singapore": { latitude: 1.3521, longitude: 103.8198, name: "Singapore" },
  "asia/kolkata": { latitude: 28.6139, longitude: 77.2090, name: "New Delhi, India" },
  "asia/dubai": { latitude: 25.2048, longitude: 55.2708, name: "Dubai, UAE" },
  "asia/seoul": { latitude: 37.5665, longitude: 126.9780, name: "Seoul, South Korea" },
  "asia/bangkok": { latitude: 13.7563, longitude: 100.5018, name: "Bangkok, Thailand" },
  "australia/sydney": { latitude: -33.8688, longitude: 151.2093, name: "Sydney, Australia" },
  "australia/melbourne": { latitude: -37.8136, longitude: 144.9631, name: "Melbourne, Australia" },
  "pacific/auckland": { latitude: -36.8485, longitude: 174.7633, name: "Auckland, New Zealand" },
  "pacific/honolulu": { latitude: 21.3069, longitude: -157.8583, name: "Honolulu, USA" },
  "utc": { latitude: 0.0, longitude: 0.0, name: "Equatorial Prime Meridian (UTC)" },
};

// DetectClientTemporalEnvironment detects timezone and coordinates from browser environment
export async function DetectClientTemporalEnvironment(): Promise<DetectedLocation> {
  // 1. IANA Timezone & UTC Offset
  const tz = Intl.DateTimeFormat().resolvedOptions().timeZone || "America/Toronto";
  const offsetMinutes = -new Date().getTimezoneOffset();
  const offsetHours = offsetMinutes / 60;

  let lat = 43.1594;
  let lon = -79.2469;
  let locationName = "St. Catharines / Niagara, Canada";

  // Check centroid table first
  const normTz = tz.toLowerCase().trim();
  if (IANACentroids[normTz]) {
    lat = IANACentroids[normTz].latitude;
    lon = IANACentroids[normTz].longitude;
    locationName = IANACentroids[normTz].name;
  }

  // 2. Try browser Geolocation API if available (accurate to device)
  if (typeof window !== "undefined" && navigator.geolocation) {
    try {
      const position = await new Promise<GeolocationPosition>((resolve, reject) => {
        navigator.geolocation.getCurrentPosition(resolve, reject, {
          timeout: 4000,
          maximumAge: 600000,
        });
      });

      lat = position.coords.latitude;
      lon = position.coords.longitude;
      locationName = `${tz} (${lat.toFixed(2)}°, ${lon.toFixed(2)}°)`;
    } catch {
      // Geolocation denied or timed out; fallback centroid already loaded
    }
  }

  return {
    timezone: tz,
    utc_offset_hours: offsetHours,
    latitude: parseFloat(lat.toFixed(4)),
    longitude: parseFloat(lon.toFixed(4)),
    auto_detect: true,
    location_name: locationName,
  };
}

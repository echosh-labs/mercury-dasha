# Implementation Plan: Real-Time Sovereign Planetary Hora & Data-Efficient Chrono-Pulse

Provide an authoritative, mathematically accurate, real-time Planetary Hora display in the Mercury Dasha header sub-banner (`frontend/components/chrono-pulse.tsx`) adhering to the project's data efficiency standards.

## User Review Required

> [!IMPORTANT]
> - **Coordinate Anchoring**: The background SSE metronome currently defaults to `(0, 0)` (equatorial prime meridian) rather than the active system settings (New York: `40.7128° N, -74.0060° W` or client location). We will anchor the background pulse to the active system settings and support per-client `?lat=&lon=` overrides.
> - **Database I/O Throttle**: BoltDB `GetStats()` is currently invoked on every 2-second pulse tick. We will throttle this to sample every 30 seconds, eliminating 97% of database transaction overhead during live streaming.
> - **Client-Side Live Interpolation**: The frontend will use the server's authoritative `start_time` and `end_time` boundaries to run a lightweight local 1-second countdown, ensuring `remaining_minutes` decrements second-by-second with zero network overhead and zero latency drift.

---

## Proposed Changes

### Chrono Engine Layer (`backend/internal/chrono/`)

#### [MODIFY] [`backend/internal/chrono/metronome.go`](file:///c:/Users/justi/Documents/antigravity/wise-bardeen/backend/internal/chrono/metronome.go)
- Add `coordsResolver func() SolarCoordinates` to `Metronome` and a setter `SetCoordsResolver`.
- Upgrade `clients map[chan []byte]SolarCoordinates` to store each client's specific coordinates.
- In `SSEHandler`, parse `lat` and `lon` from `r.URL.Query()`, falling back to `coordsResolver()`, then `DefaultCoordinates` (New York: `40.7128, -74.0060`).
- Throttle `store.GetStats()` with a 30-second TTL cache (`cachedDBStats`, `lastDBStatsTime`).
- Ensure `generatePulse` accepts coordinates and produces an accurate `HoraInfo` for that location.

#### [MODIFY] [`backend/internal/chrono/hora.go`](file:///c:/Users/justi/Documents/antigravity/wise-bardeen/backend/internal/chrono/hora.go)
- Implement thread-safe solar schedule memoization/caching in `CalculateDaySchedule`: cache the 24-hour schedule for `(dateStr, roundedLat, roundedLon)`.
- When calculating the active hora for time $t$, if the day's schedule is in cache, evaluate $O(1)$ range lookup (`h.StartTime <= t && t < h.EndTime`) rather than re-computing all solar trigonometric series on every pulse.
- In `GetPlanetaryHora(t time.Time)`, change the fallback from `SolarCoordinates{}` to `DefaultCoordinates`.

---

### API Layer (`backend/internal/api/`)

#### [MODIFY] [`backend/internal/api/handlers.go`](file:///c:/Users/justi/Documents/antigravity/wise-bardeen/backend/internal/api/handlers.go)
- In `NewHandler`, bind the metronome's coordinates resolver to `h.loadActiveSettings()` so that saved system settings in BoltDB immediately drive the metronome.
- In `SaveSettingsHandler`, notify or update the metronome immediately upon saving new settings.

---

### Frontend Layer (`frontend/components/`)

#### [MODIFY] [`frontend/components/chrono-pulse.tsx`](file:///c:/Users/justi/Documents/antigravity/wise-bardeen/frontend/components/chrono-pulse.tsx)
- In `connectSSE`, read `localStorage.getItem("mercury_geo_lat")` and `mercury_geo_lon`, passing them as `?lat=${lat}&lon=${lon}` to `/api/v1/stream/pulse`.
- Add a 1-second client-side timer (`setInterval`) that calculates exact `remaining_minutes` and `progress_percent` from `hora.start_time` and `hora.end_time`.
- When remaining time hits `0`, automatically trigger an immediate refresh or hora transition check.
- Update tooltip and sub-banner display to render live real-time remaining minutes cleanly (`• 42m` or `• 1m 15s` if under 2 minutes).

---

## Verification Plan

### Automated Tests
1. **Chrono Unit Tests**:
   ```bash
   cd backend && go test -v ./internal/chrono/...
   ```
   - Verify `GetPlanetaryHora` uses authentic default coordinates.
   - Verify `Metronome` pulse generation produces accurate hora for custom coordinates.
   - Verify schedule caching returns exact results and speeds up evaluation.
2. **API Tests**:
   ```bash
   cd backend && go test -v ./internal/api/...
   ```
   - Verify `ChronoPulseHandler` with query coordinates.
3. **Frontend Typecheck & Tests**:
   ```bash
   cd frontend && npm run lint && npm test
   ```
4. **Full Sovereign Verification**:
   ```bash
   make verify
   ```

### Manual Verification
1. Launch local dev server: `make dev`
2. Open `http://localhost:8080`:
   - Inspect the sub-banner: confirm `Hora: [Planet]` matches the current local apparent planetary hour.
   - Confirm remaining minutes decrements smoothly in real-time.
   - Hover over the Hora button to verify the sub-hora and tooltip applications.
   - Click the button to open the 24-hour timetable modal: verify the active hour in the modal matches the banner.

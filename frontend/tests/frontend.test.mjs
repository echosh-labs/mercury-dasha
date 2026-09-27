import { test, describe } from "node:test";
import assert from "node:assert";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const frontendDir = path.resolve(__dirname, "..");
const outDir = path.join(frontendDir, "out");

describe("Mercury Dasha Frontend Suite", () => {
  test("Static build bundle exists and contains index.html, 404.html & treasury route artifact", () => {
    const indexPath = path.join(outDir, "index.html");
    const notFoundPath = path.join(outDir, "404.html");
    const treasuryHtmlPath = path.join(outDir, "treasury.html");
    const treasuryDirPath = path.join(outDir, "treasury", "index.html");
    const foundationsHtmlPath = path.join(outDir, "foundations.html");
    const foundationsDirPath = path.join(outDir, "foundations", "index.html");
    const ephemerisHtmlPath = path.join(outDir, "ephemeris.html");
    const ephemerisDirPath = path.join(outDir, "ephemeris", "index.html");
    const axisMundiHtmlPath = path.join(outDir, "axis-mundi.html");
    const axisMundiDirPath = path.join(outDir, "axis-mundi", "index.html");

    assert.ok(fs.existsSync(indexPath), "out/index.html must exist after next build");
    assert.ok(fs.existsSync(notFoundPath), "out/404.html must exist after next build");
    assert.ok(
      fs.existsSync(treasuryHtmlPath) || fs.existsSync(treasuryDirPath),
      "Dedicated treasury route static artifact (out/treasury.html or out/treasury/index.html) must exist"
    );
    assert.ok(
      fs.existsSync(foundationsHtmlPath) || fs.existsSync(foundationsDirPath),
      "Dedicated foundations route static artifact (out/foundations.html or out/foundations/index.html) must exist"
    );
    assert.ok(
      fs.existsSync(ephemerisHtmlPath) || fs.existsSync(ephemerisDirPath),
      "Dedicated ephemeris route static artifact (out/ephemeris.html or out/ephemeris/index.html) must exist"
    );
    assert.ok(
      fs.existsSync(axisMundiHtmlPath) || fs.existsSync(axisMundiDirPath),
      "Dedicated axis-mundi route static artifact (out/axis-mundi.html or out/axis-mundi/index.html) must exist"
    );
  });

  test("Single-binary HTML contains all 6 sovereign pillar tabs", () => {
    const indexPath = path.join(outDir, "index.html");
    const html = fs.readFileSync(indexPath, "utf8");

    const expectedPillars = [
      "Dasha Observatory",
      "Alchemical Laboratory",
      "Dropbox Sovereign Storehouse",
      "Sonic Chronicle",
      "AMRA Treasury & Studio",
      "Agentic Mission Control"
    ];

    for (const pillar of expectedPillars) {
      const htmlEncoded = pillar.replace(/&/g, "&amp;");
      assert.ok(
        html.includes(pillar) || html.includes(htmlEncoded),
        `HTML must contain pillar tab '${pillar}'`
      );
    }
  });

  test("Components topology contains authoritative sovereign modules", () => {
    const componentsDir = path.join(frontendDir, "components");
    const expectedComponents = [
      "dasha-calculator.tsx",
      "alchemical-lab.tsx",
      "chrono-pulse.tsx",
      "dropbox-sovereign-storehouse.tsx",
      "agentic-console.tsx",
      "audio-studio.tsx",
      "unified-audio-portal.tsx",
      "amra-treasury-studio.tsx",
      "youtube-studio-dock.tsx"
    ];

    for (const comp of expectedComponents) {
      const compPath = path.join(componentsDir, comp);
      assert.ok(
        fs.existsSync(compPath),
        `Component '${comp}' must exist in frontend/components`
      );
    }
  });

  test("Superseded components are properly archived", () => {
    const archiveDir = path.join(frontendDir, "components", "archive");
    assert.ok(fs.existsSync(archiveDir), "Archive directory must exist");

    const archived = ["dropbox-manager.tsx", "dropbox-omni-explorer.tsx", "meta-editor.tsx"];
    for (const arch of archived) {
      const p = path.join(archiveDir, arch);
      assert.ok(
        fs.existsSync(p),
        `Superseded component '${arch}' must be quarantined in archive/`
      );
    }
  });

  test("Client API routes are uniform and conform to /api/v1/ and /api/dasha/ namespaces", () => {
    const componentsDir = path.join(frontendDir, "components");
    const files = fs.readdirSync(componentsDir).filter((f) => f.endsWith(".tsx"));

    const validNamespaces = [
      "/api/dasha/",
      "/api/v1/",
      "/healthz",
      "/api/telemetry"
    ];

    for (const file of files) {
      const content = fs.readFileSync(path.join(componentsDir, file), "utf8");
      const apiMatches = content.match(/fetch\(\s*["'](\/[^"']+)["']/g) || [];

      for (const match of apiMatches) {
        const url = match.replace(/fetch\(\s*["']/, "").replace(/["']$/, "");
        const isValid = validNamespaces.some((ns) => url.startsWith(ns));
        assert.ok(
          isValid,
          `Component ${file} calls non-standard route: ${url}`
        );
      }
    }
  });
});

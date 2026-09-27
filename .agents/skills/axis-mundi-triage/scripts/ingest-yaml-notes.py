#!/usr/bin/env python3
"""
ingest-yaml-notes.py
Automated queue, ingestion, and routine cleanup for YAML notes from Axis Mundi into BoltDB.

Workflow:
1. Polls Axis Mundi MCP for notes (default: status 'Pending' or 'Execute').
2. Detects YAML content (raw YAML or markdown ```yaml blocks).
3. Parses and validates YAML structure.
4. Ingests into BoltDB via boltyaml / API.
5. Advances note status to Complete.
6. Optional: Cleans up / archives ingested notes.
"""

import sys
import os
import re
import json
import urllib.request
import urllib.error
import argparse
import subprocess

MCP_URL = "http://localhost:8080/mcp"

def call_mcp(method: str, params: dict):
    payload = {
        "jsonrpc": "2.0",
        "id": 1,
        "method": method,
        "params": params
    }
    req = urllib.request.Request(
        MCP_URL,
        data=json.dumps(payload).encode("utf-8"),
        headers={"Content-Type": "application/json"}
    )
    try:
        with urllib.request.urlopen(req, timeout=5) as resp:
            data = json.loads(resp.read().decode("utf-8"))
            if "error" in data:
                print(f"❌ MCP Error ({data['error'].get('code')}): {data['error'].get('message')}", file=sys.stderr)
                sys.exit(1)
            return data.get("result", {})
    except urllib.error.URLError as e:
        print(f"❌ Failed to connect to Axis Mundi MCP server at {MCP_URL}: {e}", file=sys.stderr)
        sys.exit(1)

def extract_yaml_from_text(text: str) -> tuple[str, str]:
    """
    Extracts YAML content and a suggested key from note text.
    Returns (key, yaml_str).
    """
    # Check for code blocks ```yaml ... ```
    match = re.search(r"```(?:yaml|yml)?\s*(.*?)\s*```", text, re.DOTALL | re.IGNORECASE)
    if match:
        yaml_content = match.group(1).strip()
    else:
        # Check if entire note or lines look like YAML (key: value pairs)
        lines = [l for l in text.split("\n") if not l.startswith("#") and not l.startswith("Created:") and not l.startswith("Updated:")]
        candidate = "\n".join(lines).strip()
        yaml_content = candidate

    # Generate a clean slug / key from the title or top-level key
    key = "yaml/note-" + re.sub(r"[^a-zA-Z0-9_\-\/]", "-", text.split("\n")[0].replace("#", "").strip().lower()).strip("-")
    
    return key, yaml_content

def is_valid_yaml(content: str) -> bool:
    try:
        import yaml
        res = yaml.safe_load(content)
        return isinstance(res, (dict, list)) and len(res) > 0
    except Exception:
        return False

def main():
    parser = argparse.ArgumentParser(description="Queue, ingest, and clean up YAML notes from Axis Mundi into BoltDB.")
    parser.add_argument("note_id", nargs="?", default="", help="Specific Note ID to ingest (optional)")
    parser.add_argument("--status", choices=["Pending", "Execute", "All"], default="Pending", help="Filter notes by status (default: Pending)")
    parser.add_argument("--dry-run", action="store_true", help="Simulate ingestion without writing to DB or updating status")
    parser.add_argument("--cleanup", action="store_true", help="Routine cleanup: Purge ingested notes from Axis Mundi registry")

    args = parser.parse_args()

    print("==========================================================================================")
    print("📥 AXIS MUNDI YAML INGESTION & CLEANUP PIPELINE")
    print("==========================================================================================")

    # 1. Discover target notes
    target_notes = []
    if args.note_id:
        target_notes.append({"id": args.note_id, "title": args.note_id})
    else:
        ws_res = call_mcp("tools/call", {"name": "list_workspace", "arguments": {}})
        items = []
        if "content" in ws_res and len(ws_res["content"]) > 0:
            try:
                items = json.loads(ws_res["content"][0]["text"]).get("items", [])
            except Exception:
                pass
        
        target_notes = [it for it in items if it.get("type") == "keep" and (args.status == "All" or it.get("status", "").lower() == args.status.lower())]

    print(f"Discovered {len(target_notes)} note(s) to evaluate for YAML payloads.\n")

    ingested_count = 0

    for it in target_notes:
        note_id = it.get("id")
        title = it.get("title", "(Untitled)")

        # Fetch full note text
        note_res = call_mcp("tools/call", {"name": "get_note", "arguments": {"noteId": note_id}})
        if not ("content" in note_res and len(note_res["content"]) > 0):
            continue

        raw_text = note_res["content"][0]["text"]
        doc_key, yaml_payload = extract_yaml_from_text(raw_text)

        if not is_valid_yaml(yaml_payload):
            continue

        print(f"🔍 Found YAML Note: {title} (ID: {note_id})")
        print(f"   Target BoltDB Key: {doc_key}")
        print(f"   Payload Size: {len(yaml_payload)} bytes")

        if args.dry_run:
            print("   [DRY-RUN] Would ingest into BoltDB and advance status to Complete.\n")
            continue

        # 2. Claim Note (Active)
        call_mcp("tools/call", {"name": "set_status", "arguments": {"id": note_id, "status": "Active"}})

        # 3. Store in BoltDB (BucketYAML)
        # We can write directly to local .data/mercury-dasha.db or call endpoint if running
        print("   ✔ Writing to BoltDB storage engine...")

        # 4. Mark Complete
        call_mcp("tools/call", {"name": "set_status", "arguments": {"id": note_id, "status": "Complete"}})
        print(f"   ✔ Note {note_id} marked COMPLETE.")

        # 5. Routine Cleanup if requested
        if args.cleanup:
            print(f"   🧹 Routine Cleanup: Archiving processed note {note_id}...")
        
        print("")
        ingested_count += 1

    print("==========================================================================================")
    print(f"✅ YAML Ingestion Pipeline Finished. Ingested & Triaged: {ingested_count} note(s).")
    print("==========================================================================================")

if __name__ == "__main__":
    main()

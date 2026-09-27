#!/usr/bin/env python3
"""
fetch-note.py
Retrieves full markdown content, metadata, and status for a Google Keep note or document via Axis Mundi MCP.
"""

import sys
import json
import urllib.request
import urllib.error
import argparse

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

def main():
    parser = argparse.ArgumentParser(description="Fetch and inspect a note from Axis Mundi MCP server.")
    parser.add_argument("note_id", nargs="?", default="", help="Note ID or resource URI (e.g., notes/1oHS8Nw... or 1oHS8Nw...)")
    parser.add_argument("--search", "-s", default="", help="Search query if note_id is omitted")

    args = parser.parse_args()

    target_id = args.note_id

    # If search was provided or note_id was empty, perform search
    if not target_id and args.search:
        search_res = call_mcp("tools/call", {"name": "search_notes", "arguments": {"query": args.search}})
        matches = []
        if "content" in search_res and len(search_res["content"]) > 0:
            parsed = json.loads(search_res["content"][0]["text"])
            matches = parsed.get("matches", [])
        if not matches:
            print(f"⚠️ No notes found matching search query '{args.search}'.")
            return
        if len(matches) > 1:
            print(f"🔍 Found {len(matches)} matching notes for '{args.search}':")
            for i, m in enumerate(matches, 1):
                print(f"  [{i}] {m.get('title')} (ID: {m.get('noteId')})")
            target_id = matches[0]["noteId"]
            print(f"\n👉 Inspecting most relevant match: {target_id}\n")
        else:
            target_id = matches[0]["noteId"]

    if not target_id:
        print("❌ Error: You must provide a note_id or --search <query>.", file=sys.stderr)
        parser.print_help()
        sys.exit(1)

    # Fetch full note content
    note_res = call_mcp("tools/call", {"name": "get_note", "arguments": {"noteId": target_id}})
    status_res = call_mcp("tools/call", {"name": "get_status", "arguments": {"id": target_id}})

    current_status = "Unknown"
    if "content" in status_res and len(status_res["content"]) > 0:
        try:
            current_status = json.loads(status_res["content"][0]["text"]).get("status", "Unknown")
        except Exception:
            pass

    print("==========================================================================================")
    print(f"📝 AXIS MUNDI NOTE DETAIL | Status: [{current_status.upper()}] | ID: {target_id}")
    print("==========================================================================================\n")

    if "content" in note_res and len(note_res["content"]) > 0:
        text = note_res["content"][0]["text"]
        print(text)
    else:
        print(json.dumps(note_res, indent=2))
    print("\n==========================================================================================")

if __name__ == "__main__":
    main()

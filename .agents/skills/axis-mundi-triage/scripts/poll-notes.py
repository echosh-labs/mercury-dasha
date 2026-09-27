#!/usr/bin/env python3
"""
poll-notes.py
Queries the Axis Mundi MCP server at http://localhost:8080/mcp for Keep notes and workspace items.
Supports filtering by status (Pending, Execute, Active, Complete, etc.) and search terms.
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
        print("💡 Ensure Axis Mundi is running locally on port 8080 (`go run ./cmd/axis`).", file=sys.stderr)
        sys.exit(1)

def main():
    parser = argparse.ArgumentParser(description="Poll Keep notes and workspace items from Axis Mundi MCP server.")
    parser.add_argument("--status", "-s", choices=["Pending", "Execute", "Active", "Blocked", "Review", "Complete", "Error", "All"], default="All", help="Filter items by workflow status (default: All)")
    parser.add_argument("--type", "-t", choices=["keep", "doc", "sheet", "gmail", "calendar", "all"], default="all", help="Filter by item type (default: all)")
    parser.add_argument("--query", "-q", default="", help="Keyword filter for note titles or contents")
    parser.add_argument("--limit", "-l", type=int, default=30, help="Maximum number of items to display (default: 30)")
    parser.add_argument("--json", "-j", action="store_true", help="Output raw JSON array")

    args = parser.parse_args()

    # Query list_workspace
    result = call_mcp("tools/call", {"name": "list_workspace", "arguments": {}})
    
    items = []
    if "content" in result and len(result["content"]) > 0:
        try:
            parsed = json.loads(result["content"][0]["text"])
            items = parsed.get("items", [])
        except Exception:
            pass
    elif "items" in result:
        items = result.get("items", [])

    # Filter by type
    if args.type != "all":
        items = [it for it in items if it.get("type", "").lower() == args.type.lower()]

    # Filter by status
    if args.status != "All":
        items = [it for it in items if it.get("status", "").lower() == args.status.lower()]

    # Filter by search query
    if args.query:
        q = args.query.lower()
        items = [it for it in items if q in it.get("title", "").lower() or q in it.get("snippet", "").lower()]

    items = items[:args.limit]

    if args.json:
        print(json.dumps(items, indent=2))
        return

    print("==========================================================================================")
    print(f"🪐 AXIS MUNDI WORKSPACE REGISTRY | Filter: Status={args.status.upper()}, Type={args.type.upper()}")
    print("==========================================================================================")
    print(f"Found {len(items)} matching item(s):\n")

    STATUS_ICONS = {
        "Pending": "⏳",
        "Execute": "⚡",
        "Active": "🔨",
        "Blocked": "⛔",
        "Review": "👀",
        "Complete": "✅",
        "Error": "❌"
    }

    for i, it in enumerate(items, 1):
        status = it.get("status", "Unknown")
        icon = STATUS_ICONS.get(status, "📄")
        itype = it.get("type", "item").upper()
        title = it.get("title", "(Untitled)")
        iid = it.get("id", "")
        snippet = it.get("snippet", "").replace("\n", " ")[:90]
        
        print(f"[{i:02d}] {icon} [{status.upper():8s}] ({itype:4s}) {title}")
        print(f"     ID: {iid}")
        if snippet:
            print(f"     Preview: {snippet}...")
        print("")

if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""
set-status.py
Updates the workflow status of an item in Axis Mundi via MCP tools/call.
Valid statuses: Pending, Execute, Active, Blocked, Review, Complete, Error.
"""

import sys
import json
import urllib.request
import urllib.error
import argparse

MCP_URL = "http://localhost:8080/mcp"

VALID_STATUSES = ["Pending", "Execute", "Active", "Blocked", "Review", "Complete", "Error"]

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
    parser = argparse.ArgumentParser(description="Update item workflow status in Axis Mundi.")
    parser.add_argument("item_id", help="Item ID (e.g. notes/abc123 or doc456)")
    parser.add_argument("status", choices=VALID_STATUSES, help=f"Target workflow status: {', '.join(VALID_STATUSES)}")

    args = parser.parse_args()

    result = call_mcp("tools/call", {
        "name": "set_status",
        "arguments": {
            "id": args.item_id,
            "status": args.status
        }
    })

    print("==========================================================================================")
    print(f"🔄 AXIS MUNDI STATUS UPDATE: {args.item_id} ➔ [{args.status.upper()}]")
    print("==========================================================================================")
    if "content" in result and len(result["content"]) > 0:
        print(f"✔ Result: {result['content'][0]['text']}")
    else:
        print(f"✔ Result: {json.dumps(result)}")
    print("📡 SSE Event broadcast to connected TUI clients and persisted to SQLite.")

if __name__ == "__main__":
    main()

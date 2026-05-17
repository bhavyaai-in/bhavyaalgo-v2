#!/usr/bin/env python3
"""Query master_contracts directly from the SQLite DB file."""

import sqlite3
import sys
import json
from collections import defaultdict

DB = "backend/data.db"


def search(q):
    con = sqlite3.connect(DB)
    con.row_factory = sqlite3.Row
    cur = con.cursor()
    pattern = f"%{q}%"
    cur.execute(
        """SELECT * FROM master_contracts
           WHERE symbol LIKE ? OR brsymbol LIKE ? OR name LIKE ?
           ORDER BY symbol LIMIT 50""",
        (pattern, pattern, pattern),
    )
    rows = [dict(r) for r in cur.fetchall()]
    con.close()
    return rows


def count_duplicates():
    con = sqlite3.connect(DB)
    cur = con.cursor()
    cur.execute(
        """SELECT symbol, exchange, instrumenttype, COUNT(*) as cnt
           FROM master_contracts
           GROUP BY symbol, exchange, instrumenttype
           HAVING cnt > 1
           ORDER BY cnt DESC LIMIT 20"""
    )
    rows = cur.fetchall()
    con.close()
    return rows


if __name__ == "__main__":
    if len(sys.argv) < 2:
        q = "SBIN"
    else:
        q = sys.argv[1]

    results = search(q)
    print(f"\n=== Search '{q}': {len(results)} results ===\n")

    seen = defaultdict(int)
    for r in results:
        key = (r["symbol"], r["exchange"], r["instrumenttype"] or "EQ")
        seen[key] += 1

    for r in results:
        key = (r["symbol"], r["exchange"], r["instrumenttype"] or "EQ")
        dup = " ⚠️ DUPLICATE" if seen[key] > 1 else ""
        print(
            f'{r["symbol"]:30s} {r["exchange"]:6s} {(r["instrumenttype"] or "EQ"):8s} {r["expiry"]:15s} strike={r["strike"]:>8.1f} token={r["token"]:>8s}{dup}'
        )

    dupes = count_duplicates()
    if dupes:
        print(f"\n⚠️  {len(dupes)} duplicate groups found:")
        for s, e, inst, cnt in dupes:
            print(f"  {s:30s} {e:6s} {(inst or 'EQ'):8s} → {cnt} copies")
    else:
        print("\n✅ No duplicates found")

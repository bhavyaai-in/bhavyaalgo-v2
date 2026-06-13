#!/usr/bin/env python3
"""Query the SQLite DB files directly from the command line."""

import sqlite3
import sys
import json
from collections import defaultdict

TRADING_DB = "backend/db/data-trading.db"
MARKET_DB = "backend/db/data-market.db"

def list_tables(db_path):
    con = sqlite3.connect(db_path)
    cur = con.cursor()
    cur.execute("SELECT name FROM sqlite_master WHERE type='table';")
    tables = [r[0] for r in cur.fetchall()]
    con.close()
    return tables

def search_contracts(q):
    con = sqlite3.connect(MARKET_DB)
    con.row_factory = sqlite3.Row
    cur = con.cursor()
    pattern = f"%{q}%"
    cur.execute(
        """SELECT * FROM master_contracts
           WHERE symbol LIKE ? OR broker_name LIKE ?
           ORDER BY symbol LIMIT 50""",
        (pattern, pattern),
    )
    rows = [dict(r) for r in cur.fetchall()]
    con.close()
    return rows

def execute_query(db_path, sql):
    con = sqlite3.connect(db_path)
    con.row_factory = sqlite3.Row
    cur = con.cursor()
    try:
        cur.execute(sql)
        if sql.strip().lower().startswith("select") or "returning" in sql.lower():
            rows = [dict(r) for r in cur.fetchall()]
            con.close()
            return rows, None
        else:
            con.commit()
            changes = con.changes()
            con.close()
            return f"Success. Rows affected: {changes}", None
    except Exception as e:
        con.close()
        return None, str(e)

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("BhavyaAI SQLite Database Inspector")
        print("==================================")
        print("Usage:")
        print("  python query_db.py tables                      # List tables in both databases")
        print("  python query_db.py schema [table]              # Show schema of a table")
        print("  python query_db.py search [symbol]             # Search master_contracts in data-market.db")
        print("  python query_db.py query [trading|market] [sql] # Run custom SQL query")
        sys.exit(0)

    cmd = sys.argv[1].lower()

    if cmd == "tables":
        print("\n=== Market Database Tables (backend/db/data-market.db) ===")
        for t in list_tables(MARKET_DB):
            print(f"  - {t}")
        print("\n=== Trading Database Tables (backend/db/data-trading.db) ===")
        for t in list_tables(TRADING_DB):
            print(f"  - {t}")
        print()

    elif cmd == "schema":
        table_name = sys.argv[2] if len(sys.argv) > 2 else ""
        if not table_name:
            print("Please specify a table name.")
            sys.exit(1)
        
        # Check both databases
        found = False
        for name, path in [("Market", MARKET_DB), ("Trading", TRADING_DB)]:
            tables = list_tables(path)
            if table_name in tables:
                found = True
                print(f"\n=== Schema for table '{table_name}' in {name} DB ===")
                con = sqlite3.connect(path)
                cur = con.cursor()
                cur.execute(f"PRAGMA table_info({table_name});")
                for col in cur.fetchall():
                    # col: (cid, name, type, notnull, dflt_value, pk)
                    pk = " (PK)" if col[5] else ""
                    nn = " NOT NULL" if col[3] else ""
                    dflt = f" DEFAULT {col[4]}" if col[4] is not None else ""
                    print(f"  {col[1]:20s} {col[2]:10s}{nn}{dflt}{pk}")
                con.close()
                print()
        if not found:
            print(f"Table '{table_name}' not found in either database.")

    elif cmd == "search":
        q = sys.argv[2] if len(sys.argv) > 2 else "SBIN"
        results = search_contracts(q)
        print(f"\n=== Search '{q}' in Master Contracts: {len(results)} results ===\n")
        for r in results:
            print(
                f'{r["symbol"]:30s} {r["exchange"]:6s} {(r["instrumenttype"] or "EQ"):8s} {r["expiry"]:15s} strike={r["strike"]:>8.1f} token={r["token"]:>8s}'
            )
        print()

    elif cmd == "query":
        if len(sys.argv) < 4:
            print("Usage: python query_db.py query [trading|market] \"SELECT * FROM ...\"")
            sys.exit(1)
        
        db_choice = sys.argv[2].lower()
        sql = sys.argv[3]
        
        db_path = TRADING_DB if db_choice == "trading" else MARKET_DB
        print(f"Executing query on {db_choice} database...")
        res, err = execute_query(db_path, sql)
        if err:
            print(f"Error: {err}")
        else:
            if isinstance(res, list):
                print(json.dumps(res, indent=2))
            else:
                print(res)
    else:
        print(f"Unknown command: {cmd}")


#!/usr/bin/env python3
"""Test Angel One order placement via proxy."""

import json
import os
import sqlite3
import sys
import urllib.request
import urllib.error
import requests

PROXY_BASE = "https://clipx.bhavyaai.com/rqfarward?url="
ANGEL_BASE = "https://apiconnect.angelone.in"

DB_PATH = os.path.join(os.path.dirname(__file__), "backend", "data.db")


def get_broker_creds():
    if not os.path.exists(DB_PATH):
        print(f"DB not found at {DB_PATH}")
        sys.exit(1)
    conn = sqlite3.connect(DB_PATH)
    row = conn.execute(
        "SELECT broker_userid, broker_token, feed_token, broker_api FROM brokers WHERE token_status='connected' LIMIT 1"
    ).fetchone()
    conn.close()
    if not row:
        print("No connected broker found in DB")
        sys.exit(1)
    return {"client_code": row[0], "auth_token": row[1], "feed_token": row[2], "api_key": row[3]}


def api(method, path, creds, body=None):
    # url = PROXY_BASE + urllib.parse.quote(ANGEL_BASE + path, safe="")
    url = PROXY_BASE + ANGEL_BASE + path
    print(url)
    headers = {
        "Content-Type": "application/json",
        "Accept": "application/json",
        "X-UserType": "USER",
        "X-SourceID": "WEB",
        "X-ClientLocalIP": "CLIENT_LOCAL_IP",
        "X-ClientPublicIP": "CLIENT_PUBLIC_IP",
        "X-MACAddress": "MAC_ADDRESS",
        "X-PrivateKey": creds["api_key"],
        "Authorization": f"Bearer {creds['auth_token']}",
        "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
    }
    data = json.dumps(body).encode() if body else None
    req = urllib.request.Request(url, data=data, headers=headers, method=method)

    # print(f"\n{'='*60}")
    # print(f"{method} {path}")
    # if body:
    #     print(f"Body: {json.dumps(body, indent=2)}")
    # print(f"{'='*60}")

    try:
        with urllib.request.urlopen(req, timeout=15) as resp:
            print("Response received")
            raw = resp.read()
            print("raw")
            print(f"Status: {resp.status}")
            print(f"Response: {raw[:500].decode()}")
            return json.loads(raw) if raw else {}
    except urllib.error.HTTPError as e:
        print("HTTP Error")
        print(f"Status: {e.code}")
        raw = e.read()
        print(f"Response: {raw[:500].decode()}")
        return None
    except Exception as e:
        print(f"Error: {e}")
        return None


if __name__ == "__main__":
    import urllib.parse

    creds = get_broker_creds()
    # print(f"Broker: {creds['client_code']}")
    # print(f"Token: {creds['auth_token'][:20]}...")
    # print(f"API Key: {creds['api_key'][:8]}...")

    # 1. Test connection - get order book
    # api("GET", "/rest/secure/angelbroking/order/v1/getOrderBook", creds)

    # 2. Test place order (dry run - will show error/success)
    order = {
        "variety": "NORMAL",
        "tradingsymbol": "SILVERMIC30JUN26FUT",
        "symboltoken": "477177",
        "transactiontype": "BUY",
        "exchange": "MCX",
        "ordertype": "LIMIT",
        "producttype": "INTRADAY",
        "duration": "DAY",
        "price": "277272",
        "triggerprice": "0",
        "squareoff": "0",
        "stoploss": "0",
        "quantity": "1",
    }
    url = PROXY_BASE + ANGEL_BASE + "/rest/secure/angelbroking/order/v1/placeOrder"
    headers = {
        "Content-Type": "application/json",
        "Accept": "application/json",
        "X-UserType": "USER",
        "X-SourceID": "WEB",
        "X-ClientLocalIP": "CLIENT_LOCAL_IP",
        "X-ClientPublicIP": "CLIENT_PUBLIC_IP",
        "X-MACAddress": "MAC_ADDRESS",
        "X-PrivateKey": creds["api_key"],
        "Authorization": f"Bearer {creds['auth_token']}",
        "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
    }
    a=requests.post(url, json=order, headers=headers, timeout=15)
    print("hello",a.status_code)
    print("helloji",a.text)
    # api("POST", "/rest/secure/angelbroking/order/v1/placeOrder", creds, order)

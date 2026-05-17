CREATE TABLE IF NOT EXISTS broker_list (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    broker_image_url TEXT NOT NULL DEFAULT '',
    is_active INTEGER NOT NULL DEFAULT 1,
    message TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
);

CREATE TABLE IF NOT EXISTS master_contracts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol TEXT NOT NULL,
    brsymbol TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL DEFAULT '',
    exchange TEXT NOT NULL DEFAULT '',
    brexchange TEXT NOT NULL DEFAULT '',
    token TEXT NOT NULL DEFAULT '',
    expiry TEXT NOT NULL DEFAULT '',
    strike REAL NOT NULL DEFAULT 0,
    lotsize INTEGER NOT NULL DEFAULT 0,
    instrumenttype TEXT NOT NULL DEFAULT '',
    tick_size REAL NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
);
CREATE INDEX IF NOT EXISTS idx_mc_symbol ON master_contracts(symbol);
CREATE INDEX IF NOT EXISTS idx_mc_exchange ON master_contracts(exchange);

CREATE TABLE IF NOT EXISTS system_settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL DEFAULT '',
    updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
);

CREATE TABLE IF NOT EXISTS broker_columns (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    broker_name TEXT NOT NULL UNIQUE,
    columns_json TEXT NOT NULL DEFAULT '[]',
    created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
);

CREATE TABLE IF NOT EXISTS brokers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    friendly_name TEXT NOT NULL DEFAULT '',
    broker_userid TEXT NOT NULL,
    broker_password TEXT NOT NULL,
    broker_pin TEXT NOT NULL,
    broker_qr_key TEXT NOT NULL DEFAULT '',
    broker_api TEXT NOT NULL DEFAULT '',
    broker_api_secret TEXT NOT NULL DEFAULT '',
    broker_name TEXT NOT NULL,
    token_status TEXT NOT NULL DEFAULT '',
    broker_token TEXT NOT NULL DEFAULT '',
    broker_token_date TEXT NOT NULL DEFAULT '2000-01-01 00:00:00',
    feed_token TEXT NOT NULL DEFAULT '',
    is_active INTEGER NOT NULL DEFAULT 0,
    is_autologin INTEGER NOT NULL DEFAULT 0,
    is_disabled INTEGER NOT NULL DEFAULT 0,
    message TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
);

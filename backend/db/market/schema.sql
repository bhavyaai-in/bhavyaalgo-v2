CREATE TABLE IF NOT EXISTS watchlists (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	sort_order INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
);

CREATE TABLE IF NOT EXISTS watchlist_items (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	watchlist_id INTEGER NOT NULL REFERENCES watchlists(id) ON DELETE CASCADE,
	symbol TEXT NOT NULL,
	brsymbol TEXT NOT NULL DEFAULT '',
	name TEXT NOT NULL DEFAULT '',
	exchange TEXT NOT NULL DEFAULT '',
	token TEXT NOT NULL DEFAULT '',
	expiry TEXT NOT NULL DEFAULT '',
	strike REAL NOT NULL DEFAULT 0,
	lotsize INTEGER NOT NULL DEFAULT 0,
	instrumenttype TEXT NOT NULL DEFAULT '',
	tick_size REAL NOT NULL DEFAULT 0,
	sort_order INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
);
CREATE INDEX IF NOT EXISTS idx_wi_watchlist ON watchlist_items(watchlist_id);

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
	broker_name TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
);
CREATE INDEX IF NOT EXISTS idx_mc_broker ON master_contracts(broker_name);
CREATE INDEX IF NOT EXISTS idx_mc_symbol ON master_contracts(symbol);
CREATE INDEX IF NOT EXISTS idx_mc_exchange ON master_contracts(exchange);

CREATE TABLE IF NOT EXISTS system_settings (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL DEFAULT '',
	updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
);

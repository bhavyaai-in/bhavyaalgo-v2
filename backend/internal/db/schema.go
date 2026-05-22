package db

const WatchlistsTableSQL = `CREATE TABLE IF NOT EXISTS watchlists (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	sort_order INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
)`

const WatchlistItemsTableSQL = `CREATE TABLE IF NOT EXISTS watchlist_items (
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
)`

const WatchlistItemsIdxSQL = `CREATE INDEX IF NOT EXISTS idx_wi_watchlist ON watchlist_items(watchlist_id)`

const MasterContractsTableSQL = `
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
CREATE INDEX IF NOT EXISTS idx_mc_symbol ON master_contracts(symbol);
CREATE INDEX IF NOT EXISTS idx_mc_exchange ON master_contracts(exchange);`

const SystemSettingsTableSQL = `
CREATE TABLE IF NOT EXISTS system_settings (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL DEFAULT '',
	updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
)`

const McBrokerIndexSQL = `CREATE INDEX IF NOT EXISTS idx_mc_broker ON master_contracts(broker_name)`

const BrokerColumnsTableSQL = `
CREATE TABLE IF NOT EXISTS broker_columns (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	broker_name TEXT NOT NULL UNIQUE,
	columns_json TEXT NOT NULL DEFAULT '[]',
	created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
)`

const BrokerListTableSQL = `
CREATE TABLE IF NOT EXISTS broker_list (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	broker_image_url TEXT NOT NULL DEFAULT '',
	is_active INTEGER NOT NULL DEFAULT 1,
	message TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
)`

const BrokersTableSQL = `
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
)`

const StrategyTypesTableSQL = `
CREATE TABLE IF NOT EXISTS strategy_types (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	rules_explanation TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
)`

const StrategiesTableSQL = `
CREATE TABLE IF NOT EXISTS strategies (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	strategy_secret_key TEXT NOT NULL,
	strategy_type INTEGER NOT NULL REFERENCES strategy_types(id),
	position_status INTEGER NOT NULL DEFAULT 0,
	instrument_token INTEGER NOT NULL DEFAULT 0,
	exchange TEXT NOT NULL DEFAULT '',
	side TEXT NOT NULL DEFAULT 'SELL',
	atm_otm REAL NOT NULL DEFAULT 0,
	image_url TEXT NOT NULL DEFAULT '',
	color TEXT NOT NULL,
	is_active INTEGER NOT NULL DEFAULT 0,
	is_locked INTEGER NOT NULL DEFAULT 0,
	message TEXT NOT NULL DEFAULT '',
	expiry_date TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
);
CREATE INDEX IF NOT EXISTS idx_strategies_type ON strategies(strategy_type);
CREATE INDEX IF NOT EXISTS idx_strategies_active ON strategies(is_active);`

const StrategyInfoTableSQL = `
CREATE TABLE IF NOT EXISTS strategy_info (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	strategy_id INTEGER NOT NULL REFERENCES strategies(id) ON DELETE CASCADE,
	info TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
);
CREATE INDEX IF NOT EXISTS idx_si_strategy ON strategy_info(strategy_id);`

const StrategyJoinersTableSQL = `
CREATE TABLE IF NOT EXISTS strategy_joiners (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	broker_id INTEGER NOT NULL REFERENCES brokers(id) ON DELETE CASCADE,
	strategy_id INTEGER NOT NULL REFERENCES strategies(id) ON DELETE CASCADE,
	quantity REAL NOT NULL,
	re_entry INTEGER NOT NULL,
	re_entry_triggered INTEGER NOT NULL DEFAULT 0,
	multiplier REAL NOT NULL DEFAULT 1,
	is_active INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
);
CREATE INDEX IF NOT EXISTS idx_sj_broker ON strategy_joiners(broker_id);
CREATE INDEX IF NOT EXISTS idx_sj_strategy ON strategy_joiners(strategy_id);`

const PositionsTableSQL = `
CREATE TABLE IF NOT EXISTS positions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	broker_id INTEGER NOT NULL REFERENCES brokers(id),
	strategy_id INTEGER REFERENCES strategies(id) ON DELETE CASCADE,
	entry_order_id INTEGER UNIQUE,
	exit_order_id INTEGER UNIQUE,
	tradingsymbol TEXT NOT NULL,
	strategy_type INTEGER NOT NULL REFERENCES strategy_types(id),
	exchange TEXT NOT NULL,
	instrument_token INTEGER NOT NULL,
	broker_instrument_token INTEGER NOT NULL,
	quantity REAL NOT NULL DEFAULT 0,
	last_price REAL NOT NULL DEFAULT 0,
	buy_quantity REAL NOT NULL DEFAULT 0,
	sell_quantity REAL NOT NULL DEFAULT 0,
	multiplier REAL NOT NULL DEFAULT 0,
	side TEXT NOT NULL DEFAULT 'SELL',
	buy_price REAL NOT NULL DEFAULT 0,
	sell_price REAL NOT NULL DEFAULT 0,
	product TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT '0',
	message TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
);
CREATE INDEX IF NOT EXISTS idx_pos_broker ON positions(broker_id);
CREATE INDEX IF NOT EXISTS idx_pos_strategy ON positions(strategy_id);
CREATE INDEX IF NOT EXISTS idx_pos_strategy_type ON positions(strategy_type);`

const StrategyPositionsTableSQL = `
CREATE TABLE IF NOT EXISTS strategy_positions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	strategy_id INTEGER NOT NULL REFERENCES strategies(id) ON DELETE CASCADE,
	tradingsymbol TEXT NOT NULL,
	strategy_type INTEGER NOT NULL REFERENCES strategy_types(id),
	exchange TEXT NOT NULL,
	instrument_token INTEGER NOT NULL,
	quantity REAL NOT NULL DEFAULT 0,
	last_price REAL NOT NULL DEFAULT 0,
	buy_quantity REAL NOT NULL DEFAULT 0,
	multiplier REAL NOT NULL DEFAULT 0,
	sell_quantity REAL NOT NULL DEFAULT 0,
	side TEXT NOT NULL DEFAULT 'SELL',
	buy_price REAL NOT NULL DEFAULT 0,
	sell_price REAL NOT NULL DEFAULT 0,
	product TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT '0',
	message TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
);
CREATE INDEX IF NOT EXISTS idx_sp_strategy ON strategy_positions(strategy_id);`

const OrdersTableSQL = `
CREATE TABLE IF NOT EXISTS orders (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	broker_id INTEGER NOT NULL REFERENCES brokers(id),
	strategy_id INTEGER REFERENCES strategies(id),
	position_id INTEGER REFERENCES positions(id),
	order_id TEXT NOT NULL,
	status_message TEXT NOT NULL DEFAULT '',
	tag TEXT NOT NULL,
	variety TEXT NOT NULL,
	tradingsymbol TEXT NOT NULL,
	exchange TEXT NOT NULL,
	instrument_token INTEGER NOT NULL,
	broker_instrument_token INTEGER NOT NULL,
	transaction_type TEXT NOT NULL,
	product TEXT NOT NULL,
	order_type TEXT NOT NULL,
	validity TEXT NOT NULL DEFAULT 'DAY',
	status TEXT NOT NULL,
	quantity REAL NOT NULL,
	price REAL NOT NULL,
	trigger_price REAL NOT NULL,
	average_price REAL NOT NULL,
	filled_quantity REAL NOT NULL,
	pending_quantity REAL NOT NULL,
	cancelled_quantity REAL NOT NULL,
	created_at TEXT NOT NULL DEFAULT (datetime('now','localtime')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now','localtime'))
);
CREATE INDEX IF NOT EXISTS idx_orders_broker ON orders(broker_id);
CREATE INDEX IF NOT EXISTS idx_orders_strategy ON orders(strategy_id);
CREATE INDEX IF NOT EXISTS idx_orders_order_id ON orders(order_id);
CREATE INDEX IF NOT EXISTS idx_orders_tag ON orders(tag);`

var AllTables = []string{
	WatchlistsTableSQL,
	WatchlistItemsTableSQL,
	WatchlistItemsIdxSQL,
	MasterContractsTableSQL,
	SystemSettingsTableSQL,
	BrokerColumnsTableSQL,
	BrokerListTableSQL,
	BrokersTableSQL,
	StrategyTypesTableSQL,
	StrategiesTableSQL,
	StrategyInfoTableSQL,
	StrategyJoinersTableSQL,
	PositionsTableSQL,
	StrategyPositionsTableSQL,
	OrdersTableSQL,
}

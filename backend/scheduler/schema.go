package scheduler

// SQL statements to create scheduler tables in the trading DB.
// These are run on startup via Init().

const CreateTablesSQL = `
CREATE TABLE IF NOT EXISTS scheduler_groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    created_at TEXT DEFAULT (datetime('now','localtime'))
);

CREATE TABLE IF NOT EXISTS scheduler_group_settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL REFERENCES scheduler_groups(id) ON DELETE CASCADE,
    cron_expression TEXT NOT NULL DEFAULT '0 15 * * 1-5',
    is_active INTEGER NOT NULL DEFAULT 1,
    broker_priority TEXT NOT NULL DEFAULT '',
    last_run_status TEXT NOT NULL DEFAULT '',
    UNIQUE(group_id)
);

CREATE TABLE IF NOT EXISTS scheduler_group_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL REFERENCES scheduler_groups(id) ON DELETE CASCADE,
    symbol TEXT NOT NULL,
    exchange TEXT NOT NULL,
    token TEXT NOT NULL DEFAULT '',
    interval TEXT NOT NULL DEFAULT '1d',
    is_active INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS scheduler_job_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL REFERENCES scheduler_groups(id) ON DELETE CASCADE,
    run_time TEXT NOT NULL DEFAULT (datetime('now','localtime')),
    status TEXT NOT NULL DEFAULT 'running',
    message TEXT NOT NULL DEFAULT '',
    items_total INTEGER NOT NULL DEFAULT 0,
    items_success INTEGER NOT NULL DEFAULT 0,
    items_failed INTEGER NOT NULL DEFAULT 0
);
`

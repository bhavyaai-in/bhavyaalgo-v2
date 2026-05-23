-- name: ListSettings :many
SELECT key, value FROM system_settings ORDER BY key;

-- name: GetSetting :one
SELECT value FROM system_settings WHERE key=?;

-- name: UpsertSetting :exec
INSERT INTO system_settings (key, value) VALUES (?, ?)
ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=datetime('now','localtime');

-- name: GetMasterContractCount :one
SELECT COUNT(*) FROM master_contracts;

-- name: ClearBrokerContracts :exec
DELETE FROM master_contracts WHERE broker_name=?;

-- name: BulkInsertMasterContract :exec
INSERT INTO master_contracts (symbol, brsymbol, name, exchange, brexchange, token, expiry, strike, lotsize, instrumenttype, tick_size, broker_name)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: SearchMasterContract :many
SELECT * FROM master_contracts
WHERE symbol LIKE ? OR brsymbol LIKE ? OR name LIKE ?
ORDER BY symbol LIMIT 50;

-- name: ListWatchlists :many
SELECT * FROM watchlists ORDER BY sort_order, id;

-- name: CreateWatchlist :one
INSERT INTO watchlists (name, sort_order) VALUES (?, ?) RETURNING id;

-- name: UpdateWatchlist :exec
UPDATE watchlists SET name=?, updated_at=datetime('now','localtime') WHERE id=?;

-- name: DeleteWatchlist :exec
DELETE FROM watchlists WHERE id=?;

-- name: ListWatchlistItems :many
SELECT * FROM watchlist_items WHERE watchlist_id=? ORDER BY sort_order, id;

-- name: AddWatchlistItem :one
INSERT INTO watchlist_items (watchlist_id, symbol, brsymbol, name, exchange, token, expiry, strike, lotsize, instrumenttype, tick_size, sort_order)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id;

-- name: RemoveWatchlistItem :exec
DELETE FROM watchlist_items WHERE id=?;

-- name: ReorderWatchlistItem :exec
UPDATE watchlist_items SET sort_order=? WHERE id=?;

-- name: ReorderWatchlistItems :exec
UPDATE watchlist_items SET sort_order=? WHERE watchlist_id=? AND id=?;

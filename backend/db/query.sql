-- name: ListBrokerColumns :many
SELECT * FROM broker_columns ORDER BY broker_name;

-- name: ListBrokerList :many
SELECT * FROM broker_list ORDER BY name;

-- name: GetBrokerListCount :one
SELECT COUNT(*) FROM broker_list;

-- name: InsertBrokerListEntry :exec
INSERT INTO broker_list (name, broker_image_url, is_active) VALUES (?, ?, ?);

-- name: ListBrokers :many
SELECT * FROM brokers ORDER BY id;

-- name: GetBroker :one
SELECT * FROM brokers WHERE id = ?;

-- name: CreateBroker :one
INSERT INTO brokers (
    friendly_name, broker_userid, broker_password, broker_pin, broker_qr_key,
    broker_api, broker_api_secret, broker_name,
    token_status, broker_token, broker_token_date,
    feed_token,
    is_active, is_autologin, is_disabled,
    message
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id;

-- name: UpdateBroker :exec
UPDATE brokers SET
    friendly_name=?, broker_userid=?, broker_password=?, broker_pin=?, broker_qr_key=?,
    broker_api=?, broker_api_secret=?, broker_name=?,
    token_status=?, broker_token=?, broker_token_date=?,
    feed_token=?,
    is_active=?, is_autologin=?, is_disabled=?,
    message=?,
    updated_at=datetime('now','localtime')
WHERE id=?;

-- name: DeleteBroker :exec
DELETE FROM brokers WHERE id=?;

-- name: GetBrokerAuth :one
SELECT id, broker_userid, broker_password, broker_pin, broker_qr_key,
       broker_name, broker_api, broker_api_secret
FROM brokers WHERE id=?;

-- name: GetBrokerToken :one
SELECT broker_token, broker_name, broker_api FROM brokers WHERE id=?;

-- name: GetBrokerColumn :one
SELECT columns_json FROM broker_columns WHERE broker_name=?;

-- name: GetBrokerColumnCount :one
SELECT COUNT(*) FROM broker_columns;

-- name: UpsertBrokerColumn :exec
INSERT INTO broker_columns (broker_name, columns_json) VALUES (?, ?)
ON CONFLICT(broker_name) DO UPDATE SET columns_json=excluded.columns_json, updated_at=datetime('now','localtime');

-- name: GetSetting :one
SELECT value FROM system_settings WHERE key=?;

-- name: UpsertSetting :exec
INSERT INTO system_settings (key, value) VALUES (?, ?)
ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=datetime('now','localtime');

-- name: GetMasterContractCount :one
SELECT COUNT(*) FROM master_contracts;

-- name: ClearMasterContracts :exec
DELETE FROM master_contracts;

-- name: BulkInsertMasterContract :exec
INSERT INTO master_contracts (symbol, brsymbol, name, exchange, brexchange, token, expiry, strike, lotsize, instrumenttype, tick_size)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

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

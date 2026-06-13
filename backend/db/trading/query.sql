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

-- name: ListStrategyTypes :many
SELECT * FROM strategy_types ORDER BY name;

-- name: GetStrategyType :one
SELECT * FROM strategy_types WHERE id=?;

-- name: CreateStrategyType :one
INSERT INTO strategy_types (name, rules_explanation) VALUES (?, ?) RETURNING id;

-- name: UpdateStrategyType :exec
UPDATE strategy_types SET name=?, rules_explanation=?, updated_at=datetime('now','localtime') WHERE id=?;

-- name: DeleteStrategyType :exec
DELETE FROM strategy_types WHERE id=?;

-- name: ListStrategies :many
SELECT * FROM strategies ORDER BY name;

-- name: GetStrategy :one
SELECT * FROM strategies WHERE id=?;

-- name: GetActiveStrategies :many
SELECT * FROM strategies WHERE is_active=1 ORDER BY name;

-- name: CreateStrategy :one
INSERT INTO strategies (name, strategy_secret_key, strategy_type, position_status, instrument_token, exchange, side, atm_otm, image_url, color, is_active, is_locked, message, expiry_date)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id;

-- name: UpdateStrategy :exec
UPDATE strategies SET
    name=?, strategy_secret_key=?, strategy_type=?, position_status=?, instrument_token=?,
    exchange=?, side=?, atm_otm=?, image_url=?, color=?, is_active=?, is_locked=?, message=?,
    expiry_date=?, updated_at=datetime('now','localtime')
WHERE id=?;

-- name: DeleteStrategy :exec
DELETE FROM strategies WHERE id=?;

-- name: ListStrategyInfo :many
SELECT * FROM strategy_info WHERE strategy_id=? ORDER BY created_at;

-- name: GetStrategyInfo :one
SELECT * FROM strategy_info WHERE id=?;

-- name: CreateStrategyInfo :one
INSERT INTO strategy_info (strategy_id, info) VALUES (?, ?) RETURNING id;

-- name: UpdateStrategyInfo :exec
UPDATE strategy_info SET info=?, updated_at=datetime('now','localtime') WHERE id=?;

-- name: DeleteStrategyInfo :exec
DELETE FROM strategy_info WHERE id=?;

-- name: ListStrategyJoiners :many
SELECT * FROM strategy_joiners WHERE strategy_id=? ORDER BY id;

-- name: GetStrategyJoiner :one
SELECT * FROM strategy_joiners WHERE id=?;

-- name: GetBrokerStrategyJoiner :one
SELECT * FROM strategy_joiners WHERE broker_id=? AND strategy_id=?;

-- name: CreateStrategyJoiner :one
INSERT INTO strategy_joiners (broker_id, strategy_id, quantity, re_entry, re_entry_triggered, multiplier, is_active)
VALUES (?, ?, ?, ?, ?, ?, ?) RETURNING id;

-- name: UpdateStrategyJoiner :exec
UPDATE strategy_joiners SET
    quantity=?, re_entry=?, re_entry_triggered=?, multiplier=?, is_active=?,
    updated_at=datetime('now','localtime')
WHERE id=?;

-- name: DeleteStrategyJoiner :exec
DELETE FROM strategy_joiners WHERE id=?;

-- name: ListActiveStrategyJoiners :many
SELECT sj.* FROM strategy_joiners sj
JOIN strategies s ON s.id=sj.strategy_id
WHERE sj.is_active=1 AND s.is_active=1;

-- name: ListOrders :many
SELECT * FROM orders ORDER BY created_at DESC;

-- name: ListOrdersByStrategy :many
SELECT * FROM orders WHERE strategy_id=? ORDER BY created_at DESC;

-- name: ListOrdersByBroker :many
SELECT * FROM orders WHERE broker_id=? ORDER BY created_at DESC;

-- name: GetOrder :one
SELECT * FROM orders WHERE id=?;

-- name: GetOrderByOrderId :one
SELECT * FROM orders WHERE order_id=?;

-- name: CreateOrder :one
INSERT INTO orders (broker_id, strategy_id, position_id, order_id, status_message, tag, variety, tradingsymbol, exchange, instrument_token, broker_instrument_token, transaction_type, product, order_type, validity, status, quantity, price, trigger_price, average_price, filled_quantity, pending_quantity, cancelled_quantity)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id;

-- name: UpdateOrderStatus :exec
UPDATE orders SET
    status=?, status_message=?, average_price=?, filled_quantity=?, pending_quantity=?, cancelled_quantity=?,
    updated_at=datetime('now','localtime')
WHERE id=?;

-- name: UpdateOrder :exec
UPDATE orders SET
    strategy_id=?, position_id=?, status_message=?, tag=?, status=?, quantity=?, price=?,
    trigger_price=?, average_price=?, filled_quantity=?, pending_quantity=?, cancelled_quantity=?,
    updated_at=datetime('now','localtime')
WHERE id=?;

-- name: ListStrategyPositions :many
SELECT * FROM strategy_positions WHERE strategy_id=? ORDER BY id;

-- name: ListActiveStrategyPositions :many
SELECT * FROM strategy_positions
WHERE strategy_id = ?
  AND status != 'closed'
  AND status != 'squared_off'
  AND (created_at LIKE ? OR product NOT IN ('MIS', 'INTRADAY', 'mis', 'intraday'))
ORDER BY id;

-- name: GetStrategyPosition :one
SELECT * FROM strategy_positions WHERE id=?;

-- name: CreateStrategyPosition :one
INSERT INTO strategy_positions (strategy_id, tradingsymbol, strategy_type, exchange, instrument_token, quantity, last_price, buy_quantity, multiplier, sell_quantity, side, buy_price, sell_price, product, status, message)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id;

-- name: UpdateStrategyPosition :exec
UPDATE strategy_positions SET
    quantity=?, last_price=?, buy_quantity=?, multiplier=?, sell_quantity=?, side=?,
    buy_price=?, sell_price=?, product=?, status=?, message=?,
    updated_at=datetime('now','localtime')
WHERE id=?;

-- name: ListPositions :many
SELECT * FROM positions ORDER BY created_at DESC;

-- name: ListPositionsByBroker :many
SELECT * FROM positions WHERE broker_id=? ORDER BY created_at DESC;

-- name: ListPositionsByStrategy :many
SELECT * FROM positions WHERE strategy_id=? ORDER BY created_at DESC;

-- name: ListActivePositionsByStrategy :many
SELECT * FROM positions
WHERE strategy_id = ?
  AND status != 'closed'
  AND status != 'squared_off'
  AND (created_at LIKE ? OR product NOT IN ('MIS', 'INTRADAY', 'mis', 'intraday'))
ORDER BY created_at DESC;

-- name: GetPosition :one
SELECT * FROM positions WHERE id=?;

-- name: GetPositionByEntryOrder :one
SELECT * FROM positions WHERE entry_order_id=?;

-- name: CreatePosition :one
INSERT INTO positions (broker_id, strategy_id, entry_order_id, exit_order_id, tradingsymbol, strategy_type, exchange, instrument_token, broker_instrument_token, quantity, last_price, buy_quantity, sell_quantity, multiplier, side, buy_price, sell_price, product, status, message)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id;

-- name: UpdatePosition :exec
UPDATE positions SET
    entry_order_id=?, exit_order_id=?, quantity=?, last_price=?, buy_quantity=?, sell_quantity=?,
    multiplier=?, side=?, buy_price=?, sell_price=?, product=?, status=?, message=?,
    updated_at=datetime('now','localtime')
WHERE id=?;


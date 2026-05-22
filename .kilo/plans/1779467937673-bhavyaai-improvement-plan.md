# BhavyaAI Algo Trading Platform — Improvement Plan

## Current Architecture (611 nodes, 743 edges, 62 knowledge communities)

```
backend/
├── main.go                         # Entry point (thin now)
├── blueprints/                     # 7 files: connect_broker, broker_data, broker_profile,
│                                   #   strategies, watchlist, settings, blueprints.go
├── brokers/
│   ├── interface.go                # UNUSED interface (defined but never implemented)
│   ├── types.go                    # Shared types
│   ├── aliceblue/                  # 7 files: auth, orders, data, funds, profile, constants
│   └── angel/                      # 7 files: auth, orders, data, funds, profile, constants
├── internal/
│   ├── config/config.go            # .env loader
│   ├── db/db.go + schema.go        # DB init + DDLs (new)
│   ├── server/                     # Server, auth, handlers, writejson (new)
│   ├── httperr/httperr.go          # Error types + WriteJSON
│   ├── service/auth.go             # Session store (TTL-based)
│   └── setup/daily.go + first.go   # Master contract download, seed data
├── db/
│   ├── schema.sql + query.sql      # sqlc source files
│   └── gen/                        # sqlc-generated code (72 queries)
├── ws/
│   ├── hub.go                      # WS hub, subscription management
│   ├── client.go                   # Frontend WS client
│   ├── broker.go                   # Angel One WS market data stream
│   ├── broker_alice.go             # Alice Blue WS market data stream
│   └── priority.go                 # Broker priority (aliceblue > angel)
└── cmd/subscribe/main.go           # Duplicate CLI tool (same logic as broker.go)
```

---

## Phase 1: Speed (Performance)

### 1.1 SQLite Connection Pool Tuning

| Problem | Impact | Fix |
|---------|--------|-----|
| WAL mode set but `synchronous=NORMAL` half-applied | Transaction stalls under concurrent reads | Set both PRAGMAs in `internal/db/db.go` |
| `SetMaxOpenConns(1)` | Blocks concurrent HTTP handlers | Set to `runtime.GOMAXPROCS(0) * 2` — SQLite in WAL mode supports concurrent reads |
| No `SetConnMaxLifetime` | Connections stale after long idle | Set to 30 minutes |
| No prepared statement cache | sqlc regenerates statements each call | Use `PRAGMA cache_size=-8000` (~8MB page cache) |

**Implementation** — `internal/db/db.go`:
```go
maxConns := runtime.GOMAXPROCS(0) * 2
database.SetMaxOpenConns(maxConns)
database.SetMaxIdleConns(maxConns)
database.SetConnMaxLifetime(30 * time.Minute)
db.Exec("PRAGMA cache_size=-8000")   // 8MB page cache
db.Exec("PRAGMA busy_timeout=5000")  // 5s busy wait
db.Exec("PRAGMA temp_store=MEMORY")  // temp tables in memory
```

### 1.2 Eliminate Full Table Scans

| Problem | Location | Current | Fix |
|---------|----------|---------|-----|
| `LoadTokenSymbolMap` | `internal/db/db.go`, `blueprints/connect_broker.go` | Loads ALL `master_contracts` rows into memory on startup | Replace with lazy-loading + sync.Map cache. Load only tokens that are in active watchlists |
| `Duplicate queries` | Both `internal/db/` AND `blueprints/` define same `loadWatchlistSymbols` and `loadTokenSymbolMap` | Duplicated code, double maintenance | Remove from `blueprints/connect_broker.go`, keep only in `internal/db/` |
| `SearchMasterContract` | sqlc query | `WHERE symbol LIKE ? OR brsymbol LIKE ? OR name LIKE ?` — full table scan | Add `CREATE INDEX idx_mc_search ON master_contracts(symbol, brsymbol, name)` |
| `DELETE FROM master_contracts WHERE id NOT IN (...)` | `internal/db/db.go:initDB` | Runs on every startup, scans entire table | Remove from startup. Run once as migration. Replace with trigger-based dedup |

### 1.3 Lazy-Loaded Symbol Cache

Create `internal/db/cache.go`:
```go
type SymbolCache struct {
    mu     sync.RWMutex
    byID   map[string]string  // token -> symbol
    byKey  map[string]string  // exchange|token -> symbol
}

func NewSymbolCache() *SymbolCache

// Preload only active tokens from watchlist_items + strategies
func (c *SymbolCache) PreloadActive(ctx context.Context, q *gen.Queries, db *sql.DB) error

// Lookup with lazy-fallback to DB
func (c *SymbolCache) Get(ctx context.Context, token, exchange string) (string, error)

// Batch update from ticker data
func (c *SymbolCache) Warm(tokens []string, symbols map[string]string)
```

### 1.4 Optimize Broker Stream Hot Path

| Problem | Impact | Fix |
|---------|--------|-----|
| `parseTick` allocates `map[string]any` per tick | Heavy GC pressure (1000+ ticks/sec) | Use struct-based tick type and object pool (`sync.Pool`) |
| `BroadcastTick` locks `Hub.mu` for every tick | Contention with client (un)register | Split read/write locks. Use per-client send channels with buffered broadcast |
| `tokenSymbol` map lookup per tick | ~50ns overhead per symbol per tick | Pre-resolve symbol during subscribe and store in tick struct |
| `json.Marshal` per tick per client | CPU-heavy | Pre-encode to JSON once, broadcast raw bytes |
| Sequential broker reconnection | Full reconnection blocks market data | Immediate retry with jitter, not exponential backoff for first 3 retries |

### 1.5 `context.Background()` → Request Contexts

11 instances in `blueprints/watchlist.go` use `context.Background()` instead of `r.Context()`:
- `watchlist.go:26` — `ListWatchlists`
- `watchlist.go:46` — `CreateWatchlist`
- `watchlist.go:63` — `UpdateWatchlist`
- `watchlist.go:69` — `DeleteWatchlist`
- `watchlist.go:75` — `ListWatchlistItems`
- `watchlist.go:104` — `ListWatchlistItems`
- `watchlist.go:112` — `AddWatchlistItem`
- `watchlist.go:135` — `RemoveWatchlistItem`
- `watchlist.go:146` — `ReorderWatchlistItem`
- `watchlist.go:159` — `SearchMasterContract`

**Fix**: Replace every `context.Background()` with `r.Context()`. This enables:
- Request cancellation on client disconnect
- DB query timeouts via `r.Context().Deadline()`
- Tracing propagation

### 1.6 Add Indexes for Hot Queries

| Table | Query Pattern | Missing Index |
|-------|--------------|---------------|
| `master_contracts` | `WHERE symbol LIKE ?` | Composite `(symbol, brsymbol, name)` |
| `master_contracts` | `WHERE token = ?` | `(token)` |
| `watchlist_items` | `JOIN master_contracts ON token AND exchange` | `(token, exchange)` |
| `brokers` | `WHERE broker_name = ? AND token_status = ?` | `(broker_name, token_status)` |
| `orders` | `WHERE strategy_id = ?` | Already has `idx_orders_strategy` ✓ |
| `orders` | `WHERE order_id = ?` | Already has `idx_orders_order_id` ✓ |

---

## Phase 2: Security

### 2.1 Broker Credential Encryption (AES-256-GCM)

**Problem**: `broker_password`, `broker_pin`, `broker_qr_key`, `broker_api_secret` stored as plaintext in SQLite.

**Fix**: 
1. Add encryption key from env (`ENCRYPTION_KEY` — 32-byte hex)
2. Create `internal/crypto/aes.go`:
```go
func Encrypt(plaintext []byte, key []byte) ([]byte, error)  // returns nonce+ciphertext
func Decrypt(ciphertext []byte, key []byte) ([]byte, error)   // parses nonce+ciphertext
```
3. Encrypt on write in `handleCreateBroker` / `handleUpdateBroker`
4. Decrypt on read in `handleGetBroker` / broker auth flows
5. Add `CHECK(encryption_key IS NOT NULL)` to migration docs

### 2.2 Session Security Improvements

| Issue | Current | Fix |
|-------|---------|-----|
| Token entropy | `rand.Read` 32 bytes | Already adequate (256-bit) ✓ |
| No refresh rotation | Token valid for 24h with sliding window | Add `/api/refresh` endpoint that issues new token, invalidates old |
| No device tracking | Any token works from anywhere | Add client IP + User-Agent binding (store in session entry, verify on Get) |
| Cleanup goroutine leak | Goroutine runs forever even if SessionStore is GC'd | Add `Stop()` method, use `context.CancelFunc` |

### 2.3 Rate Limiting Expansion

| Endpoint | Risk | Fix |
|----------|------|-----|
| `POST /api/login` | Brute force | Already implemented (10/min/IP) ✓ |
| `POST /api/brokers` | Resource exhaustion | Add 30/min/IP |
| `POST /api/connect-broker` | Broker API abuse | Add 10/min/IP |
| `POST /api/broker-place-order` | Financial risk | Add 60/min/IP per broker |
| All `POST/PUT/DELETE` | General abuse | Add 120/min/IP globally |

**Fix**: Extract rate limiter to `internal/middleware/ratelimit.go` with configurable limits per route pattern. Use generics-based middleware chain.

### 2.4 CORS Hardening

**Current**: Falls back to echoing any origin.

**Fix**: 
```go
allowedOrigins := []string{
    "http://localhost:5173",        // Vite dev
    "http://localhost:8080",        // Prod same-origin
    "https://app.bhavyaai.com",     // Production
}
// DENY if origin not in list (remove fallback catch-all)
```

### 2.5 Input Validation

**Problem**: All handler request bodies are decoded directly without validation. Invalid data reaches sqlc queries.

**Fix**: Add `internal/validation/validator.go`:
```go
func ValidateBrokerCreate(req any) error  // check required fields, length limits
func ValidateStrategy(req any) error       // check numeric ranges, enum values
```
Or use `github.com/go-playground/validator/v10` with struct tags.

### 2.6 Additional Security Items

| Item | Description |
|------|-------------|
| TLS support | Add `--tls-cert` / `--tls-key` flags for HTTPS. Auto-redirect HTTP→HTTPS |
| Security headers | Add `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Strict-Transport-Security` |
| Password policy | Enforce min length 8, bcrypt cost ≥ 12 (currently default 10) |
| Token refresh URL | `POST /api/auth/refresh` with old+new token rotation |
| Audit log | Log all broker connections, orders placed, config changes |

---

## Phase 3: Modularity (Code Organization)

### 3.1 Make the Unused `Broker` Interface Actually Work

**Problem**: `brokers/interface.go` defines `Broker` interface but neither `angel.Client` nor `aliceblue.Client` implements it. The codebase uses `switch bc.brokerName` in 7+ places.

**Fix**: Implement the interface on both clients, then replace switch statements:
```go
// brokers/interface.go
type Broker interface {
    Authenticate(clientCode, brokerPin, totp string) (authToken, feedToken, string, error)
    GetMargin(authToken string) (*MarginData, error)
    PlaceOrder(authToken string, order *Order) (*OrderResponse, error)
    CancelOrder(authToken, orderID string) error
    ModifyOrder(authToken string, order *Order) error
    GetPositions(authToken string) ([]Position, error)
    GetOrderBook(authToken string) ([]Order, error)
    GetTradeBook(authToken string) ([]Trade, error)
    GetHoldings(authToken string) ([]Holding, error)
    GetHistoricalData(authToken, symbol, exchange, interval, from, to string) ([]Candle, error)
}
```

Then in `blueprints/broker_data.go`:
```go
// BEFORE: 7 switch statements
result, err := bc.angelClient.PlaceOrder(...)  // or bc.aliceClient.PlaceOrder(...)

// AFTER: One generic dispatch
result, err := bc.client.PlaceOrder(bc.authToken, orderData)
```

And `blueprints/broker_data.go:brokerClient()`:
```go
type brokerClientResult struct {
    client    brokers.Broker  // single interface
    authToken string
}
func (a *App) brokerClient(id int64) (*brokerClientResult, error) {
    // ... query DB ...
    switch brokerName {
    case "angel":
        result.client = angel.NewClient(apiKey)
    case "aliceblue":
        result.client = aliceblue.NewClient(apiKey, apiSecret)
    }
    return result, nil
}
```

### 3.2 Extract Watchlist Handlers from `blueprints/watchlist.go`

**Problem**: `blueprints/watchlist.go` is 209 lines with:
- 11 handler functions
- Multiple `context.Background()` bugs
- `sortSearchResults` helper that belongs elsewhere
- Inline JSON decoding instead of `decodeJSON`

**Fix**: Split into:
- `blueprints/watchlist/handlers.go` — HTTP handlers
- `blueprints/watchlist/service.go` — Business logic (sorting, validation)
- `blueprints/watchlist/routes.go` — Route registration

### 3.3 Create Service Layer Between HTTP Handlers and DB

**Problem**: Handlers call `a.Q.xxx(r.Context(), params)` directly — no service abstraction, no caching, no validation.

**Fix**: Create domain-based service packages:
```
internal/
  service/
    broker_service.go      # Broker CRUD + auth business logic
    strategy_service.go    # Strategy business logic
    watchlist_service.go   # Watchlist business logic
    market_data_service.go # Ticker data + symbol cache
```

Each service accepts interfaces (not concrete `*gen.Queries`) so they're testable:
```go
type BrokerRepository interface {
    GetBroker(ctx context.Context, id int64) (gen.Broker, error)
    ListBrokers(ctx context.Context) ([]gen.Broker, error)
    CreateBroker(ctx context.Context, params gen.CreateBrokerParams) (int64, error)
    UpdateBroker(ctx context.Context, params gen.UpdateBrokerParams) error
    DeleteBroker(ctx context.Context, id int64) error
}
```

### 3.4 Consolidate `cmd/subscribe/main.go` with `ws/broker.go`

**Problem**: `cmd/subscribe/main.go` duplicates:
- Angel One WebSocket connection logic
- Binary tick parsing 
- Exchange type mapping
- Subscribe/unsubscribe logic
- Watchlist token loading from DB

**Fix**: Turn `cmd/subscribe` into a thin CLI that imports `ws` package:
```go
func main() {
    db := openDB(dbPath)
    hub := ws.NewHub()
    client := ws.NewBrokerClient(clientCode, authToken, feedToken, apiKey, hub, nil)
    client.SetInitialSymbols(symbols)
    client.SetLogHandler(func(msg string) { fmt.Println(msg) })
    client.SetTickHandler(func(tick map[string]any) { printTick(tick) })
    client.Run()
}
```

### 3.5 Remove Duplicate Code

| Duplicate | Locations | Resolution |
|-----------|-----------|------------|
| `loadWatchlistSymbols` | `internal/db/db.go` + `blueprints/connect_broker.go` | Keep in `internal/db/`, remove from blueprints |
| `parseTick` binary parsing | `ws/broker.go` + `cmd/subscribe/main.go` | Export from `ws/broker.go`, import in cmd |
| `exchangeTypes` map | `ws/broker.go` + `cmd/subscribe/main.go` | Export from `ws/broker.go` |
| `groupTokens` | `ws/broker.go` + `cmd/subscribe/main.go` | Export from `ws/broker.go` |
| `writeJSON` | `internal/server/writejson.go` + `blueprints/blueprints.go` | Already consolidated via `httperr.WriteJSON` ✓ |
| `authMiddleware` | `internal/server/auth.go` + `blueprints/blueprints.go` | Different packages, acceptable for now |

### 3.6 Standardize Module Layout

```
backend/
├── cmd/
│   ├── server/main.go       # Server binary (current main.go)
│   └── subscribe/main.go    # CLI tool (refactored to use ws package)
├── internal/
│   ├── config/              # Configuration loading
│   ├── crypto/              # AES-256-GCM encryption
│   ├── db/                  # DB init, schema, symbol cache
│   ├── httperr/             # Error types + JSON helpers
│   ├── middleware/          # Auth, CORS, rate-limit, logging, recovery
│   ├── server/              # Server setup + route registration
│   ├── service/             # Session store + domain services
│   ├── validation/          # Input validation
│   └── setup/               # Seed data, master contract download
├── blueprints/              # Domain handlers (strategies, watchlists, etc.)
├── brokers/
│   ├── interface.go         # Broker interface
│   ├── aliceblue/           # Alice Blue implementation
│   ├── angel/               # Angel One implementation
│   └── types.go             # Shared types
├── db/
│   ├── migrations/          # SQL migration files (001_init.up.sql, etc.)
│   ├── query.sql            # sqlc source
│   └── gen/                 # sqlc generated code
└── ws/
    ├── hub.go               # WS hub
    ├── client.go            # Frontend client
    ├── broker.go            # Angel One stream
    ├── broker_alice.go      # Alice Blue stream
    └── priority.go          # Broker priority
```

### 3.7 Add Unit Tests

**Problem**: Zero test files across the entire project.

**Fix**: Add tests at three levels:

| Layer | Testing Strategy | Files |
|-------|-----------------|-------|
| Repository | Test sqlc queries against in-memory SQLite | `db/gen/*_test.go` |
| Service | Mock repository interface, test business logic | `internal/service/*_test.go` |
| Handler | Use `httptest.NewServer`, test HTTP + auth | `internal/server/*_test.go` |
| Broker | Mock WebSocket server, test parsing and reconnect | `ws/*_test.go` |

**Tools**: Use Go's built-in `testing` + `testify/assert` + `go-sqlite3` in-memory for DB tests.

---

## Implementation Roadmap

| Phase | Duration | Modules |
|-------|----------|---------|
| **P1-Speed** | 2-3 days | SQLite tuning, indexes, symbol cache, context fixes, tick hot-path, PRAGMA optimizations |
| **P2-Security** | 2-3 days | AES encryption, rate limiting expansion, CORS hardening, input validation, session refresh |
| **P3a-Modularity** | 3-4 days | Broker interface refactor, service layer creation, duplicate removal |
| **P3b-Modularity** | 2-3 days | cmd/subscribe consolidation, module layout, watchlist split |
| **P4-Testing** | 3-4 days | Repository tests, service tests, handler tests, broker tests |
| **P5-Polish** | 1-2 days | TLS, audit logging, security headers, migration tooling |

**Total estimated: 13-19 days**

---

## Quick Wins (Do First)

These offer highest impact with lowest risk:

1. **`context.Background()` → `r.Context()` in watchlist.go** — 11 lines changed, fixes request cancellation
2. **SQLite PRAGMA tuning** — 4 lines in `internal/db/db.go`, immediate perf gain
3. **Add missing indexes** (5 new indexes in `internal/db/schema.go`)
4. **Remove duplicate `loadWatchlistSymbols`** from `blueprints/connect_broker.go`
5. **Remove `DELETE FROM master_contracts` from startup** — one-time migration
6. **Fix `DATE('now','localtime')` → `DATE('now')`** in `internal/server/server.go` and `connect_broker.go`

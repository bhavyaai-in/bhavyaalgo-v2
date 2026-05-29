package scheduler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"bhavyaaialgo/backend/brokers/aliceblue"
	"bhavyaaialgo/backend/brokers/angel"
	"bhavyaaialgo/backend/historical"
	"database/sql"

	"github.com/robfig/cron/v3"
)

type Service struct {
	TradingDB *sql.DB
	MarketDB  *sql.DB
	cron      *cron.Cron
	entries   map[int64]cron.EntryID
	mu        sync.Mutex
	store     *historical.Store
}

func New(tradingDB, marketDB *sql.DB) *Service {
	s := &Service{
		TradingDB: tradingDB,
		MarketDB:  marketDB,
		cron:      cron.New(cron.WithLocation(time.UTC)),
		entries:   make(map[int64]cron.EntryID),
		store:     historical.GetStore(),
	}
	s.store.Init()
	return s
}

func (s *Service) InitDB() error {
	_, err := s.TradingDB.Exec(CreateTablesSQL)
	if err != nil {
		return err
	}
	_, err = s.TradingDB.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_scheduler_group_items_dedup ON scheduler_group_items(group_id, symbol, exchange)`)
	return err
}

func (s *Service) Start() {
	s.cron.Start()
	s.loadSchedules()
	log.Printf("scheduler: started, %d active schedules", len(s.entries))
}

func (s *Service) Stop() {
	s.cron.Stop()
	log.Printf("scheduler: stopped")
}

func (s *Service) loadSchedules() {
	rows, err := s.TradingDB.Query(`
		SELECT gs.group_id, gs.cron_expression, gs.broker_priority
		FROM scheduler_group_settings gs
		JOIN scheduler_groups g ON g.id = gs.group_id
		WHERE gs.is_active = 1
	`)
	if err != nil {
		log.Printf("scheduler: load error: %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var gid int64
		var cronExp, bp string
		if rows.Scan(&gid, &cronExp, &bp) == nil {
			s.addJob(gid, cronExp)
		}
	}
}

func (s *Service) addJob(gid int64, cronExp string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if eid, ok := s.entries[gid]; ok {
		s.cron.Remove(eid)
	}
	eid, err := s.cron.AddFunc(cronExp, func() { s.runGroup(gid) })
	if err != nil {
		log.Printf("scheduler: cron add error group %d: %v", gid, err)
		return
	}
	s.entries[gid] = eid
	log.Printf("scheduler: job for group %d [%s]", gid, cronExp)
}

func (s *Service) removeJob(gid int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if eid, ok := s.entries[gid]; ok {
		s.cron.Remove(eid)
		delete(s.entries, gid)
	}
}

func (s *Service) RefreshSchedule(gid int64) {
	var cronExp string
	err := s.TradingDB.QueryRow(
		`SELECT cron_expression FROM scheduler_group_settings WHERE group_id=? AND is_active=1`, gid,
	).Scan(&cronExp)
	if err != nil {
		s.removeJob(gid)
		return
	}
	s.addJob(gid, cronExp)
}

func (s *Service) RefreshAll() {
	s.mu.Lock()
	for _, eid := range s.entries {
		s.cron.Remove(eid)
	}
	s.entries = make(map[int64]cron.EntryID)
	s.mu.Unlock()
	s.loadSchedules()
}

func (s *Service) runGroup(gid int64) {
	log.Printf("scheduler: running group %d", gid)
	var logID int64
	err := s.TradingDB.QueryRow(
		`INSERT INTO scheduler_job_logs (group_id, status) VALUES (?, 'running') RETURNING id`,
		gid,
	).Scan(&logID)
	if err != nil {
		log.Printf("scheduler: log error: %v", err)
		return
	}
	var brokerPriority string
	s.TradingDB.QueryRow(`SELECT broker_priority FROM scheduler_group_settings WHERE group_id=?`, gid).Scan(&brokerPriority)

	rows, err := s.TradingDB.Query(
		`SELECT id, symbol, exchange, token, interval FROM scheduler_group_items WHERE group_id=? AND is_active=1`, gid,
	)
	if err != nil {
		s.finishLog(logID, "failed", err.Error())
		return
	}
	var items []item
	for rows.Next() {
		var it item
		if rows.Scan(&it.id, &it.symbol, &it.exchange, &it.token, &it.interval) == nil {
			items = append(items, it)
		}
	}
	rows.Close()
	if len(items) == 0 {
		s.finishLog(logID, "success", "no items")
		return
	}
	bc, err := s.getBrokerClient(brokerPriority)
	if err != nil {
		s.finishLog(logID, "failed", err.Error())
		return
	}
	total := len(items)
	success, failed := 0, 0
	errs := []string{}
	from := time.Now().AddDate(0, 0, -7).Format("2006-01-02 09:15")
	to := time.Now().Format("2006-01-02 15:30")
	for _, it := range items {
		if e := s.downloadItem(bc, it, from, to); e != nil {
			failed++
			errs = append(errs, fmt.Sprintf("%s: %v", it.symbol, e))
		} else {
			success++
		}
	}
	status := "success"
	msg := fmt.Sprintf("%d/%d ok", success, total)
	if failed > 0 {
		status = "partial"
		msg += " | err: " + strings.Join(errs, "; ")
	}
	s.TradingDB.Exec(`UPDATE scheduler_job_logs SET status=?, message=?, items_total=?, items_success=?, items_failed=? WHERE id=?`,
		status, msg, total, success, failed, logID)
	s.TradingDB.Exec(`UPDATE scheduler_group_settings SET last_run_status=? WHERE group_id=?`, status, gid)
	log.Printf("scheduler: group %d done: %s", gid, msg)
}

func (s *Service) finishLog(logID int64, status, msg string) {
	s.TradingDB.Exec(`UPDATE scheduler_job_logs SET status=?, message=? WHERE id=?`, status, msg, logID)
}

type item struct {
	id       int64
	symbol   string
	exchange string
	token    string
	interval string
}

type brokerClientResult struct {
	brokerName string
	authToken  string
	apiKey     string
	apiSecret  string
}

func (s *Service) getBrokerClient(priority string) (*brokerClientResult, error) {
	q := `SELECT broker_name, broker_token, broker_api, COALESCE(broker_api_secret,''), broker_token_date
		FROM brokers WHERE token_status='connected'`
	if priority != "" {
		q += ` AND broker_name='` + priority + `'`
	}
	q += ` ORDER BY broker_name LIMIT 1`
	var bn, bt, ba, bs, bd string
	if err := s.TradingDB.QueryRow(q).Scan(&bn, &bt, &ba, &bs, &bd); err != nil {
		return nil, fmt.Errorf("no connected broker%s", map[bool]string{true: " (" + priority + ")", false: ""}[priority != ""])
	}
	if !strings.HasPrefix(bd, time.Now().Format("2006-01-02")) {
		return nil, fmt.Errorf("token expired: %s", bn)
	}
	return &brokerClientResult{bn, bt, ba, bs}, nil
}

func (s *Service) downloadItem(bc *brokerClientResult, it item, from, to string) error {
	switch bc.brokerName {
	case "angel":
		ac := angel.NewClient(bc.apiKey)
		c, e := ac.GetHistoricalData(bc.authToken, it.token, it.exchange, it.interval, from, to)
		if e != nil {
			return e
		}
		cc := make([]historical.Candle, len(c))
		for i, v := range c {
			cc[i] = historical.Candle{v.Timestamp, v.Open, v.High, v.Low, v.Close, v.Volume}
		}
		return s.store.SaveCandles(it.symbol, it.exchange, it.token, it.interval, cc)
	case "aliceblue":
		ac := aliceblue.NewClient(bc.apiKey, bc.apiSecret)
		c, e := ac.GetHistoricalData(bc.authToken, it.token, it.interval, from, to, it.exchange)
		if e != nil {
			return e
		}
		cc := make([]historical.Candle, len(c))
		for i, v := range c {
			cc[i] = historical.Candle{v.Timestamp, v.Open, v.High, v.Low, v.Close, v.Volume}
		}
		return s.store.SaveCandles(it.symbol, it.exchange, it.token, it.interval, cc)
	default:
		return fmt.Errorf("unsupported broker: %s", bc.brokerName)
	}
}

// ─── API HANDLERS ──────────────────────────────────────────────

func (s *Service) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/scheduler/groups", s.handleListGroups)
	mux.HandleFunc("POST /api/scheduler/groups", s.handleCreateGroup)
	mux.HandleFunc("DELETE /api/scheduler/groups/{id}", s.handleDeleteGroup)
	mux.HandleFunc("GET /api/scheduler/groups/{id}/items", s.handleListItems)
	mux.HandleFunc("POST /api/scheduler/groups/{id}/items", s.handleAddItem)
	mux.HandleFunc("DELETE /api/scheduler/items/{id}", s.handleDeleteItem)
	mux.HandleFunc("GET /api/scheduler/groups/{id}/settings", s.handleGetSettings)
	mux.HandleFunc("PUT /api/scheduler/groups/{id}/settings", s.handleUpdateSettings)
	mux.HandleFunc("GET /api/scheduler/groups/{id}/logs", s.handleListLogs)
	mux.HandleFunc("POST /api/scheduler/groups/{id}/run", s.handleRunNow)
}

func (s *Service) handleListGroups(w http.ResponseWriter, r *http.Request) {
	rows, e := s.TradingDB.Query(`SELECT g.id, g.name, COALESCE(gs.cron_expression,''), COALESCE(gs.is_active,0), COALESCE(gs.broker_priority,''), COALESCE(gs.last_run_status,''), (SELECT COUNT(*) FROM scheduler_group_items WHERE group_id=g.id) FROM scheduler_groups g LEFT JOIN scheduler_group_settings gs ON gs.group_id=g.id ORDER BY g.id`)
	if e != nil { writeJSON(w, 500, map[string]string{"error": e.Error()}); return }
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id int64; var name, ce, bp, lr string; var ia, ic int
		if rows.Scan(&id, &name, &ce, &ia, &bp, &lr, &ic) == nil {
			out = append(out, map[string]any{"id": id, "name": name, "cron": ce, "is_active": ia, "broker_priority": bp, "last_run": lr, "item_count": ic})
		}
	}
	if out == nil { out = []map[string]any{} }
	writeJSON(w, 200, out)
}

func (s *Service) handleCreateGroup(w http.ResponseWriter, r *http.Request) {
	var req struct{ Name string }
	if json.NewDecoder(r.Body).Decode(&req) != nil || req.Name == "" {
		http.Error(w, "name required", 400); return
	}
	var id int64
	s.TradingDB.QueryRow(`INSERT INTO scheduler_groups (name) VALUES (?) RETURNING id`, req.Name).Scan(&id)
	s.TradingDB.Exec(`INSERT INTO scheduler_group_settings (group_id) VALUES (?)`, id)
	writeJSON(w, 200, map[string]any{"id": id, "name": req.Name})
}

func (s *Service) handleDeleteGroup(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	s.removeJob(id)
	s.TradingDB.Exec(`DELETE FROM scheduler_groups WHERE id=?`, id)
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

func (s *Service) handleListItems(w http.ResponseWriter, r *http.Request) {
	gid := parseID(r)
	rows, e := s.TradingDB.Query(`SELECT id, symbol, exchange, token, interval, is_active FROM scheduler_group_items WHERE group_id=? ORDER BY id`, gid)
	if e != nil { writeJSON(w, 500, map[string]string{"error": e.Error()}); return }
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id int64; var sym, ex, tok, iv string; var act int
		if rows.Scan(&id, &sym, &ex, &tok, &iv, &act) == nil {
			out = append(out, map[string]any{"id": id, "symbol": sym, "exchange": ex, "token": tok, "interval": iv, "is_active": act})
		}
	}
	if out == nil { out = []map[string]any{} }
	writeJSON(w, 200, out)
}

func (s *Service) handleAddItem(w http.ResponseWriter, r *http.Request) {
	gid := parseID(r)
	var req struct{ Symbol, Exchange, Token, Interval string }
	if json.NewDecoder(r.Body).Decode(&req) != nil || req.Symbol == "" {
		http.Error(w, "symbol required", 400); return
	}
	if req.Interval == "" { req.Interval = "1d" }

	var existing int
	s.TradingDB.QueryRow(
		`SELECT COUNT(*) FROM scheduler_group_items WHERE group_id=? AND symbol=? AND exchange=?`,
		gid, req.Symbol, req.Exchange,
	).Scan(&existing)
	if existing > 0 {
		writeJSON(w, 409, map[string]string{"error": "symbol already exists in group"})
		return
	}

	if req.Token == "" {
		s.MarketDB.QueryRow(`SELECT token FROM master_contracts WHERE symbol=? AND exchange=? LIMIT 1`, req.Symbol, req.Exchange).Scan(&req.Token)
	}
	var id int64
	s.TradingDB.QueryRow(`INSERT INTO scheduler_group_items (group_id, symbol, exchange, token, interval) VALUES (?,?,?,?,?) RETURNING id`, gid, req.Symbol, req.Exchange, req.Token, req.Interval).Scan(&id)
	writeJSON(w, 200, map[string]any{"id": id, "symbol": req.Symbol})
}

func (s *Service) handleDeleteItem(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	s.TradingDB.Exec(`DELETE FROM scheduler_group_items WHERE id=?`, id)
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

func (s *Service) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	gid := parseID(r)
	var ce, bp, lr string; var ia int
	e := s.TradingDB.QueryRow(`SELECT cron_expression, is_active, broker_priority, last_run_status FROM scheduler_group_settings WHERE group_id=?`, gid).Scan(&ce, &ia, &bp, &lr)
	if e != nil { writeJSON(w, 404, map[string]string{"error": "not found"}); return }
	writeJSON(w, 200, map[string]any{"cron": ce, "is_active": ia, "broker_priority": bp, "last_run": lr})
}

func (s *Service) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	gid := parseID(r)
	var req struct{ Cron string; IsActive *int; BrokerPriority string }
	if json.NewDecoder(r.Body).Decode(&req) != nil { http.Error(w, "bad request", 400); return }
	if req.Cron != "" {
		s.TradingDB.Exec(`UPDATE scheduler_group_settings SET cron_expression=? WHERE group_id=?`, req.Cron, gid)
	}
	if req.IsActive != nil {
		s.TradingDB.Exec(`UPDATE scheduler_group_settings SET is_active=? WHERE group_id=?`, *req.IsActive, gid)
	}
	s.TradingDB.Exec(`UPDATE scheduler_group_settings SET broker_priority=? WHERE group_id=?`, req.BrokerPriority, gid)
	s.RefreshSchedule(gid)
	writeJSON(w, 200, map[string]string{"status": "updated"})
}

func (s *Service) handleListLogs(w http.ResponseWriter, r *http.Request) {
	gid := parseID(r)
	rows, e := s.TradingDB.Query(`SELECT id, run_time, status, message, items_total, items_success, items_failed FROM scheduler_job_logs WHERE group_id=? ORDER BY run_time DESC LIMIT 20`, gid)
	if e != nil { writeJSON(w, 500, map[string]string{"error": e.Error()}); return }
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id int64; var rt, st, msg string; var tot, suc, fail int
		if rows.Scan(&id, &rt, &st, &msg, &tot, &suc, &fail) == nil {
			out = append(out, map[string]any{"id": id, "time": rt, "status": st, "message": msg, "total": tot, "success": suc, "failed": fail})
		}
	}
	if out == nil { out = []map[string]any{} }
	writeJSON(w, 200, out)
}

func (s *Service) handleRunNow(w http.ResponseWriter, r *http.Request) {
	gid := parseID(r)
	go s.runGroup(gid)
	writeJSON(w, 200, map[string]string{"status": "triggered"})
}

func parseID(r *http.Request) int64 {
	v := r.PathValue("id")
	id, _ := strconv.ParseInt(v, 10, 64)
	return id
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

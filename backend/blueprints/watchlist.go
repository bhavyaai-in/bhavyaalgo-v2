package blueprints

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"

	marketdb "bhavyaaialgo/backend/db/market/gen"
)

func (a *App) RegisterWatchlistRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/watchlists", a.authMiddleware(a.handleListWatchlists))
	mux.HandleFunc("POST /api/watchlists", a.authMiddleware(a.handleCreateWatchlist))
	mux.HandleFunc("PUT /api/watchlists/{id}", a.authMiddleware(a.handleUpdateWatchlist))
	mux.HandleFunc("DELETE /api/watchlists/{id}", a.authMiddleware(a.handleDeleteWatchlist))
	mux.HandleFunc("GET /api/watchlists/{id}/items", a.authMiddleware(a.handleListItems))
	mux.HandleFunc("POST /api/watchlists/{id}/items", a.authMiddleware(a.handleAddItem))
	mux.HandleFunc("DELETE /api/watchlist-items/{id}", a.authMiddleware(a.handleRemoveItem))
	mux.HandleFunc("PUT /api/watchlist-items/{id}/reorder", a.authMiddleware(a.handleReorderItem))
	mux.HandleFunc("GET /api/search-contracts", a.authMiddleware(a.handleSearchContracts))
}

func (a *App) handleListWatchlists(w http.ResponseWriter, r *http.Request) {
	list, err := a.MarketQ.ListWatchlists(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if list == nil {
		list = []marketdb.Watchlist{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (a *App) handleCreateWatchlist(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name      string `json:"name"`
		SortOrder int64  `json:"sort_order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	id, err := a.MarketQ.CreateWatchlist(r.Context(), marketdb.CreateWatchlistParams{
		Name: req.Name, SortOrder: req.SortOrder,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func (a *App) handleUpdateWatchlist(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var req struct{ Name string `json:"name"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	a.MarketQ.UpdateWatchlist(r.Context(), marketdb.UpdateWatchlistParams{Name: req.Name, ID: id})
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (a *App) handleDeleteWatchlist(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	a.MarketQ.DeleteWatchlist(r.Context(), id)
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (a *App) handleListItems(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	items, err := a.MarketQ.ListWatchlistItems(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if items == nil {
		items = []marketdb.WatchlistItem{}
	}
	writeJSON(w, http.StatusOK, items)
}

func (a *App) handleAddItem(w http.ResponseWriter, r *http.Request) {
	watchlistID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var req struct {
		Symbol         string  `json:"symbol"`
		Brsymbol       string  `json:"brsymbol"`
		Name           string  `json:"name"`
		Exchange       string  `json:"exchange"`
		Token          string  `json:"token"`
		Expiry         string  `json:"expiry"`
		Strike         float64 `json:"strike"`
		Lotsize        int64   `json:"lotsize"`
		Instrumenttype string  `json:"instrumenttype"`
		TickSize       float64 `json:"tick_size"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	existing, _ := a.MarketQ.ListWatchlistItems(r.Context(), watchlistID)
	for _, item := range existing {
		if item.Token == req.Token && item.Exchange == req.Exchange {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "symbol already exists in watchlist"})
			return
		}
	}
	nextOrder := int64(len(existing))
	id, err := a.MarketQ.AddWatchlistItem(r.Context(), marketdb.AddWatchlistItemParams{
		WatchlistID:    watchlistID,
		Symbol:         req.Symbol,
		Brsymbol:       req.Brsymbol,
		Name:           req.Name,
		Exchange:       req.Exchange,
		Token:          req.Token,
		Expiry:         req.Expiry,
		Strike:         req.Strike,
		Lotsize:        req.Lotsize,
		Instrumenttype: req.Instrumenttype,
		TickSize:       req.TickSize,
		SortOrder:      nextOrder,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func (a *App) handleRemoveItem(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	a.MarketQ.RemoveWatchlistItem(r.Context(), id)
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (a *App) handleReorderItem(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var req struct{ SortOrder int64 `json:"sort_order"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	a.MarketQ.ReorderWatchlistItem(r.Context(), marketdb.ReorderWatchlistItemParams{
		SortOrder: req.SortOrder, ID: id,
	})
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *App) handleSearchContracts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	pattern := "%" + q + "%"
	rows, err := a.MarketDB.QueryContext(r.Context(), `
		SELECT * FROM master_contracts
		WHERE symbol LIKE ? OR brsymbol LIKE ? OR name LIKE ?
		LIMIT 200
	`, pattern, pattern, pattern)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()
	var results []marketdb.MasterContract
	for rows.Next() {
		var r marketdb.MasterContract
		if err := rows.Scan(&r.ID, &r.Symbol, &r.Brsymbol, &r.Name, &r.Exchange, &r.Brexchange,
			&r.Token, &r.Expiry, &r.Strike, &r.Lotsize, &r.Instrumenttype, &r.TickSize, &r.BrokerName, &r.CreatedAt); err != nil {
			continue
		}
		results = append(results, r)
	}
	if results == nil {
		results = []marketdb.MasterContract{}
	}
	// Deduplicate by (symbol, exchange, instrumenttype) so each contract shows once
	seen := make(map[string]bool)
	var deduped []marketdb.MasterContract
	for _, r := range results {
		key := r.Symbol + "|" + r.Exchange + "|" + r.Instrumenttype
		if seen[key] {
			continue
		}
		seen[key] = true
		deduped = append(deduped, r)
	}
	// Custom sort: prefix matches first, then by exchange→inst→symbol
	upperQ := strings.ToUpper(q)
	sort.SliceStable(deduped, func(i, j int) bool {
		si, sj := strings.ToUpper(deduped[i].Symbol), strings.ToUpper(deduped[j].Symbol)
		pi := strings.HasPrefix(si, upperQ)
		pj := strings.HasPrefix(sj, upperQ)
		if pi != pj {
			return pi
		}
		ei := exchangeOrder[deduped[i].Exchange]
		if ei == 0 { ei = 99 }
		ej := exchangeOrder[deduped[j].Exchange]
		if ej == 0 { ej = 99 }
		if ei != ej { return ei < ej }
		ii := instOrder[deduped[i].Instrumenttype]
		if ii == 0 { ii = 99 }
		ij := instOrder[deduped[j].Instrumenttype]
		if ij == 0 { ij = 99 }
		if ii != ij { return ii < ij }
		return si < sj
	})
	if len(deduped) > 50 {
		deduped = deduped[:50]
	}
	writeJSON(w, http.StatusOK, deduped)
}

var exchangeOrder = map[string]int{
	"NSE": 1, "BSE": 2,
	"NFO": 3, "BFO": 4, "MCX": 5, "CDS": 6, "BCD": 7,
}

var instOrder = map[string]int{
	"EQ": 1, "": 1, "0": 1,
	"FUT": 2, "FUTSTK": 2, "FUTIDX": 2, "FUTCOM": 2, "FUTCUR": 2, "FUTENR": 2,
	"FUTBLN": 2, "FUTBAS": 2, "FUTIRC": 2, "FUTIRT": 2,
	"OPT": 3, "OPTSTK": 3, "OPTIDX": 3, "OPTCUR": 3, "OPTFUT": 3, "OPTBLN": 3, "OPTIRC": 3,
	"CE": 3, "PE": 3,
	"1": 2, "2": 2, "3": 3, "4": 3, "D": 3, "E": 3,
	"INDEX": 1, "IF": 2, "IO": 3, "SF": 2, "SO": 3,
	"UNDCUR": 1, "UNDIRC": 1, "UNDIRD": 1, "UNDIRT": 1,
	"AMXIDX": 1, "COMDTY": 1, "ETF": 1,
}

func sortSearchResults(results []marketdb.MasterContract) {
	sort.SliceStable(results, func(i, j int) bool {
		ei := exchangeOrder[results[i].Exchange]
		if ei == 0 {
			ei = 99
		}
		ej := exchangeOrder[results[j].Exchange]
		if ej == 0 {
			ej = 99
		}
		if ei != ej {
			return ei < ej
		}
		ii := instOrder[results[i].Instrumenttype]
		if ii == 0 {
			ii = 99
		}
		ij := instOrder[results[j].Instrumenttype]
		if ij == 0 {
			ij = 99
		}
		if ii != ij {
			return ii < ij
		}
		return results[i].Symbol < results[j].Symbol
	})
}

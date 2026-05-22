package blueprints

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"

	"bhavyaaialgo/backend/db/gen"
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
	list, err := a.Q.ListWatchlists(context.Background())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if list == nil {
		list = []gen.Watchlist{}
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
	id, err := a.Q.CreateWatchlist(context.Background(), gen.CreateWatchlistParams{
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
	a.Q.UpdateWatchlist(context.Background(), gen.UpdateWatchlistParams{Name: req.Name, ID: id})
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (a *App) handleDeleteWatchlist(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	a.Q.DeleteWatchlist(context.Background(), id)
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (a *App) handleListItems(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	items, err := a.Q.ListWatchlistItems(context.Background(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if items == nil {
		items = []gen.WatchlistItem{}
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
	existing, _ := a.Q.ListWatchlistItems(context.Background(), watchlistID)
	for _, item := range existing {
		if item.Token == req.Token && item.Exchange == req.Exchange {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "symbol already exists in watchlist"})
			return
		}
	}
	nextOrder := int64(len(existing))
	id, err := a.Q.AddWatchlistItem(context.Background(), gen.AddWatchlistItemParams{
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
	a.Q.RemoveWatchlistItem(context.Background(), id)
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (a *App) handleReorderItem(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var req struct{ SortOrder int64 `json:"sort_order"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	a.Q.ReorderWatchlistItem(context.Background(), gen.ReorderWatchlistItemParams{
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
	results, err := a.Q.SearchMasterContract(context.Background(), gen.SearchMasterContractParams{
		Symbol: pattern, Brsymbol: pattern, Name: pattern,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if results == nil {
		results = []gen.MasterContract{}
	}
	sortSearchResults(results)
	writeJSON(w, http.StatusOK, results)
}

var exchangeOrder = map[string]int{
	"NSE": 1, "BSE": 2,
	"NFO": 3, "BFO": 4, "MCX": 5, "CDS": 6, "BCD": 7,
}

var instOrder = map[string]int{
	"EQ": 1, "FUT": 2, "FUTSTK": 2, "OPT": 3, "OPTSTK": 3,
	"CE": 3, "PE": 3,
}

func sortSearchResults(results []gen.MasterContract) {
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

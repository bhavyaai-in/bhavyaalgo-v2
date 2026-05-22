package blueprints

import (
	"net/http"

	"bhavyaaialgo/backend/db/gen"
)

func (a *App) RegisterSettingsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/settings", a.authMiddleware(a.handleListSettings))
	mux.HandleFunc("POST /api/settings", a.authMiddleware(a.handleUpsertSetting))
}

type settingRow struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (a *App) handleListSettings(w http.ResponseWriter, r *http.Request) {
	list, err := a.Q.ListSettings(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if list == nil {
		list = []gen.ListSettingsRow{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (a *App) handleUpsertSetting(w http.ResponseWriter, r *http.Request) {
	var in settingRow
	if err := decodeJSON(r, &in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if in.Key == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "key is required"})
		return
	}
	if err := a.Q.UpsertSetting(r.Context(), gen.UpsertSettingParams{
		Key:   in.Key,
		Value: in.Value,
	}); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

package setup

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"bhavyaaialgo/backend/db/gen"
)

const (
	masterContractURL = "https://margincalculator.angelbroking.com/OpenAPI_File/files/OpenAPIScripMaster.json"
	angelSettingKey   = "master_contract_angel"
)

// DownloadMasterContract downloads & processes the Angel One master contract JSON.
// It checks last success time — if < 24h ago, skips.
// On failure, logs error and uses existing data.
func DownloadMasterContract(ctx context.Context, Q *gen.Queries) {
	// Check when we last succeeded
	lastVal, _ := Q.GetSetting(ctx, angelSettingKey)
	if lastVal != "" {
		if t, err := time.Parse(time.RFC3339, lastVal); err == nil {
			if time.Since(t) < 24*time.Hour {
				log.Printf("master contract: last download %s (<24h), skipping", t.Format("2006-01-02"))
				return
			}
		}
	}

	log.Print("master contract: downloading...")
	resp, err := http.Get(masterContractURL)
	if err != nil {
		log.Printf("master contract download failed: %v", err)
		MarkDownloadAttempt(ctx, Q, err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("master contract read failed: %v", err)
		MarkDownloadAttempt(ctx, Q, err)
		return
	}

	var rows []map[string]any
	if err := json.Unmarshal(body, &rows); err != nil {
		log.Printf("master contract parse failed: %v", err)
		MarkDownloadAttempt(ctx, Q, err)
		return
	}

	// Clear old data and insert fresh
	Q.ClearMasterContracts(ctx)
	inserted := 0
	for _, raw := range rows {
		row := processAngelRow(raw)
		if err := Q.BulkInsertMasterContract(ctx, row); err != nil {
			// skip duplicates / bad rows
			continue
		}
		inserted++
	}

	// Mark success
	now := time.Now().Format(time.RFC3339)
	Q.UpsertSetting(ctx, gen.UpsertSettingParams{
		Key: angelSettingKey, Value: now,
	})
	log.Printf("master contract: inserted %d records", inserted)
}

// MarkDownloadAttempt records a failed attempt timestamp.
func MarkDownloadAttempt(ctx context.Context, Q *gen.Queries, err error) {
	Q.UpsertSetting(ctx, gen.UpsertSettingParams{
		Key: angelSettingKey + "_last_try", Value: time.Now().Format(time.RFC3339),
	})
	Q.UpsertSetting(ctx, gen.UpsertSettingParams{
		Key: angelSettingKey + "_last_error", Value: err.Error(),
	})
}

// processAngelRow normalises a raw Angel One JSON row into insert params.
func processAngelRow(raw map[string]any) gen.BulkInsertMasterContractParams {
	get := func(k string) string {
		if v, ok := raw[k]; ok && v != nil {
			return fmt.Sprintf("%v", v)
		}
		return ""
	}
	getF := func(k string) float64 {
		if v, ok := raw[k]; ok && v != nil {
			switch x := v.(type) {
			case float64:
				return x
			case string:
				var f float64
				fmt.Sscanf(x, "%f", &f)
				return f
			}
		}
		return 0
	}
	getI := func(k string) int64 {
		if v, ok := raw[k]; ok && v != nil {
			switch x := v.(type) {
			case float64:
				return int64(x)
			case int64:
				return x
			case string:
				var i int64
				fmt.Sscanf(x, "%d", &i)
				return i
			}
		}
		return 0
	}

	symbol := get("symbol")
	symbol = strings.ReplaceAll(symbol, "-EQ", "")
	symbol = strings.ReplaceAll(symbol, "-BE", "")
	symbol = strings.ReplaceAll(symbol, "-MF", "")
	symbol = strings.ReplaceAll(symbol, "-SG", "")

	exchange := get("exch_seg")
	instType := get("instrumenttype")
	// Map NSE/BSE index types
	if instType == "AMXIDX" {
		switch exchange {
		case "NSE":
			exchange = "NSE_INDEX"
		case "BSE":
			exchange = "BSE_INDEX"
		case "MCX":
			exchange = "MCX_INDEX"
		}
	}

	return gen.BulkInsertMasterContractParams{
		Symbol:          symbol,
		Brsymbol:        symbol,
		Name:            get("name"),
		Exchange:        exchange,
		Brexchange:      exchange,
		Token:           get("token"),
		Expiry:          convertDate(get("expiry")),
		Strike:          getF("strike") / 100,
		Lotsize:         getI("lotsize"),
		Instrumenttype:  instType,
		TickSize:        getF("tick_size") / 100,
	}
}

func convertDate(s string) string {
	if s == "" {
		return ""
	}
	t, err := time.Parse("02Jan2006", strings.ToUpper(s))
	if err != nil {
		return s
	}
	return strings.ToUpper(t.Format("02-Jan-06"))
}

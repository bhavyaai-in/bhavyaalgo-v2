package setup

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bhavyaaialgo/backend/db/gen"
)

const (
	masterContractURL        = "https://margincalculator.angelbroking.com/OpenAPI_File/files/OpenAPIScripMaster.json"
	aliceContractURLTemplate = "https://v2api.aliceblueonline.com/restpy/static/contract_master/V2/%s.csv"
	MasterContractSettingKey = "master_contract_angel"
	angelSettingKey          = MasterContractSettingKey
	aliceSettingKey          = "master_contract_alice"
)

var aliceExchanges = []string{"NSE", "NFO", "CDS", "BSE", "BFO", "BCD", "MCX", "INDICES"}

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

func DownloadAliceContractMaster(ctx context.Context, Q *gen.Queries) {
	lastVal, _ := Q.GetSetting(ctx, aliceSettingKey)
	if lastVal != "" {
		if t, err := time.Parse(time.RFC3339, lastVal); err == nil {
			if time.Since(t) < 24*time.Hour {
				log.Printf("alice master contract: last download %s (<24h), skipping", t.Format("2006-01-02"))
				return
			}
		}
	}

	log.Print("alice master contract: downloading...")
	Q.ClearMasterContracts(ctx)
	totalInserted := 0

	for _, exch := range aliceExchanges {
		url := fmt.Sprintf(aliceContractURLTemplate, exch)
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("alice master contract: download %s failed: %v", exch, err)
			continue
		}
		if resp.StatusCode != 200 {
			resp.Body.Close()
			log.Printf("alice master contract: %s returned status %d", exch, resp.StatusCode)
			continue
		}

		reader := csv.NewReader(resp.Body)
		headers, err := reader.Read()
		if err != nil {
			resp.Body.Close()
			log.Printf("alice master contract: %s read headers failed: %v", exch, err)
			continue
		}

		hdr := make(map[string]int)
		for i, h := range headers {
			hdr[strings.TrimSpace(h)] = i
		}

		records, err := reader.ReadAll()
		resp.Body.Close()
		if err != nil {
			log.Printf("alice master contract: %s read failed: %v", exch, err)
			continue
		}

		inserted := 0
		for _, rec := range records {
			rows := processAliceRow(rec, hdr, exch)
			if rows == nil {
				continue
			}
			for _, row := range rows {
				if row == nil {
					continue
				}
				if err := Q.BulkInsertMasterContract(ctx, *row); err != nil {
					continue
				}
				inserted++
			}
		}
		log.Printf("alice master contract: %s = %d rows", exch, inserted)
		totalInserted += inserted
	}

	now := time.Now().Format(time.RFC3339)
	Q.UpsertSetting(ctx, gen.UpsertSettingParams{
		Key: aliceSettingKey, Value: now,
	})
	log.Printf("alice master contract: total %d records", totalInserted)
}

func processAliceRow(rec []string, hdr map[string]int, exchangeFile string) []*gen.BulkInsertMasterContractParams {
	ci := func(name string) int {
		if idx, ok := hdr[name]; ok && idx < len(rec) {
			return idx
		}
		return -1
	}
	gs := func(name string) string {
		if idx := ci(name); idx >= 0 {
			return strings.TrimSpace(rec[idx])
		}
		return ""
	}
	gi := func(name string) int64 {
		s := gs(name)
		if s == "" {
			return 0
		}
		v, _ := strconv.ParseInt(s, 10, 64)
		return v
	}
	gf := func(name string) float64 {
		s := gs(name)
		if s == "" {
			return 0
		}
		v, _ := strconv.ParseFloat(s, 64)
		return v
	}

	symbol := gs("Symbol")
	token := gs("Token")
	exch := gs("Exch")
	if symbol == "" || token == "" {
		return nil
	}

	symbolClean := symbol
	symbolClean = strings.TrimSuffix(symbolClean, "-EQ")
	symbolClean = strings.TrimSuffix(symbolClean, "-BE")
	symbolClean = strings.TrimSuffix(symbolClean, "-MF")
	symbolClean = strings.TrimSuffix(symbolClean, "-SG")

	instType := gs("Instrument Type")

	row := &gen.BulkInsertMasterContractParams{
		Symbol:         symbolClean,
		Brsymbol:       symbol,
		Name:           gs("Instrument Name"),
		Exchange:       exch,
		Brexchange:     exch,
		Token:          token,
		Expiry:         "",
		Strike:         0,
		Lotsize:        gi("Lot Size"),
		Instrumenttype: instType,
		TickSize:       gf("Tick Size"),
	}

	rows := []*gen.BulkInsertMasterContractParams{row}

	if exchangeFile == "INDICES" && exch != "" {
		idxRow := &gen.BulkInsertMasterContractParams{
			Symbol:         symbolClean,
			Brsymbol:       symbol,
			Name:           gs("Instrument Name"),
			Exchange:       exch + "_INDEX",
			Brexchange:     exch,
			Token:          "999" + token,
			Expiry:         "",
			Strike:         0,
			Lotsize:        gi("Lot Size"),
			Instrumenttype: instType,
			TickSize:       gf("Tick Size"),
		}
		rows = append(rows, idxRow)
	}

	return rows
}

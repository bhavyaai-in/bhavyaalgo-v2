package setup

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"bhavyaaialgo/backend/db/gen"
)

func SeedFromFile(ctx context.Context, Q *gen.Queries) {
	count, err := Q.GetBrokerListCount(ctx)
	if err != nil || count > 0 {
		return
	}
	data, err := os.ReadFile("seed.json")
	if err != nil {
		log.Printf("seed.json not found: %v", err)
		return
	}
	var seed struct {
		BrokerList []struct {
			Name     string `json:"name"`
			ImageURL string `json:"broker_image_url"`
			IsActive bool   `json:"is_active"`
		} `json:"broker_list"`
		Brokers []struct {
			FriendlyName    string `json:"friendly_name"`
			BrokerUserid    string `json:"broker_userid"`
			BrokerPassword  string `json:"broker_password"`
			BrokerPin       string `json:"broker_pin"`
			BrokerQrKey     string `json:"broker_qr_key"`
			BrokerAPI       string `json:"broker_api"`
			BrokerAPISecret string `json:"broker_api_secret"`
			BrokerName      string `json:"broker_name"`
			IsActive        bool   `json:"is_active"`
			IsAutologin     bool   `json:"is_autologin"`
		} `json:"brokers"`
		BrokerColumns []struct {
			BrokerName string   `json:"broker_name"`
			Columns    []string `json:"columns"`
		} `json:"broker_columns"`
	}
	if err := json.Unmarshal(data, &seed); err != nil {
		log.Printf("parse seed.json: %v", err)
		return
	}
	if len(seed.BrokerList) == 0 {
		log.Print("seed.json: broker_list is empty, skipping")
		return
	}
	for _, e := range seed.BrokerList {
		active := int64(0)
		if e.IsActive {
			active = 1
		}
		Q.InsertBrokerListEntry(ctx, gen.InsertBrokerListEntryParams{
			Name: e.Name, BrokerImageUrl: e.ImageURL, IsActive: active,
		})
	}
	for _, b := range seed.Brokers {
		active := int64(0)
		if b.IsActive {
			active = 1
		}
		autologin := int64(0)
		if b.IsAutologin {
			autologin = 1
		}
		Q.CreateBroker(ctx, gen.CreateBrokerParams{
			FriendlyName:    b.FriendlyName,
			BrokerUserid:    b.BrokerUserid,
			BrokerPassword:  b.BrokerPassword,
			BrokerPin:       b.BrokerPin,
			BrokerQrKey:     b.BrokerQrKey,
			BrokerApi:       b.BrokerAPI,
			BrokerApiSecret: b.BrokerAPISecret,
			BrokerName:      b.BrokerName,
			IsActive:        active,
			IsAutologin:     autologin,
			TokenStatus:     "",
			BrokerToken:     "",
			BrokerTokenDate: "2000-01-01 00:00:00",
			FeedToken:       "",
			IsDisabled:      0,
			Message:         "",
		})
	}
	for _, c := range seed.BrokerColumns {
		colsJSON, _ := json.Marshal(c.Columns)
		Q.UpsertBrokerColumn(ctx, gen.UpsertBrokerColumnParams{
			BrokerName: c.BrokerName, ColumnsJson: string(colsJSON),
		})
	}
	log.Printf("seeded %d broker_list, %d brokers, %d column configs", len(seed.BrokerList), len(seed.Brokers), len(seed.BrokerColumns))
}

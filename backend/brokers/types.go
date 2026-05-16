package brokers

type Order struct {
	ID              string  `json:"id"`
	Symbol          string  `json:"symbol"`
	Exchange        string  `json:"exchange"`
	TransactionType string  `json:"transaction_type"`
	Quantity        int     `json:"quantity"`
	Price           float64 `json:"price"`
	TriggerPrice    float64 `json:"trigger_price"`
	OrderType       string  `json:"order_type"`
	ProductType     string  `json:"product_type"`
	Validity        string  `json:"validity"`
	Variety         string  `json:"variety"`
	Status          string  `json:"status"`
	Message         string  `json:"message"`
}

type Position struct {
	Symbol      string  `json:"symbol"`
	Exchange    string  `json:"exchange"`
	Quantity    int     `json:"quantity"`
	BuyAvg      float64 `json:"buy_avg"`
	SellAvg     float64 `json:"sell_avg"`
	BuyQty      int     `json:"buy_qty"`
	SellQty     int     `json:"sell_qty"`
	NetQty      int     `json:"net_qty"`
	PnL         float64 `json:"pnl"`
	ProductType string  `json:"product_type"`
}

type Trade struct {
	ID              string  `json:"id"`
	OrderID         string  `json:"order_id"`
	Symbol          string  `json:"symbol"`
	Exchange        string  `json:"exchange"`
	TransactionType string  `json:"transaction_type"`
	Quantity        int     `json:"quantity"`
	Price           float64 `json:"price"`
	TradeValue      float64 `json:"trade_value"`
}

type Holding struct {
	Symbol   string  `json:"symbol"`
	Exchange string  `json:"exchange"`
	Quantity int     `json:"quantity"`
	PnL      float64 `json:"pnl"`
}

type Candle struct {
	Timestamp string  `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    int     `json:"volume"`
}

type MarginData struct {
	AvailableCash  string `json:"available_cash"`
	Collateral     string `json:"collateral"`
	M2MRealized    string `json:"m2m_realized"`
	M2MUnrealized  string `json:"m2m_unrealized"`
	UtilisedDebits string `json:"utilised_debits"`
}

type OrderResponse struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

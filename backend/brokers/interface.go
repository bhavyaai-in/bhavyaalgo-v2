package brokers

type Broker interface {
	Authenticate(clientCode, brokerPin, totp string) (string, string, string, error)
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

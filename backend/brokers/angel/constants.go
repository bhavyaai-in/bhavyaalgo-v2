package angel

type Variety string
type TransactionType string
type OrderType string
type ProductType string
type Duration string
type Exchange string

const (
	VarietyNormal   Variety = "NORMAL"
	VarietyStoploss Variety = "STOPLOSS"
	VarietyRobo     Variety = "ROBO"

	TransactionTypeBuy  TransactionType = "BUY"
	TransactionTypeSell TransactionType = "SELL"

	OrderTypeMarket         OrderType = "MARKET"
	OrderTypeLimit          OrderType = "LIMIT"
	OrderTypeStoplossLimit  OrderType = "STOPLOSS_LIMIT"
	OrderTypeStoplossMarket OrderType = "STOPLOSS_MARKET"

	ProductTypeDelivery    ProductType = "DELIVERY"
	ProductTypeCarryfwd    ProductType = "CARRYFORWARD"
	ProductTypeMargin      ProductType = "MARGIN"
	ProductTypeIntraday    ProductType = "INTRADAY"
	ProductTypeBO          ProductType = "BO"

	DurationDay Duration = "DAY"
	DurationIOC Duration = "IOC"

	ExchangeBSE Exchange = "BSE"
	ExchangeNSE Exchange = "NSE"
	ExchangeNFO Exchange = "NFO"
	ExchangeMCX Exchange = "MCX"
	ExchangeBFO Exchange = "BFO"
)

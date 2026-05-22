package aliceblue

type TransactionType string
type OrderType string
type ProductType string
type Validity string
type Exchange string
type OrderComplexity string

const (
	TransactionTypeBuy  TransactionType = "BUY"
	TransactionTypeSell TransactionType = "SELL"

	OrderTypeLimit  OrderType = "LIMIT"
	OrderTypeMarket OrderType = "MARKET"
	OrderTypeSL     OrderType = "SL"
	OrderTypeSLM    OrderType = "SLM"

	ProductTypeIntraday ProductType = "INTRADAY"
	ProductTypeLongterm ProductType = "LONGTERM"
	ProductTypeMTF      ProductType = "MTF"

	ValidityDay Validity = "DAY"
	ValidityIOC Validity = "IOC"

	ExchangeNSE Exchange = "NSE"
	ExchangeBSE Exchange = "BSE"
	ExchangeNFO Exchange = "NFO"
	ExchangeMCX Exchange = "MCX"
	ExchangeCDS Exchange = "CDS"
	ExchangeBFO Exchange = "BFO"
	ExchangeBCD Exchange = "BCD"

	OrderComplexityRegular OrderComplexity = "REGULAR"
	OrderComplexityAMO     OrderComplexity = "AMO"
	OrderComplexityCO      OrderComplexity = "CO"
	OrderComplexityBO      OrderComplexity = "BO"
)

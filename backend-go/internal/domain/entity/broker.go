package entity

import "time"

// BrokerSummary represents a summary of broker transactions for a specific stock.
type BrokerSummary struct {
	Symbol        string        `json:"symbol"`
	Timestamp     time.Time     `json:"timestamp"`
	Buyers        []BrokerTrade `json:"buyers"`
	Sellers       []BrokerTrade `json:"sellers"`
	NetStatus     string        `json:"net_status"`      // e.g., "BIG ACCUM", "NORMAL ACCUM", "NEUTRAL", etc.
	ForeignBuyPct float64       `json:"foreign_buy_pct"` // Percentage of foreign buying value
}

// BrokerTrade represents a transaction summary for a single broker.
type BrokerTrade struct {
	BrokerCode  string  `json:"broker_code"` // e.g. YP, PD, CC
	BrokerName  string  `json:"broker_name"` // Full broker name
	Volume      int64   `json:"volume"`      // Traded volume (lots or shares)
	AvgPrice    float64 `json:"avg_price"`   // Average transaction price
	Value       float64 `json:"value"`       // Total value in IDR (Volume * AvgPrice)
	Nationality string  `json:"nationality"` // "L" (Local) or "F" (Foreign)
}

package exchange

import (
	"context"
	"time"
)

// Kline represents a single K-line (candlestick) data point.
type Kline struct {
	OpenTime  time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	CloseTime time.Time
}

// Ticker represents simplified ticker information for popular symbols.
type Ticker struct {
	Symbol string
	Price  float64
	Volume float64
}

// ExchangeClient defines the interface for interacting with cryptocurrency exchanges.
type ExchangeClient interface {
	GetLatestPrice(ctx context.Context, symbol string) (float64, error)
	GetKlines(ctx context.Context, symbol string, interval string, limit int) ([]Kline, error)
	GetTopVolumeTickers(ctx context.Context, limit int) ([]Ticker, error)
}

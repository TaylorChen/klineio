package model

import (
	"testing"
	"time"
)

func TestExchangePriceInit(t *testing.T) {
	timeNow := time.Now()
	p := ExchangePrice{
		ID:        1,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
		Symbol:    "BTCUSDT",
		Exchange:  "binance",
		Price:     12345.67,
		Timestamp: 1688888888,
		Date:      timeNow,
	}
	if p.Symbol != "BTCUSDT" || p.Exchange != "binance" || p.Price != 12345.67 {
		t.Errorf("ExchangePrice fields not set correctly")
	}
}

func TestMonitorConfigInit(t *testing.T) {
	mc := MonitorConfig{
		ID:        1,
		UserID:    2,
		Symbol:    "ETHUSDT",
		Exchange:  "okex",
		Threshold: 0.2,
		Enable:    true,
	}
	if mc.Symbol != "ETHUSDT" || mc.Exchange != "okex" || !mc.Enable {
		t.Errorf("MonitorConfig fields not set correctly")
	}
}

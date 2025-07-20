package model

import (
	"time"

	"gorm.io/gorm"
)

// ExchangePrice represents a cryptocurrency price record from an exchange.
type ExchangePrice struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	Symbol    string         `gorm:"type:varchar(20);not null;index:idx_symbol_exchange_date,unique" json:"symbol"`   // Modified index
	Exchange  string         `gorm:"type:varchar(20);not null;index:idx_symbol_exchange_date,unique" json:"exchange"` // Modified index
	Price     float64        `gorm:"type:decimal(20,8);not null" json:"price"`
	Timestamp int64          `gorm:"not null;index" json:"timestamp"`                          // Unix timestamp of the price
	Date      time.Time      `gorm:"type:date;not null;index:idx_symbol_exchange_date,unique"` // New: Date part of Timestamp for daily unique key
}

// MonitorConfig represents a user's cryptocurrency monitoring configuration.
type MonitorConfig struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	Symbol    string         `gorm:"type:varchar(20);not null;uniqueIndex" json:"symbol"`
	Exchange  string         `gorm:"type:varchar(20);not null;uniqueIndex" json:"exchange"`
	Threshold float64        `gorm:"type:decimal(5,2);not null" json:"threshold"` // Percentage drop, e.g., 0.20 for 20%
	Enable    bool           `gorm:"default:true" json:"enable"`
}

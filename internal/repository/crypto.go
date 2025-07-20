package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"klineio/internal/model"
	"klineio/pkg/log"

	"go.uber.org/zap"
)

type ExchangePriceRepository interface {
	GetLatestExchangePriceBySymbolAndExchange(ctx context.Context, symbol, exchange string) (*model.ExchangePrice, error)
	GetAveragePriceForLastNDays(ctx context.Context, symbol, exchange string, days int) (float64, error)
	UpsertExchangePrice(ctx context.Context, price *model.ExchangePrice) error
}

type exchangePriceRepository struct {
	repo   *Repository
	logger *log.Logger
}

func NewExchangePriceRepository(
	repo *Repository,
	logger *log.Logger,
) ExchangePriceRepository {
	return &exchangePriceRepository{repo: repo, logger: logger}
}

func (r *exchangePriceRepository) DB(ctx context.Context) *gorm.DB {
	return r.repo.DB(ctx).Model(&model.ExchangePrice{})
}

// GetLatestExchangePriceBySymbolAndExchange retrieves the latest exchange price record for a given symbol and exchange.
func (r *exchangePriceRepository) GetLatestExchangePriceBySymbolAndExchange(ctx context.Context, symbol, exchange string) (*model.ExchangePrice, error) {
	var price model.ExchangePrice
	// Order by timestamp DESC to get the latest price for the day if multiple exist, or just the latest if only one daily record is kept.
	// The index on (symbol, exchange, date) makes the WHERE clause efficient.
	err := r.DB(ctx).Where("symbol = ? AND exchange = ?", symbol, exchange).Order("timestamp DESC").First(&price).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.WithContext(ctx).Debug("record not found for symbol and exchange",
				zap.String("symbol", symbol),
				zap.String("exchange", exchange),
			)
		}
		return nil, err
	}
	return &price, nil
}

// GetAveragePriceForLastNDays calculates the average price for a given symbol and exchange over the last N days.
func (r *exchangePriceRepository) GetAveragePriceForLastNDays(ctx context.Context, symbol, exchange string, days int) (float64, error) {
	timeThreshold := time.Now().UTC().AddDate(0, 0, -days) // Calculate threshold N days ago (UTC)

	var avgPrice float64
	// Sum prices and count records for the last N days, grouping by date to ensure distinct daily records are considered.
	// This query assumes that each (symbol, exchange, date) combination has at most one record representing the end-of-day price.
	// If multiple records per day can exist, and you want the latest for each day, you'd need a more complex subquery or a view.
	// Given the Upsert logic, this should work correctly for daily unique records.
	err := r.DB(ctx).Model(&model.ExchangePrice{}).Select("AVG(price)").
		Where("symbol = ? AND exchange = ? AND date >= ?", symbol, exchange, timeThreshold).Row().Scan(&avgPrice)

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.WithContext(ctx).Debug("no records found for average price calculation",
				zap.String("symbol", symbol),
				zap.String("exchange", exchange),
				zap.Int("days", days),
			)
		}
		return 0, err
	}

	return avgPrice, nil
}

// UpsertExchangePrice creates or updates an exchange price record based on symbol, exchange, and date.
func (r *exchangePriceRepository) UpsertExchangePrice(ctx context.Context, price *model.ExchangePrice) error {
	// Convert int64 Timestamp (milliseconds) to time.Time for Date calculation
	actualTimestamp := time.UnixMilli(price.Timestamp).UTC() // Use time.UnixMilli to correctly convert milliseconds to time.Time

	// Extract the date part from the timestamp (UTC)
	// This ensures that all prices for the same day (UTC) are upserted into one record.
	price.Date = actualTimestamp.Truncate(24 * time.Hour) // Use Truncate on time.Time

	err := r.DB(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "symbol"}, {Name: "exchange"}, {Name: "date"}},  // Conflict target: symbol, exchange, date
		DoUpdates: clause.AssignmentColumns([]string{"price", "timestamp", "updated_at"}), // Update price, timestamp, and updated_at on conflict
	}).Create(price).Error

	if err != nil {
		r.logger.WithContext(ctx).Error("failed to upsert exchange price",
			zap.Error(err),
			zap.String("symbol", price.Symbol),
			zap.String("exchange", price.Exchange),
			zap.Float64("price", price.Price),
			zap.Time("timestamp", actualTimestamp),
			zap.Time("date", price.Date),
		)
		return fmt.Errorf("failed to upsert exchange price: %w", err)
	}
	return nil
}

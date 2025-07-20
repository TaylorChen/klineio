package service

import (
	"context"
	"fmt"
	"time"

	"klineio/internal/model"
	"klineio/internal/repository"
	"klineio/pkg/exchange"
	"klineio/pkg/log"
	"klineio/pkg/notifier"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	// Interval for fetching K-lines, e.g., "1d" for 1 day
	KlineInterval = "1d"
	// Limit for fetching K-lines, e.g., 30 for 30 days
	KlineLimit = 30
)

// PriceMonitorService handles cryptocurrency price monitoring.
type PriceMonitorService struct {
	priceRepo repository.ExchangePriceRepository
	// monitorRepo      repository.MonitorConfigRepository // Removed: No longer directly used for main monitoring logic
	exchangeClients  map[string]exchange.ExchangeClient
	notifier         *notifier.DingTalkNotifier
	logger           *log.Logger
	defaultThreshold float64       // New: Default price drop threshold
	topNSymbols      int           // New: Number of top symbols to fetch
	apiRequestDelay  time.Duration // New: Delay between API requests for each symbol
}

// NewPriceMonitorService creates a new PriceMonitorService.
func NewPriceMonitorService(
	priceRepo repository.ExchangePriceRepository, // Corrected: remove pointer
	// monitorRepo repository.MonitorConfigRepository, // Removed: No longer directly used for main monitoring logic
	binanceClient *exchange.BinanceClient,
	okexClient *exchange.OKEXClient,
	notifier *notifier.DingTalkNotifier,
	logger *log.Logger,
	conf *viper.Viper,
) *PriceMonitorService {
	exchangeClients := make(map[string]exchange.ExchangeClient)
	exchangeClients["BINANCE"] = binanceClient
	exchangeClients["OKEX"] = okexClient

	return &PriceMonitorService{
		priceRepo: priceRepo, // Corrected: remove dereference
		// monitorRepo:      monitorRepo, // Removed: No longer directly used for main monitoring logic
		exchangeClients:  exchangeClients,
		notifier:         notifier,
		logger:           logger,
		defaultThreshold: conf.GetFloat64("price_monitor.default_threshold"),
		topNSymbols:      conf.GetInt("price_monitor.top_n_symbols"),
		apiRequestDelay:  time.Duration(conf.GetInt("price_monitor.api_request_delay_ms")) * time.Millisecond, // Read from config
	}
}

// RunMonitor fetches prices, calculates averages, and sends notifications if thresholds are met.
func (s *PriceMonitorService) RunMonitor(ctx context.Context) error {
	s.logger.Info("Starting price monitor run for top symbols")

	// No longer fetching monitor configs from DB as main source
	// configs, err := s.monitorRepo.GetMonitorConfigs(ctx)
	// if err != nil {
	//     s.logger.Error("Failed to get monitor configs", zap.Error(err))
	//     return fmt.Errorf("failed to get monitor configs: %w", err)
	// }

	for exchangeName, client := range s.exchangeClients {
		s.logger.Info("Fetching top symbols for exchange", zap.String("exchange", exchangeName))

		tickers, err := client.GetTopVolumeTickers(ctx, s.topNSymbols)
		if err != nil {
			s.logger.Error("Failed to get top volume tickers", zap.Error(err), zap.String("exchange", exchangeName))
			continue
		}

		if len(tickers) == 0 {
			s.logger.Warn("No top tickers data for exchange", zap.String("exchange", exchangeName))
			continue
		}

		for i, ticker := range tickers {
			// Check context before processing each ticker
			select {
			case <-ctx.Done():
				s.logger.Info("Context cancelled, stopping price monitor", zap.String("exchange", exchangeName))
				return ctx.Err()
			default:
			}

			symbol := ticker.Symbol // Use the symbol from the fetched ticker

			// 1. Fetch latest price and store it (upsert logic to ensure daily unique records)
			latestPrice, err := s.FetchAndStorePrice(ctx, client, symbol, exchangeName)
			if err != nil {
				s.logger.Error("Failed to fetch and store latest price for top symbol",
					zap.Error(err),
					zap.String("symbol", symbol),
					zap.String("exchange", exchangeName))
				continue
			}

			// 2. Get historical K-lines for average calculation
			klines, err := client.GetKlines(ctx, symbol, KlineInterval, KlineLimit)
			if err != nil {
				s.logger.Error("Failed to get klines for top symbol",
					zap.Error(err),
					zap.String("symbol", symbol),
					zap.String("exchange", exchangeName))
				continue
			}

			if len(klines) == 0 {
				s.logger.Warn("No klines data for top symbol",
					zap.String("symbol", symbol),
					zap.String("exchange", exchangeName))
				continue
			}

			// 3. Calculate 30-day average price based on close prices
			var closePrices []float64
			for _, kline := range klines {
				closePrices = append(closePrices, kline.Close)
			}
			averagePrice := s.CalculateAveragePrice(closePrices)

			// 4. Check for price drop using default threshold
			if averagePrice > 0 && latestPrice < averagePrice*(1-s.defaultThreshold) {
				dropPercentage := (1 - latestPrice/averagePrice) * 100
				title := "价格下跌警报！"
				text := fmt.Sprintf("### %s (%s) 价格下跌警报！\n\n", symbol, exchangeName) +
					fmt.Sprintf("- **当前价格**: %.4f\n", latestPrice) +
					fmt.Sprintf("- **近%d天平均价格**: %.4f\n", KlineLimit, averagePrice) +
					fmt.Sprintf("- **跌幅**: %.2f%% (阈值: %.2f%%)\n", dropPercentage, s.defaultThreshold*100) +
					fmt.Sprintf("- **来源**: 热门币种监控")

				s.logger.Info("Sending price drop alert for top symbol",
					zap.String("symbol", symbol),
					zap.String("exchange", exchangeName),
					zap.Float64("latestPrice", latestPrice),
					zap.Float64("averagePrice", averagePrice),
					zap.Float64("dropPercentage", dropPercentage))

				err = s.notifier.SendMarkdownMessage(ctx, title, text)
				if err != nil {
					s.logger.Error("Failed to send DingTalk notification for top symbol", zap.Error(err))
				}
			}

			// Introduce a delay after processing each ticker to avoid rate limits
			if s.apiRequestDelay > 0 && i < len(tickers)-1 {
				s.logger.Debug("Introducing API request delay", zap.Duration("duration", s.apiRequestDelay))
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(s.apiRequestDelay):
					// Continue after delay
				}
			}
		}
	}

	s.logger.Info("Price monitor run finished for top symbols")

	// Optional: If you still want to process custom monitor configs from DB:
	// s.logger.Info("Starting price monitor run for custom configs")
	// configs, err := s.monitorRepo.GetMonitorConfigs(ctx)
	// if err != nil {
	//     s.logger.Error("Failed to get custom monitor configs", zap.Error(err))
	//     return fmt.Errorf("failed to get custom monitor configs: %w", err)
	// }
	// // ... your existing logic for processing custom configs ...

	return nil
}

// FetchAndStorePrice fetches the latest price and stores it in the database using upsert logic.
func (s *PriceMonitorService) FetchAndStorePrice(ctx context.Context, client exchange.ExchangeClient, symbol, exchangeName string) (float64, error) {
	price, err := client.GetLatestPrice(ctx, symbol)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest price for %s from %s: %w", symbol, exchangeName, err)
	}

	currentTime := time.Now()
	// The `UpsertExchangePrice` in repository will handle the daily unique record logic.

	exchangePrice := &model.ExchangePrice{
		Symbol:    symbol,
		Exchange:  exchangeName,
		Price:     price,
		Timestamp: currentTime.UnixMilli(),
	}
	err = s.priceRepo.UpsertExchangePrice(ctx, exchangePrice)
	if err != nil {
		return 0, fmt.Errorf("failed to upsert exchange price: %w", err)
	}
	s.logger.Debug("Upserted price record", zap.String("symbol", symbol), zap.String("exchange", exchangeName), zap.Float64("price", price))

	return price, nil
}

// CalculateAveragePrice calculates the average of a slice of prices.
func (s *PriceMonitorService) CalculateAveragePrice(prices []float64) float64 {
	if len(prices) == 0 {
		return 0
	}
	var sum float64
	for _, price := range prices {
		sum += price
	}
	return sum / float64(len(prices))
}

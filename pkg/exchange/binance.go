package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"klineio/pkg/log"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const binanceAPIURL = "https://api.binance.com/api/v3"

// BinanceClient implements the ExchangeClient interface for Binance.
type BinanceClient struct {
	client *http.Client
	logger *log.Logger
}

// NewBinanceClient creates a new BinanceClient.
func NewBinanceClient(logger *log.Logger, conf *viper.Viper) *BinanceClient {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}

	proxyURL := conf.GetString("proxy.http")
	if proxyURL != "" {
		parsedProxyURL, err := url.Parse(proxyURL)
		if err != nil {
			logger.Warn("Failed to parse HTTP proxy URL", zap.Error(err), zap.String("proxy_url", proxyURL))
		} else {
			transport.Proxy = http.ProxyURL(parsedProxyURL)
		}
	}

	return &BinanceClient{
		client: &http.Client{Timeout: 60 * time.Second, Transport: transport}, // Increased timeout to 60 seconds
		logger: logger,
	}
}

// GetLatestPrice fetches the latest price for a given symbol from Binance.
func (b *BinanceClient) GetLatestPrice(ctx context.Context, symbol string) (float64, error) {
	url := fmt.Sprintf("%s/ticker/price?symbol=%s", binanceAPIURL, symbol)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := b.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("binance API returned non-OK status: %s", res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}

	var response struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return 0, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	price, err := strconv.ParseFloat(response.Price, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse price: %w", err)
	}

	return price, nil
}

// GetKlines fetches K-line data for a given symbol, interval, and limit from Binance.
func (b *BinanceClient) GetKlines(ctx context.Context, symbol string, interval string, limit int) ([]Kline, error) {
	url := fmt.Sprintf("%s/klines?symbol=%s&interval=%s&limit=%d", binanceAPIURL, symbol, interval, limit)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("binance API returned non-OK status: %s", res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var rawKlines [][]interface{}
	if err := json.Unmarshal(body, &rawKlines); err != nil {
		return nil, fmt.Errorf("failed to unmarshal klines response: %w", err)
	}

	var klines []Kline
	for _, rawKline := range rawKlines {
		openTimeMs, ok := rawKline[0].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid open time type")
		}
		open, _ := strconv.ParseFloat(rawKline[1].(string), 64)
		high, _ := strconv.ParseFloat(rawKline[2].(string), 64)
		low, _ := strconv.ParseFloat(rawKline[3].(string), 64)
		close, _ := strconv.ParseFloat(rawKline[4].(string), 64)
		volume, _ := strconv.ParseFloat(rawKline[5].(string), 64)
		closeTimeMs, ok := rawKline[6].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid close time type")
		}

		klines = append(klines, Kline{
			OpenTime:  time.Unix(0, int64(openTimeMs)*int64(time.Millisecond)),
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
			CloseTime: time.Unix(0, int64(closeTimeMs)*int64(time.Millisecond)),
		})
	}

	return klines, nil
}

// GetTopVolumeTickers fetches ticker information, sorts by volume, and returns top N.
func (b *BinanceClient) GetTopVolumeTickers(ctx context.Context, limit int) ([]Ticker, error) {
	url := fmt.Sprintf("%s/ticker/24hr", binanceAPIURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("binance API returned non-OK status: %s", res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var rawTickers []struct {
		Symbol    string `json:"symbol"`
		LastPrice string `json:"lastPrice"`
		Volume    string `json:"volume"` // Base asset volume
	}
	if err := json.Unmarshal(body, &rawTickers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tickers response: %w", err)
	}

	var tickers []Ticker
	for _, raw := range rawTickers {
		price, err := strconv.ParseFloat(raw.LastPrice, 64)
		if err != nil {
			b.logger.Warn("Failed to parse Binance ticker price", zap.Error(err), zap.String("symbol", raw.Symbol), zap.String("price", raw.LastPrice))
			continue
		}
		volume, err := strconv.ParseFloat(raw.Volume, 64)
		if err != nil {
			b.logger.Warn("Failed to parse Binance ticker volume", zap.Error(err), zap.String("symbol", raw.Symbol), zap.String("volume", raw.Volume))
			continue
		}

		// Filter for USDT pairs (or whatever base currency is desired)
		if !strings.HasSuffix(raw.Symbol, "USDT") {
			continue
		}

		tickers = append(tickers, Ticker{
			Symbol: raw.Symbol,
			Price:  price,
			Volume: volume,
		})
	}

	// Sort by volume in descending order
	sort.Slice(tickers, func(i, j int) bool {
		return tickers[i].Volume > tickers[j].Volume
	})

	// Return top N
	if len(tickers) > limit {
		return tickers[:limit], nil
	}
	return tickers, nil
}

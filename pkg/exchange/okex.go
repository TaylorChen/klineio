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

const okexAPIURL = "https://www.okx.com/api/v5/market"

// OKEXClient implements the ExchangeClient interface for OKEX.
type OKEXClient struct {
	client *http.Client
	logger *log.Logger
}

// NewOKEXClient creates a new OKEXClient.
func NewOKEXClient(logger *log.Logger, conf *viper.Viper) *OKEXClient {
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

	return &OKEXClient{
		client: &http.Client{Timeout: 60 * time.Second, Transport: transport}, // Increased timeout to 60 seconds
		logger: logger,
	}
}

// GetLatestPrice fetches the latest price for a given symbol from OKEX.
func (o *OKEXClient) GetLatestPrice(ctx context.Context, symbol string) (float64, error) {
	// OKEX现货交易对通常为 BTC-USDT 格式，需要将 symbol (例如 BTCUSDT) 转换为 BTC-USDT
	instId := symbol
	if !strings.Contains(symbol, "-") && strings.HasSuffix(symbol, "USDT") {
		instId = strings.Replace(symbol, "USDT", "-USDT", 1)
	}

	url := fmt.Sprintf("%s/tickers?instType=SPOT&instId=%s", okexAPIURL, instId)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := o.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("OKEX API returned non-OK status: %s", res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}

	var response struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			InstId string `json:"instId"`
			Last   string `json:"last"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return 0, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Code != "0" || len(response.Data) == 0 {
		return 0, fmt.Errorf("OKEX API error: %s - %s", response.Code, response.Msg)
	}

	price, err := strconv.ParseFloat(response.Data[0].Last, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse price: %w", err)
	}

	return price, nil
}

// GetKlines fetches K-line data for a given symbol, interval, and limit from OKEX.
func (o *OKEXClient) GetKlines(ctx context.Context, symbol string, interval string, limit int) ([]Kline, error) {
	// OKEX interval mapping: 1m, 3m, 5m, 15m, 30m, 1h, 2h, 4h, 6h, 12h, 1d, 2d, 3d, 5d, 1w, 1M, 3M, 6M, 1Y
	okexInterval := map[string]string{
		"1m":  "1m",
		"5m":  "5m",
		"15m": "15m",
		"30m": "30m",
		"1h":  "1H",
		"4h":  "4H",
		"1d":  "1D",
	}

	mappedInterval, ok := okexInterval[interval]
	if !ok {
		return nil, fmt.Errorf("unsupported interval for OKEX: %s", interval)
	}

	// Convert symbol (e.g., APTUSDT) to OKEX format (e.g., APT-USDT)
	instId := symbol
	if !strings.Contains(symbol, "-") && strings.HasSuffix(symbol, "USDT") {
		instId = strings.Replace(symbol, "USDT", "-USDT", 1)
	}

	url := fmt.Sprintf("%s/candles?instId=%s&bar=%s&limit=%d", okexAPIURL, instId, mappedInterval, limit)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OKEX API returned non-OK status: %s", res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response struct {
		Code string     `json:"code"`
		Msg  string     `json:"msg"`
		Data [][]string `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal klines response: %w", err)
	}

	if response.Code != "0" {
		return nil, fmt.Errorf("OKEX API error: %s - %s", response.Code, response.Msg)
	}

	var klines []Kline
	for _, rawKline := range response.Data {
		openTimeMs, _ := strconv.ParseInt(rawKline[0], 10, 64)
		open, _ := strconv.ParseFloat(rawKline[1], 64)
		high, _ := strconv.ParseFloat(rawKline[2], 64)
		low, _ := strconv.ParseFloat(rawKline[3], 64)
		close, _ := strconv.ParseFloat(rawKline[4], 64)
		volume, _ := strconv.ParseFloat(rawKline[5], 64)

		klines = append(klines, Kline{
			OpenTime:  time.Unix(0, openTimeMs*int64(time.Millisecond)),
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
			CloseTime: time.Unix(0, openTimeMs*int64(time.Millisecond)), // OKEX only provides open time timestamp in K-line data.
		})
	}

	return klines, nil
}

// GetTopVolumeTickers fetches ticker information, sorts by volume, and returns top N.
func (o *OKEXClient) GetTopVolumeTickers(ctx context.Context, limit int) ([]Ticker, error) {
	url := fmt.Sprintf("%s/tickers?instType=SPOT", okexAPIURL) // Fetch all spot tickers
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OKEX API returned non-OK status: %s", res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			InstID    string `json:"instId"` // e.g., BTC-USDT
			LastPrice string `json:"last"`
			Vol24h    string `json:"volCcy24h"` // 24h trading volume of quote currency
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tickers response: %w", err)
	}

	if response.Code != "0" {
		return nil, fmt.Errorf("OKEX API error: %s - %s", response.Code, response.Msg)
	}

	var tickers []Ticker
	for _, raw := range response.Data {
		// Filter for USDT pairs
		if !strings.HasSuffix(raw.InstID, "-USDT") {
			continue
		}

		price, err := strconv.ParseFloat(raw.LastPrice, 64)
		if err != nil {
			o.logger.Warn("Failed to parse OKEX ticker price", zap.Error(err), zap.String("instId", raw.InstID), zap.String("price", raw.LastPrice))
			continue
		}
		volume, err := strconv.ParseFloat(raw.Vol24h, 64) // Use Vol24h as volume
		if err != nil {
			o.logger.Warn("Failed to parse OKEX ticker volume", zap.Error(err), zap.String("instId", raw.InstID), zap.String("volume", raw.Vol24h))
			continue
		}

		tickers = append(tickers, Ticker{
			Symbol: strings.ReplaceAll(raw.InstID, "-", ""), // Convert BTC-USDT to BTCUSDT
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

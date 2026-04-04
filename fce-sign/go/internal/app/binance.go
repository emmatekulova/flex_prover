package app

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	binanceAPIBaseURL        = BinanceSpotAPIBaseURL()
	binanceFuturesAPIBaseURL = BinanceFuturesAPIBaseURL()
)

// BinanceProvider implements CEXProvider and ABIEncoderProvider for Binance.
type BinanceProvider struct{}

func init() {
	RegisterCEX("binance", &BinanceProvider{})
}

// FetchAndAttest fetches the current ticker price and returns a JSON attestation payload.
func (b *BinanceProvider) FetchAndAttest(apiKey, _, symbol string) ([]byte, error) {
	ticker, err := fetchBinanceTicker(apiKey, symbol)
	if err != nil {
		return nil, err
	}

	payload := BinanceAttestationPayload{
		Source:    "binance",
		Symbol:    ticker.Symbol,
		Price:     ticker.Price,
		FetchedAt: time.Now().Unix(),
		Version:   Version,
	}

	// Update observability state.
	lastBinanceSymbol = payload.Symbol
	lastBinancePrice = payload.Price
	lastBinanceAt = payload.FetchedAt

	return json.Marshal(payload)
}

// Fetch24hStats fetches 24-hour market statistics and returns a JSON attestation payload.
func (b *BinanceProvider) Fetch24hStats(_, _, symbol string) ([]byte, error) {
	stats, err := fetchBinance24hTicker(symbol)
	if err != nil {
		return nil, err
	}

	payload := Binance24hStatsAttestationPayload{
		Source:             "binance-24h",
		Symbol:             stats.Symbol,
		LastPrice:          stats.LastPrice,
		PriceChangePercent: stats.PriceChangePercent,
		Volume:             stats.Volume,
		QuoteVolume:        stats.QuoteVolume,
		OpenTime:           stats.OpenTime,
		CloseTime:          stats.CloseTime,
		FetchedAt:          time.Now().Unix(),
		Version:            Version,
	}

	lastBinanceSymbol = payload.Symbol
	lastBinancePrice = payload.LastPrice
	lastBinanceAt = payload.FetchedAt

	return json.Marshal(payload)
}

// FetchAccountPnl fetches authenticated futures account PnL and returns a JSON attestation payload.
func (b *BinanceProvider) FetchAccountPnl(apiKey, secretKey string) ([]byte, error) {
	account, err := fetchBinanceFuturesAccount(apiKey, secretKey)
	if err != nil {
		return nil, err
	}

	payload := BinanceAccountPnlAttestationPayload{
		Source:                "binance-futures",
		AccountAlias:          account.AccountAlias,
		CanTrade:              account.CanTrade,
		TotalWalletBalance:    account.TotalWalletBalance,
		TotalUnrealizedProfit: account.TotalUnrealizedProfit,
		TotalMarginBalance:    account.TotalMarginBalance,
		FetchedAt:             time.Now().Unix(),
		Version:               Version,
	}

	lastBinanceUnrealizedProfit = payload.TotalUnrealizedProfit
	lastBinanceAt = payload.FetchedAt

	return json.Marshal(payload)
}

// FetchAccountSummary fetches authenticated spot account balances and returns a JSON attestation payload.
func (b *BinanceProvider) FetchAccountSummary(apiKey, secretKey string) ([]byte, error) {
	account, err := fetchBinanceSpotAccount(apiKey, secretKey)
	if err != nil {
		return nil, err
	}
	prices, err := fetchBinanceAllSpotPrices()
	if err != nil {
		return nil, err
	}

	assets, totalUSDT, unsupported := buildBinanceAssets(account.Balances, prices)
	payload := BinanceAccountSummaryAttestationPayload{
		Source:              "binance-account",
		CanTrade:            account.CanTrade,
		CanDeposit:          account.CanDeposit,
		CanWithdraw:         account.CanWithdraw,
		EstimatedTotalUSDT:  decimalToString(totalUSDT),
		UnsupportedAssetCnt: unsupported,
		Assets:              assets,
		FetchedAt:           time.Now().Unix(),
		Version:             Version,
	}

	lastBinanceEstimatedTotalUSDT = payload.EstimatedTotalUSDT
	lastBinanceAt = payload.FetchedAt

	return json.Marshal(payload)
}

// FetchUserProfile returns a JSON-marshaled user profile attestation payload.
// Use EncodeUserProfile for the ABI-encoded form required by BinanceAttestationStore.
func (b *BinanceProvider) FetchUserProfile(apiKey, secretKey string) ([]byte, error) {
	payload, err := b.buildUserProfilePayload(apiKey, secretKey)
	if err != nil {
		return nil, err
	}
	return json.Marshal(payload)
}

// EncodeUserProfile implements ABIEncoderProvider. Returns ABI-encoded bytes for
// on-chain decoding by BinanceAttestationStore.
func (b *BinanceProvider) EncodeUserProfile(apiKey, secretKey string) ([]byte, error) {
	payload, err := b.buildUserProfilePayload(apiKey, secretKey)
	if err != nil {
		return nil, err
	}
	return abiEncodeUserProfile(*payload)
}

func (b *BinanceProvider) buildUserProfilePayload(apiKey, secretKey string) (*BinanceUserProfileAttestationPayload, error) {
	account, err := fetchBinanceSpotAccount(apiKey, secretKey)
	if err != nil {
		return nil, err
	}
	prices, err := fetchBinanceAllSpotPrices()
	if err != nil {
		return nil, err
	}

	assets, totalUSDT, unsupported := buildBinanceAssets(account.Balances, prices)
	payload := &BinanceUserProfileAttestationPayload{
		Source:              "binance-user-profile",
		UID:                 account.UID,
		AccountType:         account.AccountType,
		Permissions:         account.Permissions,
		CanTrade:            account.CanTrade,
		CanDeposit:          account.CanDeposit,
		CanWithdraw:         account.CanWithdraw,
		EstimatedTotalUSDT:  decimalToString(totalUSDT),
		UnsupportedAssetCnt: unsupported,
		Assets:              assets,
		FetchedAt:           time.Now().Unix(),
		Version:             Version,
	}

	lastBinanceEstimatedTotalUSDT = payload.EstimatedTotalUSDT
	lastBinanceAt = payload.FetchedAt

	return payload, nil
}

// buildBinanceAssets computes per-asset USD values from spot balances and prices.
// Returns the enriched asset list, total estimated USDT, and count of unsupported assets.
func buildBinanceAssets(balances []BinanceSpotBalance, prices map[string]*big.Float) ([]BinanceAccountAssetSummary, *big.Float, int) {
	assets := make([]BinanceAccountAssetSummary, 0)
	totalUSDT := new(big.Float)
	unsupported := 0

	for _, balance := range balances {
		free, ok := new(big.Float).SetString(balance.Free)
		if !ok {
			continue
		}
		locked, ok := new(big.Float).SetString(balance.Locked)
		if !ok {
			continue
		}

		qty := new(big.Float).Add(free, locked)
		if qty.Sign() == 0 {
			continue
		}

		estimated := new(big.Float)
		asset := strings.ToUpper(strings.TrimSpace(balance.Asset))
		switch asset {
		case "USDT", "USDC", "BUSD", "FDUSD", "TUSD":
			estimated.Copy(qty)
		default:
			price, found := prices[asset+"USDT"]
			if !found {
				unsupported++
				continue
			}
			estimated.Mul(qty, price)
		}

		totalUSDT.Add(totalUSDT, estimated)
		assets = append(assets, BinanceAccountAssetSummary{
			Asset:         asset,
			Total:         decimalToString(qty),
			EstimatedUSDT: decimalToString(estimated),
		})
	}

	return assets, totalUSDT, unsupported
}

func fetchBinanceTicker(apiKey, symbol string) (*BinanceTickerPriceResponse, error) {
	endpoint := fmt.Sprintf("%s/api/v3/ticker/price?symbol=%s", strings.TrimRight(binanceAPIBaseURL, "/"), url.QueryEscape(symbol))

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Optionally add auth header if key is provided.
	if apiKey != "" {
		req.Header.Set("X-MBX-APIKEY", apiKey)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Binance data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("binance returned %d: %s", resp.StatusCode, string(b))
	}

	var ticker BinanceTickerPriceResponse
	if err := json.NewDecoder(resp.Body).Decode(&ticker); err != nil {
		return nil, fmt.Errorf("failed to decode Binance response: %w", err)
	}

	if ticker.Symbol == "" || ticker.Price == "" {
		return nil, fmt.Errorf("binance response missing symbol/price")
	}

	return &ticker, nil
}

func fetchBinanceFuturesAccount(apiKey, secretKey string) (*BinanceFuturesAccountResponse, error) {
	apiKey = strings.TrimSpace(apiKey)
	secretKey = strings.TrimSpace(secretKey)
	if apiKey == "" || secretKey == "" {
		return nil, fmt.Errorf("apiKey and secretKey are required for account PnL")
	}

	timestamp := time.Now().UnixMilli()
	query := fmt.Sprintf("timestamp=%d&recvWindow=5000", timestamp)
	signature := signBinanceQuery(secretKey, query)

	endpoint := fmt.Sprintf("%s/fapi/v2/account?%s&signature=%s", strings.TrimRight(binanceFuturesAPIBaseURL, "/"), query, signature)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("X-MBX-APIKEY", apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Binance account data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("binance futures returned %d: %s", resp.StatusCode, string(b))
	}

	var account BinanceFuturesAccountResponse
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return nil, fmt.Errorf("failed to decode Binance futures account response: %w", err)
	}

	if account.TotalWalletBalance == "" || account.TotalUnrealizedProfit == "" {
		return nil, fmt.Errorf("binance futures account response missing pnl fields")
	}

	return &account, nil
}

func fetchBinance24hTicker(symbol string) (*Binance24hrTickerResponse, error) {
	endpoint := fmt.Sprintf("%s/api/v3/ticker/24hr?symbol=%s", strings.TrimRight(binanceAPIBaseURL, "/"), url.QueryEscape(symbol))

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Binance 24h data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("binance 24h returned %d: %s", resp.StatusCode, string(b))
	}

	var stats Binance24hrTickerResponse
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode Binance 24h response: %w", err)
	}

	if stats.Symbol == "" || stats.LastPrice == "" {
		return nil, fmt.Errorf("binance 24h response missing symbol/lastPrice")
	}

	return &stats, nil
}

func fetchBinanceSpotAccount(apiKey, secretKey string) (*BinanceSpotAccountResponse, error) {
	apiKey = strings.TrimSpace(apiKey)
	secretKey = strings.TrimSpace(secretKey)
	if apiKey == "" || secretKey == "" {
		return nil, fmt.Errorf("apiKey and secretKey are required for account summary")
	}

	timestamp := time.Now().UnixMilli()
	query := fmt.Sprintf("timestamp=%d&recvWindow=5000", timestamp)
	signature := signBinanceQuery(secretKey, query)

	endpoint := fmt.Sprintf("%s/api/v3/account?%s&signature=%s", strings.TrimRight(binanceAPIBaseURL, "/"), query, signature)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("X-MBX-APIKEY", apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Binance account data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("binance account returned %d: %s", resp.StatusCode, string(b))
	}

	var account BinanceSpotAccountResponse
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return nil, fmt.Errorf("failed to decode Binance account response: %w", err)
	}

	return &account, nil
}

func fetchBinanceAllSpotPrices() (map[string]*big.Float, error) {
	endpoint := fmt.Sprintf("%s/api/v3/ticker/price", strings.TrimRight(binanceAPIBaseURL, "/"))
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Binance spot prices: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("binance ticker price returned %d: %s", resp.StatusCode, string(b))
	}

	var entries []BinanceTickerPriceEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("failed to decode Binance ticker price response: %w", err)
	}

	result := make(map[string]*big.Float, len(entries))
	for _, entry := range entries {
		p, ok := new(big.Float).SetString(entry.Price)
		if !ok {
			continue
		}
		result[entry.Symbol] = p
	}

	return result, nil
}

func signBinanceQuery(secret, query string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(query))
	return fmt.Sprintf("%x", mac.Sum(nil))
}

func decimalToString(v *big.Float) string {
	if v == nil {
		return "0"
	}
	return strings.TrimRight(strings.TrimRight(v.Text('f', 8), "0"), ".")
}

// fetch-positions-bitget fetches the authenticated user's current futures
// positions from Bitget and prints them as a JSON object. This is used by the
// Next.js frontend to populate the position-selection step before the user
// chooses which positions to include in an individual-trades attestation.
//
// Usage:
//
//	go run ./cmd/fetch-positions-bitget \
//	  -apiKey <BITGET_API_KEY> \
//	  -secretKey <BITGET_SECRET_KEY> \
//	  -passphrase <BITGET_PASSPHRASE>
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type bitgetAsset struct {
	Symbol            string `json:"symbol"`
	HoldSide          string `json:"holdSide"`
	Total             string `json:"total"`
	MarkPrice         string `json:"markPrice"`
	MarketPrice       string `json:"marketPrice"`
	AvgOpenPrice      string `json:"avgOpenPrice"`
	OpenPriceAvg      string `json:"openPriceAvg"`
	UnrealizedPL      string `json:"unrealizedPL"`
	LiquidationPrice  string `json:"liquidationPrice"`
	MarginCoin        string `json:"marginCoin"`
}

type bitgetAssetsResponse struct {
	Code string          `json:"code"`
	Data json.RawMessage `json:"data"`
}

type bitgetTicker struct {
	Symbol string `json:"symbol"`
	LastPr string `json:"lastPr"`
}

type bitgetTickersResponse struct {
	Code string         `json:"code"`
	Data []bitgetTicker `json:"data"`
}

type tradePosition struct {
	Asset     string `json:"asset"`
	Quantity  string `json:"quantity"`
	PriceUSDT string `json:"priceUsdt"`
	ValueUSDT string `json:"valueUsdt"`
}

type positionsOutput struct {
	Exchange  string          `json:"exchange"`
	Positions []tradePosition `json:"positions"`
	FetchedAt int64           `json:"fetchedAt"`
}

const baseURL = "https://api.bitget.com"

func main() {
	_ = godotenv.Load()
	_ = godotenv.Load("../../.env")

	apiKey := flag.String("apiKey", "", "Bitget API key (required)")
	secretKey := flag.String("secretKey", "", "Bitget secret key (required)")
	passphrase := flag.String("passphrase", "", "Bitget passphrase (required)")
	flag.Parse()

	if strings.TrimSpace(*apiKey) == "" || strings.TrimSpace(*secretKey) == "" || strings.TrimSpace(*passphrase) == "" {
		fmt.Fprintln(os.Stderr, "-apiKey, -secretKey, and -passphrase are required")
		os.Exit(1)
	}

	positions, err := fetchFuturesPositions(*apiKey, *secretKey, *passphrase)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fetch positions: %v\n", err)
		os.Exit(1)
	}

	out := positionsOutput{
		Exchange:  "bitget",
		Positions: positions,
		FetchedAt: time.Now().Unix(),
	}

	enc := json.NewEncoder(os.Stdout)
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "encode output: %v\n", err)
		os.Exit(1)
	}
}

func fetchFuturesPositions(apiKey, secretKey, passphrase string) ([]tradePosition, error) {
	var positions []tradePosition
	seen := make(map[string]struct{})
	var lastErr error
	hadSuccess := false

	productTypes := []string{"USDT-FUTURES", "USDC-FUTURES", "COIN-FUTURES"}
	for _, productType := range productTypes {
		entries, err := fetchFuturesPositionEntries(apiKey, secretKey, passphrase, productType)
		if err != nil {
			lastErr = err
			continue
		}
		hadSuccess = true
		for _, entry := range entries {
			position := convertBitgetPosition(entry)
			if position == nil {
				continue
			}
			if _, ok := seen[position.Asset]; ok {
				continue
			}
			seen[position.Asset] = struct{}{}
			positions = append(positions, *position)
		}
	}

	if len(positions) == 0 {
		if hadSuccess {
			return []tradePosition{}, nil
		}
		if lastErr != nil {
			return nil, lastErr
		}
		return []tradePosition{}, nil
	}
	return positions, nil
}

func fetchFuturesPositionEntries(apiKey, secretKey, passphrase, productType string) ([]bitgetAsset, error) {
	path := "/api/v2/mix/position/all-position?productType=" + url.QueryEscape(productType)
	timestamp := timestamp()
	sign := signRequest(secretKey, timestamp, "GET", path, "")

	endpoint := baseURL + path
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("ACCESS-KEY", apiKey)
	req.Header.Set("ACCESS-SIGN", sign)
	req.Header.Set("ACCESS-TIMESTAMP", timestamp)
	req.Header.Set("ACCESS-PASSPHRASE", passphrase)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bitget positions returned %d: %s", resp.StatusCode, string(b))
	}

	var result bitgetAssetsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.Code != "00000" {
		return nil, fmt.Errorf("bitget positions error code %s", result.Code)
	}

	return decodeBitgetPositions(result.Data)
}

func decodeBitgetPositions(raw json.RawMessage) ([]bitgetAsset, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}

	var entries []bitgetAsset
	if err := json.Unmarshal(raw, &entries); err == nil {
		return entries, nil
	}

	var wrapper map[string]json.RawMessage
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return nil, err
	}
	for _, key := range []string{"list", "data", "rows", "result"} {
		if nested, ok := wrapper[key]; ok {
			if decoded, err := decodeBitgetPositions(nested); err == nil && len(decoded) > 0 {
				return decoded, nil
			}
		}
	}
	return nil, fmt.Errorf("unsupported Bitget position response shape")
}

func convertBitgetPosition(asset bitgetAsset) *tradePosition {
	symbol := strings.ToUpper(strings.TrimSpace(asset.Symbol))
	if symbol == "" {
		return nil
	}

	quantity := parseBitgetFloat(asset.Total)
	if quantity.Sign() == 0 {
		return nil
	}

	price := firstNonEmpty(asset.MarkPrice, asset.MarketPrice, asset.AvgOpenPrice, asset.OpenPriceAvg)
	priceUSDT := parseBitgetFloat(price)
	if priceUSDT.Sign() == 0 {
		priceUSDT = new(big.Float)
	}

	valueUSDT := new(big.Float).Mul(quantity, priceUSDT)
	holdSide := strings.ToUpper(strings.TrimSpace(asset.HoldSide))
	assetName := symbol
	if holdSide != "" {
		assetName = symbol + "-" + holdSide
	}

	return &tradePosition{
		Asset:     assetName,
		Quantity:  quantity.Text('f', 8),
		PriceUSDT: priceUSDT.Text('f', 6),
		ValueUSDT: valueUSDT.Text('f', 2),
	}
}

func parseBitgetFloat(input string) *big.Float {
	value, ok := new(big.Float).SetString(strings.TrimSpace(input))
	if !ok {
		return new(big.Float)
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func signRequest(secretKey, ts, method, path, body string) string {
	prehash := ts + method + path + body
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(prehash))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func timestamp() string {
	return fmt.Sprintf("%d", time.Now().UnixMilli())
}

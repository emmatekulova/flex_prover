package app

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"sign-extension/internal/base"
	"strings"
	"time"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// state holds the mutable state for the extension.
// The framework serializes all handler calls, so no additional locking is needed.
var (
	privateKey *secp256k1.PrivateKey
	signPort   string
	httpClient = http.DefaultClient

	binanceAPIBaseURL = BinanceSpotAPIBaseURL()
	binanceFuturesAPIBaseURL = BinanceFuturesAPIBaseURL()
	lastBinanceSymbol string
	lastBinancePrice  string
	lastBinanceAt     int64
	lastBinanceUnrealizedProfit string
	lastBinanceEstimatedTotalUSDT string
)

// SetSignPort sets the sign port for communicating with the TEE node.
func SetSignPort(port string) {
	signPort = port
}

// Register registers the handlers and initial state with the framework.
func Register(f *base.Framework) {
	f.Handle(OpTypeKey, OpCommandUpdate, handleKeyUpdate)
	f.Handle(OpTypeKey, OpCommandSign, handleKeySign)
	f.Handle(OpTypeMarket, OpCommandBinanceFetchAndAttest, handleBinanceFetchAndAttest)
	f.Handle(OpTypeMarket, OpCommandBinance24hStats, handleBinance24hStats)
	f.Handle(OpTypeMarket, OpCommandBinanceAccountPnl, handleBinanceAccountPnl)
	f.Handle(OpTypeMarket, OpCommandBinanceAccountSummary, handleBinanceAccountSummary)
	f.Handle(OpTypeMarket, OpCommandBinanceUserProfile, handleBinanceUserProfile)
}

// ReportState returns a JSON snapshot of the current state.
func ReportState() json.RawMessage {
	hasKey := privateKey != nil
	data, _ := json.Marshal(map[string]interface{}{
		"hasKey":            hasKey,
		"version":           Version,
		"lastBinanceSymbol": lastBinanceSymbol,
		"lastBinancePrice":  lastBinancePrice,
		"lastBinanceAt":     lastBinanceAt,
		"lastBinanceUnrealizedProfit": lastBinanceUnrealizedProfit,
		"lastBinanceEstimatedTotalUsdt": lastBinanceEstimatedTotalUSDT,
	})
	return data
}

// handleBinanceUserProfile fetches the authenticated Binance spot account, enriches it
// with UID, account type, and permissions, computes per-asset USD values, then returns
// ABI-encoded (payload, signature) signed by the TEE node key.
func handleBinanceUserProfile(_ string) (data *string, status int, err error) {
	account, fetchErr := fetchBinanceSpotAccount()
	if fetchErr != nil {
		return nil, 0, fetchErr
	}

	prices, pricesErr := fetchBinanceAllSpotPrices()
	if pricesErr != nil {
		return nil, 0, pricesErr
	}

	assets := make([]BinanceAccountAssetSummary, 0)
	totalUSDT := new(big.Float)
	unsupported := 0

	for _, balance := range account.Balances {
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

	payload := BinanceUserProfileAttestationPayload{
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

	payloadBytes, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return nil, 0, fmt.Errorf("failed to marshal user profile payload: %v", marshalErr)
	}

	signature, signErr := signViaNode(payloadBytes)
	if signErr != nil {
		return nil, 0, fmt.Errorf("signing failed: %v", signErr)
	}

	encoded, abiErr := abiEncodeTwo(payloadBytes, signature)
	if abiErr != nil {
		return nil, 0, fmt.Errorf("ABI encoding failed: %v", abiErr)
	}

	lastBinanceEstimatedTotalUSDT = payload.EstimatedTotalUSDT
	lastBinanceAt = payload.FetchedAt

	dataHex := base.BytesToHex(encoded)
	return &dataHex, 1, nil
}

func handleBinanceAccountSummary(_ string) (data *string, status int, err error) {
	account, fetchErr := fetchBinanceSpotAccount()
	if fetchErr != nil {
		return nil, 0, fetchErr
	}

	prices, pricesErr := fetchBinanceAllSpotPrices()
	if pricesErr != nil {
		return nil, 0, pricesErr
	}

	assets := make([]BinanceAccountAssetSummary, 0)
	totalUSDT := new(big.Float)
	unsupported := 0

	for _, balance := range account.Balances {
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

	payloadBytes, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return nil, 0, fmt.Errorf("failed to marshal account summary payload: %v", marshalErr)
	}

	signature, signErr := signViaNode(payloadBytes)
	if signErr != nil {
		return nil, 0, fmt.Errorf("signing failed: %v", signErr)
	}

	encoded, abiErr := abiEncodeTwo(payloadBytes, signature)
	if abiErr != nil {
		return nil, 0, fmt.Errorf("ABI encoding failed: %v", abiErr)
	}

	lastBinanceEstimatedTotalUSDT = payload.EstimatedTotalUSDT
	lastBinanceAt = payload.FetchedAt

	dataHex := base.BytesToHex(encoded)
	return &dataHex, 1, nil
}

// handleBinanceAccountPnl fetches authenticated Binance futures account metrics,
// builds an account-PnL payload, signs it via the TEE node key, and returns
// ABI-encoded (payload, signature).
func handleBinanceAccountPnl(_ string) (data *string, status int, err error) {
	account, fetchErr := fetchBinanceFuturesAccount()
	if fetchErr != nil {
		return nil, 0, fetchErr
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

	payloadBytes, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return nil, 0, fmt.Errorf("failed to marshal account pnl payload: %v", marshalErr)
	}

	signature, signErr := signViaNode(payloadBytes)
	if signErr != nil {
		return nil, 0, fmt.Errorf("signing failed: %v", signErr)
	}

	encoded, abiErr := abiEncodeTwo(payloadBytes, signature)
	if abiErr != nil {
		return nil, 0, fmt.Errorf("ABI encoding failed: %v", abiErr)
	}

	lastBinanceUnrealizedProfit = payload.TotalUnrealizedProfit
	lastBinanceAt = payload.FetchedAt

	dataHex := base.BytesToHex(encoded)
	return &dataHex, 1, nil
}

func handleBinance24hStats(msg string) (data *string, status int, err error) {
	if msg == "" {
		return nil, 0, fmt.Errorf("originalMessage is empty")
	}

	msgBytes, hexErr := base.HexToBytes(msg)
	if hexErr != nil {
		return nil, 0, fmt.Errorf("invalid hex in originalMessage: %v", hexErr)
	}

	request, parseErr := parseBinanceFetchRequest(msgBytes)
	if parseErr != nil {
		return nil, 0, parseErr
	}

	stats, fetchErr := fetchBinance24hTicker(request.Symbol)
	if fetchErr != nil {
		return nil, 0, fetchErr
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

	payloadBytes, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return nil, 0, fmt.Errorf("failed to marshal 24h stats payload: %v", marshalErr)
	}

	signature, signErr := signViaNode(payloadBytes)
	if signErr != nil {
		return nil, 0, fmt.Errorf("signing failed: %v", signErr)
	}

	encoded, abiErr := abiEncodeTwo(payloadBytes, signature)
	if abiErr != nil {
		return nil, 0, fmt.Errorf("ABI encoding failed: %v", abiErr)
	}

	lastBinanceSymbol = payload.Symbol
	lastBinancePrice = payload.LastPrice
	lastBinanceAt = payload.FetchedAt

	dataHex := base.BytesToHex(encoded)
	return &dataHex, 1, nil
}

// handleBinanceFetchAndAttest fetches ticker price data from Binance, builds an
// attestation payload, signs it via the TEE node's private key, and returns
// ABI-encoded (payload, signature).
func handleBinanceFetchAndAttest(msg string) (data *string, status int, err error) {
	if msg == "" {
		return nil, 0, fmt.Errorf("originalMessage is empty")
	}

	msgBytes, hexErr := base.HexToBytes(msg)
	if hexErr != nil {
		return nil, 0, fmt.Errorf("invalid hex in originalMessage: %v", hexErr)
	}

	request, parseErr := parseBinanceFetchRequest(msgBytes)
	if parseErr != nil {
		return nil, 0, parseErr
	}

	ticker, fetchErr := fetchBinanceTicker(request.Symbol)
	if fetchErr != nil {
		return nil, 0, fetchErr
	}

	payload := BinanceAttestationPayload{
		Source:    "binance",
		Symbol:    ticker.Symbol,
		Price:     ticker.Price,
		FetchedAt: time.Now().Unix(),
		Version:   Version,
	}

	payloadBytes, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return nil, 0, fmt.Errorf("failed to marshal attestation payload: %v", marshalErr)
	}

	signature, signErr := signViaNode(payloadBytes)
	if signErr != nil {
		return nil, 0, fmt.Errorf("signing failed: %v", signErr)
	}

	encoded, abiErr := abiEncodeTwo(payloadBytes, signature)
	if abiErr != nil {
		return nil, 0, fmt.Errorf("ABI encoding failed: %v", abiErr)
	}

	lastBinanceSymbol = payload.Symbol
	lastBinancePrice = payload.Price
	lastBinanceAt = payload.FetchedAt

	dataHex := base.BytesToHex(encoded)
	return &dataHex, 1, nil
}

// handleKeyUpdate decrypts the original message using the TEE node's key, then
// stores the decrypted value as an ECDSA private key.
func handleKeyUpdate(msg string) (data *string, status int, err error) {
	if msg == "" {
		return nil, 0, fmt.Errorf("originalMessage is empty")
	}

	// originalMessage is a hex string (hexutil.Bytes JSON serialization).
	// Hex-decode to get the raw ECIES ciphertext bytes.
	ciphertext, hexErr := base.HexToBytes(msg)
	if hexErr != nil {
		return nil, 0, fmt.Errorf("invalid hex in originalMessage: %v", hexErr)
	}

	// Decrypt via TEE node — sends ciphertext bytes (JSON-serialized as base64).
	keyBytes, decryptErr := decryptViaNode(ciphertext)
	if decryptErr != nil {
		return nil, 0, fmt.Errorf("decryption failed: %v", decryptErr)
	}

	privKey, parseErr := parseSecp256k1PrivateKey(keyBytes)
	if parseErr != nil {
		return nil, 0, fmt.Errorf("invalid private key: %v", parseErr)
	}

	privateKey = privKey
	log.Printf("private key updated")
	return nil, 1, nil
}

// handleKeySign signs the original message with the stored private key.
// Returns the message and signature in data as ABI-encoded (bytes, bytes).
func handleKeySign(msg string) (data *string, status int, err error) {
	if privateKey == nil {
		return nil, 0, fmt.Errorf("no private key stored")
	}

	if msg == "" {
		return nil, 0, fmt.Errorf("originalMessage is empty")
	}

	msgBytes, hexErr := base.HexToBytes(msg)
	if hexErr != nil {
		return nil, 0, fmt.Errorf("invalid hex in originalMessage: %v", hexErr)
	}

	sig, signErr := signECDSA(privateKey, msgBytes)
	if signErr != nil {
		return nil, 0, fmt.Errorf("signing failed: %v", signErr)
	}

	encoded, abiErr := abiEncodeTwo(msgBytes, sig)
	if abiErr != nil {
		return nil, 0, fmt.Errorf("ABI encoding failed: %v", abiErr)
	}

	dataHex := base.BytesToHex(encoded)
	return &dataHex, 1, nil
}

// decryptViaNode calls the TEE node's /decrypt endpoint.
// ciphertext is the raw ECIES ciphertext bytes; it is JSON-serialized as base64
// in the request, matching the tee-node's DecryptRequest.EncryptedMessage []byte field.
// Returns the decrypted plaintext bytes (also base64-serialized by tee-node).
func decryptViaNode(ciphertext []byte) ([]byte, error) {
	url := fmt.Sprintf("http://localhost:%s/decrypt", signPort)
	reqBody, _ := json.Marshal(DecryptRequest{EncryptedMessage: ciphertext})

	resp, err := httpClient.Post(url, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("node returned %d: %s", resp.StatusCode, string(b))
	}

	var dr DecryptResponse
	if err := json.NewDecoder(resp.Body).Decode(&dr); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return dr.DecryptedMessage, nil
}

// signViaNode calls the TEE node's /sign endpoint.
// message is raw bytes and is JSON-serialized as base64 in the request body.
func signViaNode(message []byte) ([]byte, error) {
	url := fmt.Sprintf("http://localhost:%s/sign", signPort)
	reqBody, _ := json.Marshal(SignRequest{Message: message})

	resp, err := httpClient.Post(url, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("node returned %d: %s", resp.StatusCode, string(b))
	}

	var sr SignResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(sr.Signature) == 0 {
		return nil, fmt.Errorf("empty signature from node")
	}

	return sr.Signature, nil
}

func parseBinanceFetchRequest(raw []byte) (*BinanceFetchRequest, error) {
	var req BinanceFetchRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return nil, fmt.Errorf("invalid request payload: expected JSON {\"symbol\":\"...\"}")
	}

	req.Symbol = strings.TrimSpace(strings.ToUpper(req.Symbol))
	if req.Symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}

	return &req, nil
}

func fetchBinanceTicker(symbol string) (*BinanceTickerPriceResponse, error) {
	endpoint := fmt.Sprintf("%s/api/v3/ticker/price?symbol=%s", strings.TrimRight(binanceAPIBaseURL, "/"), url.QueryEscape(symbol))

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Optionally add auth headers if credentials are provided.
	// Note: Full signed requests require HMAC-SHA256; for now we just add the key header.
	apiKey := BinanceAPIKey()
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

func fetchBinanceFuturesAccount() (*BinanceFuturesAccountResponse, error) {
	apiKey := strings.TrimSpace(BinanceAPIKey())
	apiSecret := strings.TrimSpace(BinanceSecretKey())
	if apiKey == "" || apiSecret == "" {
		return nil, fmt.Errorf("BINANCE_API_KEY and BINANCE_SECRET_KEY are required for account PnL")
	}

	timestamp := time.Now().UnixMilli()
	query := fmt.Sprintf("timestamp=%d&recvWindow=5000", timestamp)
	signature := signBinanceQuery(apiSecret, query)

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

func signBinanceQuery(secret, query string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(query))
	return fmt.Sprintf("%x", mac.Sum(nil))
}

func fetchBinanceSpotAccount() (*BinanceSpotAccountResponse, error) {
	apiKey := strings.TrimSpace(BinanceAPIKey())
	apiSecret := strings.TrimSpace(BinanceSecretKey())
	if apiKey == "" || apiSecret == "" {
		return nil, fmt.Errorf("BINANCE_API_KEY and BINANCE_SECRET_KEY are required for account summary")
	}

	timestamp := time.Now().UnixMilli()
	query := fmt.Sprintf("timestamp=%d&recvWindow=5000", timestamp)
	signature := signBinanceQuery(apiSecret, query)

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

func decimalToString(v *big.Float) string {
	if v == nil {
		return "0"
	}
	return strings.TrimRight(strings.TrimRight(v.Text('f', 8), "0"), ".")
}

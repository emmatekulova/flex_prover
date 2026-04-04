package app

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
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
	f.Handle(OpTypeMarket, OpCommandBinanceFuturesDetails, handleBinanceFuturesDetails)
	f.Handle(OpTypeMarket, OpCommandBitgetAccountSummary, handleBitgetAccountSummary)
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

func handleBinanceFuturesDetails(_ string) (data *string, status int, err error) {
	// 1. Fetch Open Positions
	openPositions, fetchOpenErr := fetchBinanceFuturesPositionRisk()
	if fetchOpenErr != nil {
		return nil, 0, fmt.Errorf("failed to fetch open positions: %v", fetchOpenErr)
	}

	// 2. Fetch Recent Trade History (last 50 trades)
	recentTrades, fetchTradesErr := fetchBinanceFuturesTrades()
	if fetchTradesErr != nil {
		return nil, 0, fmt.Errorf("failed to fetch recent trades: %v", fetchTradesErr)
	}

	details := make([]BinanceFuturesPositionDetail, 0)

	// Process Open Positions
	for _, p := range openPositions {
		amt, _ := new(big.Float).SetString(p.PositionAmt)
		if amt != nil && amt.Sign() != 0 {
			side := "LONG"
			if amt.Sign() < 0 {
				side = "SHORT"
			}
			if p.PositionSide != "BOTH" && p.PositionSide != "" {
				side = p.PositionSide
			}

			// Calculate PNL % (approximate)
			entry, _ := new(big.Float).SetString(p.EntryPrice)
			mark, _ := new(big.Float).SetString(p.MarkPrice)
			pnlPerc := "0"
			if entry != nil && entry.Sign() != 0 && mark != nil {
				diff := new(big.Float).Sub(mark, entry)
				if side == "SHORT" {
					diff.Neg(diff)
				}
				perc := new(big.Float).Quo(diff, entry)
				perc.Mul(perc, big.NewFloat(100))

				// Factor in leverage
				lev, _ := new(big.Float).SetString(p.Leverage)
				if lev != nil {
					perc.Mul(perc, lev)
				}
				pnlPerc = decimalToString(perc)
			}

			details = append(details, BinanceFuturesPositionDetail{
				Symbol:           p.Symbol,
				Side:             side,
				EntryPrice:       p.EntryPrice,
				UnrealizedPnl:    p.UnRealizedProfit,
				PnlPercentage:    pnlPerc,
				LiquidationPrice: p.LiquidationPrice,
				MarkPrice:        p.MarkPrice,
				Timestamp:        time.Now().Unix(),
				IsClosed:         false,
			})
		}
	}

	// Process Closed Trades (last 10 unique closed events to avoid huge payload)
	processedTrades := 0
	for i := len(recentTrades) - 1; i >= 0 && processedTrades < 10; i-- {
		t := recentTrades[i]
		pnl, _ := new(big.Float).SetString(t.RealizedPnl)
		if pnl != nil && pnl.Sign() != 0 {
			side := t.Side
			if t.PositionSide != "BOTH" && t.PositionSide != "" {
				side = t.PositionSide
			}

			details = append(details, BinanceFuturesPositionDetail{
				Symbol:      t.Symbol,
				Side:        side,
				EntryPrice:  "---", // Hard to reconstruct accurately without more API calls
				ExitPrice:   t.Price,
				RealizedPnl: t.RealizedPnl,
				MarginAsset: t.MarginAsset,
				// Pnl % is tricky for closed trades without knowing margin/entry.
				// For now we'll mark it as closed.
				PnlPercentage: "---",
				MarkPrice:     t.Price,
				Timestamp:     t.Time / 1000,
				IsClosed:      true,
			})
			processedTrades++
		}
	}

	payload := BinanceFuturesDetailsAttestationPayload{
		Source:    "binance-futures-details",
		Positions: details,
		FetchedAt: time.Now().Unix(),
		Version:   Version,
	}

	payloadBytes, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return nil, 0, fmt.Errorf("failed to marshal payload: %v", marshalErr)
	}

	signature, signErr := signViaNode(payloadBytes)
	if signErr != nil {
		return nil, 0, fmt.Errorf("signing failed: %v", signErr)
	}

	encoded, abiErr := abiEncodeTwo(payloadBytes, signature)
	if abiErr != nil {
		return nil, 0, fmt.Errorf("ABI encoding failed: %v", abiErr)
	}

	dataHex := base.BytesToHex(encoded)
	return &dataHex, 1, nil
}

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

func handleBitgetAccountSummary(_ string) (data *string, status int, err error) {
	account, fetchErr := fetchBitgetSpotAccount()
	if fetchErr != nil {
		return nil, 0, fetchErr
	}

	prices, pricesErr := fetchBitgetAllSpotPrices()
	if pricesErr != nil {
		// Fallback to Binance prices if Bitget fails, or just return error
		prices, pricesErr = fetchBinanceAllSpotPrices()
		if pricesErr != nil {
			return nil, 0, pricesErr
		}
	}

	assets := make([]BitgetAssetSummary, 0)
	totalUSDT := new(big.Float)
	unsupported := 0

	for _, asset := range account.Data {
		avail, _ := new(big.Float).SetString(asset.Available)
		frozen, _ := new(big.Float).SetString(asset.Frozen)
		locked, _ := new(big.Float).SetString(asset.Locked)

		if avail == nil {
			avail = new(big.Float)
		}
		if frozen == nil {
			frozen = new(big.Float)
		}
		if locked == nil {
			locked = new(big.Float)
		}

		qty := new(big.Float).Add(avail, frozen)
		qty.Add(qty, locked)

		if qty.Sign() == 0 {
			continue
		}

		estimated := new(big.Float)
		coin := strings.ToUpper(strings.TrimSpace(asset.Coin))
		switch coin {
		case "USDT", "USDC", "BUSD", "FDUSD", "TUSD":
			estimated.Copy(qty)
		default:
			price, found := prices[coin+"USDT"]
			if !found {
				unsupported++
				continue
			}
			estimated.Mul(qty, price)
		}

		totalUSDT.Add(totalUSDT, estimated)
		assets = append(assets, BitgetAssetSummary{
			Asset:         coin,
			Total:         decimalToString(qty),
			EstimatedUSDT: decimalToString(estimated),
		})
	}

	payload := BitgetAccountSummaryAttestationPayload{
		Source:              "bitget-account",
		EstimatedTotalUSDT:  decimalToString(totalUSDT),
		UnsupportedAssetCnt: unsupported,
		Assets:              assets,
		FetchedAt:           time.Now().Unix(),
		Version:             Version,
	}

	payloadBytes, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return nil, 0, fmt.Errorf("failed to marshal bitget account payload: %v", marshalErr)
	}

	log.Printf("[BITGET] Final payload to be signed: %s", string(payloadBytes))

	signature, signErr := signViaNode(payloadBytes)
	if signErr != nil {
		return nil, 0, fmt.Errorf("signing failed: %v", signErr)
	}

	encoded, abiErr := abiEncodeTwo(payloadBytes, signature)
	if abiErr != nil {
		return nil, 0, fmt.Errorf("ABI encoding failed: %v", abiErr)
	}

	dataHex := base.BytesToHex(encoded)
	return &dataHex, 1, nil
}

func fetchBitgetSpotAccount() (*BitgetSpotAccountResponse, error) {
	apiKey := strings.TrimSpace(BitgetAPIKey())
	apiSecret := strings.TrimSpace(BitgetSecretKey())
	passphrase := strings.TrimSpace(BitgetPassphrase())

	if apiKey == "" || apiSecret == "" || passphrase == "" {
		return nil, fmt.Errorf("BITGET_API_KEY, BITGET_SECRET_KEY, and BITGET_PASSPHRASE are required")
	}

	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	method := "GET"
	requestPath := "/api/v2/spot/account/assets"
	
	sign := signBitget(apiSecret, timestamp, method, requestPath, "")

	endpoint := BitgetAPIBaseURL() + requestPath
	req, err := http.NewRequest(method, endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("ACCESS-KEY", apiKey)
	req.Header.Set("ACCESS-SIGN", sign)
	req.Header.Set("ACCESS-TIMESTAMP", timestamp)
	req.Header.Set("ACCESS-PASSPHRASE", passphrase)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bitget returned %d: %s", resp.StatusCode, string(b))
	}

	var res BitgetSpotAccountResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	if res.Code != "00000" {
		return nil, fmt.Errorf("bitget error: %s (msg: %s)", res.Code, res.Msg)
	}

	return &res, nil
}

func fetchBitgetAllSpotPrices() (map[string]*big.Float, error) {
	endpoint := BitgetAPIBaseURL() + "/api/v2/spot/market/tickers"
	resp, err := httpClient.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bitget tickers returned %d", resp.StatusCode)
	}

	var res struct {
		Code string `json:"code"`
		Data []struct {
			Symbol   string `json:"symbol"`
			LastPr   string `json:"lastPr"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	result := make(map[string]*big.Float)
	for _, entry := range res.Data {
		p, ok := new(big.Float).SetString(entry.LastPr)
		if ok {
			result[entry.Symbol] = p
		}
	}
	return result, nil
}

func signBitget(secret, timestamp, method, requestPath, body string) string {
	message := timestamp + method + requestPath + body
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
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
		return nil, err
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

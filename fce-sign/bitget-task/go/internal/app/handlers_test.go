package app

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"sign-extension/internal/base"
)

func setupTestServer(mockNodeURL string) *base.Server {
	// Override the signPort and httpClient for testing.
	parts := strings.Split(mockNodeURL, ":")
	port := parts[len(parts)-1]

	// Save and restore globals.
	origSignPort := signPort
	origClient := httpClient
	signPort = port
	httpClient = http.DefaultClient

	srv := base.New("0", port, Version, Register, ReportState)

	// Restore the original signPort so Register doesn't re-read env.
	signPort = origSignPort
	httpClient = origClient

	// Re-set for this test.
	signPort = port

	return srv
}

func makeActionBody(opType, opCommand, originalMessage string) string {
	df := map[string]interface{}{
		"instructionId":   "0x0000000000000000000000000000000000000000000000000000000000000001",
		"teeId":           "0x0000000000000000000000000000000001",
		"timestamp":       1234567890,
		"opType":          opType,
		"opCommand":       opCommand,
		"originalMessage": originalMessage,
	}
	dfJSON, _ := json.Marshal(df)

	action := map[string]interface{}{
		"data": map[string]interface{}{
			"id":            "0x0000000000000000000000000000000000000000000000000000000000000001",
			"type":          "instruction",
			"submissionTag": "submit",
			"message":       base.BytesToHex(dfJSON),
		},
	}
	body, _ := json.Marshal(action)
	return string(body)
}

func opTypeHex(s string) string {
	return base.VersionToHex(s) // reuse the same stringToBytes32Hex
}

func TestActionKeyUpdateAndSign(t *testing.T) {
	// Reset state.
	privateKey = nil

	privKeyBytes := big.NewInt(12345).FillBytes(make([]byte, 32))

	mockNode := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/decrypt" {
			var req DecryptRequest
			json.NewDecoder(r.Body).Decode(&req)
			json.NewEncoder(w).Encode(DecryptResponse{DecryptedMessage: privKeyBytes})
			return
		}
		http.Error(w, "not found", 404)
	}))
	defer mockNode.Close()

	// Point signPort and httpClient at mock.
	parts := strings.Split(mockNode.URL, ":")
	testPort := parts[len(parts)-1]
	signPort = testPort
	httpClient = http.DefaultClient

	srv := base.New("0", testPort, Version, Register, ReportState)

	// Step 1: Update key.
	updateBody := makeActionBody(opTypeHex("KEY"), opTypeHex("UPDATE"), base.BytesToHex([]byte("encrypteddata")))
	req := httptest.NewRequest(http.MethodPost, "/action", strings.NewReader(updateBody))
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		t.Fatalf("update: status %d, body: %s", resp.StatusCode, body)
	}

	var updateResult base.ActionResult
	if err := json.Unmarshal(body, &updateResult); err != nil {
		t.Fatalf("unmarshal update result: %v", err)
	}
	if updateResult.Status != 1 {
		t.Fatalf("update failed: status=%d log=%v", updateResult.Status, updateResult.Log)
	}

	// Step 2: Sign a message.
	messageHex := base.BytesToHex([]byte("hello"))
	signBody := makeActionBody(opTypeHex("KEY"), opTypeHex("SIGN"), messageHex)
	req = httptest.NewRequest(http.MethodPost, "/action", strings.NewReader(signBody))
	w = httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	resp = w.Result()
	body, _ = io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		t.Fatalf("sign: status %d, body: %s", resp.StatusCode, body)
	}

	var signResult base.ActionResult
	if err := json.Unmarshal(body, &signResult); err != nil {
		t.Fatalf("unmarshal sign result: %v", err)
	}
	if signResult.Status != 1 {
		t.Fatalf("sign failed: status=%d log=%v", signResult.Status, signResult.Log)
	}
	if signResult.Data == nil {
		t.Fatal("sign result data is nil")
	}

	// Decode the ABI-encoded (message, signature).
	dataBytes, err := base.HexToBytes(*signResult.Data)
	if err != nil {
		t.Fatalf("hex decode result data: %v", err)
	}

	msg, sig, err := abiDecodeTwo(dataBytes)
	if err != nil {
		t.Fatalf("abi decode: %v", err)
	}

	if string(msg) != "hello" {
		t.Errorf("message mismatch: got %q, want %q", string(msg), "hello")
	}
	if len(sig) != 65 {
		t.Errorf("expected 65-byte signature, got %d", len(sig))
	}
}

func TestActionSignWithoutKey(t *testing.T) {
	privateKey = nil
	signPort = "9999"

	srv := base.New("0", signPort, Version, Register, ReportState)

	messageHex := base.BytesToHex([]byte("hello"))
	body := makeActionBody(opTypeHex("KEY"), opTypeHex("SIGN"), messageHex)
	req := httptest.NewRequest(http.MethodPost, "/action", strings.NewReader(body))
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	var result base.ActionResult
	json.NewDecoder(w.Result().Body).Decode(&result)
	if result.Status != 0 {
		t.Errorf("expected status 0 (error), got %d", result.Status)
	}
	if result.Log == nil || !strings.Contains(*result.Log, "no private key") {
		t.Errorf("expected 'no private key' error, got %v", result.Log)
	}
}

func TestActionUnknownOperation(t *testing.T) {
	privateKey = nil
	signPort = "9999"

	srv := base.New("0", signPort, Version, Register, ReportState)

	body := makeActionBody(opTypeHex("UNKNOWN"), opTypeHex("OP"), "0xdeadbeef")
	req := httptest.NewRequest(http.MethodPost, "/action", strings.NewReader(body))
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("expected 501, got %d", w.Code)
	}
}

func TestActionUpdateEmptyMessage(t *testing.T) {
	privateKey = nil
	signPort = "9999"

	srv := base.New("0", signPort, Version, Register, ReportState)

	body := makeActionBody(opTypeHex("KEY"), opTypeHex("UPDATE"), "")
	req := httptest.NewRequest(http.MethodPost, "/action", strings.NewReader(body))
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	var result base.ActionResult
	json.NewDecoder(w.Result().Body).Decode(&result)
	if result.Status != 0 {
		t.Errorf("expected status 0, got %d", result.Status)
	}
	if result.Log == nil || !strings.Contains(*result.Log, "originalMessage is empty") {
		t.Errorf("expected 'originalMessage is empty' error, got %v", result.Log)
	}
}

func TestActionMethodNotAllowed(t *testing.T) {
	srv := base.New("0", "9999", Version, Register, ReportState)

	req := httptest.NewRequest(http.MethodGet, "/action", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestStateEndpoint(t *testing.T) {
	privateKey = nil

	srv := base.New("0", "9999", Version, Register, ReportState)

	req := httptest.NewRequest(http.MethodGet, "/state", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	var resp base.StateResponse
	json.NewDecoder(w.Result().Body).Decode(&resp)
	if resp.StateVersion == "" {
		t.Error("stateVersion is empty")
	}

	var state map[string]interface{}
	json.Unmarshal(resp.State, &state)
	if state["hasKey"] != false {
		t.Error("expected hasKey=false")
	}
}

func TestStateMethodNotAllowed(t *testing.T) {
	srv := base.New("0", "9999", Version, Register, ReportState)

	req := httptest.NewRequest(http.MethodPost, "/state", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestActionDecryptionFailure(t *testing.T) {
	privateKey = nil

	mockNode := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"message":"decryption error"}`)
	}))
	defer mockNode.Close()

	parts := strings.Split(mockNode.URL, ":")
	signPort = parts[len(parts)-1]
	httpClient = http.DefaultClient

	srv := base.New("0", signPort, Version, Register, ReportState)

	body := makeActionBody(opTypeHex("KEY"), opTypeHex("UPDATE"), base.BytesToHex([]byte("baddata")))
	req := httptest.NewRequest(http.MethodPost, "/action", strings.NewReader(body))
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	var result base.ActionResult
	json.NewDecoder(w.Result().Body).Decode(&result)
	if result.Status != 0 {
		t.Errorf("expected status 0, got %d", result.Status)
	}
	if result.Log == nil || !strings.Contains(*result.Log, "decryption failed") {
		t.Errorf("expected decryption failure, got %v", result.Log)
	}
}

func TestActionBinanceFetchAndAttest(t *testing.T) {
	privateKey = nil
	lastBinanceSymbol = ""
	lastBinancePrice = ""
	lastBinanceAt = 0

	origBinanceBaseURL := binanceAPIBaseURL
	defer func() {
		binanceAPIBaseURL = origBinanceBaseURL
	}()

	expectedSig := []byte{0x01, 0x02, 0x03, 0x04}

	mockNode := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v3/ticker/price":
			symbol := r.URL.Query().Get("symbol")
			if symbol != "BTCUSDT" {
				http.Error(w, "unknown symbol", http.StatusBadRequest)
				return
			}
			json.NewEncoder(w).Encode(BinanceTickerPriceResponse{Symbol: "BTCUSDT", Price: "65000.12"})
		case "/sign":
			var req SignRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			json.NewEncoder(w).Encode(SignResponse{Message: req.Message, Signature: expectedSig})
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer mockNode.Close()

	parts := strings.Split(mockNode.URL, ":")
	signPort = parts[len(parts)-1]
	httpClient = http.DefaultClient
	binanceAPIBaseURL = mockNode.URL

	srv := base.New("0", signPort, Version, Register, ReportState)

	reqPayload := []byte(`{"symbol":"BTCUSDT"}`)
	body := makeActionBody(opTypeHex("MARKET"), opTypeHex("BINANCE_FETCH_AND_ATTEST"), base.BytesToHex(reqPayload))
	req := httptest.NewRequest(http.MethodPost, "/action", strings.NewReader(body))
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	resp := w.Result()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		t.Fatalf("status %d, body: %s", resp.StatusCode, respBody)
	}

	var result base.ActionResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if result.Status != 1 {
		t.Fatalf("expected success status=1, got %d log=%v", result.Status, result.Log)
	}
	if result.Data == nil {
		t.Fatal("result data is nil")
	}

	encoded, err := base.HexToBytes(*result.Data)
	if err != nil {
		t.Fatalf("hex decode result data: %v", err)
	}

	payloadBytes, sig, err := abiDecodeTwo(encoded)
	if err != nil {
		t.Fatalf("abi decode result data: %v", err)
	}

	if string(sig) != string(expectedSig) {
		t.Fatalf("signature mismatch: got %x want %x", sig, expectedSig)
	}

	var payload BinanceAttestationPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		t.Fatalf("payload json decode: %v", err)
	}

	if payload.Source != "binance" {
		t.Errorf("source mismatch: got %q", payload.Source)
	}
	if payload.Symbol != "BTCUSDT" {
		t.Errorf("symbol mismatch: got %q", payload.Symbol)
	}
	if payload.Price != "65000.12" {
		t.Errorf("price mismatch: got %q", payload.Price)
	}
	if payload.Version != Version {
		t.Errorf("version mismatch: got %q", payload.Version)
	}
	if payload.FetchedAt <= 0 {
		t.Errorf("invalid fetchedAt: %d", payload.FetchedAt)
	}
}

func TestActionBinanceFetchAndAttestSignFailure(t *testing.T) {
	privateKey = nil

	origBinanceBaseURL := binanceAPIBaseURL
	defer func() {
		binanceAPIBaseURL = origBinanceBaseURL
	}()

	mockNode := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v3/ticker/price":
			json.NewEncoder(w).Encode(BinanceTickerPriceResponse{Symbol: "BTCUSDT", Price: "65000.12"})
		case "/sign":
			http.Error(w, `{"message":"sign error"}`, http.StatusInternalServerError)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer mockNode.Close()

	parts := strings.Split(mockNode.URL, ":")
	signPort = parts[len(parts)-1]
	httpClient = http.DefaultClient
	binanceAPIBaseURL = mockNode.URL

	srv := base.New("0", signPort, Version, Register, ReportState)

	reqPayload := []byte(`{"symbol":"BTCUSDT"}`)
	body := makeActionBody(opTypeHex("MARKET"), opTypeHex("BINANCE_FETCH_AND_ATTEST"), base.BytesToHex(reqPayload))
	req := httptest.NewRequest(http.MethodPost, "/action", strings.NewReader(body))
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	var result base.ActionResult
	json.NewDecoder(w.Result().Body).Decode(&result)
	if result.Status != 0 {
		t.Errorf("expected status 0, got %d", result.Status)
	}
	if result.Log == nil || !strings.Contains(*result.Log, "signing failed") {
		t.Errorf("expected signing failed error, got %v", result.Log)
	}
}

func TestActionBinanceAccountPnlAndAttest(t *testing.T) {
	privateKey = nil
	lastBinanceUnrealizedProfit = ""

	origFuturesBaseURL := binanceFuturesAPIBaseURL
	defer func() {
		binanceFuturesAPIBaseURL = origFuturesBaseURL
	}()

	origAPIKey := os.Getenv("BINANCE_API_KEY")
	origAPISecret := os.Getenv("BINANCE_SECRET_KEY")
	defer func() {
		_ = os.Setenv("BINANCE_API_KEY", origAPIKey)
		_ = os.Setenv("BINANCE_SECRET_KEY", origAPISecret)
	}()

	_ = os.Setenv("BINANCE_API_KEY", "test-api-key")
	_ = os.Setenv("BINANCE_SECRET_KEY", "test-secret")

	expectedSig := []byte{0x11, 0x22, 0x33}

	mockNode := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/fapi/v2/account":
			if r.Header.Get("X-MBX-APIKEY") != "test-api-key" {
				http.Error(w, "missing api key", http.StatusUnauthorized)
				return
			}
			if r.URL.Query().Get("signature") == "" {
				http.Error(w, "missing signature", http.StatusUnauthorized)
				return
			}
			json.NewEncoder(w).Encode(BinanceFuturesAccountResponse{
				AccountAlias:          "abc",
				CanTrade:              true,
				TotalWalletBalance:    "100.00",
				TotalUnrealizedProfit: "5.25",
				TotalMarginBalance:    "105.25",
			})
		case "/sign":
			var req SignRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			json.NewEncoder(w).Encode(SignResponse{Message: req.Message, Signature: expectedSig})
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer mockNode.Close()

	parts := strings.Split(mockNode.URL, ":")
	signPort = parts[len(parts)-1]
	httpClient = http.DefaultClient
	binanceFuturesAPIBaseURL = mockNode.URL

	srv := base.New("0", signPort, Version, Register, ReportState)
	body := makeActionBody(opTypeHex("MARKET"), opTypeHex("BINANCE_ACCOUNT_PNL"), base.BytesToHex([]byte("{}")))
	req := httptest.NewRequest(http.MethodPost, "/action", strings.NewReader(body))
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	resp := w.Result()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		t.Fatalf("status %d, body: %s", resp.StatusCode, respBody)
	}

	var result base.ActionResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if result.Status != 1 {
		t.Fatalf("expected success status=1, got %d log=%v", result.Status, result.Log)
	}
	if result.Data == nil {
		t.Fatal("result data is nil")
	}

	encoded, err := base.HexToBytes(*result.Data)
	if err != nil {
		t.Fatalf("hex decode result data: %v", err)
	}

	payloadBytes, sig, err := abiDecodeTwo(encoded)
	if err != nil {
		t.Fatalf("abi decode result data: %v", err)
	}

	if string(sig) != string(expectedSig) {
		t.Fatalf("signature mismatch: got %x want %x", sig, expectedSig)
	}

	var payload BinanceAccountPnlAttestationPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		t.Fatalf("payload json decode: %v", err)
	}

	if payload.Source != "binance-futures" {
		t.Errorf("source mismatch: got %q", payload.Source)
	}
	if payload.TotalUnrealizedProfit != "5.25" {
		t.Errorf("unrealized pnl mismatch: got %q", payload.TotalUnrealizedProfit)
	}
	if payload.TotalWalletBalance != "100.00" {
		t.Errorf("wallet balance mismatch: got %q", payload.TotalWalletBalance)
	}
}

func TestActionBinanceAccountPnlMissingCredentials(t *testing.T) {
	privateKey = nil

	origAPIKey := os.Getenv("BINANCE_API_KEY")
	origAPISecret := os.Getenv("BINANCE_SECRET_KEY")
	defer func() {
		_ = os.Setenv("BINANCE_API_KEY", origAPIKey)
		_ = os.Setenv("BINANCE_SECRET_KEY", origAPISecret)
	}()

	_ = os.Unsetenv("BINANCE_API_KEY")
	_ = os.Unsetenv("BINANCE_SECRET_KEY")

	srv := base.New("0", "9999", Version, Register, ReportState)
	body := makeActionBody(opTypeHex("MARKET"), opTypeHex("BINANCE_ACCOUNT_PNL"), base.BytesToHex([]byte("{}")))
	req := httptest.NewRequest(http.MethodPost, "/action", strings.NewReader(body))
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	var result base.ActionResult
	json.NewDecoder(w.Result().Body).Decode(&result)
	if result.Status != 0 {
		t.Errorf("expected status 0, got %d", result.Status)
	}
	if result.Log == nil || !strings.Contains(*result.Log, "BINANCE_API_KEY and BINANCE_SECRET_KEY are required") {
		t.Errorf("expected missing credentials error, got %v", result.Log)
	}
}

func TestActionBinance24hStatsAndAttest(t *testing.T) {
	privateKey = nil

	origSpotBaseURL := binanceAPIBaseURL
	defer func() {
		binanceAPIBaseURL = origSpotBaseURL
	}()

	expectedSig := []byte{0x9a, 0xbc, 0xde}

	mockNode := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v3/ticker/24hr":
			symbol := r.URL.Query().Get("symbol")
			if symbol != "BTCUSDT" {
				http.Error(w, "unknown symbol", http.StatusBadRequest)
				return
			}
			json.NewEncoder(w).Encode(Binance24hrTickerResponse{
				Symbol:             "BTCUSDT",
				LastPrice:          "67000.01",
				PriceChangePercent: "2.15",
				Volume:             "12345.67",
				QuoteVolume:        "800000000.12",
				OpenTime:           1712100000000,
				CloseTime:          1712186400000,
			})
		case "/sign":
			var req SignRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			json.NewEncoder(w).Encode(SignResponse{Message: req.Message, Signature: expectedSig})
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer mockNode.Close()

	parts := strings.Split(mockNode.URL, ":")
	signPort = parts[len(parts)-1]
	httpClient = http.DefaultClient
	binanceAPIBaseURL = mockNode.URL

	srv := base.New("0", signPort, Version, Register, ReportState)
	reqPayload := []byte(`{"symbol":"BTCUSDT"}`)
	body := makeActionBody(opTypeHex("MARKET"), opTypeHex("BINANCE_24H_STATS"), base.BytesToHex(reqPayload))
	req := httptest.NewRequest(http.MethodPost, "/action", strings.NewReader(body))
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	resp := w.Result()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		t.Fatalf("status %d, body: %s", resp.StatusCode, respBody)
	}

	var result base.ActionResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if result.Status != 1 {
		t.Fatalf("expected success status=1, got %d log=%v", result.Status, result.Log)
	}
	if result.Data == nil {
		t.Fatal("result data is nil")
	}

	encoded, err := base.HexToBytes(*result.Data)
	if err != nil {
		t.Fatalf("hex decode result data: %v", err)
	}

	payloadBytes, sig, err := abiDecodeTwo(encoded)
	if err != nil {
		t.Fatalf("abi decode result data: %v", err)
	}

	if string(sig) != string(expectedSig) {
		t.Fatalf("signature mismatch: got %x want %x", sig, expectedSig)
	}

	var payload Binance24hStatsAttestationPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		t.Fatalf("payload json decode: %v", err)
	}

	if payload.Source != "binance-24h" {
		t.Errorf("source mismatch: got %q", payload.Source)
	}
	if payload.Symbol != "BTCUSDT" {
		t.Errorf("symbol mismatch: got %q", payload.Symbol)
	}
	if payload.LastPrice != "67000.01" {
		t.Errorf("lastPrice mismatch: got %q", payload.LastPrice)
	}
}

func TestActionBinanceAccountSummaryAndAttest(t *testing.T) {
	privateKey = nil
	lastBinanceEstimatedTotalUSDT = ""

	origSpotBaseURL := binanceAPIBaseURL
	defer func() {
		binanceAPIBaseURL = origSpotBaseURL
	}()

	origAPIKey := os.Getenv("BINANCE_API_KEY")
	origAPISecret := os.Getenv("BINANCE_SECRET_KEY")
	defer func() {
		_ = os.Setenv("BINANCE_API_KEY", origAPIKey)
		_ = os.Setenv("BINANCE_SECRET_KEY", origAPISecret)
	}()

	_ = os.Setenv("BINANCE_API_KEY", "test-api-key")
	_ = os.Setenv("BINANCE_SECRET_KEY", "test-secret")

	expectedSig := []byte{0xaa, 0xbb, 0xcc}

	mockNode := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v3/account":
			if r.Header.Get("X-MBX-APIKEY") != "test-api-key" {
				http.Error(w, "missing api key", http.StatusUnauthorized)
				return
			}
			json.NewEncoder(w).Encode(BinanceSpotAccountResponse{
				CanTrade:    true,
				CanDeposit:  true,
				CanWithdraw: true,
				Balances: []BinanceSpotBalance{
					{Asset: "USDT", Free: "100.0", Locked: "0"},
					{Asset: "BTC", Free: "0.1", Locked: "0"},
				},
			})
		case "/api/v3/ticker/price":
			json.NewEncoder(w).Encode([]BinanceTickerPriceEntry{
				{Symbol: "BTCUSDT", Price: "65000"},
			})
		case "/sign":
			var req SignRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			json.NewEncoder(w).Encode(SignResponse{Message: req.Message, Signature: expectedSig})
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer mockNode.Close()

	parts := strings.Split(mockNode.URL, ":")
	signPort = parts[len(parts)-1]
	httpClient = http.DefaultClient
	binanceAPIBaseURL = mockNode.URL

	srv := base.New("0", signPort, Version, Register, ReportState)
	body := makeActionBody(opTypeHex("MARKET"), opTypeHex("BINANCE_ACCOUNT_SUMMARY"), base.BytesToHex([]byte("{}")))
	req := httptest.NewRequest(http.MethodPost, "/action", strings.NewReader(body))
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	resp := w.Result()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		t.Fatalf("status %d, body: %s", resp.StatusCode, respBody)
	}

	var result base.ActionResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if result.Status != 1 {
		t.Fatalf("expected success status=1, got %d log=%v", result.Status, result.Log)
	}
	if result.Data == nil {
		t.Fatal("result data is nil")
	}

	encoded, err := base.HexToBytes(*result.Data)
	if err != nil {
		t.Fatalf("hex decode result data: %v", err)
	}

	payloadBytes, sig, err := abiDecodeTwo(encoded)
	if err != nil {
		t.Fatalf("abi decode result data: %v", err)
	}

	if string(sig) != string(expectedSig) {
		t.Fatalf("signature mismatch: got %x want %x", sig, expectedSig)
	}

	var payload BinanceAccountSummaryAttestationPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		t.Fatalf("payload json decode: %v", err)
	}

	if payload.Source != "binance-account" {
		t.Errorf("source mismatch: got %q", payload.Source)
	}
	if payload.EstimatedTotalUSDT == "" {
		t.Errorf("estimated total usdt should not be empty")
	}
	if len(payload.Assets) == 0 {
		t.Errorf("assets should not be empty")
	}
}

func TestActionBinanceFuturesDetails(t *testing.T) {
	privateKey = nil

	origFuturesBaseURL := binanceFuturesAPIBaseURL
	defer func() {
		binanceFuturesAPIBaseURL = origFuturesBaseURL
	}()

	origAPIKey := os.Getenv("BINANCE_API_KEY")
	origAPISecret := os.Getenv("BINANCE_SECRET_KEY")
	defer func() {
		_ = os.Setenv("BINANCE_API_KEY", origAPIKey)
		_ = os.Setenv("BINANCE_SECRET_KEY", origAPISecret)
	}()

	_ = os.Setenv("BINANCE_API_KEY", "test-api-key")
	_ = os.Setenv("BINANCE_SECRET_KEY", "test-secret")

	expectedSig := []byte{0x55, 0x66, 0x77}

	mockNode := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/fapi/v2/positionRisk":
			json.NewEncoder(w).Encode([]BinanceFuturesPositionRiskResponse{
				{
					Symbol:           "BTCUSDT",
					PositionAmt:      "0.1",
					EntryPrice:       "60000",
					MarkPrice:        "61000",
					UnRealizedProfit: "100",
					LiquidationPrice: "40000",
					Leverage:         "10",
				},
			})
		case "/fapi/v1/userTrades":
			json.NewEncoder(w).Encode([]BinanceFuturesTradeResponse{
				{
					Symbol:      "ETHUSDT",
					Side:        "SELL",
					Price:       "3500",
					RealizedPnl: "50",
					Time:        1712232000000,
					MarginAsset: "USDT",
				},
			})
		case "/sign":
			var req SignRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			json.NewEncoder(w).Encode(SignResponse{Message: req.Message, Signature: expectedSig})
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer mockNode.Close()

	parts := strings.Split(mockNode.URL, ":")
	signPort = parts[len(parts)-1]
	httpClient = http.DefaultClient
	binanceFuturesAPIBaseURL = mockNode.URL

	srv := base.New("0", signPort, Version, Register, ReportState)
	body := makeActionBody(opTypeHex("MARKET"), opTypeHex("BINANCE_FUTURES_DETAILS"), base.BytesToHex([]byte("{}")))
	req := httptest.NewRequest(http.MethodPost, "/action", strings.NewReader(body))
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	resp := w.Result()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		t.Fatalf("status %d, body: %s", resp.StatusCode, respBody)
	}

	var result base.ActionResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if result.Status != 1 {
		t.Fatalf("expected success status=1, got %d log=%v", result.Status, result.Log)
	}
	if result.Data == nil {
		t.Fatal("result data is nil")
	}

	encoded, err := base.HexToBytes(*result.Data)
	if err != nil {
		t.Fatalf("hex decode result data: %v", err)
	}

	payloadBytes, sig, err := abiDecodeTwo(encoded)
	if err != nil {
		t.Fatalf("abi decode result data: %v", err)
	}

	if string(sig) != string(expectedSig) {
		t.Fatalf("signature mismatch: got %x want %x", sig, expectedSig)
	}

	var payload BinanceFuturesDetailsAttestationPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		t.Fatalf("payload json decode: %v", err)
	}

	if payload.Source != "binance-futures-details" {
		t.Errorf("source mismatch: got %q", payload.Source)
	}
	if len(payload.Positions) == 0 {
		t.Errorf("positions should not be empty")
	}

	// Verify one open and one closed position
	hasOpen := false
	hasClosed := false
	for _, p := range payload.Positions {
		if p.IsClosed {
			hasClosed = true
			if p.Symbol != "ETHUSDT" {
				t.Errorf("closed symbol mismatch: %s", p.Symbol)
			}
		} else {
			hasOpen = true
			if p.Symbol != "BTCUSDT" {
				t.Errorf("open symbol mismatch: %s", p.Symbol)
			}
		}
	}
	if !hasOpen || !hasClosed {
		t.Errorf("expected both open and closed positions, got open=%v closed=%v", hasOpen, hasClosed)
	}
}

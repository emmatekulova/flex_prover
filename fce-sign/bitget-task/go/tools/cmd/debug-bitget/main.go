package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load("../../.env")

	apiKey := os.Getenv("BITGET_API_KEY")
	apiSecret := os.Getenv("BITGET_SECRET_KEY")
	passphrase := os.Getenv("BITGET_PASSPHRASE")

	if apiKey == "" || apiSecret == "" || passphrase == "" {
		log.Fatal("ERROR: Missing API credentials in .env")
	}

	fmt.Println("🔍 Bitget Deep Scan...")

	// 1. Try Spot Assets
	tryEndpoint(apiKey, apiSecret, passphrase, "SPOT", "/api/v2/spot/account/assets")

	// 2. Try Funding Assets
	tryEndpoint(apiKey, apiSecret, passphrase, "FUNDING", "/api/v2/account/funding-assets")

	// 3. Try All Account Balance (Overview)
	tryEndpoint(apiKey, apiSecret, passphrase, "OVERVIEW", "/api/v2/account/all-account-balance")

	// 4. Try Futures (USDT-M)
	tryEndpoint(apiKey, apiSecret, passphrase, "FUTURES (USDT-M)", "/api/v2/mix/account/accounts?productType=USDT-FUTURES")
}

func tryEndpoint(key, secret, pass, label, path string) {
	fmt.Printf("\n--- Testing %s (%s) ---\n", label, path)
	
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	method := "GET"
	
	sig := signBitget(secret, timestamp, method, path, "")

	url := "https://api.bitget.com" + path
	req, _ := http.NewRequest(method, url, nil)
	req.Header.Set("ACCESS-KEY", key)
	req.Header.Set("ACCESS-SIGN", sig)
	req.Header.Set("ACCESS-TIMESTAMP", timestamp)
	req.Header.Set("ACCESS-PASSPHRASE", pass)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Connection Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Response: %s\n", string(body))
}

func signBitget(secret, timestamp, method, path, body string) string {
	message := timestamp + method + path + body
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

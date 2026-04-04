package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
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

	fmt.Println("🚀 Fetching Bitget Futures Deep Data (USDT-M)...")

	// 1. Open Positions
	fmt.Println("\n--- [1/3] OPEN POSITIONS ---")
	fetchAndPrint(apiKey, apiSecret, passphrase, "/api/v2/mix/position/all-position", "?productType=USDT-FUTURES")

	// 2. Closed Positions History (Summarized)
	// This shows "singular closed positions" with PnL
	fmt.Println("\n--- [2/3] CLOSED POSITIONS HISTORY ---")
	fetchAndPrint(apiKey, apiSecret, passphrase, "/api/v2/mix/position/history-position", "?productType=USDT-FUTURES&limit=10")

	// 3. Recent Fills (Individual trade executions)
	fmt.Println("\n--- [3/3] RECENT TRADE FILLS ---")
	fetchAndPrint(apiKey, apiSecret, passphrase, "/api/v2/mix/order/fills", "?productType=USDT-FUTURES&limit=10")
}

func fetchAndPrint(key, secret, pass, path, query string) {
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	method := "GET"
	
	message := timestamp + method + path + query
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	url := "https://api.bitget.com" + path + query
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

	bodyBytes, _ := io.ReadAll(resp.Body)
	
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, bodyBytes, "", "  "); err == nil {
		fmt.Println(prettyJSON.String())
	} else {
		fmt.Println(string(bodyBytes))
	}
}

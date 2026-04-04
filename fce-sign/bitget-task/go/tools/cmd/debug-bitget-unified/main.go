package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type FuturesDetailPanel struct {
	ID               string  `json:"id"`
	Status           string  `json:"status"` // "Open" or "Closed"
	Symbol           string  `json:"symbol"`
	Direction        string  `json:"direction"`
	Leverage         *string `json:"leverage,omitempty"`
	SizeUsdt         string  `json:"sizeUsdt"`
	EntryPrice       string  `json:"entryPrice"`
	ExitPrice        string  `json:"exitPrice"`
	Duration         string  `json:"duration"`
	Timestamp        int64   `json:"timestamp"`
	AbsolutePnl      string  `json:"absolutePnl"`
	RelativePnl      string  `json:"relativePnl"`
	CurrentPrice     string  `json:"currentPrice"`
	LiquidationPrice string  `json:"liquidationPrice,omitempty"`
	IsClosed         bool    `json:"isClosed"`
}

func main() {
	_ = godotenv.Load("../../.env")
	apiKey := os.Getenv("BITGET_API_KEY")
	apiSecret := os.Getenv("BITGET_SECRET_KEY")
	passphrase := os.Getenv("BITGET_PASSPHRASE")

	if apiKey == "" { log.Fatal("ERROR: Missing API credentials") }

	results := []FuturesDetailPanel{}
	tickers, _ := fetchTickers()

	// 1. Open Positions
	openRaw, _ := fetchPrivate(apiKey, apiSecret, passphrase, "/api/v2/mix/position/all-position", "?productType=USDT-FUTURES")
	var openRes struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(openRaw, &openRes); err == nil {
		for _, p := range openRes.Data {
			if getFloat(p, "total") == 0 { continue }
			symbol := getString(p, "symbol")
			cTime, _ := strconv.ParseInt(getString(p, "cTime"), 10, 64)
			
			qty := getFloat(p, "total")
			entry := getFloat(p, "openPriceAvg")
			sizeUsdt := qty * entry
			lev := getString(p, "leverage") + "x"

			posId := getString(p, "positionId")
			if posId == "" || posId == "0" {
				posId = getString(p, "posId")
			}

			results = append(results, FuturesDetailPanel{
				ID:               posId,
				Status:           "Open",
				Symbol:           symbol,
				Direction:        strings.Title(getString(p, "holdSide")),
				Leverage:         &lev,
				SizeUsdt:         fmt.Sprintf("%.2f", sizeUsdt),
				EntryPrice:       getString(p, "openPriceAvg"),
				ExitPrice:        "Active",
				Duration:         formatDuration(time.Since(time.UnixMilli(cTime))),
				Timestamp:        cTime / 1000,
				AbsolutePnl:      getString(p, "unrealizedPL"),
				RelativePnl:      fmt.Sprintf("%.2f%%", getFloat(p, "pnlRate")*100),
				CurrentPrice:     tickers[symbol],
				LiquidationPrice: getString(p, "liquidationPrice"),
				IsClosed:         false,
			})
		}
	}

	// 2. Closed History
	historyRaw, _ := fetchPrivate(apiKey, apiSecret, passphrase, "/api/v2/mix/position/history-position", "?productType=USDT-FUTURES&limit=20")
	var historyRes struct {
		Data struct {
			List []map[string]interface{} `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(historyRaw, &historyRes); err == nil {
		for _, p := range historyRes.Data.List {
			symbol := getString(p, "symbol")
			ct, _ := strconv.ParseInt(getString(p, "ctime"), 10, 64)
			ut, _ := strconv.ParseInt(getString(p, "utime"), 10, 64)
			
			qty := getFloat(p, "openTotalPos")
			entry := getFloat(p, "openAvgPrice")
			exit := getFloat(p, "closeAvgPrice")
			sizeUsdt := qty * entry
			
			side := getString(p, "holdSide")
			relPnl := 0.0
			if entry > 0 {
				if strings.ToLower(side) == "long" {
					relPnl = (exit - entry) / entry * 100
				} else {
					relPnl = (entry - exit) / entry * 100
				}
			}

			results = append(results, FuturesDetailPanel{
				ID:           getString(p, "positionId"),
				Status:       "Closed",
				Symbol:       symbol,
				Direction:    strings.Title(side),
				Leverage:     nil,
				SizeUsdt:     fmt.Sprintf("%.2f", sizeUsdt),
				EntryPrice:   getString(p, "openAvgPrice"),
				ExitPrice:    getString(p, "closeAvgPrice"),
				Duration:     formatDuration(time.UnixMilli(ut).Sub(time.UnixMilli(ct))),
				Timestamp:    ut / 1000,
				AbsolutePnl:  getString(p, "pnl"),
				RelativePnl:  fmt.Sprintf("%.2f%%", relPnl),
				CurrentPrice: tickers[symbol],
				IsClosed:     true,
			})
		}
	}

	finalJSON, _ := json.MarshalIndent(results, "", "  ")
	fmt.Println(string(finalJSON))
}

func fetchPrivate(key, secret, pass, path, query string) ([]byte, error) {
	ts := fmt.Sprintf("%d", time.Now().UnixMilli())
	sig := signBitget(secret, ts, "GET", path, query, "")
	req, _ := http.NewRequest("GET", "https://api.bitget.com"+path+query, nil)
	req.Header.Set("ACCESS-KEY", key)
	req.Header.Set("ACCESS-SIGN", sig)
	req.Header.Set("ACCESS-TIMESTAMP", ts)
	req.Header.Set("ACCESS-PASSPHRASE", pass)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func fetchTickers() (map[string]string, error) {
	m := make(map[string]string)
	resp, _ := http.Get("https://api.bitget.com/api/v2/mix/market/tickers?productType=USDT-FUTURES")
	var res struct { Data []struct{ Symbol, LastPr string } }
	json.NewDecoder(resp.Body).Decode(&res)
	for _, d := range res.Data { m[d.Symbol] = d.LastPr }
	return m, nil
}

func signBitget(secret, ts, method, path, query, body string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + method + path + query + body))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil { return fmt.Sprintf("%v", v) }
	return ""
}

func getFloat(m map[string]interface{}, key string) float64 {
	f, _ := strconv.ParseFloat(getString(m, key), 64)
	return f
}

func formatDuration(d time.Duration) string {
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}

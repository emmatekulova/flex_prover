// fetch-positions-binance fetches the authenticated user's current spot holdings
// from Binance and prints them as a JSON object.  This is used by the Next.js
// frontend to populate the position-selection step before the user chooses which
// holdings to include in an individual-trades attestation.
//
// Usage:
//
//	go run ./cmd/fetch-positions-binance \
//	  -apiKey <BINANCE_API_KEY> \
//	  -secretKey <BINANCE_SECRET_KEY>
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type spotBalance struct {
	Asset  string `json:"asset"`
	Free   string `json:"free"`
	Locked string `json:"locked"`
}

type spotAccountResponse struct {
	Balances []spotBalance `json:"balances"`
}

type tickerEntry struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
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

const baseURL = "https://api.binance.com"

func main() {
	_ = godotenv.Load()
	_ = godotenv.Load("../../.env")

	apiKey := flag.String("apiKey", "", "Binance API key (required)")
	secretKey := flag.String("secretKey", "", "Binance secret key (required)")
	flag.Parse()

	if strings.TrimSpace(*apiKey) == "" || strings.TrimSpace(*secretKey) == "" {
		fmt.Fprintln(os.Stderr, "-apiKey and -secretKey are required")
		os.Exit(1)
	}

	account, err := fetchSpotAccount(*apiKey, *secretKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fetch account: %v\n", err)
		os.Exit(1)
	}

	prices, err := fetchAllSpotPrices()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fetch prices: %v\n", err)
		os.Exit(1)
	}

	stablecoins := map[string]struct{}{
		"USDT": {}, "USDC": {}, "BUSD": {}, "FDUSD": {}, "TUSD": {},
	}

	var positions []tradePosition
	for _, bal := range account.Balances {
		free, okF := new(big.Float).SetString(bal.Free)
		locked, okL := new(big.Float).SetString(bal.Locked)
		if !okF {
			free = new(big.Float)
		}
		if !okL {
			locked = new(big.Float)
		}
		qty := new(big.Float).Add(free, locked)
		if qty.Sign() == 0 {
			continue
		}

		asset := strings.ToUpper(bal.Asset)

		var priceUSDT *big.Float
		if _, isStable := stablecoins[asset]; isStable {
			priceUSDT = big.NewFloat(1)
		} else {
			p, ok := prices[asset+"USDT"]
			if !ok {
				continue
			}
			priceUSDT = p
		}

		valueUSDT := new(big.Float).Mul(qty, priceUSDT)
		positions = append(positions, tradePosition{
			Asset:     asset,
			Quantity:  qty.Text('f', 8),
			PriceUSDT: priceUSDT.Text('f', 6),
			ValueUSDT: valueUSDT.Text('f', 2),
		})
	}

	out := positionsOutput{
		Exchange:  "binance",
		Positions: positions,
		FetchedAt: time.Now().Unix(),
	}

	enc := json.NewEncoder(os.Stdout)
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "encode output: %v\n", err)
		os.Exit(1)
	}
}

func fetchSpotAccount(apiKey, secretKey string) (*spotAccountResponse, error) {
	timestamp := time.Now().UnixMilli()
	query := fmt.Sprintf("timestamp=%d&recvWindow=5000", timestamp)
	sig := signQuery(secretKey, query)
	endpoint := fmt.Sprintf("%s/api/v3/account?%s&signature=%s", baseURL, query, sig)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-MBX-APIKEY", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("binance account returned %d: %s", resp.StatusCode, string(b))
	}

	var account spotAccountResponse
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return nil, err
	}
	return &account, nil
}

func fetchAllSpotPrices() (map[string]*big.Float, error) {
	resp, err := http.Get(baseURL + "/api/v3/ticker/price")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("binance prices returned %d: %s", resp.StatusCode, string(b))
	}

	var entries []tickerEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, err
	}

	result := make(map[string]*big.Float, len(entries))
	for _, e := range entries {
		p, ok := new(big.Float).SetString(e.Price)
		if ok {
			result[e.Symbol] = p
		}
	}
	return result, nil
}

func signQuery(secret, query string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(query))
	return fmt.Sprintf("%x", mac.Sum(nil))
}

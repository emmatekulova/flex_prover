package app

// DecryptRequest is sent to the TEE node's /decrypt endpoint.
// EncryptedMessage is []byte so it JSON-marshals as base64, matching the tee-node's
// DecryptRequest which also uses []byte.
type DecryptRequest struct {
	EncryptedMessage []byte `json:"encryptedMessage"`
}

// DecryptResponse is returned from the TEE node's /decrypt endpoint.
// DecryptedMessage is []byte so it JSON-unmarshals from base64, matching the tee-node's
// DecryptResponse which also uses []byte.
type DecryptResponse struct {
	DecryptedMessage []byte `json:"decryptedMessage"`
}

// SignRequest is sent to the TEE node's /sign endpoint.
// Message is []byte so it JSON-marshals as base64.
type SignRequest struct {
	Message []byte `json:"message"`
}

// SignResponse is returned from the TEE node's /sign endpoint.
// Message and Signature are []byte so they JSON-unmarshal from base64.
type SignResponse struct {
	Message   []byte `json:"message"`
	Signature []byte `json:"signature"`
}

// CEXRequest is the unified message payload for all MARKET handlers.
// Credentials must be encrypted with the TEE node's public key (ECIES) and
// hex-encoded. The TEE decrypts them internally; they are never stored or logged.
// Public endpoints (ticker, 24h stats) do not require encryptedCredentials.
type CEXRequest struct {
	CEX                  string `json:"cex"`                            // e.g. "binance", "bybit"
	EncryptedCredentials string `json:"encryptedCredentials,omitempty"` // hex-encoded ECIES ciphertext of CEXCredentials JSON
	Symbol               string `json:"symbol,omitempty"`               // for ticker/stats requests
}

// CEXCredentials holds plaintext credentials, used only inside the TEE after decryption.
type CEXCredentials struct {
	APIKey    string `json:"apiKey"`
	SecretKey string `json:"secretKey"`
}

// BinanceFetchRequest is kept for reference; use CEXRequest for all handlers.
type BinanceFetchRequest struct {
	Symbol string `json:"symbol"`
}

// BinanceTickerPriceResponse matches Binance /api/v3/ticker/price response.
type BinanceTickerPriceResponse struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

// BinanceAttestationPayload is what the TEE signs and returns to callers.
type BinanceAttestationPayload struct {
	Source    string `json:"source"`
	Symbol    string `json:"symbol"`
	Price     string `json:"price"`
	FetchedAt int64  `json:"fetchedAt"`
	Version   string `json:"version"`
}

// Binance24hrTickerResponse matches key fields from Binance 24h ticker endpoint.
type Binance24hrTickerResponse struct {
	Symbol             string `json:"symbol"`
	LastPrice          string `json:"lastPrice"`
	PriceChangePercent string `json:"priceChangePercent"`
	Volume             string `json:"volume"`
	QuoteVolume        string `json:"quoteVolume"`
	OpenTime           int64  `json:"openTime"`
	CloseTime          int64  `json:"closeTime"`
}

// Binance24hStatsAttestationPayload is what the TEE signs for public 24h market stats.
type Binance24hStatsAttestationPayload struct {
	Source             string `json:"source"`
	Symbol             string `json:"symbol"`
	LastPrice          string `json:"lastPrice"`
	PriceChangePercent string `json:"priceChangePercent"`
	Volume             string `json:"volume"`
	QuoteVolume        string `json:"quoteVolume"`
	OpenTime           int64  `json:"openTime"`
	CloseTime          int64  `json:"closeTime"`
	FetchedAt          int64  `json:"fetchedAt"`
	Version            string `json:"version"`
}

type BinanceSpotBalance struct {
	Asset  string `json:"asset"`
	Free   string `json:"free"`
	Locked string `json:"locked"`
}

type BinanceSpotAccountResponse struct {
	UID         int64                `json:"uid"`
	AccountType string               `json:"accountType"`
	Permissions []string             `json:"permissions"`
	CanTrade    bool                 `json:"canTrade"`
	CanDeposit  bool                 `json:"canDeposit"`
	CanWithdraw bool                 `json:"canWithdraw"`
	Balances    []BinanceSpotBalance `json:"balances"`
}

type BinanceTickerPriceEntry struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

type BinanceAccountAssetSummary struct {
	Asset         string `json:"asset"`
	Total         string `json:"total"`
	EstimatedUSDT string `json:"estimatedUsdt"`
}

type BinanceAccountSummaryAttestationPayload struct {
	Source              string                      `json:"source"`
	CanTrade            bool                        `json:"canTrade"`
	CanDeposit          bool                        `json:"canDeposit"`
	CanWithdraw         bool                        `json:"canWithdraw"`
	EstimatedTotalUSDT  string                      `json:"estimatedTotalUsdt"`
	UnsupportedAssetCnt int                         `json:"unsupportedAssetCount"`
	Assets              []BinanceAccountAssetSummary `json:"assets"`
	FetchedAt           int64                       `json:"fetchedAt"`
	Version             string                      `json:"version"`
}

// BinanceUserProfileAttestationPayload is what the TEE signs for the enriched user profile.
// It includes account identity (UID, type, permissions) and portfolio snapshot.
type BinanceUserProfileAttestationPayload struct {
	Source              string                       `json:"source"`
	UID                 int64                        `json:"uid"`
	AccountType         string                       `json:"accountType"`
	Permissions         []string                     `json:"permissions"`
	CanTrade            bool                         `json:"canTrade"`
	CanDeposit          bool                         `json:"canDeposit"`
	CanWithdraw         bool                         `json:"canWithdraw"`
	EstimatedTotalUSDT  string                       `json:"estimatedTotalUsdt"`
	UnsupportedAssetCnt int                          `json:"unsupportedAssetCount"`
	Assets              []BinanceAccountAssetSummary `json:"assets"`
	FetchedAt           int64                        `json:"fetchedAt"`
	Version             string                       `json:"version"`
}

// BinanceFuturesAccountResponse matches key fields from Binance USD-M futures account endpoint.
type BinanceFuturesAccountResponse struct {
	AccountAlias          string `json:"accountAlias"`
	CanTrade              bool   `json:"canTrade"`
	TotalWalletBalance    string `json:"totalWalletBalance"`
	TotalUnrealizedProfit string `json:"totalUnrealizedProfit"`
	TotalMarginBalance    string `json:"totalMarginBalance"`
}

// BinanceAccountPnlAttestationPayload is what the TEE signs for authenticated account PnL.
type BinanceAccountPnlAttestationPayload struct {
	Source                string `json:"source"`
	AccountAlias          string `json:"accountAlias"`
	CanTrade              bool   `json:"canTrade"`
	TotalWalletBalance    string `json:"totalWalletBalance"`
	TotalUnrealizedProfit string `json:"totalUnrealizedProfit"`
	TotalMarginBalance    string `json:"totalMarginBalance"`
	FetchedAt             int64  `json:"fetchedAt"`
	Version               string `json:"version"`
}

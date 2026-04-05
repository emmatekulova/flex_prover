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

// BinanceFetchRequest is the originalMessage payload for Binance attestation.
// The message bytes should be a JSON object: {"symbol":"BTCUSDT"}.
type BinanceFetchRequest struct {
	Symbol string `json:"symbol"`
}

// BinanceAuthenticatedRequest carries optional Binance credentials for
// authenticated account-style handlers.
type BinanceAuthenticatedRequest struct {
	APIKey    string `json:"apiKey"`
	SecretKey string `json:"secretKey"`
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

// BinanceProfileGrowthRequest is the originalMessage payload for growth attestation.
type BinanceProfileGrowthRequest struct {
	APIKey     string `json:"apiKey"`
	SecretKey  string `json:"secretKey"`
	Wallet     string `json:"wallet"`     // connected on-chain wallet address
	WindowDays int    `json:"windowDays"` // 7 or 30; handler defaults to 7 if ≤ 0
}

// BinanceSnapshotPoint is one daily portfolio snapshot.
type BinanceSnapshotPoint struct {
	Date     string `json:"date"`     // "YYYY-MM-DD"
	TotalBTC string `json:"totalBtc"` // totalAssetOfBtc from Binance
}

// BinanceProfileGrowthPayload is what the TEE signs for the growth attestation.
// Financial detail is kept minimal — only growthPercent is included.
type BinanceProfileGrowthPayload struct {
	Source        string               `json:"source"`        // "binance-profile-growth"
	Wallet        string               `json:"wallet"`
	WindowDays    int                  `json:"windowDays"`
	StartSnapshot BinanceSnapshotPoint `json:"startSnapshot"`
	EndSnapshot   BinanceSnapshotPoint `json:"endSnapshot"`
	GrowthPercent string               `json:"growthPercent"`
	FetchedAt     int64                `json:"fetchedAt"`
	Version       string               `json:"version"`
}

// BinanceAccountSnapshotResponse is returned by GET /sapi/v1/accountSnapshot.
type BinanceAccountSnapshotResponse struct {
	Code        int                 `json:"code"`
	SnapshotVos []BinanceSnapshotVo `json:"snapshotVos"`
}

// BinanceSnapshotVo is one element in the snapshotVos array.
type BinanceSnapshotVo struct {
	Data       BinanceSnapshotData `json:"data"`
	UpdateTime int64               `json:"updateTime"` // millisecond epoch
}

// BinanceSnapshotData holds the portfolio value for one snapshot.
type BinanceSnapshotData struct {
	TotalAssetOfBtc string `json:"totalAssetOfBtc"`
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

// ── Bitget types ──────────────────────────────────────────────────────────────

// BitgetProfileGrowthRequest is the originalMessage payload for Bitget attestation.
type BitgetProfileGrowthRequest struct {
	APIKey     string `json:"apiKey"`
	SecretKey  string `json:"secretKey"`
	Passphrase string `json:"passphrase"`
	Wallet     string `json:"wallet"`
	WindowDays int    `json:"windowDays"`
}

// BitgetSnapshotPoint is a single portfolio value snapshot for Bitget.
type BitgetSnapshotPoint struct {
	Date      string `json:"date"`
	TotalUSDT string `json:"totalUsdt"`
}

// BitgetProfileGrowthPayload is what the TEE signs for a Bitget attestation.
// Field names match BinanceProfileGrowthPayload so the shared route handler parses both identically.
type BitgetProfileGrowthPayload struct {
	Source        string              `json:"source"`        // "bitget-profile-growth"
	Wallet        string              `json:"wallet"`
	WindowDays    int                 `json:"windowDays"`
	StartSnapshot BitgetSnapshotPoint `json:"startSnapshot"`
	EndSnapshot   BitgetSnapshotPoint `json:"endSnapshot"`
	GrowthPercent string              `json:"growthPercent"` // "0.00" — Bitget has no snapshot history API
	TotalUSDT     string              `json:"totalUsdt"`
	FetchedAt     int64               `json:"fetchedAt"`
	Version       string              `json:"version"`
}

// BitgetAsset is one entry from GET /api/v2/spot/account/assets.
type BitgetAsset struct {
	CoinName  string `json:"coinName"`
	Available string `json:"available"`
	Frozen    string `json:"frozen"`
	Locked    string `json:"locked"`
}

// BitgetAssetsResponse wraps the Bitget assets endpoint response.
type BitgetAssetsResponse struct {
	Code string        `json:"code"`
	Data []BitgetAsset `json:"data"`
}

// BitgetTicker is one entry from GET /api/v2/spot/market/tickers.
type BitgetTicker struct {
	Symbol string `json:"symbol"` // e.g. "BTCUSDT"
	LastPr string `json:"lastPr"` // last traded price
}

// BitgetTickersResponse wraps the Bitget tickers endpoint response.
type BitgetTickersResponse struct {
	Code string         `json:"code"`
	Data []BitgetTicker `json:"data"`
}

// BitgetFuturesAccount is one entry from GET /api/v2/mix/account/accounts.
type BitgetFuturesAccount struct {
	MarginCoin  string `json:"marginCoin"` // e.g. "USDT", "BTC"
	AccountEquity string `json:"accountEquity"`
	UsdtEquity  string `json:"usdtEquity"`  // USDT-denominated equity (present for all product types)
}

// BitgetFuturesAccountsResponse wraps the Bitget futures accounts endpoint response.
type BitgetFuturesAccountsResponse struct {
	Code string                 `json:"code"`
	Data []BitgetFuturesAccount `json:"data"`
}

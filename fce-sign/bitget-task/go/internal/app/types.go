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

// BinanceFuturesPositionDetail represents a single open or closed futures position.
type BinanceFuturesPositionDetail struct {
	Symbol           string `json:"symbol"`
	Side             string `json:"side"` // LONG, SHORT, or BOTH
	EntryPrice       string `json:"entryPrice"`
	ExitPrice        string `json:"exitPrice,omitempty"`
	RealizedPnl      string `json:"realizedPnl,omitempty"`
	UnrealizedPnl    string `json:"unrealizedPnl,omitempty"`
	MarginAsset      string `json:"marginAsset"`
	PnlPercentage    string `json:"pnlPercentage"`
	LiquidationPrice string `json:"liquidationPrice,omitempty"`
	MarkPrice        string `json:"markPrice"`
	Duration         string `json:"duration,omitempty"` // e.g. "2h 30m"
	Timestamp        int64  `json:"timestamp"`          // trade time or fetch time
	IsClosed         bool   `json:"isClosed"`
}

// BinanceFuturesDetailsAttestationPayload is the TEE-signed payload for multiple positions.
type BinanceFuturesDetailsAttestationPayload struct {
	Source    string                         `json:"source"`
	Positions []BinanceFuturesPositionDetail `json:"positions"`
	FetchedAt int64                          `json:"fetchedAt"`
	Version   string                         `json:"version"`
}

// BinanceFuturesTradeResponse matches Binance /fapi/v1/userTrades response items.
type BinanceFuturesTradeResponse struct {
	Symbol       string `json:"symbol"`
	ID           int64  `json:"id"`
	OrderID      int64  `json:"orderId"`
	Side         string `json:"side"`
	Price        string `json:"price"`
	Qty          string `json:"qty"`
	RealizedPnl  string `json:"realizedPnl"`
	MarginAsset  string `json:"marginAsset"`
	QuoteQty     string `json:"quoteQty"`
	Commission   string `json:"commission"`
	Time         int64  `json:"time"`
	PositionSide string `json:"positionSide"`
}

// BitgetFuturesPositionRiskResponse matches Binance /fapi/v2/positionRisk response items.
type BinanceFuturesPositionRiskResponse struct {
	Symbol           string `json:"symbol"`
	PositionAmt      string `json:"positionAmt"`
	EntryPrice       string `json:"entryPrice"`
	MarkPrice        string `json:"markPrice"`
	UnRealizedProfit string `json:"unRealizedProfit"`
	LiquidationPrice string `json:"liquidationPrice"`
	Leverage         string `json:"leverage"`
	MarginType       string `json:"marginType"`
	IsIsolated       string `json:"isIsolated"`
	PositionSide     string `json:"positionSide"`
}

type BitgetSpotAsset struct {
	Coin             string `json:"coin"`
	Available        string `json:"available"`
	Frozen           string `json:"frozen"`
	Locked           string `json:"locked"`
	LimitAvailable   string `json:"limitAvailable"`
	UnlimitAvailable string `json:"unlimitAvailable"`
}

type BitgetSpotAccountResponse struct {
	Code        string            `json:"code"`
	Msg         string            `json:"msg"`
	RequestTime int64             `json:"requestTime"`
	Data        []BitgetSpotAsset `json:"data"`
}

type BitgetAssetSummary struct {
	Asset         string `json:"asset"`
	Total         string `json:"total"`
	EstimatedUSDT string `json:"estimatedUsdt"`
}

type BitgetAccountSummaryAttestationPayload struct {
	Source              string               `json:"source"`
	EstimatedTotalUSDT  string               `json:"estimatedTotalUsdt"`
	UnsupportedAssetCnt int                  `json:"unsupportedAssetCount"`
	Assets              []BitgetAssetSummary `json:"assets"`
	FetchedAt           int64                `json:"fetchedAt"`
	Version             string               `json:"version"`
}


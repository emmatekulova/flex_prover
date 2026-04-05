package app

import "os"

// Version is the SemVer version of this extension.
const Version = "0.1.0"

// OPType and OPCommand constants — must match the bytes32 constants in InstructionSender.sol.
const (
	OpTypeKey       = "KEY"
	OpCommandUpdate = "UPDATE"
	OpCommandSign   = "SIGN"

	OpTypeMarket                    = "MARKET"
	OpCommandBinanceFetchAndAttest  = "BINANCE_FETCH_AND_ATTEST"
	OpCommandBinance24hStats        = "BINANCE_24H_STATS"
	OpCommandBinanceAccountPnl      = "BINANCE_ACCOUNT_PNL"
	OpCommandBinanceAccountSummary  = "BINANCE_ACCOUNT_SUMMARY"
	OpCommandBinanceUserProfile     = "BINANCE_USER_PROFILE"
	OpCommandBinanceProfileGrowth   = "BINANCE_PROFILE_GROWTH"

	OpCommandBitgetProfileGrowth = "BITGET_PROFILE_GROWTH"

	OpCommandBinanceIndividualTrades = "BINANCE_INDIVIDUAL_TRADES"
	OpCommandBitgetIndividualTrades  = "BITGET_INDIVIDUAL_TRADES"
)

// BinanceAPIKey returns the Binance API key from environment, if set.
func BinanceAPIKey() string {
	return os.Getenv("BINANCE_API_KEY")
}

// BinanceSecretKey returns the Binance secret key from environment, if set.
func BinanceSecretKey() string {
	return os.Getenv("BINANCE_SECRET_KEY")
}

// BinanceSpotAPIBaseURL returns spot API base URL, defaulting to production.
func BinanceSpotAPIBaseURL() string {
	if v := os.Getenv("BINANCE_SPOT_API_BASE_URL"); v != "" {
		return v
	}
	return "https://api.binance.com"
}

// BinanceFuturesAPIBaseURL returns futures API base URL, defaulting to production.
func BinanceFuturesAPIBaseURL() string {
	if v := os.Getenv("BINANCE_FUTURES_API_BASE_URL"); v != "" {
		return v
	}
	return "https://fapi.binance.com"
}

// BitgetAPIBaseURL returns the Bitget API base URL, defaulting to production.
func BitgetAPIBaseURL() string {
	if v := os.Getenv("BITGET_API_BASE_URL"); v != "" {
		return v
	}
	return "https://api.bitget.com"
}


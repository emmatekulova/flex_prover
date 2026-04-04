package app

import "os"

// Version is the SemVer version of this extension.
const Version = "0.1.0"

// OPType and OPCommand constants — must match the bytes32 constants in InstructionSender.sol.
const (
	OpTypeKey       = "KEY"
	OpCommandUpdate = "UPDATE"
	OpCommandSign   = "SIGN"

	OpTypeMarket = "MARKET"

	// Generic op commands (CEX-agnostic, preferred for new integrations):
	OpCommandFetchAndAttest = "FETCH_AND_ATTEST"
	OpCommand24hStats       = "24H_STATS"
	OpCommandAccountPnl     = "ACCOUNT_PNL"
	OpCommandAccountSummary = "ACCOUNT_SUMMARY"
	OpCommandUserProfile    = "USER_PROFILE"

	// Binance-prefixed aliases kept for backward compatibility with deployed InstructionSender contracts:
	OpCommandBinanceFetchAndAttest = "BINANCE_FETCH_AND_ATTEST"
	OpCommandBinance24hStats       = "BINANCE_24H_STATS"
	OpCommandBinanceAccountPnl     = "BINANCE_ACCOUNT_PNL"
	OpCommandBinanceAccountSummary = "BINANCE_ACCOUNT_SUMMARY"
	OpCommandBinanceUserProfile    = "BINANCE_USER_PROFILE"
)

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

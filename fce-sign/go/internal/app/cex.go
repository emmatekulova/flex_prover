package app

import (
	"encoding/json"
	"fmt"
	"sign-extension/internal/base"
	"strings"
)

// CEXProvider is the interface every exchange adapter must implement.
// Each method receives credentials (may be empty for public endpoints) and
// returns raw JSON bytes of the attestation payload; the handler signs and
// ABI-encodes them.
//
// Security: credentials arrive as plaintext in the TEE-processed message.
// This is intentional — TEE execution is isolated and the Flare protocol
// encrypts messages in transit. Credentials are never stored or logged.
type CEXProvider interface {
	FetchAndAttest(apiKey, secretKey, symbol string) ([]byte, error)
	Fetch24hStats(apiKey, secretKey, symbol string) ([]byte, error)
	FetchAccountPnl(apiKey, secretKey string) ([]byte, error)
	FetchAccountSummary(apiKey, secretKey string) ([]byte, error)
	FetchUserProfile(apiKey, secretKey string) ([]byte, error)
	FetchPortfolioGrowth(apiKey, secretKey string, lookbackDays int) ([]byte, error)
}

// ABIEncoderProvider is an optional interface for providers (like Binance) that
// require structured ABI encoding for on-chain decoding of user profile data.
// When a provider implements this, handleCEXUserProfile calls EncodeUserProfile
// instead of FetchUserProfile.
type ABIEncoderProvider interface {
	CEXProvider
	EncodeUserProfile(apiKey, secretKey string) ([]byte, error)
}

var cexRegistry = map[string]CEXProvider{}

// RegisterCEX registers a CEX provider under a lowercase name.
// Called from each provider's init() function.
func RegisterCEX(name string, p CEXProvider) {
	cexRegistry[strings.ToLower(name)] = p
}

func lookupCEX(name string) (CEXProvider, error) {
	p, ok := cexRegistry[strings.ToLower(strings.TrimSpace(name))]
	if !ok {
		return nil, fmt.Errorf("unknown CEX provider %q", name)
	}
	return p, nil
}

// parseCEXRequest hex-decodes and JSON-unmarshals the handler message into a
// CEXRequest. Returns an error if the message is empty, malformed, or missing
// the required cex field.
func parseCEXRequest(msg string) (*CEXRequest, error) {
	if msg == "" {
		return nil, fmt.Errorf("originalMessage is empty")
	}
	msgBytes, err := base.HexToBytes(msg)
	if err != nil {
		return nil, fmt.Errorf("invalid hex in originalMessage: %v", err)
	}
	var req CEXRequest
	if err := json.Unmarshal(msgBytes, &req); err != nil {
		return nil, fmt.Errorf("invalid request payload: expected JSON {\"cex\":\"...\", ...}")
	}
	req.CEX = strings.ToLower(strings.TrimSpace(req.CEX))
	req.Symbol = strings.ToUpper(strings.TrimSpace(req.Symbol))
	if req.CEX == "" {
		return nil, fmt.Errorf("cex field is required")
	}
	return &req, nil
}

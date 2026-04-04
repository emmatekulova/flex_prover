package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sign-extension/internal/base"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// state holds the mutable state for the extension.
// The framework serializes all handler calls, so no additional locking is needed.
var (
	privateKey *secp256k1.PrivateKey
	signPort   string
	httpClient = http.DefaultClient

	lastBinanceSymbol             string
	lastBinancePrice              string
	lastBinanceAt                 int64
	lastBinanceUnrealizedProfit   string
	lastBinanceEstimatedTotalUSDT string
)

// SetSignPort sets the sign port for communicating with the TEE node.
func SetSignPort(port string) {
	signPort = port
}

// Register registers the handlers and initial state with the framework.
func Register(f *base.Framework) {
	f.Handle(OpTypeKey, OpCommandUpdate, handleKeyUpdate)
	f.Handle(OpTypeKey, OpCommandSign, handleKeySign)

	// Generic op commands (CEX-agnostic):
	f.Handle(OpTypeMarket, OpCommandFetchAndAttest, handleCEXFetchAndAttest)
	f.Handle(OpTypeMarket, OpCommand24hStats, handleCEX24hStats)
	f.Handle(OpTypeMarket, OpCommandAccountPnl, handleCEXAccountPnl)
	f.Handle(OpTypeMarket, OpCommandAccountSummary, handleCEXAccountSummary)
	f.Handle(OpTypeMarket, OpCommandUserProfile, handleCEXUserProfile)

	// Binance-prefixed aliases for backward compatibility with deployed InstructionSender contracts:
	f.Handle(OpTypeMarket, OpCommandBinanceFetchAndAttest, handleCEXFetchAndAttest)
	f.Handle(OpTypeMarket, OpCommandBinance24hStats, handleCEX24hStats)
	f.Handle(OpTypeMarket, OpCommandBinanceAccountPnl, handleCEXAccountPnl)
	f.Handle(OpTypeMarket, OpCommandBinanceAccountSummary, handleCEXAccountSummary)
	f.Handle(OpTypeMarket, OpCommandBinanceUserProfile, handleCEXUserProfile)
}

// ReportState returns a JSON snapshot of the current state.
func ReportState() json.RawMessage {
	hasKey := privateKey != nil
	data, _ := json.Marshal(map[string]interface{}{
		"hasKey":                       hasKey,
		"version":                      Version,
		"lastBinanceSymbol":            lastBinanceSymbol,
		"lastBinancePrice":             lastBinancePrice,
		"lastBinanceAt":                lastBinanceAt,
		"lastBinanceUnrealizedProfit":  lastBinanceUnrealizedProfit,
		"lastBinanceEstimatedTotalUsdt": lastBinanceEstimatedTotalUSDT,
	})
	return data
}

// resolveCredentials decrypts the credentials from the CEXRequest.
// If encryptedCredentials is empty, returns empty strings (valid for public endpoints).
// The decrypted JSON must be: {"apiKey":"...","secretKey":"..."}
func resolveCredentials(req *CEXRequest) (apiKey, secretKey string, err error) {
	if req.EncryptedCredentials == "" {
		return "", "", nil
	}
	ciphertext, err := base.HexToBytes(req.EncryptedCredentials)
	if err != nil {
		return "", "", fmt.Errorf("invalid encryptedCredentials hex: %v", err)
	}
	plaintext, err := decryptViaNode(ciphertext)
	if err != nil {
		return "", "", fmt.Errorf("failed to decrypt credentials: %v", err)
	}
	var creds CEXCredentials
	if err := json.Unmarshal(plaintext, &creds); err != nil {
		return "", "", fmt.Errorf("decrypted credentials are not valid JSON: %v", err)
	}
	return creds.APIKey, creds.SecretKey, nil
}

// signAndEncode signs payloadBytes via the TEE node and returns ABI-encoded (payload, signature).
func signAndEncode(payloadBytes []byte) (*string, int, error) {
	signature, err := signViaNode(payloadBytes)
	if err != nil {
		return nil, 0, fmt.Errorf("signing failed: %v", err)
	}
	encoded, err := abiEncodeTwo(payloadBytes, signature)
	if err != nil {
		return nil, 0, fmt.Errorf("ABI encoding failed: %v", err)
	}
	hex := base.BytesToHex(encoded)
	return &hex, 1, nil
}

func handleCEXFetchAndAttest(msg string) (data *string, status int, err error) {
	req, err := parseCEXRequest(msg)
	if err != nil {
		return nil, 0, err
	}
	apiKey, _, err := resolveCredentials(req)
	if err != nil {
		return nil, 0, err
	}
	provider, err := lookupCEX(req.CEX)
	if err != nil {
		return nil, 0, err
	}
	payloadBytes, err := provider.FetchAndAttest(apiKey, "", req.Symbol)
	if err != nil {
		return nil, 0, err
	}
	return signAndEncode(payloadBytes)
}

func handleCEX24hStats(msg string) (data *string, status int, err error) {
	req, err := parseCEXRequest(msg)
	if err != nil {
		return nil, 0, err
	}
	provider, err := lookupCEX(req.CEX)
	if err != nil {
		return nil, 0, err
	}
	payloadBytes, err := provider.Fetch24hStats("", "", req.Symbol)
	if err != nil {
		return nil, 0, err
	}
	return signAndEncode(payloadBytes)
}

func handleCEXAccountPnl(msg string) (data *string, status int, err error) {
	req, err := parseCEXRequest(msg)
	if err != nil {
		return nil, 0, err
	}
	apiKey, secretKey, err := resolveCredentials(req)
	if err != nil {
		return nil, 0, err
	}
	provider, err := lookupCEX(req.CEX)
	if err != nil {
		return nil, 0, err
	}
	payloadBytes, err := provider.FetchAccountPnl(apiKey, secretKey)
	if err != nil {
		return nil, 0, err
	}
	return signAndEncode(payloadBytes)
}

func handleCEXAccountSummary(msg string) (data *string, status int, err error) {
	req, err := parseCEXRequest(msg)
	if err != nil {
		return nil, 0, err
	}
	apiKey, secretKey, err := resolveCredentials(req)
	if err != nil {
		return nil, 0, err
	}
	provider, err := lookupCEX(req.CEX)
	if err != nil {
		return nil, 0, err
	}
	payloadBytes, err := provider.FetchAccountSummary(apiKey, secretKey)
	if err != nil {
		return nil, 0, err
	}
	return signAndEncode(payloadBytes)
}

func handleCEXUserProfile(msg string) (data *string, status int, err error) {
	req, err := parseCEXRequest(msg)
	if err != nil {
		return nil, 0, err
	}
	apiKey, secretKey, err := resolveCredentials(req)
	if err != nil {
		return nil, 0, err
	}
	provider, err := lookupCEX(req.CEX)
	if err != nil {
		return nil, 0, err
	}

	var payloadBytes []byte
	// Use structured ABI encoding if the provider supports it (required by BinanceAttestationStore).
	if enc, ok := provider.(ABIEncoderProvider); ok {
		payloadBytes, err = enc.EncodeUserProfile(apiKey, secretKey)
	} else {
		payloadBytes, err = provider.FetchUserProfile(apiKey, secretKey)
	}
	if err != nil {
		return nil, 0, err
	}
	return signAndEncode(payloadBytes)
}

// handleKeyUpdate decrypts the original message using the TEE node's key, then
// stores the decrypted value as an ECDSA private key.
func handleKeyUpdate(msg string) (data *string, status int, err error) {
	if msg == "" {
		return nil, 0, fmt.Errorf("originalMessage is empty")
	}

	ciphertext, hexErr := base.HexToBytes(msg)
	if hexErr != nil {
		return nil, 0, fmt.Errorf("invalid hex in originalMessage: %v", hexErr)
	}

	keyBytes, decryptErr := decryptViaNode(ciphertext)
	if decryptErr != nil {
		return nil, 0, fmt.Errorf("decryption failed: %v", decryptErr)
	}

	privKey, parseErr := parseSecp256k1PrivateKey(keyBytes)
	if parseErr != nil {
		return nil, 0, fmt.Errorf("invalid private key: %v", parseErr)
	}

	privateKey = privKey
	log.Printf("private key updated")
	return nil, 1, nil
}

// handleKeySign signs the original message with the stored private key.
// Returns the message and signature in data as ABI-encoded (bytes, bytes).
func handleKeySign(msg string) (data *string, status int, err error) {
	if privateKey == nil {
		return nil, 0, fmt.Errorf("no private key stored")
	}

	if msg == "" {
		return nil, 0, fmt.Errorf("originalMessage is empty")
	}

	msgBytes, hexErr := base.HexToBytes(msg)
	if hexErr != nil {
		return nil, 0, fmt.Errorf("invalid hex in originalMessage: %v", hexErr)
	}

	sig, signErr := signECDSA(privateKey, msgBytes)
	if signErr != nil {
		return nil, 0, fmt.Errorf("signing failed: %v", signErr)
	}

	encoded, abiErr := abiEncodeTwo(msgBytes, sig)
	if abiErr != nil {
		return nil, 0, fmt.Errorf("ABI encoding failed: %v", abiErr)
	}

	dataHex := base.BytesToHex(encoded)
	return &dataHex, 1, nil
}

// decryptViaNode calls the TEE node's /decrypt endpoint.
func decryptViaNode(ciphertext []byte) ([]byte, error) {
	url := fmt.Sprintf("http://localhost:%s/decrypt", signPort)
	reqBody, _ := json.Marshal(DecryptRequest{EncryptedMessage: ciphertext})

	resp, err := httpClient.Post(url, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("node returned %d: %s", resp.StatusCode, string(b))
	}

	var dr DecryptResponse
	if err := json.NewDecoder(resp.Body).Decode(&dr); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return dr.DecryptedMessage, nil
}

// signViaNode calls the TEE node's /sign endpoint.
func signViaNode(message []byte) ([]byte, error) {
	url := fmt.Sprintf("http://localhost:%s/sign", signPort)
	reqBody, _ := json.Marshal(SignRequest{Message: message})

	resp, err := httpClient.Post(url, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("node returned %d: %s", resp.StatusCode, string(b))
	}

	var sr SignResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(sr.Signature) == 0 {
		return nil, fmt.Errorf("empty signature from node")
	}

	return sr.Signature, nil
}

// attest triggers the TEE to fetch data from a CEX and produce an on-chain attestation.
//
// Usage (once the stack is fully deployed and running):
//
//	# Public endpoint — no credentials needed:
//	go run ./cmd/attest -mode ticker -symbol BTCUSDT
//
//	# Authenticated endpoint — API key required:
//	go run ./cmd/attest -mode profile -apiKey YOUR_KEY -secretKey YOUR_SECRET
//
// The tool:
//  1. Fetches the TEE node's ECIES public key from the proxy
//  2. Encrypts your credentials with that key
//  3. Sends the instruction on-chain via InstructionSender
//  4. Polls the proxy until the TEE returns the result
//  5. Prints the instruction ID, payload, and TEE signature
package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"math/big"
	"os"
	"strings"
	"time"

	"sign-tools/app"
	"sign-tools/app/contract"
	"sign-tools/base"
	"sign-tools/base/fccutils"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/flare-foundation/go-flare-common/pkg/logger"
	teetypes "github.com/flare-foundation/tee-node/pkg/types"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

// cexRequest mirrors the Go extension type for JSON encoding.
type cexRequest struct {
	CEX                  string `json:"cex"`
	EncryptedCredentials string `json:"encryptedCredentials,omitempty"`
	Symbol               string `json:"symbol,omitempty"`
	LookbackDays         int    `json:"lookbackDays,omitempty"`
}

// cexCredentials is encrypted and embedded in the request.
type cexCredentials struct {
	APIKey    string `json:"apiKey"`
	SecretKey string `json:"secretKey"`
}

func main() {
	_ = godotenv.Load()
	_ = godotenv.Load("../../.env")

	defaultInstructionSender := os.Getenv("INSTRUCTION_SENDER")

	defaultProxyURL := os.Getenv("TUNNEL_URL")
	if defaultProxyURL == "" {
		defaultProxyURL = base.DefaultExtensionProxyURL
	}

	af := flag.String("a", base.DefaultAddressesFile, "deployed addresses JSON file")
	cf := flag.String("c", base.DefaultChainNodeURL, "chain RPC URL")
	pf := flag.String("p", defaultProxyURL, "extension proxy URL (e.g. http://localhost:6676 or tunnel URL)")
	isf := flag.String("instructionSender", defaultInstructionSender, "InstructionSender contract address (or set INSTRUCTION_SENDER in .env)")
	modef := flag.String("mode", "ticker", "attestation mode: ticker | stats | pnl | account | profile | growth")
	cexf := flag.String("cex", "binance", "CEX provider name")
	symbolf := flag.String("symbol", "BTCUSDT", "trading pair symbol (ticker/stats modes)")
	lookbackDaysf := flag.Int("lookbackDays", 7, "lookback window in days for growth mode (1-29)")
	apiKeyf := flag.String("apiKey", "", "CEX API key (required for pnl/account/profile/growth)")
	secretKeyf := flag.String("secretKey", "", "CEX secret key (required for pnl/account/profile/growth)")
	timeoutf := flag.Duration("timeout", 120*time.Second, "poll timeout for TEE result")
	flag.Parse()

	if *isf == "" {
		logger.Fatal("--instructionSender is required (or set INSTRUCTION_SENDER in .env)")
	}

	instructionSenderAddr := common.HexToAddress(*isf)

	s, err := base.DefaultSupport(*af, *cf)
	if err != nil {
		logger.Fatal(err)
	}

	if err := run(s, instructionSenderAddr, *pf, *modef, *cexf, *symbolf, *lookbackDaysf, *apiKeyf, *secretKeyf, *timeoutf); err != nil {
		logger.Fatal(err)
	}
}

func run(
	s *base.Support,
	instructionSenderAddr common.Address,
	proxyURL, mode, cex, symbol string,
	lookbackDays int,
	apiKey, secretKey string,
	timeout time.Duration,
) error {
	mode = strings.ToLower(strings.TrimSpace(mode))

	// Step 0: ensure the InstructionSender contract has its extension ID set.
	// This is idempotent — it is skipped if the ID is already set.
	logger.Infof("Ensuring extension ID is set on InstructionSender %s...", instructionSenderAddr.Hex())
	if err := app.SetExtensionId(s, instructionSenderAddr); err != nil {
		return errors.Errorf("set extension ID: %s", err)
	}

	// Step 1: fetch the TEE node's public key and encrypt credentials if needed.
	var encryptedCredsHex string
	if apiKey != "" || secretKey != "" {
		logger.Infof("Fetching TEE public key to encrypt credentials...")
		teeInfo, err := fccutils.TeeInfo(proxyURL)
		if err != nil {
			return errors.Errorf("fetch TEE info: %s", err)
		}

		ecdsaPub, err := teetypes.ParsePubKey(teeInfo.MachineData.PublicKey)
		if err != nil {
			return errors.Errorf("parse TEE public key: %s", err)
		}

		eciesPub := &ecies.PublicKey{
			X:      ecdsaPub.X,
			Y:      ecdsaPub.Y,
			Curve:  ecies.DefaultCurve,
			Params: ecies.ECIES_AES128_SHA256,
		}

		credsJSON, _ := json.Marshal(cexCredentials{APIKey: apiKey, SecretKey: secretKey})
		ciphertext, err := ecies.Encrypt(rand.Reader, eciesPub, credsJSON, nil, nil)
		if err != nil {
			return errors.Errorf("ECIES encrypt credentials: %s", err)
		}

		encryptedCredsHex = "0x" + hex.EncodeToString(ciphertext)
		logger.Infof("Credentials encrypted (%d bytes ciphertext)", len(ciphertext))
	}

	// Step 2: build the CEXRequest message.
	req := cexRequest{
		CEX:                  cex,
		EncryptedCredentials: encryptedCredsHex,
		Symbol:               strings.ToUpper(strings.TrimSpace(symbol)),
	}
	if mode == "growth" {
		if lookbackDays < 1 || lookbackDays > 29 {
			return errors.Errorf("lookbackDays must be in [1,29]")
		}
		req.LookbackDays = lookbackDays
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return errors.Errorf("marshal request: %s", err)
	}
	logger.Infof("Sending %s attestation for CEX=%s...", mode, cex)

	// Step 3: pick the InstructionSender function and send on-chain.
	contractFunc, err := selectContractFunc(mode)
	if err != nil {
		return err
	}

	fee := app.DefaultFee
	txHash, instructionID, err := app.SendAttestWithTxHash(s, instructionSenderAddr, reqBytes, fee, contractFunc)
	if err != nil {
		return errors.Errorf("send attest: %s", err)
	}
	logger.Infof("Transaction submitted: %s", txHash.Hex())
	logger.Infof("  Explorer: https://coston2-explorer.flare.network/tx/%s", txHash.Hex())
	logger.Infof("Instruction ID: %s", instructionID.Hex())

	// Step 4: poll the proxy for the TEE result.
	logger.Infof("Waiting for TEE result (timeout=%s)...", timeout)
	actionResp, err := fccutils.ActionResult(proxyURL, instructionID)
	if err != nil {
		return errors.Errorf("poll result: %s", err)
	}

	if actionResp.Result.Status == 0 {
		return errors.Errorf("TEE returned error: %s", actionResp.Result.Log)
	}

	// Step 5: decode and print the result.
	resultBytes := []byte(actionResp.Result.Data)
	payloadBytes, sigBytes, err := abiDecodeTwo(resultBytes)
	if err != nil {
		return errors.Errorf("ABI decode result: %s", err)
	}

	logger.Infof("=== Attestation Result ===")
	logger.Infof("Instruction ID: %s", instructionID.Hex())
	logger.Infof("Signature (hex): %s", hex.EncodeToString(sigBytes))
	logger.Infof("Payload: %s", string(payloadBytes))

	return nil
}

// selectContractFunc maps a mode string to the corresponding InstructionSender method.
func selectContractFunc(mode string) (func(*contract.InstructionSender, *bind.TransactOpts, []byte) (*ethtypes.Transaction, error), error) {
	switch mode {
	case "ticker":
		return func(s *contract.InstructionSender, opts *bind.TransactOpts, msg []byte) (*ethtypes.Transaction, error) {
			return s.FetchBinanceAndAttest(opts, msg)
		}, nil
	case "stats":
		return func(s *contract.InstructionSender, opts *bind.TransactOpts, msg []byte) (*ethtypes.Transaction, error) {
			return s.FetchBinance24hStatsAndAttest(opts, msg)
		}, nil
	case "pnl":
		return func(s *contract.InstructionSender, opts *bind.TransactOpts, msg []byte) (*ethtypes.Transaction, error) {
			return s.FetchBinanceAccountPnlAndAttest(opts, msg)
		}, nil
	case "account":
		return func(s *contract.InstructionSender, opts *bind.TransactOpts, msg []byte) (*ethtypes.Transaction, error) {
			return s.FetchBinanceAccountSummaryAndAttest(opts, msg)
		}, nil
	case "profile":
		return func(s *contract.InstructionSender, opts *bind.TransactOpts, msg []byte) (*ethtypes.Transaction, error) {
			return s.FetchBinanceUserProfileAndAttest(opts, msg)
		}, nil
	case "growth":
		return func(s *contract.InstructionSender, opts *bind.TransactOpts, msg []byte) (*ethtypes.Transaction, error) {
			return s.FetchBinanceUserProfileAndAttest(opts, msg)
		}, nil
	default:
		return nil, errors.Errorf("unknown mode %q — use: ticker, stats, pnl, account, profile, growth", mode)
	}
}

// abiDecodeTwo decodes ABI-encoded (bytes, bytes).
func abiDecodeTwo(data []byte) ([]byte, []byte, error) {
	if len(data) < 64 {
		return nil, nil, errors.Errorf("data too short for (bytes,bytes): %d bytes", len(data))
	}
	offset1 := new(big.Int).SetBytes(data[0:32]).Uint64()
	offset2 := new(big.Int).SetBytes(data[32:64]).Uint64()

	readBytes := func(offset uint64) ([]byte, error) {
		if int(offset)+32 > len(data) {
			return nil, errors.Errorf("offset %d out of range", offset)
		}
		length := new(big.Int).SetBytes(data[offset : offset+32]).Uint64()
		start := offset + 32
		if int(start+length) > len(data) {
			return nil, errors.Errorf("length %d exceeds data at offset %d", length, offset)
		}
		return data[start : start+length], nil
	}

	a, err := readBytes(offset1)
	if err != nil {
		return nil, nil, err
	}
	b, err := readBytes(offset2)
	if err != nil {
		return nil, nil, err
	}
	return a, b, nil
}

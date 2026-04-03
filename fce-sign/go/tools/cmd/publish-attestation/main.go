// publish-attestation fetches a Binance user profile attestation from the local extension,
// then calls BinanceAttestationStore.publishAttestation() on-chain so the TEE-signed
// payload becomes a permanent, verifiable on-chain record.
//
// Usage:
//
//	go run ./cmd/publish-attestation \
//	  [-url http://127.0.0.1:8883/action] \
//	  [-store 0x<deployed BinanceAttestationStore address>] \
//	  [-a ../../config/coston2/deployed-addresses.json] \
//	  [-c https://coston2-api.flare.network/ext/C/rpc]
//
// If -store is not given the tool reads the forge artifact and deploys a new
// BinanceAttestationStore before publishing. Set ATTESTATION_STORE in .env to
// reuse a previously deployed instance without the -store flag.
package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sign-tools/base"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/flare-foundation/go-flare-common/pkg/logger"
	"github.com/joho/godotenv"
)

// attestationStoreABIStr is the ABI for BinanceAttestationStore.
const attestationStoreABIStr = `[
  {
    "type": "function",
    "name": "publishAttestation",
    "inputs": [
      {"name": "payload",   "type": "bytes"},
      {"name": "signature", "type": "bytes"}
    ],
    "outputs": [],
    "stateMutability": "nonpayable"
  },
  {
    "type": "event",
    "name": "AttestationPublished",
    "inputs": [
      {"name": "teeAddress", "type": "address", "indexed": true},
      {"name": "payload",    "type": "bytes",   "indexed": false},
      {"name": "signature",  "type": "bytes",   "indexed": false},
      {"name": "timestamp",  "type": "uint256", "indexed": false}
    ],
    "anonymous": false
  }
]`

// forgeArtifact matches the relevant fields of a forge build output JSON.
type forgeArtifact struct {
	Bytecode struct {
		Object string `json:"object"`
	} `json:"bytecode"`
}

func main() {
	// Load .env (try cwd, then project root relative to go/tools/).
	_ = godotenv.Load()
	_ = godotenv.Load("../../.env")

	defaultStore := os.Getenv("ATTESTATION_STORE")

	extURL := flag.String("url", "http://127.0.0.1:8883/action", "extension /action endpoint")
	storeAddr := flag.String("store", defaultStore, "deployed BinanceAttestationStore address (deploy new one if empty)")
	artifactPath := flag.String("artifact", "../../contract/out/BinanceAttestationStore.sol/BinanceAttestationStore.json", "forge artifact path (used when deploying)")
	af := flag.String("a", base.DefaultAddressesFile, "deployed addresses file")
	cf := flag.String("c", base.DefaultChainNodeURL, "chain RPC URL")
	flag.Parse()

	// ── Step 1: fetch attestation from local extension ──────────────────────
	logger.Infof("Fetching Binance user profile attestation from extension...")
	payload, signature, err := fetchAttestationFromExtension(*extURL)
	if err != nil {
		logger.Fatalf("fetch attestation: %v", err)
	}
	logger.Infof("  Payload (%d bytes): %s", len(payload), string(payload))
	logger.Infof("  Signature (%d bytes): %s", len(signature), hex.EncodeToString(signature))

	// ── Step 2: connect to chain ─────────────────────────────────────────────
	support, err := base.DefaultSupport(*af, *cf)
	if err != nil {
		logger.Fatalf("chain support: %v", err)
	}
	from := crypto.PubkeyToAddress(support.Prv.PublicKey)
	logger.Infof("  Wallet: %s", from.Hex())

	parsedABI, err := abi.JSON(strings.NewReader(attestationStoreABIStr))
	if err != nil {
		logger.Fatalf("parse ABI: %v", err)
	}

	// ── Step 3: deploy store if no address given ─────────────────────────────
	var contractAddr common.Address
	if *storeAddr == "" {
		logger.Infof("No -store address given, deploying BinanceAttestationStore...")
		contractAddr, err = deployAttestationStore(support, parsedABI, *artifactPath)
		if err != nil {
			logger.Fatalf("deploy: %v", err)
		}
		logger.Infof("  BinanceAttestationStore deployed at: %s", contractAddr.Hex())
		logger.Infof("  Set ATTESTATION_STORE=%s in .env to reuse.", contractAddr.Hex())
	} else {
		contractAddr = common.HexToAddress(*storeAddr)
		logger.Infof("Using existing BinanceAttestationStore at %s", contractAddr.Hex())
	}

	// ── Step 4: call publishAttestation on-chain ─────────────────────────────
	logger.Infof("Publishing attestation on-chain...")
	txHash, err := callPublishAttestation(support, parsedABI, contractAddr, payload, signature)
	if err != nil {
		logger.Fatalf("publishAttestation: %v", err)
	}
	logger.Infof("  TX hash: %s", txHash.Hex())

	// ── Step 5: show recovered TEE address ───────────────────────────────────
	msgHash := crypto.Keccak256(payload)
	sig65 := normalizeV(signature)
	pub, recErr := crypto.SigToPub(msgHash, sig65)
	if recErr == nil {
		teeAddr := crypto.PubkeyToAddress(*pub)
		logger.Infof("  Recovered TEE address: %s", teeAddr.Hex())
	}

	logger.Infof("✅ Attestation published on-chain!")
	fmt.Printf("store=%s\ntx=%s\n", contractAddr.Hex(), txHash.Hex())
}

// fetchAttestationFromExtension calls the local extension and returns (payload, signature).
func fetchAttestationFromExtension(endpoint string) ([]byte, []byte, error) {
	opCommand := "BINANCE_USER_PROFILE"
	reqBytes := []byte("{}")

	df := dataFixed{
		InstructionID:   "0x0000000000000000000000000000000000000000000000000000000000000001",
		TeeID:           "0x0000000000000000000000000000000000000000",
		Timestamp:       time.Now().Unix(),
		OpType:          stringToBytes32Hex("MARKET"),
		OpCommand:       stringToBytes32Hex(opCommand),
		OriginalMessage: bytesToHex(reqBytes),
	}
	dfBytes, _ := json.Marshal(df)

	body := actionRequest{Data: actionData{
		ID:            "0x0000000000000000000000000000000000000000000000000000000000000001",
		Type:          "instruction",
		SubmissionTag: "publish-attestation",
		Message:       bytesToHex(dfBytes),
	}}
	bodyBytes, _ := json.Marshal(body)

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("POST extension: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("extension returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result actionResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, nil, fmt.Errorf("decode response: %w", err)
	}
	if result.Status != 1 {
		logMsg := ""
		if result.Log != nil {
			logMsg = *result.Log
		}
		return nil, nil, fmt.Errorf("extension status=%d log=%s", result.Status, logMsg)
	}
	if result.Data == nil {
		return nil, nil, fmt.Errorf("extension returned nil data")
	}

	encoded, err := hexToBytes(*result.Data)
	if err != nil {
		return nil, nil, fmt.Errorf("hex decode data: %w", err)
	}
	return abiDecodeTwo(encoded)
}

// deployAttestationStore reads the forge artifact and deploys the contract.
func deployAttestationStore(s *base.Support, parsedABI abi.ABI, artifactPath string) (common.Address, error) {
	absPath, err := filepath.Abs(artifactPath)
	if err != nil {
		return common.Address{}, fmt.Errorf("resolve artifact path: %w", err)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return common.Address{}, fmt.Errorf("read forge artifact at %s: %w\nRun `forge build` in contract/ first.", absPath, err)
	}

	var artifact forgeArtifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		return common.Address{}, fmt.Errorf("parse forge artifact: %w", err)
	}

	bytecodeHex := strings.TrimPrefix(artifact.Bytecode.Object, "0x")
	if bytecodeHex == "" {
		return common.Address{}, fmt.Errorf("forge artifact has empty bytecode — run `forge build` in contract/")
	}
	bytecode, err := hex.DecodeString(bytecodeHex)
	if err != nil {
		return common.Address{}, fmt.Errorf("decode bytecode: %w", err)
	}

	opts, err := bind.NewKeyedTransactorWithChainID(s.Prv, s.ChainID)
	if err != nil {
		return common.Address{}, fmt.Errorf("transactor: %w", err)
	}

	addr, tx, _, err := bind.DeployContract(opts, parsedABI, bytecode, s.ChainClient)
	if err != nil {
		return common.Address{}, fmt.Errorf("deploy contract: %w", err)
	}

	receipt, err := bind.WaitMined(context.Background(), s.ChainClient, tx)
	if err != nil {
		return common.Address{}, fmt.Errorf("wait deploy: %w", err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return common.Address{}, fmt.Errorf("deploy tx reverted")
	}
	return addr, nil
}

// callPublishAttestation calls publishAttestation(payload, signature) on the store contract.
func callPublishAttestation(s *base.Support, parsedABI abi.ABI, storeAddr common.Address, payload, signature []byte) (common.Hash, error) {
	opts, err := bind.NewKeyedTransactorWithChainID(s.Prv, s.ChainID)
	if err != nil {
		return common.Hash{}, fmt.Errorf("transactor: %w", err)
	}

	contract := bind.NewBoundContract(storeAddr, parsedABI, s.ChainClient, s.ChainClient, s.ChainClient)
	tx, err := contract.Transact(opts, "publishAttestation", payload, signature)
	if err != nil {
		return common.Hash{}, fmt.Errorf("transact: %w", err)
	}

	receipt, err := bind.WaitMined(context.Background(), s.ChainClient, tx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("wait mined: %w", err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return common.Hash{}, fmt.Errorf("publishAttestation tx reverted")
	}
	return tx.Hash(), nil
}

// normalizeV converts v from 27/28 → 0/1 for go-ethereum SigToPub.
func normalizeV(sig []byte) []byte {
	if len(sig) != 65 {
		return sig
	}
	out := make([]byte, 65)
	copy(out, sig)
	if out[64] >= 27 {
		out[64] -= 27
	}
	return out
}

// ── local extension wire types (mirrors test-binance-attest) ────────────────

type actionRequest struct {
	Data actionData `json:"data"`
}

type actionData struct {
	ID            string `json:"id"`
	Type          string `json:"type"`
	SubmissionTag string `json:"submissionTag"`
	Message       string `json:"message"`
}

type dataFixed struct {
	InstructionID   string `json:"instructionId"`
	TeeID           string `json:"teeId"`
	Timestamp       int64  `json:"timestamp"`
	OpType          string `json:"opType"`
	OpCommand       string `json:"opCommand"`
	OriginalMessage string `json:"originalMessage"`
}

type actionResult struct {
	ID      string  `json:"id"`
	Status  int     `json:"status"`
	Log     *string `json:"log"`
	OpType  string  `json:"opType"`
	OpCommand string `json:"opCommand"`
	Data    *string `json:"data"`
}

func stringToBytes32Hex(s string) string {
	b := make([]byte, 32)
	copy(b, []byte(s))
	return "0x" + hex.EncodeToString(b)
}

func hexToBytes(h string) ([]byte, error) {
	h = strings.TrimPrefix(h, "0x")
	if len(h)%2 == 1 {
		h = "0" + h
	}
	return hex.DecodeString(h)
}

func bytesToHex(b []byte) string {
	return "0x" + hex.EncodeToString(b)
}

func abiDecodeTwo(data []byte) ([]byte, []byte, error) {
	if len(data) < 64 {
		return nil, nil, fmt.Errorf("data too short for ABI-encoded (bytes, bytes)")
	}
	offsetA := new(big.Int).SetBytes(data[0:32]).Int64()
	offsetB := new(big.Int).SetBytes(data[32:64]).Int64()

	a, err := abiReadBytes(data, offsetA)
	if err != nil {
		return nil, nil, err
	}
	b, err := abiReadBytes(data, offsetB)
	if err != nil {
		return nil, nil, err
	}
	return a, b, nil
}

func abiReadBytes(data []byte, offset int64) ([]byte, error) {
	if offset+32 > int64(len(data)) {
		return nil, fmt.Errorf("offset %d out of range", offset)
	}
	length := new(big.Int).SetBytes(data[offset : offset+32]).Int64()
	start := offset + 32
	if start+length > int64(len(data)) {
		return nil, fmt.Errorf("bytes data out of range")
	}
	return data[start : start+length], nil
}

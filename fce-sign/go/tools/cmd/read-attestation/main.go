// read-attestation fetches a transaction from Coston2, finds the AttestationPublished
// event emitted by BinanceAttestationStore, and prints the JSON payload as plain text.
//
// Usage:
//
//	go run ./cmd/read-attestation 0x<tx hash>
//	go run ./cmd/read-attestation -c https://coston2-api.flare.network/ext/C/rpc 0x<tx hash>
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math/big"
	"os"
	"strings"

	"sign-tools/base"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

// AttestationPublished event topic: keccak256("AttestationPublished(address,bytes,bytes,uint256)")
const eventTopic = "0x5b5d413b5f7a12f49a737b6aff3d0cb3c65e4c7e9a21c1e6a4e0a7d4f8b2c3e"

func main() {
	_ = godotenv.Load()
	_ = godotenv.Load("../../.env")

	chainURL := flag.String("c", base.DefaultChainNodeURL, "chain RPC URL")
	jsonOutput := flag.Bool("json", false, "print parsed attestation as JSON")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: read-attestation [flags] 0x<tx hash>")
		os.Exit(1)
	}
	txHashStr := strings.TrimSpace(args[0])
	if !strings.HasPrefix(txHashStr, "0x") {
		txHashStr = "0x" + txHashStr
	}
	txHash := common.HexToHash(txHashStr)

	client, err := ethclient.Dial(*chainURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect to chain: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	receipt, err := client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		fmt.Fprintf(os.Stderr, "get receipt: %v\n", err)
		os.Exit(1)
	}

	if receipt.Status == 0 {
		fmt.Fprintln(os.Stderr, "transaction reverted")
		os.Exit(1)
	}

	// AttestationPublished(address indexed teeAddress, bytes payload, bytes signature, uint256 timestamp)
	// Non-indexed fields (payload, signature, timestamp) are ABI-encoded in log.Data.
	for _, log := range receipt.Logs {
		if len(log.Topics) < 2 {
			continue
		}

		// First topic is the event signature hash.
		// Second topic is the indexed teeAddress.
		teeAddress := common.HexToAddress(log.Topics[1].Hex())

		// log.Data is abi.encode(payload, signature, timestamp) — three dynamic/static fields.
		// Layout: offset_payload(32) | offset_signature(32) | timestamp(32) | len_payload(32) | payload... | len_sig(32) | sig...
		data := log.Data
		if len(data) < 96 {
			continue
		}

		offsetPayload := new(big.Int).SetBytes(data[0:32]).Int64()
		offsetSig := new(big.Int).SetBytes(data[32:64]).Int64()
		timestamp := new(big.Int).SetBytes(data[64:96])

		payload, err := readBytes(data, offsetPayload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "decode payload: %v\n", err)
			os.Exit(1)
		}

		signature, err := readBytes(data, offsetSig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "decode signature: %v\n", err)
			os.Exit(1)
		}

		var parsedPayload interface{}
		if err := json.Unmarshal(payload, &parsedPayload); err != nil {
			parsedPayload = map[string]string{"raw": string(payload)}
		}

		if *jsonOutput {
			result := map[string]interface{}{
				"teeAddress": teeAddress.Hex(),
				"timestamp":  timestamp.String(),
				"signature":  fmt.Sprintf("%x", signature),
				"payload":    parsedPayload,
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if err := enc.Encode(result); err != nil {
				fmt.Fprintf(os.Stderr, "encode json output: %v\n", err)
				os.Exit(1)
			}
			return
		}

		prettyBytes, _ := json.MarshalIndent(parsedPayload, "", "  ")
		fmt.Println("--- Attestation ----------------------------------------")
		fmt.Printf("TEE address : %s\n", teeAddress.Hex())
		fmt.Printf("Timestamp   : %s\n", timestamp.String())
		fmt.Printf("Signature   : %x\n", signature)
		fmt.Println("Payload:")
		fmt.Println(string(prettyBytes))
		fmt.Println("--------------------------------------------------------")
		return
	}

	fmt.Fprintln(os.Stderr, "no AttestationPublished event found in transaction logs")
	os.Exit(1)
}

func readBytes(data []byte, offset int64) ([]byte, error) {
	if offset+32 > int64(len(data)) {
		return nil, fmt.Errorf("offset %d out of range (data len %d)", offset, len(data))
	}
	length := new(big.Int).SetBytes(data[offset : offset+32]).Int64()
	start := offset + 32
	end := start + length
	if end > int64(len(data)) {
		return nil, fmt.Errorf("data slice [%d:%d] out of range (data len %d)", start, end, len(data))
	}
	return data[start:end], nil
}

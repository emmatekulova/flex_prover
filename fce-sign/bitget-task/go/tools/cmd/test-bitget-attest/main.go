package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"
)

type action struct {
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
	ID            string  `json:"id"`
	SubmissionTag string  `json:"submissionTag"`
	Status        int     `json:"status"`
	Log           *string `json:"log"`
	OpType        string  `json:"opType"`
	OpCommand     string  `json:"opCommand"`
	Version       string  `json:"version"`
	Data          *string `json:"data"`
}

func main() {
	endpoint := flag.String("url", "http://127.0.0.1:8883/action", "extension /action endpoint")
	mode := flag.String("mode", "account", "test mode: account")
	flag.Parse()

	opCommand := "BITGET_ACCOUNT_SUMMARY"
	reqBytes := []byte("{}")

	switch strings.ToLower(strings.TrimSpace(*mode)) {
	case "account":
		// Account summary handler ignores originalMessage for now.
		reqBytes = []byte("{}")
		opCommand = "BITGET_ACCOUNT_SUMMARY"
	default:
		panic("invalid -mode, use account")
	}

	df := dataFixed{
		InstructionID:   "0x0000000000000000000000000000000000000000000000000000000000000001",
		TeeID:           "0x0000000000000000000000000000000000000000",
		Timestamp:       time.Now().Unix(),
		OpType:          stringToBytes32Hex("MARKET"),
		OpCommand:       stringToBytes32Hex(opCommand),
		OriginalMessage: bytesToHex(reqBytes),
	}
	dfBytes, _ := json.Marshal(df)

	body := action{Data: actionData{
		ID:            "0x0000000000000000000000000000000000000000000000000000000000000001",
		Type:          "instruction",
		SubmissionTag: "manual-bitget-attest-check",
		Message:       bytesToHex(dfBytes),
	}}

	bodyBytes, _ := json.Marshal(body)
	httpReq, err := http.NewRequest(http.MethodPost, *endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		panic(err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("status=%d body=%s", resp.StatusCode, string(respBody)))
	}

	var result actionResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		panic(err)
	}

	if result.Status != 1 {
		logMsg := ""
		if result.Log != nil {
			logMsg = *result.Log
		}
		panic(fmt.Sprintf("action failed status=%d log=%s", result.Status, logMsg))
	}
	if result.Data == nil {
		panic("action succeeded but data is nil")
	}

	encoded, err := hexToBytes(*result.Data)
	if err != nil {
		panic(err)
	}
	payloadBytes, signatureBytes, err := abiDecodeTwo(encoded)
	if err != nil {
		panic(err)
	}

	fmt.Println("✅ Bitget attestation + TEE sign succeeded")
	fmt.Printf("mode=%s\n", strings.ToLower(strings.TrimSpace(*mode)))
	fmt.Printf("opType=%s opCommand=%s status=%d\n", bytes32HexToString(result.OpType), bytes32HexToString(result.OpCommand), result.Status)
	fmt.Printf("signature_len=%d\n", len(signatureBytes))
	fmt.Printf("payload=%s\n", string(payloadBytes))
}

func stringToBytes32Hex(s string) string {
	b := make([]byte, 32)
	copy(b, []byte(s))
	return "0x" + hex.EncodeToString(b)
}

func bytes32HexToString(h string) string {
	b, err := hexToBytes(h)
	if err != nil {
		return ""
	}
	end := len(b)
	for end > 0 && b[end-1] == 0 {
		end--
	}
	return string(b[:end])
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

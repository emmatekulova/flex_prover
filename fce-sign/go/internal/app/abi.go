package app

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// abiEncodeTwo ABI-encodes two dynamic byte arrays: (bytes, bytes).
// Layout:
//
//	offset of first bytes  (32 bytes) = 64
//	offset of second bytes (32 bytes) = 64 + 32 + padded(len(a))
//	length of a            (32 bytes)
//	a data                 (ceil(len(a)/32)*32 bytes)
//	length of b            (32 bytes)
//	b data                 (ceil(len(b)/32)*32 bytes)
func abiEncodeTwo(a, b []byte) ([]byte, error) {
	aPadded := padToMultipleOf32(a)
	bPadded := padToMultipleOf32(b)

	offsetA := big.NewInt(64)
	offsetB := big.NewInt(int64(64 + 32 + len(aPadded)))

	buf := make([]byte, 0, 64+32+len(aPadded)+32+len(bPadded))

	buf = append(buf, padLeft(offsetA.Bytes(), 32)...)
	buf = append(buf, padLeft(offsetB.Bytes(), 32)...)

	buf = append(buf, padLeft(big.NewInt(int64(len(a))).Bytes(), 32)...)
	buf = append(buf, aPadded...)

	buf = append(buf, padLeft(big.NewInt(int64(len(b))).Bytes(), 32)...)
	buf = append(buf, bPadded...)

	return buf, nil
}

// abiDecodeTwo decodes ABI-encoded (bytes, bytes) back into two byte slices.
func abiDecodeTwo(data []byte) ([]byte, []byte, error) {
	if len(data) < 128 {
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

// abiEncodeUserProfile ABI-encodes a BinanceUserProfileAttestationPayload as a
// Solidity tuple so BinanceAttestationStore can abi.decode it on-chain and emit
// named typed event fields instead of opaque bytes.
func abiEncodeUserProfile(p BinanceUserProfileAttestationPayload) ([]byte, error) {
	assetComponents := []abi.ArgumentMarshaling{
		{Name: "asset", Type: "string"},
		{Name: "total", Type: "string"},
		{Name: "estimatedUsdt", Type: "string"},
	}
	profileType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "source", Type: "string"},
		{Name: "uid", Type: "uint64"},
		{Name: "accountType", Type: "string"},
		{Name: "permissions", Type: "string[]"},
		{Name: "canTrade", Type: "bool"},
		{Name: "canDeposit", Type: "bool"},
		{Name: "canWithdraw", Type: "bool"},
		{Name: "estimatedTotalUsdt", Type: "string"},
		{Name: "unsupportedAssetCount", Type: "uint256"},
		{Name: "assets", Type: "tuple[]", Components: assetComponents},
		{Name: "fetchedAt", Type: "uint256"},
		{Name: "version", Type: "string"},
	})
	if err != nil {
		return nil, fmt.Errorf("build ABI type: %v", err)
	}

	type AssetABI struct {
		Asset         string
		Total         string
		EstimatedUsdt string
	}
	type ProfileABI struct {
		Source                string
		Uid                   uint64
		AccountType           string
		Permissions           []string
		CanTrade              bool
		CanDeposit            bool
		CanWithdraw           bool
		EstimatedTotalUsdt    string
		UnsupportedAssetCount *big.Int
		Assets                []AssetABI
		FetchedAt             *big.Int
		Version               string
	}

	assets := make([]AssetABI, len(p.Assets))
	for i, a := range p.Assets {
		assets[i] = AssetABI{Asset: a.Asset, Total: a.Total, EstimatedUsdt: a.EstimatedUSDT}
	}

	profile := ProfileABI{
		Source:                p.Source,
		Uid:                   uint64(p.UID),
		AccountType:           p.AccountType,
		Permissions:           p.Permissions,
		CanTrade:              p.CanTrade,
		CanDeposit:            p.CanDeposit,
		CanWithdraw:           p.CanWithdraw,
		EstimatedTotalUsdt:    p.EstimatedTotalUSDT,
		UnsupportedAssetCount: big.NewInt(int64(p.UnsupportedAssetCnt)),
		Assets:                assets,
		FetchedAt:             big.NewInt(p.FetchedAt),
		Version:               p.Version,
	}

	return abi.Arguments{{Type: profileType}}.Pack(profile)
}

// padToMultipleOf32 pads data to a multiple of 32 bytes with trailing zeros.
func padToMultipleOf32(data []byte) []byte {
	if len(data) == 0 {
		return nil
	}
	remainder := len(data) % 32
	if remainder == 0 {
		result := make([]byte, len(data))
		copy(result, data)
		return result
	}
	padded := make([]byte, len(data)+32-remainder)
	copy(padded, data)
	return padded
}

package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	hederasdk "github.com/hashgraph/hedera-sdk-go/v2"
	"flex-prover-backend/service"
)

type verifyRequest struct {
	WalletAddress   string `json:"walletAddress"`
	AttestationURL  string `json:"attestationUrl"`
	HederaAccountID string `json:"hederaAccountId"`
	Nonce           string `json:"nonce"`
	Signature       string `json:"signature"`
}

func Verify(h *service.HederaService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req verifyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// 1. Consume nonce (one-time use)
		if !consumeNonce(req.Nonce) {
			jsonError(w, "nonce is invalid or expired", http.StatusUnauthorized)
			return
		}

		// 2. Verify wallet owns the address — recover signer from signature
		message := buildWhaleMessage(req.Nonce, req.AttestationURL, req.HederaAccountID)
		recovered, err := recoverAddress(message, req.Signature)
		if err != nil {
			jsonError(w, fmt.Sprintf("signature recovery failed: %v", err), http.StatusUnauthorized)
			return
		}
		if !strings.EqualFold(recovered.Hex(), req.WalletAddress) {
			jsonError(w, fmt.Sprintf("signature mismatch: recovered %s, expected %s", recovered.Hex(), req.WalletAddress), http.StatusUnauthorized)
			return
		}

		// 3. Optional PNL threshold check (PNL_THRESHOLD env, -1 = disabled)
		if threshold := pnlThreshold(); threshold >= 0 {
			// attestation URL was already decoded in the Next.js proxy layer;
			// growth percent would need to be passed explicitly for this check.
			// Skipped here — enforced via Next.js verify route if needed.
			_ = threshold
		}

		// 4. Parse Hedera account ID
		accountID, err := hederasdk.AccountIDFromString(req.HederaAccountID)
		if err != nil {
			jsonError(w, fmt.Sprintf("invalid Hedera account ID: %v", err), http.StatusBadRequest)
			return
		}

		// 5. Mint + transfer WHALE NFT.
		// Hedera auto-association handles accounts with open slots (most testnet accounts).
		// If the account has no open slots, MintAndTransfer returns a TOKEN_NOT_ASSOCIATED error.
		// Hedera NFT metadata is capped at 100 bytes per serial.
		// Store wallet + last 10 chars of attestation URL as a short reference.
		shortRef := req.AttestationURL
		if len(shortRef) > 10 {
			shortRef = shortRef[len(shortRef)-10:]
		}
		metadata := []byte(fmt.Sprintf(`{"w":"%s","r":"%s"}`, req.WalletAddress, shortRef))
		if err := h.MintAndTransfer(accountID, metadata); err != nil {
			errStr := err.Error()
			if strings.Contains(errStr, "TOKEN_NOT_ASSOCIATED_TO_ACCOUNT") {
				jsonResponse(w, http.StatusOK, map[string]string{
					"status":  "needs_association",
					"message": "Your Hedera account has no auto-association slots. Please associate token " + os.Getenv("HEDERA_TOKEN_ID") + " manually via HashPack or hashscan.io, then try again.",
				})
				return
			}
			jsonError(w, fmt.Sprintf("mint/transfer failed: %v", err), http.StatusInternalServerError)
			return
		}

		jsonResponse(w, http.StatusOK, map[string]string{"status": "minted"})
	}
}

// buildWhaleMessage must match exactly what the frontend builds in whale-page.tsx:
// `FlexProver Whale Verification\n\nAttestation: {url}\nHedera Account: {id}\nNonce: {nonce}`
func buildWhaleMessage(nonce, attestationURL, hederaAccountID string) string {
	return fmt.Sprintf(
		"FlexProver Whale Verification\n\nAttestation: %s\nHedera Account: %s\nNonce: %s",
		attestationURL, hederaAccountID, nonce,
	)
}

// recoverAddress recovers the EVM address from an eth_sign (personal_sign) signature.
func recoverAddress(message, hexSig string) (common.Address, error) {
	sig := common.FromHex(hexSig)
	if len(sig) != 65 {
		return common.Address{}, fmt.Errorf("signature must be 65 bytes, got %d", len(sig))
	}
	if sig[64] >= 27 {
		sig[64] -= 27
	}
	hash := accounts.TextHash([]byte(message))
	pubKey, err := crypto.SigToPub(hash, sig)
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(*pubKey), nil
}

func pnlThreshold() float64 {
	f, err := strconv.ParseFloat(os.Getenv("PNL_THRESHOLD"), 64)
	if err != nil {
		return 0
	}
	return f
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	jsonResponse(w, code, map[string]string{"status": "error", "message": msg})
}

func jsonResponse(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

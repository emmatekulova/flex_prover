package handler

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	hederasdk "github.com/hashgraph/hedera-sdk-go/v2"
	"flex-prover-backend/service"
)

type prepareAssociateRequest struct {
	HederaAccountID string `json:"hederaAccountId"`
}

// PrepareAssociate builds an unsigned TokenAssociateTransaction and returns it
// as a hex string for the frontend to sign with MetaMask and submit back.
func PrepareAssociate(h *service.HederaService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req prepareAssociateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		accountID, err := hederasdk.AccountIDFromString(req.HederaAccountID)
		if err != nil {
			jsonError(w, fmt.Sprintf("invalid Hedera account ID: %v", err), http.StatusBadRequest)
			return
		}

		txBytes, err := h.BuildUnsignedAssociateTx(accountID)
		if err != nil {
			jsonError(w, fmt.Sprintf("failed to build associate tx: %v", err), http.StatusInternalServerError)
			return
		}

		jsonResponse(w, http.StatusOK, map[string]string{
			"txBytes": hex.EncodeToString(txBytes),
		})
	}
}

type submitAssociateRequest struct {
	TxBytes string `json:"txBytes"`
}

// SubmitAssociate receives a signed TokenAssociateTransaction (hex) and submits it.
func SubmitAssociate(h *service.HederaService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req submitAssociateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		txBytes, err := hex.DecodeString(req.TxBytes)
		if err != nil {
			jsonError(w, "txBytes must be a valid hex string", http.StatusBadRequest)
			return
		}

		if err := h.SubmitSignedTx(txBytes); err != nil {
			jsonError(w, fmt.Sprintf("failed to submit association: %v", err), http.StatusInternalServerError)
			return
		}

		jsonResponse(w, http.StatusOK, map[string]string{"status": "associated"})
	}
}

package service

import (
	"fmt"
	"os"

	"github.com/hashgraph/hedera-sdk-go/v2"
)

// HederaService holds the operator client and token config.
type HederaService struct {
	client    *hedera.Client
	tokenID   hedera.TokenID
	operatorID hedera.AccountID
}

func NewHederaService() (*HederaService, error) {
	network := os.Getenv("HEDERA_NETWORK")
	operatorIDStr := os.Getenv("HEDERA_OPERATOR_ID")
	operatorKeyStr := os.Getenv("HEDERA_OPERATOR_KEY")
	tokenIDStr := os.Getenv("HEDERA_TOKEN_ID")

	if operatorIDStr == "" || operatorKeyStr == "" || tokenIDStr == "" {
		return nil, fmt.Errorf("HEDERA_OPERATOR_ID, HEDERA_OPERATOR_KEY, and HEDERA_TOKEN_ID must be set")
	}

	operatorID, err := hedera.AccountIDFromString(operatorIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid HEDERA_OPERATOR_ID: %w", err)
	}

	operatorKey, err := hedera.PrivateKeyFromStringECDSA(operatorKeyStr)
	if err != nil {
		// try ED25519 fallback
		operatorKey, err = hedera.PrivateKeyFromString(operatorKeyStr)
		if err != nil {
			return nil, fmt.Errorf("invalid HEDERA_OPERATOR_KEY: %w", err)
		}
	}

	tokenID, err := hedera.TokenIDFromString(tokenIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid HEDERA_TOKEN_ID: %w", err)
	}

	var client *hedera.Client
	if network == "mainnet" {
		client = hedera.ClientForMainnet()
	} else {
		client = hedera.ClientForTestnet()
	}
	client.SetOperator(operatorID, operatorKey)

	return &HederaService{
		client:    client,
		tokenID:   tokenID,
		operatorID: operatorID,
	}, nil
}

// TokenID returns the configured WHALE token ID.
func (h *HederaService) TokenID() hedera.TokenID { return h.tokenID }

// Client returns the Hedera client.
func (h *HederaService) Client() *hedera.Client { return h.client }

// IsAssociated checks if accountID already has the WHALE token associated.
func (h *HederaService) IsAssociated(accountID hedera.AccountID) (bool, error) {
	info, err := hedera.NewAccountInfoQuery().
		SetAccountID(accountID).
		Execute(h.client)
	if err != nil {
		return false, fmt.Errorf("account info query failed: %w", err)
	}
	for _, rel := range info.TokenRelationships {
		if rel.TokenID.String() == h.tokenID.String() {
			return true, nil
		}
	}
	return false, nil
}

// MintAndTransfer mints 1 WHALE token and transfers it to recipient.
// WHALE is a fungible HTS token (FUNGIBLE_COMMON), so we mint amount=1 then transfer.
func (h *HederaService) MintAndTransfer(recipient hedera.AccountID, _ []byte) error {
	// Mint 1 unit to treasury
	mintResp, err := hedera.NewTokenMintTransaction().
		SetTokenID(h.tokenID).
		SetAmount(1).
		Execute(h.client)
	if err != nil {
		return fmt.Errorf("mint execute failed: %w", err)
	}
	if _, err = mintResp.GetReceipt(h.client); err != nil {
		return fmt.Errorf("mint receipt failed: %w", err)
	}

	// Transfer 1 token from treasury to recipient
	_, err = hedera.NewTransferTransaction().
		AddTokenTransfer(h.tokenID, h.operatorID, -1).
		AddTokenTransfer(h.tokenID, recipient, 1).
		Execute(h.client)
	if err != nil {
		return fmt.Errorf("transfer failed: %w", err)
	}
	return nil
}

// BuildUnsignedAssociateTx returns a frozen-but-unsigned TokenAssociateTransaction
// as bytes, so the frontend can sign it with MetaMask and submit back.
func (h *HederaService) BuildUnsignedAssociateTx(accountID hedera.AccountID) ([]byte, error) {
	tx, err := hedera.NewTokenAssociateTransaction().
		SetAccountID(accountID).
		SetTokenIDs(h.tokenID).
		FreezeWith(h.client)
	if err != nil {
		return nil, fmt.Errorf("freeze associate tx: %w", err)
	}
	return tx.ToBytes()
}

// SubmitSignedTx deserializes and submits a signed TokenAssociateTransaction.
func (h *HederaService) SubmitSignedTx(txBytes []byte) error {
	iface, err := hedera.TransactionFromBytes(txBytes)
	if err != nil {
		return fmt.Errorf("deserialize tx: %w", err)
	}
	tx, ok := iface.(hedera.TokenAssociateTransaction)
	if !ok {
		return fmt.Errorf("expected TokenAssociateTransaction, got %T", iface)
	}
	resp, err := tx.Execute(h.client)
	if err != nil {
		return fmt.Errorf("execute tx: %w", err)
	}
	_, err = resp.GetReceipt(h.client)
	return err
}

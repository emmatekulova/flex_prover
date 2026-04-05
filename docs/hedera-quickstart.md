# Hedera Quickstart

Everything you need to run the Hedera layer of FlexProver from zero.

---

## Prerequisites

- Node.js 18+
- A funded Hedera Testnet account — get one free at [portal.hedera.com](https://portal.hedera.com)
- Your account ID (`0.0.XXXXXX`) and private key (DER-encoded hex or PEM)

---

## Environment Variables

Add these to your `.env` (or `.env.local` for the frontend):

```env
# Hedera account (operator — pays for transactions)
HEDERA_ACCOUNT_ID=0.0.123456
HEDERA_PRIVATE_KEY=302e...    # DER-encoded hex private key from portal

# Created once at deploy time (see Setup below)
HEDERA_TOPIC_ID=0.0.789012        # HCS audit log topic
HEDERA_TOKEN_ID=0.0.789013        # HTS FlexCard NFT token
HEDERA_TREASURY_KEY=302e...       # Key that controls the token treasury

# Network
HEDERA_NETWORK=testnet            # or mainnet
```

---

## One-Time Setup (run once per deployment)

Install the SDK:

```bash
npm install @hashgraph/sdk
```

Run the setup script to create the HCS topic and HTS token:

```bash
npx ts-node scripts/hedera-setup.ts
```

This script does three things:

1. **Creates the HCS topic** — the public audit log channel
2. **Creates the HTS token** — an NFT collection with KYC key, pause key, and a 5% royalty fee
3. **Prints the topic ID and token ID** — copy these into your `.env`

**What the setup script creates:**

```typescript
// HCS topic — no submit key means anyone can post (public audit log)
const topicTx = await new TopicCreateTransaction()
  .setTopicMemo("FlexProver proof audit log")
  .execute(client)

// HTS NFT token with compliance controls
const tokenTx = await new TokenCreateTransaction()
  .setTokenName("FlexCard")
  .setTokenSymbol("FLEX")
  .setTokenType(TokenType.NonFungibleUnique)
  .setTreasuryAccountId(operatorId)
  .setAdminKey(operatorKey)
  .setKycKey(operatorKey)          // required for KYC grants
  .setPauseKey(operatorKey)        // allows pausing expired badges
  .setFreezeKey(operatorKey)       // allows freezing sanctioned accounts
  .setSupplyKey(operatorKey)       // required for minting
  .setCustomFees([
    new CustomRoyaltyFee()
      .setNumerator(5).setDenominator(100)   // 5% royalty
      .setFallbackFee(new CustomFixedFee().setHbarAmount(new Hbar(1)))
      .setFeeCollectorAccountId(operatorId)
  ])
  .execute(client)
```

---

## Running the Backend (with Hedera)

The backend handles HCS posting and HTS minting after a Flare proof is verified.

```bash
cd backend
cp .env.example .env        # fill in HEDERA_* vars
npm install
npm run dev
```

The relevant endpoint is `POST /proof/verified` — called internally after the Flare TEE attestation is confirmed. It:
1. Posts the proof hash + metadata to HCS
2. Mints an HTS NFT to the user's linked Hedera account
3. Grants KYC flag on the token for that account
4. Returns the Mirror Node URL for the HCS message

---

## Verifying on Mirror Node

After a proof is issued, query the audit log:

```bash
# All messages on the topic
curl "https://testnet.mirrornode.hedera.com/api/v1/topics/0.0.789012/messages"

# NFT info for a specific serial number
curl "https://testnet.mirrornode.hedera.com/api/v1/tokens/0.0.789013/nfts/1"
```

The frontend also calls these endpoints to show proof history on a wallet's profile.

---

## Hashscan (Block Explorer)

View everything on [hashscan.io](https://hashscan.io/testnet):

- Your topic: `https://hashscan.io/testnet/topic/0.0.789012`
- Your token: `https://hashscan.io/testnet/token/0.0.789013`
- A specific transaction: `https://hashscan.io/testnet/transaction/{txId}`

---

## Key SDK Operations Reference

```typescript
import {
  Client,
  AccountId,
  PrivateKey,
  TopicMessageSubmitTransaction,
  TokenMintTransaction,
  TokenGrantKycTransaction,
  TokenPauseTransaction,
  TransferTransaction,
  NftId,
} from "@hashgraph/sdk"

// Client setup
const client = Client.forTestnet()
client.setOperator(
  AccountId.fromString(process.env.HEDERA_ACCOUNT_ID!),
  PrivateKey.fromStringDer(process.env.HEDERA_PRIVATE_KEY!)
)

// Post proof to HCS
await new TopicMessageSubmitTransaction()
  .setTopicId(HEDERA_TOPIC_ID)
  .setMessage(JSON.stringify(proofPayload))
  .execute(client)

// Mint FlexCard NFT
const mintReceipt = await (
  await new TokenMintTransaction()
    .setTokenId(HEDERA_TOKEN_ID)
    .setMetadata([Buffer.from(JSON.stringify({ tier, roi, proofHash, expiry }))])
    .execute(client)
).getReceipt(client)

const serialNumber = mintReceipt.serials[0].toNumber()

// Grant KYC + transfer to user
await new TokenGrantKycTransaction()
  .setTokenId(HEDERA_TOKEN_ID)
  .setAccountId(userHederaAccountId)
  .execute(client)

await new TransferTransaction()
  .addNftTransfer(
    new NftId(HEDERA_TOKEN_ID, serialNumber),
    TREASURY_ACCOUNT_ID,
    userHederaAccountId
  )
  .execute(client)

// Pause badge on expiry
await new TokenPauseTransaction()
  .setTokenId(HEDERA_TOKEN_ID)
  .execute(client)
```

---

## Notes

- All testnet HBAR is free from the [Hedera faucet](https://portal.hedera.com)
- The treasury key should be kept offline in production — use a separate operator key for daily minting
- Mirror Node queries require no authentication on testnet
- For the demo video: show the Hashscan page for the topic with messages, and the NFT page showing the minted FlexCard

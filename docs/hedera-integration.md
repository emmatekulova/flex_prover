# Hedera Integration — FlexProver

## Why Hedera Is Here

FlexProver already solves the hard part: a Flare TEE fetches real Binance trade data inside a hardware enclave, attests it, and writes the result on-chain. What Hedera adds is two things the core stack does not natively cover:

1. **An immutable, queryable audit log** of every proof issuance — without smart contracts.
2. **A native NFT badge** (the FlexCard) that can carry compliance metadata and a royalty fee schedule out of the box.

Both fit into the existing flow as **add-on layers**: the TEE still does all the sensitive work, and Hedera receives signed outputs from it.

---

## What We Use

### 1. Hedera Consensus Service (HCS)

**What it is:** A public, ordered, timestamped message bus. Anyone can post a message to a topic. The Hedera network timestamps and sequences it, and it is readable forever via the Mirror Node API.

**What we post:** Immediately after a proof is verified on Flare (Coston2), the backend posts a structured message to a dedicated HCS topic:

```json
{
  "proofHash": "0x8f3a...e2d1",
  "walletAddress": "0x1234...5678",
  "ensName": "alice.eth",
  "tier": "gold",
  "roi": 47,
  "flareTxHash": "0xabc...def",
  "issuedAt": 1775224933
}
```

**Why this matters:**
- Creates a permanent, third-party-verifiable receipt for every credential
- Anyone can query `mirror.hashgraph.com/api/v1/topics/{topicId}/messages` to audit all issued proofs
- Zero smart contracts needed — pure SDK call
- Adds an independent timestamp outside of Flare, useful if either chain has downtime

**SDK call (no Solidity):**

```typescript
import { Client, TopicMessageSubmitTransaction } from "@hashgraph/sdk"

await new TopicMessageSubmitTransaction()
  .setTopicId(HEDERA_TOPIC_ID)
  .setMessage(JSON.stringify(auditPayload))
  .execute(client)
```

---

### 2. Hedera Token Service (HTS) — FlexCard NFT

**What it is:** Hedera's native token layer. Creates and manages fungible and non-fungible tokens at the protocol level — no EVM, no Solidity.

**What we do:** Each verified proof mints one HTS NFT representing the FlexCard badge:

| HTS Feature | How FlexProver Uses It |
|---|---|
| NFT metadata | Tier, ROI, proof hash, expiry encoded in token memo |
| KYC grant | Account receives a KYC flag (Binance KYC proven via the TEE flow) |
| Token pause | Badge can be paused if credential expires or is disputed |
| Royalty fee | 5% royalty on secondary transfers — keeps the badge non-trivially transferable |
| Account freeze | Freeze badge if wallet linked to sanctioned address |

**Minting flow:**

```typescript
import { TokenMintTransaction, NftId } from "@hashgraph/sdk"

// 1. Token already created at deploy time with KYC key, pause key, royalty fee
// 2. On proof verified:
const mintTx = await new TokenMintTransaction()
  .setTokenId(FLEX_CARD_TOKEN_ID)
  .setMetadata([Buffer.from(JSON.stringify({ tier, roi, proofHash, expiry }))])
  .execute(client)

// 3. Grant KYC to recipient wallet
await new TokenGrantKycTransaction()
  .setTokenId(FLEX_CARD_TOKEN_ID)
  .setAccountId(recipientHederaAccountId)
  .execute(client)

// 4. Transfer NFT to recipient
await new TransferTransaction()
  .addNftTransfer(new NftId(FLEX_CARD_TOKEN_ID, serialNumber), TREASURY_ID, recipientId)
  .execute(client)
```

---

## Which Bounty Tracks This Qualifies For

### Track A: "No Solidity Allowed" — $3,000 (up to 3 × $1,000)

**Requirement mapping:**

| Requirement | How We Meet It |
|---|---|
| Hedera JS/TS SDK only, no Solidity | `@hashgraph/sdk` throughout. Zero `.sol` files on Hedera. |
| Two native Hedera services | HCS (audit log) + HTS (FlexCard NFT) |
| Public GitHub repo with README | This repo |
| ≤5 min demo video | Show: HCS topic message posted after proof, HTS NFT minted, Mirror Node query |

**Optional enhancements we hit:**
- Mirror Node REST API — frontend queries all proofs for a wallet address
- HCS for data integrity — audit trail of every credential issuance
- Coherent end-to-end UX — the Hedera steps are invisible to the user; they just see "badge minted"

**This is the primary target track.** The "no Solidity" framing is a perfect fit: FlexProver's Hedera layer is 100% SDK-driven.

---

### Track B: "Tokenization on Hedera" — $2,500 (up to 2 × $1,250)

**What the "real-world asset" is here:**

A verified trading performance credential is an asset in the same sense as an invoice or a fund share. It represents a proven claim — audited by hardware — that a specific wallet achieved a specific ROI over a specific period. It has:
- A face value (tier: bronze / silver / gold)
- An expiry (180 days)
- An issuer (the Flare TEE, hardware-attested)
- Compliance controls (KYC grant = Binance-verified human behind it)

**Requirement mapping:**

| Requirement | How We Meet It |
|---|---|
| Create/manage tokens via HTS | HTS NFT creation, minting, KYC, pause, royalty fee |
| Deploy on Hedera Testnet | All transactions on testnet |
| Source code in public repo | This repo |
| Demo: creation + lifecycle op | Create token → mint → KYC grant → transfer → pause on expiry |

**Optional enhancements we hit:**
- KYC grants (Binance KYC → HTS KYC flag)
- Account freezing (expired credentials)
- Royalty fee schedule on FlexCard transfers
- Oracle integration possible: Chainlink for USD value of ROI

---

## Data Flow with Hedera

```
User proves trade
       │
       ▼
Flare TEE attests
       │
       ├──► Flare Coston2 tx (primary proof, on-chain)
       │
       ├──► HCS message posted (audit log, Mirror Node queryable)
       │
       └──► HTS NFT minted → KYC granted → transferred to user wallet
                                                │
                                                ▼
                                    User's HBAR account holds FlexCard NFT
                                    Anyone can verify via Mirror Node
```

---

## Why Not Earlier?

Hedera was considered and dropped early in design because we didn't want to add chains for their own sake. The reason it makes sense now:

- HCS genuinely adds something: a permanent, chain-independent audit log with sub-second finality that doesn't require a smart contract
- HTS genuinely adds something: native compliance controls (KYC, pause, freeze, royalty) that would require multiple Solidity contracts to replicate on EVM chains

It is not decorative. Both services replace things that would otherwise need custom contract code.

---

## Services Summary

| Service | Purpose | SDK calls used |
|---|---|---|
| HCS | Immutable proof audit log | `TopicCreateTransaction`, `TopicMessageSubmitTransaction` |
| HTS | FlexCard NFT with compliance | `TokenCreateTransaction`, `TokenMintTransaction`, `TokenGrantKycTransaction`, `TokenPauseTransaction`, `TransferTransaction` |
| Mirror Node | Frontend proof history query | REST `GET /api/v1/topics/{id}/messages` |

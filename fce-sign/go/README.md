# TEE Extension — Binance Portfolio Growth Attestation (Go)

A Flare TEE extension that fetches your Binance spot account's daily portfolio snapshots inside a Trusted Execution Environment, computes your portfolio growth percentage over the past 7 or 30 days, links the result to your on-chain wallet, signs it with the TEE's key, and publishes a permanent verifiable record on Flare Coston2 devnet.

---

## What it does

```
On-chain caller
  → publish-attestation tool sends credentials + wallet to extension
    → TEE fetches daily portfolio snapshots from Binance /sapi/v1/accountSnapshot
      → computes growthPercent from first to last snapshot (BTC-denominated)
        → signs JSON payload with TEE secp256k1 key
          → publish-attestation calls BinanceAttestationStore.publishAttestation()
            → AttestationPublished event emitted on Coston2 (payload as readable JSON)
```

Each run produces an `AttestationPublished` event. Compare `growthPercent` across multiple events over time to track portfolio performance with TEE-backed proof.

### What gets attested

```json
{
  "source": "binance-profile-growth",
  "wallet": "0xYourWalletAddress",
  "windowDays": 7,
  "startSnapshot": { "date": "2026-03-28", "totalBtc": "0.12345678" },
  "endSnapshot":   { "date": "2026-04-04", "totalBtc": "0.13456789" },
  "growthPercent": "9.00",
  "fetchedAt": 1712345678,
  "version": "0.1.0"
}
```

The `AttestationPublished` event contains:
- `teeAddress` — Ethereum address recovered from the TEE's ECDSA signature (verifiable against Flare TEE machine registry)
- `payload` — human-readable JSON string (shown as text in Coston2 explorer)
- `signature` — 65-byte secp256k1 sig over `keccak256(payload)`
- `timestamp` — block timestamp

---

## Prerequisites

- **Docker & Docker Compose**
- **Go 1.23+** and **Foundry** (`forge`, `cast`)
- **Cloudflared** or ngrok (to expose local port to the internet)
- **A funded Coston2 wallet** (C2FLR for gas + TEE fees — get from [Coston2 faucet](https://faucet.flare.network/))
- **Binance account** with API key enabled (read-only permissions are sufficient)

---

## Required environment variables

Copy `.env.example` to `.env` and fill in:

| Variable | Purpose |
|----------|---------|
| `PRIVATE_KEY` | Funded Coston2 private key, **no 0x prefix** |
| `INITIAL_OWNER` | Your Coston2 wallet address (`0x...`) |
| `INSTRUCTION_SENDER` | Set after Step 1 |
| `EXTENSION_ID` | Set after Step 2 |
| `TUNNEL_URL` | Public URL from cloudflared/ngrok |
| `ATTESTATION_STORE` | Optional — reuse an existing store contract |

> **Binance credentials are never stored in env vars.** Pass them directly as CLI flags at runtime.

Also copy the proxy config:
```bash
cp config/proxy/extension_proxy.toml.example config/proxy/extension_proxy.toml
# Fill in [db] section with your Coston2 C-chain indexer credentials
```

---

## One-time setup

### Step 1: Deploy the InstructionSender contract

```bash
cd fce-sign/go/tools
go run ./cmd/deploy-contract
```

Add the printed address to `.env`:
```bash
INSTRUCTION_SENDER="0x<printed address>"
```

### Step 2: Register the extension

```bash
go run ./cmd/register-extension --instructionSender $INSTRUCTION_SENDER
```

Add the printed extension ID to `.env`:
```bash
EXTENSION_ID="0x<printed id>"
```

### Step 3: Start the Docker stack

```bash
# from project root
docker compose build
docker compose up -d
```

Wait for the proxy to be ready:
```bash
until curl -sf http://localhost:6676/info >/dev/null 2>&1; do sleep 2; done
echo "proxy ready"
```

### Step 4: Start a tunnel

In a separate terminal (keep it running):
```bash
cloudflared tunnel --url http://localhost:6676
# or: ngrok http 6676
```

Add the printed URL to `.env`:
```bash
TUNNEL_URL="https://<your-tunnel>.trycloudflare.com"
```

### Step 5: Allow the TEE version

```bash
go run ./cmd/allow-tee-version -p http://localhost:6676
```

### Step 6: Register the TEE machine

```bash
go run ./cmd/register-tee -p http://localhost:6676 -l
# -l = local test mode (fake attestation token, required on Coston2 testnet)
```

---

## Local test (no chain required)

Verify the TEE extension produces a valid growth attestation without touching the chain:

```bash
cd fce-sign/go/tools
go run ./cmd/test-binance-attest \
  -apiKey    YOUR_BINANCE_API_KEY \
  -secretKey YOUR_BINANCE_SECRET_KEY \
  -wallet    0xYOUR_WALLET_ADDRESS
```

Optional flags: `-days 30` (default 7), `-url http://127.0.0.1:8883/action`

Example output:
```
✅ Binance profile growth attestation succeeded
wallet=0xYour... days=7
signature_len=65
payload={"source":"binance-profile-growth","wallet":"0xYour...","windowDays":7,"startSnapshot":{"date":"2026-03-28","totalBtc":"0.12345678"},"endSnapshot":{"date":"2026-04-04","totalBtc":"0.13456789"},"growthPercent":"9.00","fetchedAt":1712345678,"version":"0.1.0"}
```

---

## The main command — publish on-chain

```bash
cd fce-sign/go/tools
go run ./cmd/publish-attestation \
  -apiKey    YOUR_BINANCE_API_KEY \
  -secretKey YOUR_BINANCE_SECRET_KEY \
  -wallet    0xYOUR_WALLET_ADDRESS
```

Optional flags:
- `-days 30` — use 30-day window instead of 7
- `-store 0x<addr>` — reuse an existing BinanceAttestationStore (skip deploy; or set `ATTESTATION_STORE` in `.env`)

Example output:
```
Fetching Binance profile growth attestation from extension (window=7 days)...
  Payload (312 bytes): {"source":"binance-profile-growth",...}
  Signature (65 bytes): 7147...
No -store address given, deploying BinanceAttestationStore...
  BinanceAttestationStore deployed at: 0x<address>
  Set ATTESTATION_STORE=0x<address> in .env to reuse.
Publishing attestation on-chain...
  TX hash: 0x<hash>
  Recovered TEE address: 0x<tee address>
✅ Attestation published on-chain!
```

---

## How the TEE flow works

```
CLI tool (publish-attestation)
  │
  ├─ POSTs credentials + wallet to extension /action endpoint
  │    (inside Docker, runs in hardware TEE)
  │
  └─ Extension handler (handleBinanceProfileGrowth):
       1. Calls Binance /sapi/v1/accountSnapshot?type=SPOT&limit=7
       2. Computes growthPercent = (endBTC - startBTC) / startBTC × 100
       3. Builds JSON payload with wallet + growth data
       4. Signs keccak256(payload) with TEE secp256k1 key via internal /sign endpoint
       5. Returns ABI-encoded (payload, signature)

publish-attestation then:
  1. Decodes (payload, signature)
  2. Deploys BinanceAttestationStore if needed
  3. Calls publishAttestation(payloadString, signature) on-chain
  4. Contract recovers TEE address via ecrecover and emits AttestationPublished
```

The TEE address is verifiable against the Flare TEE machine registry — anyone can confirm the data came from a genuine TEE.

---

## How on-chain submission works

`BinanceAttestationStore` is a standalone Solidity contract on Coston2. Anyone can submit a valid `(payload, signature)` pair:

1. Contract calls `ecrecover(keccak256(payload), signature)` to recover the TEE address
2. Emits `AttestationPublished(teeAddress, payload, signature, timestamp)`
3. The `payload` field is `bytes` containing UTF-8 JSON — decode the hex as UTF-8 to read it

View your attestation on the [Coston2 explorer](https://coston2-explorer.flare.network/):
```
https://coston2-explorer.flare.network/tx/0x<tx hash>
```
Look for the `AttestationPublished` event. The `payload` log field shows the JSON directly.

---

## Devnet prerequisites

- **Coston2 faucet**: get free C2FLR at [faucet.flare.network](https://faucet.flare.network/)
- **Chain RPC**: `https://coston2-api.flare.network/ext/C/rpc` (default, no config needed)
- The Docker stack must be running and the tunnel must be active before running `publish-attestation`

---

## Tools reference

All commands run from `fce-sign/go/tools/`.

| Command | Purpose |
|---------|---------|
| `test-binance-attest` | Direct handler test — no chain required |
| `publish-attestation` | **Main command** — fetch growth, deploy store if needed, publish on-chain |
| `deploy-contract` | One-time: deploy InstructionSender |
| `register-extension` | One-time: register extension on Flare TEE registry |
| `allow-tee-version` | One-time: whitelist TEE code hash + platform |
| `register-tee` | One-time: register TEE machine |
| `run-test` | E2E test: ECIES key update + sign + verify |

---

## Troubleshooting

| Error | Fix |
|-------|-----|
| `apiKey and secretKey are required` | Pass `-apiKey` and `-secretKey` flags |
| `wallet address is required` | Pass `-wallet 0x...` flag |
| `not enough snapshots returned` | Binance account may have no trading history; check account has been active |
| `extension ID not set` | Run `go run ./cmd/run-test` first (calls `setExtensionId`) |
| `501 unsupported op type` | Rebuild Docker image: `docker compose build extension-tee && docker compose up -d extension-tee` |
| Binance `400` on snapshot | Account may not have spot wallet history; ensure spot account is activated |
| `FeeTooLow` revert | Increase `FEE_WEI` in `.env` |
| Tunnel drops | Restart cloudflared, update `TUNNEL_URL`, re-run steps 5–6 |
| `ecrecover failed` | Signature length must be exactly 65 bytes |

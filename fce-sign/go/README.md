# TEE Extension — Binance Attestation (Go)

A TEE extension that fetches live Binance account and market data inside a Trusted Execution Environment, signs it with the TEE's key, and publishes the attestation on-chain.

## What it does

```
On-chain caller
  → InstructionSender.fetchBinanceUserProfileAndAttest()
    → TEE receives instruction
      → fetches account data from Binance API (inside TEE)
      → signs JSON payload with TEE key
      → returns ABI-encoded (payload, signature)
        → publish-attestation tool calls BinanceAttestationStore.publishAttestation()
          → AttestationPublished event emitted on Coston2
```

The `AttestationPublished` event contains:
- `teeAddress` — Ethereum address recovered from the TEE's ECDSA signature
- `payload` — raw JSON (UID, account type, balances, total USD value, etc.)
- `signature` — 65-byte secp256k1 sig over `keccak256(payload)`
- `timestamp` — block timestamp

Anyone can verify the payload was produced inside a genuine TEE by checking `teeAddress` against the Flare TEE machine registry.

---

## Prerequisites

- **Docker & Docker Compose**
- **Go 1.23+** and **Foundry** (`forge`, `cast`)
- **Cloudflared** or ngrok (to tunnel local port to internet)
- **A funded Coston2 wallet** (C2FLR for gas + TEE fees)

---

## Setup

### Step 0: Configure environment

```bash
cp .env.example .env
```

Fill in `.env`:
```bash
PRIVATE_KEY="<your funded Coston2 private key, no 0x>"
INITIAL_OWNER="0x<your wallet address>"
LANGUAGE=go
```

> **API keys are no longer stored in `.env`.** They are passed per-request to the
> `attest` tool and ECIES-encrypted with the TEE's public key before being sent
> on-chain. The TEE decrypts them inside the secure enclave — they never travel
> in plaintext outside the TEE.

Also copy the proxy config:
```bash
cp config/proxy/extension_proxy.toml.example config/proxy/extension_proxy.toml
# Fill in [db] section with your Coston2 C-chain indexer credentials
```

### Step 1: Deploy the InstructionSender contract

```bash
cd go/tools
go run ./cmd/deploy-contract
```

Copy the printed address to `.env`:
```bash
INSTRUCTION_SENDER="0x<printed address>"
```

### Step 2: Register the extension

```bash
go run ./cmd/register-extension --instructionSender $INSTRUCTION_SENDER
```

Copy the printed extension ID to `.env`:
```bash
EXTENSION_ID="0x<printed id>"
```

### Step 3: Start the Docker stack

```bash
# from project root
docker compose build
docker compose up -d
```

Wait for the proxy:
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

Copy the printed URL to `.env`:
```bash
TUNNEL_URL="https://<your-tunnel>.trycloudflare.com"
```

### Step 5: Allow TEE version

```bash
cd go/tools
go run ./cmd/allow-tee-version -p http://localhost:6676
```

### Step 6: Register the TEE machine

```bash
go run ./cmd/register-tee -p http://localhost:6676 -l
# -l = local test mode (fake attestation token, required on Coston2 testnet)
```

### Step 7: Run the end-to-end test

```bash
go run ./cmd/run-test --instructionSender $INSTRUCTION_SENDER -p http://localhost:6676
```

---

## Triggering a CEX attestation on-chain

After completing the one-time setup (steps 0–7), use the `attest` tool to request a
signed TEE attestation at any time:

```bash
cd go/tools

# Public endpoint — no API key required:
go run ./cmd/attest -mode ticker -symbol BTCUSDT

# Authenticated — credentials are ECIES-encrypted before sending on-chain:
go run ./cmd/attest -mode profile -apiKey YOUR_KEY -secretKey YOUR_SECRET
go run ./cmd/attest -mode account -apiKey YOUR_KEY -secretKey YOUR_SECRET
go run ./cmd/attest -mode pnl    -apiKey YOUR_KEY -secretKey YOUR_SECRET
go run ./cmd/attest -mode growth -lookbackDays 7 -apiKey YOUR_KEY -secretKey YOUR_SECRET
go run ./cmd/attest -mode stats  -symbol ETHUSDT
```

The tool prints the transaction hash, instruction ID, payload, and TEE signature:
```
INF Credentials encrypted (113 bytes ciphertext)
INF Sending profile attestation for CEX=binance...
INF Transaction submitted: 0x<txHash>
INF Explorer: https://coston2-explorer.flare.network/tx/0x<txHash>
INF Instruction ID: 0x<instructionId>
INF Waiting for TEE result...
INF === Attestation Result ===
INF Instruction ID:    0x<instructionId>
INF Signature (hex):   7147...
INF Payload:           {"source":"binance-user-profile","uid":...}
```

**Available modes:**

| Mode | Binance endpoint | Auth |
|------|-----------------|------|
| `ticker` | `/api/v3/ticker/price` | No |
| `stats` | `/api/v3/ticker/24hr` | No |
| `account` | `/api/v3/account` | Yes |
| `pnl` | `/fapi/v2/account` | Yes + futures |
| `profile` | `/api/v3/account` (full profile) | Yes |
| `growth` | `/sapi/v1/accountSnapshot?type=SPOT` | Yes |

## Local handler test (no chain)

Test any handler directly against the running extension without sending a chain transaction:

```bash
cd go/tools

# Current ticker price
go run ./cmd/test-binance-attest -mode ticker -symbol BTCUSDT

# 24h market stats
go run ./cmd/test-binance-attest -mode stats -symbol BTCUSDT

# Spot account balances (pass keys directly — NOT encrypted in this test tool)
go run ./cmd/test-binance-attest -mode account -apiKey YOUR_KEY -secretKey YOUR_SECRET

# Futures PnL
go run ./cmd/test-binance-attest -mode pnl -apiKey YOUR_KEY -secretKey YOUR_SECRET

# Full user profile
go run ./cmd/test-binance-attest -mode profile -apiKey YOUR_KEY -secretKey YOUR_SECRET
```

> **Warning**: `test-binance-attest` sends credentials as plaintext hex for local testing only.
> Use `cmd/attest` for production — it ECIES-encrypts credentials with the TEE's public key.

Each mode prints the signed payload and signature. Example for `profile`:
```
✅ CEX attestation + TEE sign succeeded
cex=binance mode=profile
payload={"source":"binance-user-profile","uid":1228038409,"accountType":"SPOT",...}
signature_len=65
```

---

## Publishing attestations on-chain

`BinanceAttestationStore` is a standalone contract that verifies TEE signatures and emits permanent on-chain events. Anyone can submit a valid (payload, signature) pair.

### Deploy the store and publish in one step

```bash
cd go/tools
go run ./cmd/publish-attestation
```

Output:
```
Fetching Binance user profile attestation from extension...
  Payload (398 bytes): {"source":"binance-user-profile","uid":...}
  Signature (65 bytes): 7147...
No -store address given, deploying BinanceAttestationStore...
  BinanceAttestationStore deployed at: 0x<address>
  Set ATTESTATION_STORE=0x<address> in .env to reuse.
Publishing attestation on-chain...
  TX hash: 0x<hash>
  Recovered TEE address: 0x<tee address>
✅ Attestation published on-chain!
```

Add the store address to `.env` so subsequent runs skip deployment:
```bash
ATTESTATION_STORE="0x<address>"
```

### View on Coston2 explorer

```
https://coston2-explorer.flare.network/tx/0x<tx hash>
```

Look for the `AttestationPublished` event in the transaction logs. The decoded event shows the TEE address, payload, and timestamp.

### Re-use an existing store

```bash
go run ./cmd/publish-attestation -store 0x<address>
# or set ATTESTATION_STORE in .env
```

---

## Supported op commands

All handlers live under `OP_TYPE = MARKET`. The message payload is a JSON `CEXRequest`:

```json
{
  "cex": "binance",
  "encryptedCredentials": "0x<ecies-ciphertext-hex>",
  "symbol": "BTCUSDT"
}
```

Generic op commands (preferred for new integrations):

| `opCommand` | Data fetched | Auth |
|---|---|---|
| `FETCH_AND_ATTEST` | Current ticker price | No |
| `24H_STATS` | 24h market stats | No |
| `ACCOUNT_SUMMARY` | Spot balances + total USDT | Yes |
| `ACCOUNT_PNL` | Futures wallet + unrealised PnL | Yes + futures |
| `USER_PROFILE` | UID, account type, permissions, balances | Yes |

Binance-prefixed aliases (backward compatible with deployed contracts):
`BINANCE_FETCH_AND_ATTEST`, `BINANCE_24H_STATS`, `BINANCE_ACCOUNT_SUMMARY`,
`BINANCE_ACCOUNT_PNL`, `BINANCE_USER_PROFILE` — all route to the same handlers.

All handlers return ABI-encoded `(bytes payload, bytes signature)` where `signature` is a 65-byte secp256k1 ECDSA signature over `keccak256(payload)`.

---

## Contracts

| Contract | Purpose |
|---|---|
| `InstructionSender.sol` | Sends instructions to the TEE; one function per op command |
| `BinanceAttestationStore.sol` | Verifies TEE sig on-chain, emits `AttestationPublished` event |

### Regenerating Go bindings after changing InstructionSender.sol

```bash
# Compile
cd contract && forge build

# Extract ABI + bytecode
jq -r '.abi' out/InstructionSender.sol/InstructionSender.json > ../go/tools/app/contract/InstructionSender.abi
jq -r '.bytecode.object' out/InstructionSender.sol/InstructionSender.json > ../go/tools/app/contract/InstructionSender.bin

# Regenerate Go bindings
cd ../go/tools && go generate ./...
```

---

## Extending this for your hackathon project

Modify files in `internal/app/` to build your own TEE extension:

| File | What to change |
|------|----------------|
| `internal/app/handlers.go` | Add your handler functions, register them in `Register()` |
| `internal/app/config.go` | Add your `OpType`/`OpCommand` constants (must match Solidity) |
| `internal/app/types.go` | Add request/response types |
| `contract/InstructionSender.sol` | Add a function for each new op command |

Handler signature:
```go
func myHandler(msg string) (data *string, status int, err error) {
    // msg: hex-encoded originalMessage from the on-chain instruction
    // Return: ABI-encoded result hex, status (0=error, 1=success, >=2=pending), error
    return &dataHex, 1, nil
}
```

The `internal/base/` package is framework infrastructure — don't modify it.

---

## Tools reference

All commands run from `go/tools/`. They read `PRIVATE_KEY`, `CHAIN_URL`, and `ADDRESSES_FILE` from `.env` at the repo root.

| Command | Purpose | One-time? |
|---|---|---|
| `deploy-contract` | Deploy InstructionSender | Yes |
| `register-extension` | Register extension on Flare TEE registry | Yes |
| `allow-tee-version` | Whitelist TEE code hash + platform | Yes |
| `register-tee` | Register TEE machine (pre-reg → attest → produce) | Yes |
| `run-test` | End-to-end test (ECIES key update + sign + verify) | Yes |
| `attest` | **Trigger a CEX attestation on-chain** (ECIES-encrypts API keys) | **Ongoing** |
| `test-binance-attest` | Direct handler test, no chain, plaintext keys (local only) | Dev |
| `publish-attestation` | Fetch profile + deploy store + publish on-chain | Ongoing |

---

## Troubleshooting

| Error | Fix |
|-------|-----|
| `extension ID not set` | Run `go run ./cmd/run-test` (calls `setExtensionId` first) |
| `501 unsupported op type` | Rebuild Docker image: `docker compose build extension-tee && docker compose up -d extension-tee` |
| Binance 400 error | Symbol format should be `BTCUSDT` not `BTC/USDT` |
| `FeeTooLow` revert | Increase `FEE_WEI` in `.env` |
| Tunnel drops | Restart cloudflared, update `TUNNEL_URL`, re-run steps 5–6 |
| `ecrecover failed` | Signature length must be exactly 65 bytes |

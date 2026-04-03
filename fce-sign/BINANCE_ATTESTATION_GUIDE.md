# Binance Price Attestation via TEE — Complete Setup Guide

This guide walks you through deploying and running the TEE extension that fetches Binance ticker data and returns it with a cryptographic signature from the TEE.

## Quick Overview

The flow is:

```
1. Deploy InstructionSender contract → get contract address
2. Register extension on registry → get extension ID  
3. Start Docker stack (TEE + extension) → extension listens on :8080
4. Start tunnel (cloudflared) → expose extension to internet
5. Register TEE machine → attests publicly on-chain
6. Call contract.fetchBinanceAndAttest(symbol) → Binance data + TEE signature
```

## Prerequisites

- **Docker & Docker Compose**
- **Go 1.23+** (for deployment tools; Nix shell covers this: `nix-shell shell.nix`)
- **Foundry** (`forge`, `cast` for contracts)
- **Cloudflared** or ngrok (to tunnel local port to internet)
- **A funded Coston2 wallet** (C2FLR for gas + TEE fees)
- **jq** (for JSON parsing)

Everything is in the Nix shell:
```bash
nix-shell shell.nix   # or: nix develop
```

---

## Step 0: Configure Environment

Copy the example env:
```bash
cp .env.example .env
```

Edit `.env` and fill in:

```bash
# REQUIRED: Your funded Coston2 private key (no 0x prefix)
PRIVATE_KEY="abc123def456..."

# REQUIRED: Your wallet address (derived from PRIVATE_KEY)
INITIAL_OWNER="0x1234567890abcdef..."

# OPTIONAL: Binance API credentials (leave empty for public API)
BINANCE_API_KEY=""
BINANCE_SECRET_KEY=""

# Use go, python, or typescript
LANGUAGE=go

# Other values filled in after each step
INSTRUCTION_SENDER=""
EXTENSION_ID=""
TUNNEL_URL=""
```

Also update the proxy config:
```bash
cp config/proxy/extension_proxy.toml.example config/proxy/extension_proxy.toml
```

Edit `config/proxy/extension_proxy.toml` and fill in the `[db]` section with your Coston2 C-chain indexer credentials.

---

## Step 1: Deploy the InstructionSender Contract

The contract is what callers interact with on-chain. It sends instructions to your TEE machines.

```bash
cd go/tools
go run ./cmd/deploy-contract
```

Output will be:
```
Deployed InstructionSender at: 0x1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a
Verification: https://coston2-explorer.flare.network/address/0x1a2b3c4d5e6f...
```

**Copy the address** and update `.env`:
```bash
INSTRUCTION_SENDER="0x1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a"
```

---

## Step 2: Register the Extension

Register your extension with the registry so the TEE system recognizes it:

```bash
go run ./cmd/register-extension --instructionSender 0x1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a
```

Output will be:
```
Registered extension with ID: 0x0000000000000000000000000000000000000000000000000000000000000001
```

**Update `.env`**:
```bash
EXTENSION_ID="0x0000000000000000000000000000000000000000000000000000000000000001"
```

---

## Step 3: Start the Docker Stack

Now build and start the TEE node + your extension in one container:

```bash
docker compose build
docker compose up -d
```

Wait for the proxy to be healthy:
```bash
until curl -sf http://localhost:6676/info >/dev/null 2>&1; do sleep 2; done
echo "Extension proxy is ready"
```

The extension listens internally; the proxy exposes it on port 6676.

---

## Step 4: Start a Tunnel

Expose your local proxy to the internet. In a separate terminal:

**Using cloudflared** (no account needed):
```bash
cloudflared tunnel --url http://localhost:6676
```

**Or using ngrok** (requires signup):
```bash
ngrok http 6676
```

You'll see:
```
Cloudflare quick tunnel URL: https://abc123def456.trycloudflare.com
```

**Copy the URL** and update `.env`:
```bash
TUNNEL_URL="https://abc123def456.trycloudflare.com"
```

> **Keep this terminal open** — the tunnel must stay running for the entire session.

---

## Step 5: Allow TEE Version

Tell the registry that this TEE version's code is allowed:

```bash
cd go/tools
go run ./cmd/allow-tee-version -p http://localhost:6676
```

---

## Step 6: Register the TEE Machine

Register your TEE as a machine available for your extension:

```bash
go run ./cmd/register-tee -p http://localhost:6676 -l
```

The `-l` flag means "local test mode" (uses a fake attestation token; required on Coston2 testnet).

---

## Step 7: Run the End-to-End Test

Test the full flow: deploy private key, fetch Binance data, verify signature:

```bash
go run ./cmd/run-test -p http://localhost:6676
```

Output should show:
```
✓ setExtensionId
✓ updateKey (encrypted private key)
✓ Binance fetch + attestation
✓ sign verification passed
```

---

## How It Works: Binance Attestation

When you call the contract:

```solidity
bytes32 tx = instructionSender.fetchBinanceAndAttest(
    abi.encode(json.encode({"symbol":"BTCUSDT"}))
);
// Wait for TEE to process...
// Result: (attestationPayload, TEESignature)
```

The TEE:

1. **Receives** the instruction (`MARKET/BINANCE_FETCH_AND_ATTEST`)
2. **Parses** the symbol from the message
3. **Fetches** real-time price from Binance API:
   ```
   https://api.binance.com/api/v3/ticker/price?symbol=BTCUSDT
   ```
4. **Builds** an attestation payload:
   ```json
   {
     "source": "binance",
     "symbol": "BTCUSDT",
     "price": "65000.12",
     "fetchedAt": 1712188890,
     "version": "0.1.0"
   }
   ```
5. **Signs** the payload with the TEE's private key (via `/sign` endpoint)
6. **Returns** ABI-encoded `(payload, signature)` as proof

On-chain, the caller can verify the signature against the TEE's public key to confirm:
- Price came fresh from Binance
- TEE attested to it
- Data wasn't tampered with

---

## Using Binance API Credentials (Optional)

If you want **authenticated requests** (higher rate limits):

1. Create a Binance API key at https://www.binance.com/en/account/api-management
2. Set in `.env`:
   ```bash
   BINANCE_API_KEY="your-key-here"
   BINANCE_SECRET_KEY="your-secret-here"
   ```
3. Restart the extension:
   ```bash
   docker compose restart extension-go
   ```

The extension will now include `X-MBX-APIKEY` header in Binance requests. (Full request signing with HMAC-SHA256 can be added if needed.)

---

## Testing Locally (Without Coston2)

To test the handler locally without on-chain deployment:

```bash
cd go
go test ./internal/app/...
```

This runs unit tests with mocked Binance and TEE sign endpoints.

---

## Troubleshooting

| Issue | Solution |
|-------|----------|
| "extension ID not set" | Run `setExtensionId()` on the contract first (`run-test` does this) |
| Binance 400 error | Wrong symbol format (should be `BTCUSDT` not `BTC/USDT`) |
| Signature verification fails | Ensure TEE machine is registered and healthy (`docker logs extension-go`) |
| Tunnel connection drops | Restart cloudflared and update `TUNNEL_URL` in `.env`, then re-run steps 5-6 |
| Docker stack won't start | Check proxy logs: `docker compose logs ext-proxy` |

---

## Cleanup

Stop the stack:
```bash
docker compose down
```

Full reset (rebuild images):
```bash
docker compose down --rmi local
rm -f .env config/proxy/extension_proxy.toml
```

---

## Next: Deploy to Production

On a live testnet/mainnet:
- Use real attestation (remove `-l` flag in `register-tee`)
- Use a real tunnel URL (not cloudflared quick tunnel)
- Scale to multiple TEE machines for availability
- Monitor `/state` endpoint for stale data

See [README.md](README.md) for the full architecture.

# flex_prover

Trustless proof of trading performance using TEE attestation on Flare.

## Problem

Traders often need to prove their performance (PnL, win rate, etc.) to others — for credibility, fund management, or competitions. Current options are either too broad or too fragile:

- **Sharing read-only API keys** exposes your entire account: balances, positions, order history — far more than you want to reveal.
- **Screenshots** are trivially faked.
- **Self-reported stats** require trust in the person claiming them.

## Solution

Flex Prover uses a **Trusted Execution Environment (TEE)** on Flare's Confidential Compute (FCC) framework to fetch specific trading data from Binance and attest it on-chain. The TEE guarantees that the data was genuinely retrieved from Binance's API and was not tampered with — no trust in the prover required.

You choose exactly what to share. Nothing more.

### How It Works

```
User                    Flare (on-chain)              TEE Enclave                 Binance
 │                           │                            │                          │
 │  1. Encrypt API creds     │                            │                          │
 │   with TEE public key     │                            │                          │
 │──────────────────────────>│  2. Instruction emitted    │                          │
 │                           │───────────────────────────>│                          │
 │                           │                            │  3. Decrypt creds        │
 │                           │                            │  4. Call Binance API ───>│
 │                           │                            │  5. Receive stats  <─────│
 │                           │                            │  6. Extract requested    │
 │                           │                            │     metrics only         │
 │                           │                            │  7. Sign & return ──────>│
 │                           │  8. Attested result stored │                          │
 │  9. Anyone can verify <───│                            │                          │
```

1. The user encrypts their Binance read-only API credentials using the TEE's public key (ECIES) and sends them on-chain via the `InstructionSender` contract.
2. The TEE extension decrypts the credentials inside the secure enclave — they are never visible in plaintext on-chain or to any third party.
3. The TEE calls Binance's API, fetches the requested trading stats, and extracts only the specific metrics the user wants to share (e.g. 30-day PnL).
4. The result is signed and written back on-chain as an attested data point.
5. Anyone can verify the attestation: the TEE hardware proves the code wasn't tampered with, and the open-source handler code proves the data came from Binance.

### Trust Model

The security relies on two pillars:

- **TEE attestation** — The hardware (via Flare's FCC framework) cryptographically proves that the code running inside the enclave matches a known, registered version. No one — not even the server operator — can modify what runs inside.
- **Open-source handler code** — Anyone can inspect the extension source to confirm it genuinely calls Binance's API and doesn't fabricate data. The registered TEE version hash ties the running code to the auditable source.

Together: *"I can see what the code does + the TEE proves that exact code ran = I trust the output."*

## Setup and run

### frontend

Requires Docker, Foundry, and a funded Coston2 wallet. See the [FCC guide](https://dev.flare.network/fcc/guides/sign-extension) for detailed prerequisites.

```bash
# 1. Configure
cp .env.example .env
cp config/proxy/extension_proxy.toml.example config/proxy/extension_proxy.toml
# Edit .env with your private key and settings

# 2. Deploy contract
cd fce-sign/go/tools
go run ./cmd/deploy-contract
# → copy the printed address into INSTRUCTION_SENDER in .env

# 3. Register extension
go run ./cmd/register-extension
# → copy the printed extension ID into EXTENSION_ID in .env

# 4. Start the stack
cd ../../..   # back to repo root
docker compose build && docker compose up -d

# 5. Expose via tunnel
cloudflared tunnel --url http://localhost:6676
# → copy the https://... URL into TUNNEL_URL in .env

# 6. Register TEE version + machine
cd fce-sign/go/tools
go run ./cmd/allow-tee-version
go run ./cmd/register-tee

# 7. Run an attestation (from repo root)
#    The attest command automatically sets the extension ID on first run.
./run-attest.sh -mode growth -lookbackDays 7 -apiKey YOUR_KEY -secretKey YOUR_SECRET
# or, using flags directly from the tools directory:
cd fce-sign/go/tools
go run ./cmd/attest -mode growth -lookbackDays 7 -apiKey YOUR_KEY -secretKey YOUR_SECRET
```

### Contract verification on Flare Coston2 Explorer

After deploying `InstructionSender`, verify it on [Coston2 Explorer](https://coston2-explorer.flare.network/) so that the ABI and source are publicly visible:

```bash
# From fce-sign/contract/:
forge verify-contract \
  --chain-id 114 \
  --rpc-url https://coston2-api.flare.network/ext/C/rpc \
  --etherscan-api-key any \
  --verifier blockscout \
  --verifier-url https://coston2-explorer.flare.network/api/ \
  <INSTRUCTION_SENDER_ADDRESS> \
  InstructionSender.sol:InstructionSender \
  --constructor-args $(cast abi-encode "constructor(address,address)" <TEE_EXTENSION_REGISTRY> <TEE_MACHINE_REGISTRY>)
```

Replace `<INSTRUCTION_SENDER_ADDRESS>`, `<TEE_EXTENSION_REGISTRY>`, and `<TEE_MACHINE_REGISTRY>` with the values from your `.env` and `config/coston2/deployed-addresses.json`.

## Status

Early development — built on Flare's FCC testnet (Coston2). The FCC framework is still under active development by Flare.

## License

MIT

## Quick start

1. Go to frontend:
	- `cd frontend`

2. Install dependencies:
	- `npm install`

3. Run development server:
	- `npm run dev`

## LAN dev access (optional)

If you test from another device on your local network, add your local IP to `allowedDevOrigins` in [frontend/next.config.mjs](frontend/next.config.mjs):

- `allowedDevOrigins: ['192.168.x.x']`

## Built On

Adapted from here: https://dev.flare.network/fcc/guides/sign-extension#step-5-add-tee-version

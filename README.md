# flex_prover

Trustless proof of trading performance using TEE attestation on Flare.

## Problem

Traders are seeking to showcase their performance to the public or potential employers and face a common issue: the lack of a verifiable, trustless standard for sharing authenticated trades.

The current options all fail in the same ways:

- **Sharing read-only API keys** gives the other party access to your entire account: every position, every balance, full order history, linked addresses.
- **Screenshots** prove nothing. Any number on a screenshot can be edited in thirty seconds.
- **Self-reported stats** are the same problem with more steps. A spreadsheet you filled in is not a proof, it is a claim.
- **Third-party trackers** require you to permanently connect your exchange account to an external service, which then owns that relationship indefinitely.

## Solution

Flex Prover uses a **Trusted Execution Environment (TEE)** on Flare's Confidential Compute (FCC) framework to fetch specific trading data from Binance and attest it on-chain. The TEE guarantees that the data was genuinely retrieved from Binance's API and was not tampered with — no trust in the prover required.

You choose exactly what to share. Nothing more.

### UML
<img width="1253" height="611" alt="UML_diag" src="https://github.com/user-attachments/assets/1288ffe1-c5ea-4091-850c-c6844e669f9f" />


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
cd <language>/tools
# run deploy-contract command for your language

# 3. Register extension
# run register-extension command

# 4. Start the stack
docker compose build && docker compose up -d

# 5. Expose via tunnel
cloudflared tunnel --url http://localhost:6676

# 6. Register TEE version + machine
# run allow-tee-version and register-tee commands

# 7. Test
# run run-test command
```

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

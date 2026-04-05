<p align="center">
  <img src="images/cover_flexProver.png" alt="FlexProver Cover" width="800">
</p>

# Flex Prover

### Trustless, Verifiable Trading Performance via Flare Confidential Compute (FCC)

Flex Prover is a decentralized platform that enables traders to generate cryptographically verifiable proofs of their trading performance without compromising their security or privacy. By leveraging Trusted Execution Environments (TEEs) on the Flare Network, Flex Prover ensures that trading data is fetched directly from exchanges (Binance, Bitget) and attested on-chain in a tamper-proof manner.

---

## The Problem

Traders who wish to showcase their performance to investors, employers, or the public currently face four major hurdles:
1. **Security Risks**: Sharing "read-only" API keys exposes entire account histories and balances.
2. **Fragile Trust**: Screenshots are trivial to manipulate using basic browser inspection tools.
3. **Lack of Standard**: Self-reported spreadsheets offer no proof of origin.
4. **Third-Party Dependency**: Connecting accounts to centralized trackers often leads to data harvesting and privacy loss.

## The Flex Prover Solution

Flex Prover uses **Flare's Confidential Compute (FCC)** framework to move the trust from individuals to hardware-secured enclaves.

- **Selective Disclosure**: Choose exactly what to share (e.g., 30-day PnL % or specific holdings) and nothing more.
- **Hardware-Level Security**: API credentials are encrypted via ECIES and only decrypted inside the TEE; they are never visible to Flex Prover operators or the blockchain.
- **Immutable Attestation**: Every "flex" is signed by the TEE's unique hardware key and stored on the Flare blockchain for anyone to verify.

---

## Architecture & Logic

### System Components

1.  **TEE Enclave (`fce-sign/go`)**: A secure Go-based backend running inside an FCC-compatible enclave. It handles:
    *   ECIES decryption of exchange credentials.
    *   Secure API communication with Binance and Bitget.
    *   Calculation of performance metrics (Growth %, PnL, Asset Valuation).
    *   Signing of the resulting attestation payload using a hardware-protected key.
2.  **Smart Contracts (`fce-sign/contract`)**:
    *   `InstructionSender.sol`: Orchestrates the flow between users and TEE machines.
    *   `BinanceAttestationStore.sol`: A permanent on-chain registry for verified proofs.
3.  **Frontend (`frontend`)**: A Next.js 14 application for:
    *   **Multi-Chain Connectivity**: Integrated with **Reown AppKit** (formerly WalletConnect) to support seamless wallet connection across both **EVM** (Flare, Ethereum, Hedera) and **non-EVM** (**Solana**) ecosystems.
    *   **Reown Authentication (SIWX)**: Implements secure, cross-chain identity verification using Sign-In With X (SIWX). This allows users to bind their exchange performance to a unified identity regardless of the underlying blockchain.
    *   Step-by-step wizard for credential encryption and proof generation.
    *   Visual "Proof Card" generation and verification tools.

### UML
<img width="1253" height="611" alt="UML_diag" src="https://github.com/user-attachments/assets/1288ffe1-c5ea-4091-850c-c6844e669f9f" />


### Reown AppKit Integration

Flex Prover leverages the **Reown AppKit SDK** to provide a world-class on-chain UX:
- **Ecosystem Interoperability**: Users can connect wallets from disparate ecosystems (e.g., MetaMask for Flare/EVM and Phantom for Solana) within the same interface.
- **Unified Auth**: Using **SIWX**, we provide a sybil-resistant authentication flow that works natively across Ethereum (SIWE) and Solana.
- **Social Login & On-Ramp**: AppKit's extensible features ensure Flex Prover remains accessible to both crypto-native and retail traders.


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


Security is rooted in the **Flare FCC Framework**:
1. **Hardware Integrity**: TEE attestation proves the exact code version running in the enclave.
2. **Open Source Auditor**: The `fce-sign/go` handler code is open for inspection to confirm it only extracts the requested metrics.
3. **Data Sovereignty**: Credentials never leave the enclave's encrypted memory and are never persisted.

### Hedera Token Service (HTS) Implementation

| HTS Feature | When | For What |
|---|---|---|
| **TokenMintTransaction** | When a user submits a verified Binance proof. | Mints **1 WHALE** token to the treasury account. |
| **TransferTransaction** | Immediately after the minting is successful. | Sends **1 WHALE** from the treasury to the user's Hedera account. |
| **FUNGIBLE_COMMON Token** | Created once at deployment time. | The official **WHALE** badge token itself (`0.0.8515448`). |

---

## Features

*   **Multi-Exchange Support**: Full integration for **Binance** (Spot & Futures) and **Bitget** (Spot & Futures).
*   **Performance Metrics**:
    *   **Portfolio Growth**: 7-day or 30-day BTC/USDT denominated growth.
    *   **Individual Trades**: Attest specific open positions and their current value.
    *   **Account Summary**: Proof of account type, permissions, and total estimated value.
*   **On-Chain Verification**: Simple `ecrecover` logic in Solidity allows any third party to verify the TEE signature.

---

## Setup & Development

### Prerequisites
*   [Foundry](https://book.getfoundry.sh/) (Forge/Cast)
*   [Go 1.23+](https://golang.org/)
*   [Docker](https://www.docker.com/)
*   Funded account on **Coston2 Testnet**
  
See the [FCC guide](https://dev.flare.network/fcc/guides/sign-extension) for detailed prerequisites.


### 1. Cloudflare Tunnel
In your terminal, run the following commands to install and start the tunnel:

```bash
sudo pacman -S cloudflared
cloudflared tunnel --url http://localhost:6676
```
Leave this process running. Copy the generated URL from the output and add it to your .env file.

### 2. Docker

Open a new terminal at the project root and execute:
```bash
docker compose build
docker compose up
```
Leave this process running.

### 3. Frontend

Open a new terminal, navigate to the flex../frontend directory, and run:
```bash
npm install
npm run dev
```

### 4. Backend

Open a new terminal, navigate to the flex../backend directory, and run:
```bash
go run .
```


## License

This project is licensed under the MIT License.

## Built On

Developed for **ETH Global Cannes 2026**. Built using:
- **[Flare Network FCC Framework](https://dev.flare.network/fcc/overview)**: Trustless computation and hardware-secured enclaves.
- **[Reown AppKit SDK](https://reown.com)**: Advanced multi-chain wallet connectivity and cross-chain authentication (WalletConnect).

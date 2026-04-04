# Reown AppKit + SIWX Integration Design

**Date:** 2026-04-04  
**Project:** FlexProver — Proof of Whale  
**Scope:** Replace mock wallet connection with Reown AppKit (wagmi adapter) + server-verified SIWX authentication

---

## Summary

Upgrade the frontend wallet connection from a mock `connectWallet()` stub to a production-ready Reown AppKit integration. The user must connect their wallet **and** sign a SIWX (Sign-In with X) message. The signature is verified server-side in a Next.js API route before the wallet address is trusted. The verified address is stored in a signed iron-session cookie and proxied to the Flare TEE backend — the TEE never receives an unverified address from the client.

Multi-chain support covers Flare (primary, chain ID 14), Ethereum mainnet, and Solana mainnet to satisfy multi-chain prize requirements.

---

## Packages

```
@reown/appkit
@reown/appkit-adapter-wagmi
@reown/appkit-adapter-solana
@reown/appkit-siwx
wagmi
viem
@solana/web3.js
tweetnacl          — ed25519 signature verification for Solana SIWX
iron-session
```

---

## File Structure

### New files

| File | Purpose |
|------|---------|
| `frontend/lib/reown.ts` | AppKit config: chains, adapters, SIWX plugin, project ID |
| `frontend/app/providers.tsx` | Client-side `<WagmiProvider>` + `<QueryClientProvider>` wrapper |
| `frontend/hooks/use-wallet.ts` | Hook exposing `address`, `isConnected`, `isAuthenticated`, `open`, `disconnect` |
| `frontend/app/api/auth/siwx/route.ts` | SIWX nonce generation + signature verification |
| `frontend/app/api/flare/pnl/route.ts` | Proxy to Flare TEE — reads address from session, never from client body |
| `frontend/lib/session.ts` | iron-session config and helper types |

### Modified files

| File | Change |
|------|--------|
| `frontend/app/layout.tsx` | Wrap `{children}` with `<Providers>` |
| `frontend/app/page.tsx` | Replace mock wallet state + `connectWallet()` with `useWallet()` hook; replace "Connect MetaMask" button with AppKit-driven button |

---

## Architecture

### AppKit Configuration (`lib/reown.ts`)

- Wagmi adapter configured with:
  - **Flare mainnet** (chain ID `14`) — added as a custom chain; set as `defaultNetwork`
  - **Ethereum mainnet** — via wagmi built-in
- Solana adapter configured with Solana mainnet
- SIWX plugin initialised with `required: true` — this causes AppKit to **immediately** trigger the SIWX signature prompt after wallet selection, so the user never sees a separate "Sign Message" step. Implemented via `new ReownAuthentication({ required: true })` (or equivalent `createSIWX` config depending on final AppKit API).
- SIWX plugin configured with custom `getNonce` / `verifyMessage` callbacks that call our `/api/auth/siwx` routes, not Reown's cloud endpoint.
- Reown project ID read from `NEXT_PUBLIC_REOWN_PROJECT_ID` env variable

### Provider Tree (`app/providers.tsx`)

```
<WagmiProvider config={wagmiConfig}>
  <QueryClientProvider client={queryClient}>
    <AppKitProvider>
      {children}
    </AppKitProvider>
  </QueryClientProvider>
</WagmiProvider>
```

Must be a `"use client"` component. Imported into `layout.tsx` which stays a server component.

---

## Data Flow

```
1. User clicks "Connect"
   → open() from useAppKit() triggers AppKit modal

2. User selects wallet (MetaMask, WalletConnect, Phantom, etc.) and connects

3. AppKit SIWX plugin (automatic):
   a. GET  /api/auth/siwx?action=nonce
      — server generates cryptographically random nonce
      — stores nonce in iron-session cookie (10-min expiry)
      — returns { nonce }
   b. AppKit constructs SIWX message with nonce, domain, address, chainId
   c. User signs message in their wallet
   d. POST /api/auth/siwx  { message, signature }
      — server verifies signature using verifySiweMessage (viem/siwe)
      — checks nonce matches session to prevent replay attacks
      — writes { address, chainId } to iron-session cookie
      — returns { address }

4. useWallet hook:
   — polls GET /api/auth/siwx?action=session on mount
   — exposes { address, isAuthenticated: true } once session is set
   — page.tsx replaces hardcoded "0x1234...5678" with real address

5. User reaches Step 4 (Generate Proof):
   — Frontend calls POST /api/flare/pnl { tradeId, binanceApiKey }
   — API route reads address from iron-session (ignores any client-supplied address)
   — Forwards { address, tradeId, binanceApiKey } to Flare TEE endpoint
   — Returns TEE response to client
```

---

## SIWX API Route (`/api/auth/siwx/route.ts`)

### `GET ?action=nonce`
- Generate 16-byte random nonce (`crypto.randomUUID()` or `generateSiweNonce()` from viem)
- Save to iron-session: `session.nonce = nonce`
- Return `{ nonce }`

### `POST` (verify) — multi-chain

Read `{ message, signature, chainId }` from request body. Branch on chain type:

**EVM (Flare, Ethereum) — `chainId` is a number (e.g. `14`, `1`):**
- Call `verifySiweMessage({ message, signature })` from `viem/siwe`
- Validate: signature valid, nonce matches `session.nonce`, domain matches, not expired

**Solana — `chainId` is the string `"solana:5eykt4UsFv8P8NJdTREpY1vzqKqZKvdp"` (mainnet):**
- Decode the base58 public key (wallet address) using `@solana/web3.js` `PublicKey`
- Decode the base58 signature
- Verify the ed25519 signature over the raw UTF-8 message bytes using `tweetnacl`'s `sign.detached.verify`
- Validate nonce matches `session.nonce`

Both paths on success: clear nonce, write `{ address, chainId }` to session, return `{ address }`.  
Both paths on failure: return `401` with `{ error: "Invalid signature" }`.

### `GET ?action=session`
- Return `{ address: session.address ?? null }`

### `DELETE`
- Clear iron-session, return `200`

---

## useWallet Hook Interface

```ts
interface UseWalletReturn {
  address: string | null        // verified address; null until SIWX complete
  isConnected: boolean          // AppKit wallet modal connected
  isAuthenticated: boolean      // SIWX signature verified server-side
  open: () => void              // opens AppKit modal
  disconnect: () => void        // disconnects AppKit + DELETE /api/auth/siwx
}
```

### Hydration safety

The hook **must** use a `mounted` guard to prevent hydration mismatches. The server always renders the logged-out state; the client reads the real session only after mount:

```ts
const [mounted, setMounted] = useState(false)
useEffect(() => { setMounted(true) }, [])

// All derived state gates on mounted:
const address = mounted ? session?.address ?? null : null
const isAuthenticated = mounted ? !!session?.address : false
```

The session is fetched from `/api/auth/siwx?action=session` inside a `useEffect` (not during SSR). AppKit's own state (`useAppKitAccount`) is also only read after mount for the same reason.

---

## Session Storage (`lib/session.ts`)

Uses `iron-session` with a server-side secret (`SESSION_SECRET` env variable, min 32 chars):

```ts
interface SessionData {
  nonce?: string       // temporary, cleared after verification
  address?: string     // verified EVM address (post-SIWX)
  chainId?: number
}
```

Cookie is `httpOnly`, `secure` in production, `sameSite: lax`.

---

## UI Changes (`app/page.tsx`)

Step 1 of the wizard:

- **Before:** `<Button onClick={connectWallet}>Connect MetaMask</Button>` with hardcoded `walletAddress` and `ensName`
- **After:** `<Button onClick={open}>Connect Wallet</Button>` (calls `useWallet().open`) or `<appkit-button />` web component
- Connected state reads `address` and `isAuthenticated` from `useWallet()`
- ENS lookup: once `address` is available, call `getEnsName` via wagmi's `useEnsName` hook

---

## Environment Variables

```
NEXT_PUBLIC_REOWN_PROJECT_ID=   # from cloud.reown.com
SESSION_SECRET=                  # random 32+ char string for iron-session
FLARE_TEE_ENDPOINT=              # URL of the Flare TEE backend
```

---

## Flare TEE Proxy (`/api/flare/pnl/route.ts`)

The proxy enforces that **only a server-verified address reaches the TEE**:

```
POST /api/flare/pnl  { tradeId, binanceApiKey }   ← client sends NO address

1. Read iron-session cookie
2. If session.address is missing → return 401 { error: "Unauthorized" }
3. Forward { address: session.address, tradeId, binanceApiKey } to FLARE_TEE_ENDPOINT
4. Return TEE response to client
```

The `address` field from the client request body is explicitly ignored even if present.

---

## Security Properties

- The Flare TEE receives the wallet address **only from the server-side session**, never from the client request body
- `/api/flare/pnl` returns `401` immediately if session is missing or has no verified address
- Nonce is single-use and time-limited (10 min) — prevents replay attacks
- Session cookie is `httpOnly` — not accessible to client-side JS
- Domain and origin are validated in the SIWX message to prevent cross-site signature reuse
- `required: true` on SIWX ensures the signature step cannot be skipped after wallet connect
- Solana ed25519 signatures are verified with `tweetnacl` — the route does not silently pass unverified Solana addresses

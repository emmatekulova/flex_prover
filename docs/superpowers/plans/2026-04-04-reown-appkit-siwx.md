# Reown AppKit + SIWX Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the mock `connectWallet()` stub in FlexProver with a production Reown AppKit integration that requires SIWX signature verification before the wallet address is trusted by the server.

**Architecture:** The AppKit `SIWXConfig` interface drives authentication — `createMessage` fetches a server nonce and builds a SIWE/CAIP-122 message, `addSession` POSTs the signature to our API route for server-side verification, and `getSessions` reads back the iron-session cookie. The `/api/flare/pnl` proxy reads the wallet address exclusively from the iron-session cookie, never from the client request body.

**Tech Stack:** `@reown/appkit`, `@reown/appkit-adapter-wagmi`, `@reown/appkit-adapter-solana`, `@reown/appkit-siwx`, `wagmi`, `viem` (SIWE helpers + `verifyMessage`), `@solana/web3.js` + `tweetnacl` + `bs58` (Solana ed25519 verification), `iron-session` (signed httpOnly cookie), `vitest` (unit tests on crypto helpers).

---

## File Map

| File | Status | Responsibility |
|------|--------|----------------|
| `frontend/lib/session.ts` | **Create** | iron-session config + `getSession()` helper |
| `frontend/lib/siwx-verify.ts` | **Create** | Pure `verifyEvmSignature` + `verifySolanaSignature` + `extractNonce` helpers |
| `frontend/lib/reown.ts` | **Create** | AppKit init: Flare chain, wagmi adapter, Solana adapter, `SIWXConfig` |
| `frontend/app/providers.tsx` | **Create** | `"use client"` wrapper: `WagmiProvider` + `QueryClientProvider` |
| `frontend/app/layout.tsx` | **Modify** | Wrap `{children}` with `<Providers>` |
| `frontend/app/api/auth/siwx/route.ts` | **Create** | GET nonce/session, POST verify (EVM + Solana), DELETE |
| `frontend/app/api/flare/pnl/route.ts` | **Create** | Proxy to Flare TEE — address from session only |
| `frontend/hooks/use-wallet.ts` | **Create** | `useWallet()` hook with `mounted` guard |
| `frontend/app/page.tsx` | **Modify** | Wire real wallet state; replace mock connect button |
| `frontend/vitest.config.ts` | **Create** | Vitest config with `jsdom` environment + `@` alias |
| `frontend/vitest.setup.ts` | **Create** | `@testing-library/jest-dom` import |
| `frontend/__tests__/siwx-verify.test.ts` | **Create** | Unit tests for crypto verification helpers |
| `frontend/__tests__/flare-pnl.test.ts` | **Create** | Unit test for TEE proxy auth guard |
| `frontend/.env.local` | **Create** | Environment variable stubs |

---

## Task 1: Install Dependencies + Test Setup

**Files:**
- Modify: `frontend/package.json` (via npm install)
- Create: `frontend/vitest.config.ts`
- Create: `frontend/vitest.setup.ts`

- [ ] **Step 1.1: Install runtime packages**

```bash
cd frontend
npm install @reown/appkit @reown/appkit-adapter-wagmi @reown/appkit-adapter-solana wagmi viem @solana/web3.js tweetnacl bs58 iron-session
```

Expected: no peer-dependency errors. If you see a wagmi/viem version conflict, check `@reown/appkit`'s peer deps and pin accordingly.

- [ ] **Step 1.2: Install dev packages for testing**

```bash
cd frontend
npm install -D vitest @vitejs/plugin-react @testing-library/react @testing-library/jest-dom jsdom
```

- [ ] **Step 1.3: Create `frontend/vitest.config.ts`**

```typescript
import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  test: {
    environment: 'jsdom',
    setupFiles: ['./vitest.setup.ts'],
    globals: true,
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, '.'),
    },
  },
})
```

- [ ] **Step 1.4: Create `frontend/vitest.setup.ts`**

```typescript
import '@testing-library/jest-dom'
```

- [ ] **Step 1.5: Add test script to `frontend/package.json`**

Open `frontend/package.json` and add to `"scripts"`:
```json
"test": "vitest run",
"test:watch": "vitest"
```

- [ ] **Step 1.6: Verify test runner works**

```bash
cd frontend
npm test
```

Expected: `No test files found` (exits 0 or with a "no tests" warning — either is fine at this stage).

---

## Task 2: Session Config (`lib/session.ts`)

**Files:**
- Create: `frontend/lib/session.ts`

- [ ] **Step 2.1: Create `frontend/lib/session.ts`**

```typescript
import { getIronSession, type IronSession } from 'iron-session'
import { cookies } from 'next/headers'

export interface SessionData {
  nonce?: string
  address?: string
  chainId?: string
}

export const sessionOptions = {
  password: process.env.SESSION_SECRET!,
  cookieName: 'flex-prover-session',
  cookieOptions: {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax' as const,
    maxAge: 60 * 60 * 24 * 7, // 1 week
  },
}

export async function getSession(): Promise<IronSession<SessionData>> {
  return getIronSession<SessionData>(await cookies(), sessionOptions)
}
```

---

## Task 3: Crypto Verification Helpers, TDD (`lib/siwx-verify.ts`)

These are pure functions — no Next.js dependencies — making them easy to unit test.

**Files:**
- Create: `frontend/lib/siwx-verify.ts`
- Create: `frontend/__tests__/siwx-verify.test.ts`

- [ ] **Step 3.1: Write the failing tests first**

Create `frontend/__tests__/siwx-verify.test.ts`:

```typescript
import { describe, it, expect } from 'vitest'
import { verifyEvmSignature, verifySolanaSignature, extractNonce } from '@/lib/siwx-verify'
import { privateKeyToAccount } from 'viem/accounts'
import { createSiweMessage } from 'viem/siwe'
import nacl from 'tweetnacl'
import bs58 from 'bs58'
import { PublicKey } from '@solana/web3.js'

// Deterministic test account (well-known anvil key — never use in production)
const TEST_PRIVATE_KEY =
  '0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80' as const
const account = privateKeyToAccount(TEST_PRIVATE_KEY)

function makeEvmMessage(nonce: string) {
  return createSiweMessage({
    domain: 'localhost',
    address: account.address,
    statement: 'Sign in to FlexProver to verify your identity.',
    uri: 'http://localhost:3000',
    version: '1',
    chainId: 14,
    nonce,
  })
}

function makeSolanaMessage(address: string, nonce: string) {
  return [
    `localhost wants you to sign in with your account:`,
    address,
    '',
    'Sign in to FlexProver to verify your identity.',
    '',
    'URI: http://localhost:3000',
    'Version: 1',
    'Chain ID: solana:5eykt4UsFv8P8NJdTREpY1vzqKqZKvdp',
    `Nonce: ${nonce}`,
    `Issued At: 2026-04-04T00:00:00.000Z`,
  ].join('\n')
}

// --- extractNonce ---
describe('extractNonce', () => {
  it('extracts nonce from a SIWE message', () => {
    const msg = makeEvmMessage('abc123')
    expect(extractNonce(msg)).toBe('abc123')
  })

  it('extracts nonce from a CAIP-122 message', () => {
    const msg = makeSolanaMessage('someaddress', 'xyz789')
    expect(extractNonce(msg)).toBe('xyz789')
  })

  it('returns null when no nonce line exists', () => {
    expect(extractNonce('no nonce here')).toBeNull()
  })
})

// --- verifyEvmSignature ---
describe('verifyEvmSignature', () => {
  it('returns the address for a valid EVM signature', async () => {
    const message = makeEvmMessage('testnonce')
    const signature = await account.signMessage({ message })
    const result = await verifyEvmSignature(message, signature)
    expect(result?.toLowerCase()).toBe(account.address.toLowerCase())
  })

  it('returns null for a tampered message', async () => {
    const message = makeEvmMessage('testnonce')
    const signature = await account.signMessage({ message })
    const tampered = message.replace('testnonce', 'BADNONCE')
    const result = await verifyEvmSignature(tampered, signature)
    expect(result).toBeNull()
  })

  it('returns null for a garbage signature', async () => {
    const message = makeEvmMessage('testnonce')
    const result = await verifyEvmSignature(message, '0xdeadbeef')
    expect(result).toBeNull()
  })
})

// --- verifySolanaSignature ---
describe('verifySolanaSignature', () => {
  it('returns the address for a valid Solana ed25519 signature', async () => {
    const keypair = nacl.sign.keyPair.fromSeed(new Uint8Array(32).fill(42))
    const address = new PublicKey(keypair.publicKey).toBase58()
    const message = makeSolanaMessage(address, 'solnonce')
    const messageBytes = new TextEncoder().encode(message)
    const sigBytes = nacl.sign.detached(messageBytes, keypair.secretKey)
    const signature = bs58.encode(sigBytes)

    const result = await verifySolanaSignature(message, signature)
    expect(result).toBe(address)
  })

  it('returns null when the signature does not match the address in the message', async () => {
    const keypair = nacl.sign.keyPair.fromSeed(new Uint8Array(32).fill(42))
    const wrongKeypair = nacl.sign.keyPair.fromSeed(new Uint8Array(32).fill(99))
    const address = new PublicKey(keypair.publicKey).toBase58()
    const message = makeSolanaMessage(address, 'solnonce')
    const messageBytes = new TextEncoder().encode(message)
    // Sign with the WRONG key
    const sigBytes = nacl.sign.detached(messageBytes, wrongKeypair.secretKey)
    const signature = bs58.encode(sigBytes)

    const result = await verifySolanaSignature(message, signature)
    expect(result).toBeNull()
  })

  it('returns null for a garbage Solana signature', async () => {
    const keypair = nacl.sign.keyPair.fromSeed(new Uint8Array(32).fill(42))
    const address = new PublicKey(keypair.publicKey).toBase58()
    const message = makeSolanaMessage(address, 'solnonce')
    const result = await verifySolanaSignature(message, 'notvalidbase58!!!')
    expect(result).toBeNull()
  })
})
```

- [ ] **Step 3.2: Run tests — confirm they all fail**

```bash
cd frontend
npm test
```

Expected: all 8 tests FAIL with `Cannot find module '@/lib/siwx-verify'`.

- [ ] **Step 3.3: Create `frontend/lib/siwx-verify.ts`**

```typescript
import { verifyMessage } from 'viem'
import { parseSiweMessage } from 'viem/siwe'

/**
 * Extracts the nonce from a SIWE or CAIP-122 message string.
 * Both formats use the line "Nonce: <value>".
 */
export function extractNonce(message: string): string | null {
  const match = message.match(/^Nonce: (.+)$/m)
  return match?.[1]?.trim() ?? null
}

/**
 * Verifies an EVM SIWE signature.
 * Returns the checksummed address from the message if valid, null otherwise.
 */
export async function verifyEvmSignature(
  message: string,
  signature: string,
): Promise<string | null> {
  try {
    const parsed = parseSiweMessage(message)
    if (!parsed.address) return null
    const valid = await verifyMessage({
      address: parsed.address as `0x${string}`,
      message,
      signature: signature as `0x${string}`,
    })
    return valid ? parsed.address : null
  } catch {
    return null
  }
}

/**
 * Verifies a Solana ed25519 SIWX signature.
 * The address (base58 public key) is parsed from the second line of the CAIP-122 message.
 * Returns the address if valid, null otherwise.
 */
export async function verifySolanaSignature(
  message: string,
  signature: string,
): Promise<string | null> {
  try {
    const { PublicKey } = await import('@solana/web3.js')
    const nacl = (await import('tweetnacl')).default
    const bs58 = (await import('bs58')).default

    // CAIP-122 format: second line is the account address
    const lines = message.split('\n')
    const address = lines[1]?.trim()
    if (!address) return null

    const publicKey = new PublicKey(address)
    const messageBytes = new TextEncoder().encode(message)
    const signatureBytes = bs58.decode(signature)

    const valid = nacl.sign.detached.verify(
      messageBytes,
      signatureBytes,
      publicKey.toBytes(),
    )
    return valid ? address : null
  } catch {
    return null
  }
}
```

- [ ] **Step 3.4: Run tests — confirm they all pass**

```bash
cd frontend
npm test
```

Expected: 8 tests PASS.

---

## Task 4: AppKit Configuration (`lib/reown.ts`)

**Files:**
- Create: `frontend/lib/reown.ts`

- [ ] **Step 4.1: Create `frontend/lib/reown.ts`**

This file must only be imported in `"use client"` contexts. It calls `createAppKit` which registers the global AppKit instance and the `<appkit-button>` web component.

```typescript
import { createAppKit, type SIWXConfig, type SIWXMessage, type SIWXSession } from '@reown/appkit'
import { WagmiAdapter } from '@reown/appkit-adapter-wagmi'
import { SolanaAdapter } from '@reown/appkit-adapter-solana/react'
import { mainnet, solana } from '@reown/appkit/networks'
import { defineChain } from 'viem'
import { createSiweMessage } from 'viem/siwe'

const projectId = process.env.NEXT_PUBLIC_REOWN_PROJECT_ID!

// Flare Mainnet — not in wagmi/AppKit's built-in network list
export const flare = defineChain({
  id: 14,
  name: 'Flare',
  nativeCurrency: { name: 'Flare', symbol: 'FLR', decimals: 18 },
  rpcUrls: {
    default: { http: ['https://flare-api.flare.network/ext/C/rpc'] },
  },
  blockExplorers: {
    default: { name: 'Flarescan', url: 'https://flarescan.com' },
  },
})

// AppKit network object for Flare (CAIP format)
const flareNetwork = {
  id: 14,
  name: 'Flare',
  caipNetworkId: 'eip155:14' as const,
  chainNamespace: 'eip155' as const,
  nativeCurrency: { name: 'Flare', symbol: 'FLR', decimals: 18 },
  rpcUrls: {
    default: { http: ['https://flare-api.flare.network/ext/C/rpc'] },
  },
  blockExplorers: {
    default: { name: 'Flarescan', url: 'https://flarescan.com' },
  },
}

// Wagmi adapter — EVM chains (Flare + Ethereum)
const wagmiAdapter = new WagmiAdapter({
  networks: [flareNetwork, mainnet],
  projectId,
})

// Solana adapter
const solanaAdapter = new SolanaAdapter()

// Custom SIWX config — calls our own API routes, not Reown cloud
const siwx: SIWXConfig = {
  /**
   * Called by AppKit to construct the message the user will sign.
   * Fetches a one-time nonce from our server to prevent replay attacks.
   */
  createMessage: async (input: SIWXMessage.Input): Promise<SIWXMessage> => {
    const res = await fetch('/api/auth/siwx?action=nonce')
    if (!res.ok) throw new Error('Failed to fetch nonce from server')
    const { nonce } = (await res.json()) as { nonce: string }

    const domain = window.location.host
    const uri = window.location.origin
    const issuedAt = new Date().toISOString()

    const isEvm = input.chainId.startsWith('eip155:')
    let messageString: string

    if (isEvm) {
      const chainIdNum = parseInt(input.chainId.split(':')[1]!)
      messageString = createSiweMessage({
        domain,
        address: input.accountAddress as `0x${string}`,
        statement: 'Sign in to FlexProver to verify your identity.',
        uri,
        version: '1',
        chainId: chainIdNum,
        nonce,
        issuedAt,
      })
    } else {
      // CAIP-122 format for Solana
      messageString = [
        `${domain} wants you to sign in with your account:`,
        input.accountAddress,
        '',
        'Sign in to FlexProver to verify your identity.',
        '',
        `URI: ${uri}`,
        'Version: 1',
        `Chain ID: ${input.chainId}`,
        `Nonce: ${nonce}`,
        `Issued At: ${issuedAt}`,
      ].join('\n')
    }

    return {
      ...input,
      nonce,
      issuedAt,
      domain,
      uri,
      toString: () => messageString,
    }
  },

  /**
   * Called after the user signs the message.
   * Posts the signature to our API for server-side verification.
   * Throws on failure so AppKit can surface the error to the user.
   */
  addSession: async (session: SIWXSession): Promise<void> => {
    const res = await fetch('/api/auth/siwx', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        message: session.data.toString(),
        signature: session.signature,
        chainId: session.data.chainId,
      }),
    })
    if (!res.ok) {
      const body = await res.json().catch(() => ({}))
      throw new Error((body as { error?: string }).error ?? 'SIWX verification failed')
    }
  },

  /**
   * Called by AppKit to check if the user is already authenticated.
   * Reads the iron-session cookie via our API route.
   */
  getSessions: async (chainId: string, address: string): Promise<SIWXSession[]> => {
    const res = await fetch('/api/auth/siwx?action=session')
    if (!res.ok) return []
    const { address: storedAddress, chainId: storedChainId } = (await res.json()) as {
      address?: string
      chainId?: string
    }

    if (!storedAddress) return []

    // Case-insensitive comparison for EVM; exact for Solana (base58 is case-sensitive)
    const isEvm = chainId.startsWith('eip155:')
    const matches = isEvm
      ? storedAddress.toLowerCase() === address.toLowerCase()
      : storedAddress === address

    if (!matches) return []

    // Return a minimal session — AppKit only checks sessions.length > 0
    return [
      {
        data: {
          accountAddress: storedAddress,
          chainId: (storedChainId ?? chainId) as `${string}:${string}`,
          domain: window.location.host,
          uri: window.location.origin,
          version: '1',
          nonce: '',
          issuedAt: new Date().toISOString(),
          toString: () => '',
        },
        signature: '',
      },
    ]
  },

  /**
   * Called when the user disconnects or explicitly signs out.
   */
  revokeSession: async (_chainId: string, _address: string): Promise<void> => {
    await fetch('/api/auth/siwx', { method: 'DELETE' })
  },

  /**
   * Called by AppKit to replace sessions in bulk (e.g. on chain switch).
   * If empty, clear the server session. Non-empty means addSession already stored it.
   */
  setSessions: async (sessions: SIWXSession[]): Promise<void> => {
    if (sessions.length === 0) {
      await fetch('/api/auth/siwx', { method: 'DELETE' })
    }
  },

  /** Forces the signature step immediately after wallet selection. */
  getRequired: () => true,

  /** Clear session when the user disconnects from AppKit. */
  signOutOnDisconnect: true,
}

// Export wagmi config for WagmiProvider
export const wagmiConfig = wagmiAdapter.wagmiConfig

// Initialise AppKit — registers <appkit-button> web component globally
createAppKit({
  adapters: [wagmiAdapter, solanaAdapter],
  networks: [flareNetwork, mainnet, solana],
  defaultNetwork: flareNetwork,
  projectId,
  siwx,
  features: { analytics: false },
})
```

- [ ] **Step 4.2: Verify TypeScript compiles**

```bash
cd frontend
npx tsc --noEmit
```

Expected: no errors. If you see `SIWXMessage.Input` or `SIWXSession` type errors, check the installed version of `@reown/appkit` and adjust the type imports — the interface is exported from `@reown/appkit` or `@reown/appkit-core`.

---

## Task 5: Client Provider Tree (`app/providers.tsx`)

**Files:**
- Create: `frontend/app/providers.tsx`

- [ ] **Step 5.1: Create `frontend/app/providers.tsx`**

This is a `"use client"` component. It imports `lib/reown.ts` (which calls `createAppKit`) and wraps children in the wagmi + query client providers. Importing `lib/reown.ts` here is what registers AppKit on the client.

```typescript
'use client'

import '@/lib/reown' // side-effect: initialises AppKit + registers <appkit-button>
import { WagmiProvider } from 'wagmi'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { wagmiConfig } from '@/lib/reown'
import { type ReactNode, useState } from 'react'

export function Providers({ children }: { children: ReactNode }) {
  // QueryClient must be created per-component to avoid sharing across requests
  const [queryClient] = useState(() => new QueryClient())

  return (
    <WagmiProvider config={wagmiConfig}>
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    </WagmiProvider>
  )
}
```

- [ ] **Step 5.2: Install `@tanstack/react-query` (wagmi peer dependency)**

```bash
cd frontend
npm install @tanstack/react-query
```

---

## Task 6: Wrap Layout with Providers (`app/layout.tsx`)

**Files:**
- Modify: `frontend/app/layout.tsx`

- [ ] **Step 6.1: Update `frontend/app/layout.tsx`**

The layout file stays a server component — only `Providers` is a client component.

Replace the entire file content with:

```typescript
import type { Metadata } from 'next'
import { Geist, Geist_Mono } from 'next/font/google'
import { Analytics } from '@vercel/analytics/next'
import { Providers } from './providers'
import './globals.css'

const _geist = Geist({ subsets: ['latin'] })
const _geistMono = Geist_Mono({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: 'FlexProver - Proof of Whale',
  description:
    'Sybil-resistant reputation engine. Bind your Binance PNL to your ENS identity using secure Flare TEE enclaves.',
  generator: 'v0.app',
  icons: {
    icon: [
      { url: '/icon-light-32x32.png', media: '(prefers-color-scheme: light)' },
      { url: '/icon-dark-32x32.png', media: '(prefers-color-scheme: dark)' },
      { url: '/icon.svg', type: 'image/svg+xml' },
    ],
    apple: '/apple-icon.png',
  },
}

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body className="font-sans antialiased">
        <Providers>
          {children}
        </Providers>
        <Analytics />
      </body>
    </html>
  )
}
```

---

## Task 7: SIWX API Route (`app/api/auth/siwx/route.ts`)

**Files:**
- Create: `frontend/app/api/auth/siwx/route.ts`

- [ ] **Step 7.1: Create `frontend/app/api/auth/siwx/route.ts`**

```typescript
import { type NextRequest, NextResponse } from 'next/server'
import { generateSiweNonce } from 'viem/siwe'
import { getSession } from '@/lib/session'
import { verifyEvmSignature, verifySolanaSignature, extractNonce } from '@/lib/siwx-verify'

// GET /api/auth/siwx?action=nonce  — generates a one-time nonce
// GET /api/auth/siwx?action=session — returns the current verified address
export async function GET(request: NextRequest) {
  const action = request.nextUrl.searchParams.get('action')

  if (action === 'nonce') {
    const session = await getSession()
    const nonce = generateSiweNonce()
    session.nonce = nonce
    await session.save()
    return NextResponse.json({ nonce })
  }

  if (action === 'session') {
    const session = await getSession()
    return NextResponse.json({
      address: session.address ?? null,
      chainId: session.chainId ?? null,
    })
  }

  return NextResponse.json({ error: 'Invalid action' }, { status: 400 })
}

// POST /api/auth/siwx — verifies signature (EVM or Solana) and stores session
export async function POST(request: NextRequest) {
  const body = (await request.json()) as {
    message: string
    signature: string
    chainId: string
  }
  const { message, signature, chainId } = body

  const session = await getSession()

  if (!session.nonce) {
    return NextResponse.json({ error: 'No nonce in session. Request a new nonce first.' }, { status: 400 })
  }

  // Verify the nonce in the message matches the session nonce (replay protection)
  const messageNonce = extractNonce(message)
  if (messageNonce !== session.nonce) {
    return NextResponse.json({ error: 'Nonce mismatch' }, { status: 401 })
  }

  const isEvm = typeof chainId === 'string' && chainId.startsWith('eip155:')
  const verifiedAddress = isEvm
    ? await verifyEvmSignature(message, signature)
    : await verifySolanaSignature(message, signature)

  if (!verifiedAddress) {
    return NextResponse.json({ error: 'Invalid signature' }, { status: 401 })
  }

  // Nonce is single-use — clear it and store the verified address
  session.nonce = undefined
  session.address = verifiedAddress
  session.chainId = chainId
  await session.save()

  return NextResponse.json({ address: verifiedAddress })
}

// DELETE /api/auth/siwx — clears the session (sign out)
export async function DELETE() {
  const session = await getSession()
  session.destroy()
  return NextResponse.json({ ok: true })
}
```

- [ ] **Step 7.2: Smoke-test the nonce endpoint manually**

Start the dev server in a separate terminal:
```bash
cd frontend
npm run dev
```

Then in another terminal:
```bash
# Should return { nonce: "<random string>" } and set a cookie
curl -c /tmp/flex-cookies.txt \
  "http://localhost:3000/api/auth/siwx?action=nonce"

# Should return { address: null }
curl -b /tmp/flex-cookies.txt \
  "http://localhost:3000/api/auth/siwx?action=session"
```

Expected: both return JSON with 200.

---

## Task 8: Flare TEE Proxy Route, TDD (`app/api/flare/pnl/route.ts`)

**Files:**
- Create: `frontend/app/api/flare/pnl/route.ts`
- Create: `frontend/__tests__/flare-pnl.test.ts`

- [ ] **Step 8.1: Write the failing test**

Create `frontend/__tests__/flare-pnl.test.ts`:

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest'

// We test the auth-guard logic in isolation by mocking iron-session and fetch.
// The test imports the actual POST handler from the route file.

vi.mock('@/lib/session', () => ({
  getSession: vi.fn(),
}))

// Mock next/headers (required by iron-session inside Next.js routes)
vi.mock('next/headers', () => ({
  cookies: vi.fn(() => Promise.resolve(new Map())),
}))

import { POST } from '@/app/api/flare/pnl/route'
import { getSession } from '@/lib/session'

const mockGetSession = vi.mocked(getSession)

function makeRequest(body: object) {
  return new Request('http://localhost:3000/api/flare/pnl', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
}

describe('POST /api/flare/pnl', () => {
  beforeEach(() => {
    vi.resetAllMocks()
  })

  it('returns 401 when no session address exists', async () => {
    mockGetSession.mockResolvedValue({ address: undefined } as never)
    const req = makeRequest({ tradeId: '1', binanceApiKey: 'test' })
    const res = await POST(req as never)
    expect(res.status).toBe(401)
    const body = await res.json()
    expect(body.error).toBe('Unauthorized')
  })

  it('does not use address from request body when session is absent', async () => {
    mockGetSession.mockResolvedValue({ address: undefined } as never)
    const req = makeRequest({
      tradeId: '1',
      binanceApiKey: 'test',
      address: '0xmalicious', // attacker-supplied address — must be ignored
    })
    const res = await POST(req as never)
    expect(res.status).toBe(401)
  })

  it('forwards verified session address to TEE when authenticated', async () => {
    const verifiedAddress = '0xabc123'
    mockGetSession.mockResolvedValue({ address: verifiedAddress } as never)

    const mockTeeResponse = { proof: 'hash123' }
    global.fetch = vi.fn().mockResolvedValue(
      new Response(JSON.stringify(mockTeeResponse), { status: 200 }),
    )
    process.env.FLARE_TEE_ENDPOINT = 'http://tee.local/verify'

    const req = makeRequest({ tradeId: 'trade-5', binanceApiKey: 'key-abc' })
    const res = await POST(req as never)
    expect(res.status).toBe(200)

    // Verify the TEE received the session address, not anything from the client
    expect(global.fetch).toHaveBeenCalledWith(
      'http://tee.local/verify',
      expect.objectContaining({
        method: 'POST',
        body: expect.stringContaining(verifiedAddress),
      }),
    )
    const body = await res.json()
    expect(body).toEqual(mockTeeResponse)
  })
})
```

- [ ] **Step 8.2: Run — confirm tests fail**

```bash
cd frontend
npm test
```

Expected: 3 tests FAIL with `Cannot find module '@/app/api/flare/pnl/route'`.

- [ ] **Step 8.3: Create `frontend/app/api/flare/pnl/route.ts`**

```typescript
import { type NextRequest, NextResponse } from 'next/server'
import { getSession } from '@/lib/session'

export async function POST(request: NextRequest) {
  const session = await getSession()

  // Enforce: only serve requests with a server-verified address
  if (!session.address) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  const body = (await request.json()) as {
    tradeId: string
    binanceApiKey: string
  }

  const teeEndpoint = process.env.FLARE_TEE_ENDPOINT
  if (!teeEndpoint) {
    return NextResponse.json({ error: 'TEE endpoint not configured' }, { status: 500 })
  }

  // The address comes exclusively from the session — the client body is never trusted
  const teeRes = await fetch(teeEndpoint, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      address: session.address,
      tradeId: body.tradeId,
      binanceApiKey: body.binanceApiKey,
    }),
  })

  const data = await teeRes.json()
  return NextResponse.json(data, { status: teeRes.status })
}
```

- [ ] **Step 8.4: Run tests — confirm all pass**

```bash
cd frontend
npm test
```

Expected: all 11 tests PASS (8 from siwx-verify + 3 from flare-pnl).

---

## Task 9: `useWallet` Hook (`hooks/use-wallet.ts`)

**Files:**
- Create: `frontend/hooks/use-wallet.ts`

- [ ] **Step 9.1: Create `frontend/hooks/use-wallet.ts`**

The `mounted` guard prevents hydration mismatches: the server renders the logged-out state (all `null`/`false`), and only after the client mounts does it read the real session cookie and AppKit state.

```typescript
'use client'

import { useEffect, useState } from 'react'
import { useAppKit, useAppKitAccount } from '@reown/appkit/react'
import { useDisconnect, useEnsName } from 'wagmi'

export interface UseWalletReturn {
  /** Verified wallet address. null until SIWX is complete. */
  address: string | null
  /** Whether AppKit reports a wallet as connected (pre-SIWX). */
  isConnected: boolean
  /** Whether SIWX has been verified server-side. Use this to gate protected steps. */
  isAuthenticated: boolean
  /** ENS name for EVM addresses. null for Solana or if no ENS found. */
  ensName: string | null
  /** Opens the AppKit wallet modal. */
  open: () => void
  /** Disconnects wallet and invalidates the server-side session. */
  disconnect: () => void
}

export function useWallet(): UseWalletReturn {
  const [mounted, setMounted] = useState(false)
  const [sessionAddress, setSessionAddress] = useState<string | null>(null)

  const { open } = useAppKit()
  const { address: appKitAddress, isConnected: appKitConnected } = useAppKitAccount()
  const { disconnect: wagmiDisconnect } = useDisconnect()

  // Step 1: mark mounted to unlock client-side state
  useEffect(() => {
    setMounted(true)
  }, [])

  // Step 2: fetch the server-side session whenever mount state or connection changes
  useEffect(() => {
    if (!mounted) return
    fetch('/api/auth/siwx?action=session')
      .then((r) => r.json())
      .then((data: { address?: string | null }) => {
        setSessionAddress(data.address ?? null)
      })
      .catch(() => setSessionAddress(null))
  }, [mounted, appKitConnected])

  // Gate all derived state on mounted — server always sees null/false
  const address = mounted ? sessionAddress : null
  const isConnected = mounted ? (appKitConnected ?? false) : false
  const isAuthenticated = mounted ? !!sessionAddress : false

  // ENS lookup — only for EVM (0x) addresses
  const { data: ensData } = useEnsName({
    address: (address?.startsWith('0x') ? address : undefined) as
      | `0x${string}`
      | undefined,
    query: { enabled: !!address && address.startsWith('0x') },
  })

  const disconnect = () => {
    wagmiDisconnect()
    fetch('/api/auth/siwx', { method: 'DELETE' }).catch(() => null)
    setSessionAddress(null)
  }

  return {
    address,
    isConnected,
    isAuthenticated,
    ensName: ensData ?? null,
    open,
    disconnect,
  }
}
```

---

## Task 10: Wire Up `app/page.tsx`

**Files:**
- Modify: `frontend/app/page.tsx`

- [ ] **Step 10.1: Replace mock wallet state with `useWallet()`**

At the top of the `FlexProver` component, find the block (around line 97–101):

```typescript
  // Wallet state
  const [walletConnected, setWalletConnected] = useState(false)
  const [walletAddress] = useState("0x1234...5678")
  const [ensName] = useState("flexer.eth")
  const [hasEns] = useState(true)
```

Replace it with:

```typescript
  // Wallet state — driven by Reown AppKit + SIWX
  const {
    address: walletAddress,
    isAuthenticated: walletConnected,
    ensName,
    open: openWalletModal,
    disconnect: disconnectWallet,
  } = useWallet()
  const hasEns = !!ensName
```

Add the import at the top of the file (after existing imports):

```typescript
import { useWallet } from '@/hooks/use-wallet'
```

Also remove the now-unused `connectWallet` function (around line 186–188):

```typescript
  const connectWallet = () => {
    setWalletConnected(true)
  }
```

Delete those 3 lines entirely.

- [ ] **Step 10.2: Replace "Connect MetaMask" button**

Find the button (around line 471–477):

```typescript
                          <Button
                            onClick={connectWallet}
                            className="w-full h-14 bg-secondary hover:bg-secondary/80 border border-border text-foreground"
                          >
                            <Wallet className="w-5 h-5 mr-3" />
                            Connect MetaMask
                          </Button>
```

Replace with:

```typescript
                          <Button
                            onClick={openWalletModal}
                            className="w-full h-14 bg-secondary hover:bg-secondary/80 border border-border text-foreground"
                          >
                            <Wallet className="w-5 h-5 mr-3" />
                            Connect Wallet
                          </Button>
```

- [ ] **Step 10.3: Verify the app renders without hydration errors**

```bash
cd frontend
npm run dev
```

Open `http://localhost:3000` in the browser. Open DevTools console.

Expected:
- No `Hydration mismatch` warnings in the console
- Step 1 shows "Connect Wallet" button
- Clicking it opens the AppKit modal with Flare pre-selected
- After connecting + signing, the wallet address appears in Step 1
- The step advances only after SIWX is complete (because `walletConnected = isAuthenticated`)

---

## Task 11: Environment Variables

**Files:**
- Create: `frontend/.env.local`

- [ ] **Step 11.1: Create `frontend/.env.local`**

```bash
# Get your project ID from https://cloud.reown.com
NEXT_PUBLIC_REOWN_PROJECT_ID=your_project_id_here

# Random 32+ character string — generate with: openssl rand -hex 32
SESSION_SECRET=replace_with_32_plus_char_random_string

# Your Flare TEE backend endpoint
FLARE_TEE_ENDPOINT=http://localhost:8080/verify
```

- [ ] **Step 11.2: Verify `.env.local` is in `.gitignore`**

```bash
cd frontend
grep '.env.local' .gitignore
```

Expected: `.env.local` is listed. If not:

```bash
echo '.env.local' >> .gitignore
```

---

## Self-Review Notes

**Spec coverage check:**
- [x] `@reown/appkit-adapter-wagmi` — Task 4
- [x] SIWX with `required: true` via `getRequired: () => true` — Task 4
- [x] Flare as primary chain (defaultNetwork) — Task 4
- [x] Ethereum + Solana multi-chain — Task 4
- [x] AppKit modal via `useWallet().open` — Tasks 9 + 10
- [x] EVM SIWX verification (`verifyMessage` from viem) — Task 3 + 7
- [x] Solana ed25519 verification (`tweetnacl`) — Task 3 + 7
- [x] Nonce mismatch rejection (replay protection) — Task 7
- [x] iron-session `httpOnly` cookie — Task 2
- [x] Hydration `mounted` guard — Task 9
- [x] TEE proxy reads address from session only — Task 8
- [x] TEE proxy returns 401 with no session — Task 8 (tested)
- [x] `.env.local` with all three env vars — Task 11
- [x] Unit tests for crypto helpers (8 tests) — Task 3
- [x] Unit tests for TEE proxy auth guard (3 tests) — Task 8

**Type consistency:** `UseWalletReturn` interface defined in Task 9 and consumed unchanged in Task 10. `SessionData` defined in Task 2 and used in Tasks 7 and 8. `verifyEvmSignature` / `verifySolanaSignature` / `extractNonce` defined in Task 3 and imported in Task 7.

**No placeholders:** All code is complete and copy-pasteable.

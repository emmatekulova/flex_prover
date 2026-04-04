import { createAppKit, type SIWXConfig, type SIWXMessage, type SIWXSession } from '@reown/appkit'
import { WagmiAdapter } from '@reown/appkit-adapter-wagmi'
import { SolanaAdapter } from '@reown/appkit-adapter-solana/react'
import { mainnet, solana } from '@reown/appkit/networks'
import { type AppKitNetwork } from '@reown/appkit/networks'
import { type CaipNetworkId } from '@reown/appkit-common'
import { defineChain } from 'viem'
import { createSiweMessage } from 'viem/siwe'

const projectId = process.env.NEXT_PUBLIC_REOWN_PROJECT_ID!

const STALE_WC_RECOVERY_FLAG = 'flexprover:wc-stale-recovered'

function getErrorMessage(reason: unknown): string {
  if (typeof reason === 'string') return reason
  if (reason && typeof reason === 'object' && 'message' in reason) {
    const msg = (reason as { message?: unknown }).message
    if (typeof msg === 'string') return msg
  }
  return ''
}

function isStaleWalletConnectTopicError(message: string): boolean {
  const normalized = message.toLowerCase()
  return normalized.includes('no matching key') && normalized.includes("session topic doesn't exist")
}

function clearWalletConnectCache(): void {
  if (typeof window === 'undefined') return

  try {
    const keysToRemove: string[] = []
    for (let i = 0; i < window.localStorage.length; i += 1) {
      const key = window.localStorage.key(i)
      if (!key) continue
      if (key.startsWith('wc@') || key.toLowerCase().includes('walletconnect')) {
        keysToRemove.push(key)
      }
    }
    keysToRemove.forEach((key) => window.localStorage.removeItem(key))
  } catch {
    // noop
  }
}

if (typeof window !== 'undefined') {
  window.addEventListener('unhandledrejection', (event) => {
    const message = getErrorMessage(event.reason)
    if (!isStaleWalletConnectTopicError(message)) return

    const alreadyRecovered = window.sessionStorage.getItem(STALE_WC_RECOVERY_FLAG) === '1'
    if (alreadyRecovered) return

    clearWalletConnectCache()
    window.sessionStorage.setItem(STALE_WC_RECOVERY_FLAG, '1')
    window.location.reload()
  })
}

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

// AppKit network object — derived from the viem chain definition to avoid duplicate RPC/explorer URLs
const flareNetwork: AppKitNetwork = {
  ...flare,
  caipNetworkId: 'eip155:14' as const,
  chainNamespace: 'eip155' as const,
}

// Wagmi adapter — EVM chains (Flare + Ethereum)
const wagmiAdapter = new WagmiAdapter({
  networks: [flareNetwork, mainnet],
  projectId,
  ssr: true,
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
    const issuedAtDate = new Date()
    const issuedAt = issuedAtDate.toISOString()

    const isEvm = input.chainId.startsWith('eip155:')
    let messageString: string

    if (isEvm) {
      const chainIdNum = parseInt(input.chainId.split(':')[1]!, 10)
      messageString = createSiweMessage({
        domain,
        address: input.accountAddress as `0x${string}`,
        statement: 'Sign in to FlexProver to verify your identity.',
        uri,
        version: '1',
        chainId: chainIdNum,
        nonce,
        issuedAt: issuedAtDate,
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
      version: '1',
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
        message: session.message,
        signature: session.signature,
        chainId: session.data.chainId,
        data: session.data,
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
    const {
      address: storedAddress,
      chainId: storedChainId,
      message: storedMessage,
      signature: storedSignature,
      data: storedData,
    } = (await res.json()) as {
      address?: string
      chainId?: string
      message?: string
      signature?: string
      data?: SIWXSession['data']
    }

    if (!storedAddress) return []

    // Case-insensitive comparison for EVM; exact for Solana (base58 is case-sensitive)
    const isEvm = chainId.startsWith('eip155:')
    const matches = isEvm
      ? storedAddress.toLowerCase() === address.toLowerCase()
      : storedAddress === address

    if (!matches) return []

    if (!storedMessage || !storedSignature || !storedData) return []

    return [
      {
        data: {
          ...storedData,
          accountAddress: storedAddress,
          chainId: (storedChainId ?? chainId) as CaipNetworkId,
        },
        message: storedMessage,
        signature: storedSignature,
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
declare global {
  var __flexProverAppKitInitialized__: boolean | undefined
}

if (!globalThis.__flexProverAppKitInitialized__) {
  createAppKit({
    adapters: [wagmiAdapter, solanaAdapter],
    networks: [flareNetwork, mainnet, solana],
    defaultNetwork: flareNetwork,
    projectId,
    siwx,
    metadata: {
      name: 'FlexProver',
      description: 'Sybil-resistant reputation engine. Bind your CEX PNL to your wallet identity using secure Flare TEE enclaves.',
      url: process.env.NEXT_PUBLIC_APP_URL ?? 'http://localhost:3000',
      icons: [`${process.env.NEXT_PUBLIC_APP_URL ?? 'http://localhost:3000'}/icon.svg`],
    },
    features: { analytics: false },
  })
  globalThis.__flexProverAppKitInitialized__ = true
}

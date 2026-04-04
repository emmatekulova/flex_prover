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
    const msg = makeEvmMessage('abc12345')
    expect(extractNonce(msg)).toBe('abc12345')
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
    // Use Uint8Array.from to ensure same-realm Uint8Array (jsdom cross-realm fix)
    const messageBytes = Uint8Array.from(new TextEncoder().encode(message))
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
    // Use Uint8Array.from to ensure same-realm Uint8Array (jsdom cross-realm fix)
    const messageBytes = Uint8Array.from(new TextEncoder().encode(message))
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

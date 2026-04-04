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
    // Use Uint8Array.from to ensure same-realm Uint8Array (avoids jsdom cross-realm instanceof issues)
    const messageBytes = Uint8Array.from(new TextEncoder().encode(message))
    const signatureBytes = Uint8Array.from(bs58.decode(signature))

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

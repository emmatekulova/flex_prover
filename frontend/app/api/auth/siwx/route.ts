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

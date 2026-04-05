import { NextResponse } from "next/server"
import { decodeAbiParameters, parseAbiParameters, type Hex } from "viem"

// Non-indexed fields of AttestationPublished event:
// event AttestationPublished(address indexed teeAddress, bytes payload, bytes signature, uint256 timestamp)
const ATTESTATION_LOG_PARAMS = parseAbiParameters("bytes payload, bytes signature, uint256 timestamp")

// Coston2 Blockscout explorer base URL
const COSTON2_EXPLORER = "https://coston2-explorer.flare.network"

function resolveWhaleBackendUrl(): string {
  const explicit =
    process.env.WHALE_BACKEND_URL?.trim() ||
    process.env.NEXT_PUBLIC_BACKEND_URL?.trim()

  if (explicit) return explicit

  const flareEndpoint = process.env.FLARE_TEE_ENDPOINT?.trim()
  if (flareEndpoint) {
    try {
      const url = new URL(flareEndpoint)
      url.pathname = url.pathname.replace(/\/verify\/?$/, "")
      return url.toString().replace(/\/$/, "")
    } catch {
      // ignore invalid URL, fallback below
    }
  }

  return "http://localhost:8080"
}

/**
 * Extracts a 0x-prefixed 32-byte tx hash from a Coston2 explorer URL.
 * Accepts both /tx/0x... and /transactions/0x... path formats.
 */
function extractTxHash(attestationUrl: string): string | null {
  try {
    const { pathname } = new URL(attestationUrl)
    const m = pathname.match(/\/(?:tx|transaction)\/(0x[0-9a-fA-F]{64})/i)
    return m?.[1] ?? null
  } catch {
    return null
  }
}

/**
 * Fetches the first log of a Coston2 transaction via the Blockscout v2 API,
 * ABI-decodes the log data as (bytes payload, bytes signature, uint256 timestamp),
 * JSON-parses the payload bytes, and returns the wallet address embedded in the proof.
 *
 * Returns { wallet: "0x..." } on success, or { error: "reason" } on any failure.
 */
async function resolveProofWallet(
  attestationUrl: string,
): Promise<{ wallet: string } | { error: string }> {
  const txHash = extractTxHash(attestationUrl)
  if (!txHash) {
    return { error: "Could not extract a valid transaction hash from the attestation URL" }
  }

  // Blockscout v2 — returns logs array under `items`
  let logsRes: Response
  try {
    logsRes = await fetch(`${COSTON2_EXPLORER}/api/v2/transactions/${txHash}/logs`, {
      cache: "no-store",
    })
  } catch {
    return { error: "Failed to reach Coston2 explorer — check network connectivity" }
  }

  if (!logsRes.ok) {
    return { error: `Coston2 explorer returned HTTP ${logsRes.status} for transaction logs` }
  }

  const logsJson = (await logsRes.json()) as {
    items?: Array<{ data: string; topics?: string[] }>
  }

  const firstLog = logsJson.items?.[0]
  if (!firstLog?.data || firstLog.data === "0x") {
    return { error: "Transaction has no log data — is this the right transaction?" }
  }

  // ABI-decode the log data: (bytes payload, bytes signature, uint256 timestamp)
  let payloadHex: Hex
  try {
    const [decoded] = decodeAbiParameters(ATTESTATION_LOG_PARAMS, firstLog.data as Hex)
    // viem returns `bytes` as a 0x-prefixed hex string
    payloadHex = decoded as Hex
  } catch (err) {
    return { error: `Failed to ABI-decode log data: ${String(err)}` }
  }

  // payloadHex is the raw JSON bytes as 0x-prefixed hex — decode to UTF-8
  let proofJson: { wallet?: unknown }
  try {
    const jsonText = Buffer.from(payloadHex.slice(2), "hex").toString("utf-8")
    proofJson = JSON.parse(jsonText)
  } catch {
    return { error: "Decoded log payload is not valid JSON" }
  }

  if (typeof proofJson.wallet !== "string" || !proofJson.wallet) {
    return { error: "Proof payload does not contain a wallet field" }
  }

  return { wallet: proofJson.wallet }
}

export async function POST(request: Request) {
  const backendUrl = resolveWhaleBackendUrl()

  let body: string
  try {
    body = await request.text()
  } catch {
    return NextResponse.json({ message: "Failed to read request body" }, { status: 400 })
  }

  // Parse the request to extract walletAddress and attestationUrl for the ownership check
  let parsed: { walletAddress?: string; attestationUrl?: string } = {}
  try {
    parsed = JSON.parse(body)
  } catch {
    return NextResponse.json({ message: "Request body is not valid JSON" }, { status: 400 })
  }

  const { walletAddress, attestationUrl } = parsed

  if (!walletAddress || !attestationUrl) {
    return NextResponse.json(
      { message: "walletAddress and attestationUrl are required" },
      { status: 400 },
    )
  }

  // Decode the on-chain proof and verify the wallet address matches the connected wallet
  const proofResult = await resolveProofWallet(attestationUrl)

  if ("error" in proofResult) {
    return NextResponse.json(
      { message: `Could not read proof: ${proofResult.error}` },
      { status: 422 },
    )
  }

  // Case-insensitive EVM address comparison
  if (proofResult.wallet.toLowerCase() !== walletAddress.toLowerCase()) {
    return NextResponse.json(
      {
        message:
          `Proof wallet mismatch: the attestation belongs to ${proofResult.wallet}, ` +
          `but your connected wallet is ${walletAddress}. ` +
          `Connect the wallet you used when generating the proof.`,
      },
      { status: 403 },
    )
  }

  // Wallet matches — proxy to the whale backend
  try {
    const upstream = await fetch(`${backendUrl}/verify`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body,
      cache: "no-store",
    })

    const text = await upstream.text()

    if (!upstream.ok) {
      return NextResponse.json(
        { message: text || "Verification request failed on whale backend" },
        { status: upstream.status },
      )
    }

    return new NextResponse(text, {
      status: 200,
      headers: { "content-type": upstream.headers.get("content-type") || "application/json" },
    })
  } catch {
    return NextResponse.json(
      {
        message:
          "Cannot reach whale backend. Set WHALE_BACKEND_URL (or FLARE_TEE_ENDPOINT/NEXT_PUBLIC_BACKEND_URL) and ensure backend is running.",
      },
      { status: 502 },
    )
  }
}

import { execFile } from "node:child_process"
import { promisify } from "node:util"
import path from "node:path"
import { NextResponse } from "next/server"

import type { AttestationApiResponse, AttestationSubmitRequest } from "@/lib/attestation"

export const runtime = "nodejs"

const execFileAsync = promisify(execFile)
const TX_HASH_REGEX = /^tx=(0x[a-fA-F0-9]+)$/m

interface ReadAttestationOutput {
  teeAddress: string
  timestamp: string
  signature: string
  payload: {
    wallet?: string
    startSnapshot?: { date?: string }
    endSnapshot?: { date?: string }
    growthPercent?: string
  }
}

function parseDateToUnixSeconds(input: string): number {
  const parsed = Date.parse(`${input}T00:00:00Z`)
  return Number.isNaN(parsed) ? 0 : Math.floor(parsed / 1000)
}

function toolsDir(): string {
  return path.resolve(process.cwd(), "..", "fce-sign", "go", "tools")
}

export async function POST(request: Request) {
  try {
    const body = (await request.json()) as AttestationSubmitRequest
    const apiKey = body.apiKey?.trim()
    const secretKey = body.secretKey?.trim()
    const wallet = body.wallet?.trim()
    const windowDays = body.windowDays && body.windowDays > 0 ? body.windowDays : 7

    if (!apiKey || !secretKey || !wallet) {
      return NextResponse.json({ error: "apiKey, secretKey, and wallet are required" }, { status: 400 })
    }

    const publishArgs = [
      "run",
      "./cmd/publish-attestation",
      "-apiKey", apiKey,
      "-secretKey", secretKey,
      "-wallet", wallet,
      "-days", String(windowDays),
    ]

    const publishResult = await execFileAsync("go", publishArgs, {
      cwd: toolsDir(),
      timeout: 180000,
      maxBuffer: 1024 * 1024,
    })

    const txMatch = publishResult.stdout.match(TX_HASH_REGEX)
    if (!txMatch) {
      return NextResponse.json(
        { error: "publish-attestation completed but tx hash was not found in output" },
        { status: 502 },
      )
    }

    const txHash = txMatch[1]
    const readResult = await execFileAsync(
      "go",
      ["run", "./cmd/read-attestation", "-json", txHash],
      {
        cwd: toolsDir(),
        timeout: 120000,
        maxBuffer: 1024 * 1024,
      },
    )

    const parsed = JSON.parse(readResult.stdout) as ReadAttestationOutput
    const payload = parsed.payload ?? {}

    const startDate = payload.startSnapshot?.date ?? ""
    const endDate = payload.endSnapshot?.date ?? ""

    const response: AttestationApiResponse = {
      result: {
        txHash,
        attestedWallet: payload.wallet ?? "",
        providedWallet: wallet,
        startDate,
        startTimestamp: parseDateToUnixSeconds(startDate),
        endDate,
        endTimestamp: parseDateToUnixSeconds(endDate),
        profitPercent: payload.growthPercent ?? "",
      },
    }

    return NextResponse.json(response)
  } catch (error) {
    if (error && typeof error === "object" && "stderr" in error) {
      const stderr = String((error as { stderr?: string }).stderr ?? "").trim()
      const sanitized = stderr || "attestation command execution failed"
      return NextResponse.json({ error: sanitized }, { status: 500 })
    }

    return NextResponse.json({ error: "Unknown backend error" }, { status: 500 })
  }
}

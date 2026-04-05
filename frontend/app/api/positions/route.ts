import { execFile } from "node:child_process"
import { promisify } from "node:util"
import path from "node:path"
import { NextResponse } from "next/server"

import type { PositionsFetchResponse } from "@/lib/attestation"

export const runtime = "nodejs"

const execFileAsync = promisify(execFile)

function toolsDir(): string {
  return path.resolve(process.cwd(), "..", "fce-sign", "go", "tools")
}

interface PositionsFetchRequest {
  exchange?: "binance" | "bitget"
  apiKey: string
  secretKey: string
  passphrase?: string
}

export async function POST(request: Request) {
  try {
    const body = (await request.json()) as PositionsFetchRequest
    const exchange = body.exchange ?? "binance"
    const apiKey = body.apiKey?.trim()
    const secretKey = body.secretKey?.trim()
    const passphrase = body.passphrase?.trim()

    if (!apiKey || !secretKey) {
      return NextResponse.json({ error: "apiKey and secretKey are required" }, { status: 400 })
    }
    if (exchange === "bitget" && !passphrase) {
      return NextResponse.json({ error: "passphrase is required for Bitget" }, { status: 400 })
    }

    let args: string[]
    if (exchange === "bitget") {
      args = [
        "run",
        "./cmd/fetch-positions-bitget",
        "-apiKey", apiKey,
        "-secretKey", secretKey,
        "-passphrase", passphrase!,
      ]
    } else {
      args = [
        "run",
        "./cmd/fetch-positions-binance",
        "-apiKey", apiKey,
        "-secretKey", secretKey,
      ]
    }

    const { stdout } = await execFileAsync("go", args, {
      cwd: toolsDir(),
      timeout: 60000,
      maxBuffer: 1024 * 1024,
    })

    const data = JSON.parse(stdout) as PositionsFetchResponse
    return NextResponse.json({
      ...data,
      positions: Array.isArray(data.positions) ? data.positions : [],
    })
  } catch (error) {
    if (error && typeof error === "object" && "stderr" in error) {
      const stderr = String((error as { stderr?: string }).stderr ?? "").trim()
      const sanitized = stderr || "fetch-positions command execution failed"
      return NextResponse.json({ error: sanitized }, { status: 500 })
    }
    return NextResponse.json({ error: "Unknown backend error" }, { status: 500 })
  }
}

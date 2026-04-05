import { NextResponse } from "next/server"

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

export async function GET() {
  const backendUrl = resolveWhaleBackendUrl()

  try {
    const upstream = await fetch(`${backendUrl}/nonce`, {
      method: "GET",
      cache: "no-store",
    })

    const text = await upstream.text()

    if (!upstream.ok) {
      return NextResponse.json(
        { message: text || "Failed to fetch nonce from whale backend" },
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

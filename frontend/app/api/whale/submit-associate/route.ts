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
    } catch { /* fallback */ }
  }
  return "http://localhost:8080"
}

export async function POST(request: Request) {
  const backendUrl = resolveWhaleBackendUrl()
  try {
    const body = await request.text()
    const upstream = await fetch(`${backendUrl}/submit-associate`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body,
      cache: "no-store",
    })
    const text = await upstream.text()
    return new NextResponse(text, {
      status: upstream.status,
      headers: { "content-type": upstream.headers.get("content-type") || "application/json" },
    })
  } catch {
    return NextResponse.json({ message: "Cannot reach whale backend." }, { status: 502 })
  }
}

"use client"

import { useEffect, useState } from "react"
import { motion, AnimatePresence } from "framer-motion"
import {
  Crown, Shield, CheckCircle2, AlertTriangle, Loader2,
  Wallet, Sparkles, ArrowLeft, RefreshCw, Link2,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { useSignMessage } from "wagmi"
import { recoverPublicKey, hashMessage, toHex, fromHex } from "viem"

const BACKEND_URL = "/api/whale"
const MIRROR_NODE = "https://testnet.mirrornode.hedera.com"

// Exported for tests
export function buildWhaleMessage(nonce: string, attestationUrl: string, hederaAccountId: string): string {
  return `FlexProver Whale Verification\n\nAttestation: ${attestationUrl}\nHedera Account: ${hederaAccountId}\nNonce: ${nonce}`
}

// Exported for tests
export function parseHederaAccountId(id: string): boolean {
  return /^\d+\.\d+\.\d+$/.test(id) && id.split(".").every(p => p.length > 0)
}

function isEvmAddress(address: string): boolean {
  return /^0x[0-9a-fA-F]{40}$/.test(address)
}

// Compress an uncompressed secp256k1 public key (65 bytes, 0x04 prefix) to 33 bytes
function compressPublicKey(uncompressed: `0x${string}`): string {
  const bytes = fromHex(uncompressed, "bytes")
  // bytes[0] = 0x04, bytes[1..32] = x, bytes[33..64] = y
  const prefix = (bytes[64] ?? 0) % 2 === 0 ? 0x02 : 0x03
  const compressed = new Uint8Array(33)
  compressed[0] = prefix
  compressed.set(bytes.slice(1, 33), 1)
  return toHex(compressed).slice(2) // strip 0x
}

type VerifyState = "idle" | "associating" | "verifying" | "minted" | "error"
type HederaLookupState = "idle" | "loading" | "found" | "not_found"

interface MirrorAccount {
  account?: string
  key?: { _type: string; key: string } | null
  max_automatic_token_associations?: number
}

interface WhalePageProps {
  walletAddress: string | null
  onNavigateToCreate?: () => void
}

export function WhalePage({ walletAddress, onNavigateToCreate }: WhalePageProps) {
  const [attestationUrl, setAttestationUrl] = useState("")
  const [hederaAccountId, setHederaAccountId] = useState("")
  const [hederaLookup, setHederaLookup] = useState<HederaLookupState>("idle")
  const [verifyState, setVerifyState] = useState<VerifyState>("idle")
  const [errorMsg, setErrorMsg] = useState("")
  const [statusMsg, setStatusMsg] = useState("")

  // Cache the recovered public key across the verify flow
  const [recoveredPubKey, setRecoveredPubKey] = useState<string | null>(null)

  const { signMessageAsync } = useSignMessage()

  const isEvm = !!walletAddress && isEvmAddress(walletAddress)
  const isSolana = !!walletAddress && !isEvmAddress(walletAddress)

  // Auto-lookup Hedera account ID from EVM address via Mirror Node
  const runLookup = (address: string) => {
    setHederaLookup("loading")
    setHederaAccountId("")
    fetch(`${MIRROR_NODE}/api/v1/accounts/${address}`)
      .then(r => r.ok ? r.json() as Promise<MirrorAccount> : Promise.resolve(null))
      .then(data => {
        if (data?.account) {
          setHederaAccountId(data.account)
          setHederaLookup("found")
        } else {
          setHederaLookup("not_found")
        }
      })
      .catch(() => setHederaLookup("not_found"))
  }

  useEffect(() => {
    if (!isEvm || !walletAddress) return
    runLookup(walletAddress)
  }, [walletAddress, isEvm])

  const canSubmit =
    isEvm &&
    attestationUrl.startsWith("http") &&
    parseHederaAccountId(hederaAccountId) &&
    verifyState !== "verifying" &&
    verifyState !== "associating"

  // Step 1: sign + get nonce. Returns { nonce, signature, pubKey }
  const signVerification = async () => {
    if (!walletAddress) throw new Error("No wallet connected")

    const nonceRes = await fetch(`${BACKEND_URL}/nonce`)
    if (!nonceRes.ok) {
      const body = await nonceRes.json().catch(() => ({ message: "Failed to fetch nonce" })) as { message?: string }
      throw new Error(body.message ?? "Failed to fetch nonce")
    }
    const { nonce } = await nonceRes.json() as { nonce: string }
    const message = buildWhaleMessage(nonce, attestationUrl, hederaAccountId)
    const signature = await signMessageAsync({ message })

    // Recover compressed secp256k1 public key for Hedera signing
    const msgHash = hashMessage(message)
    const uncompressed = await recoverPublicKey({ hash: msgHash, signature })
    const compressed = compressPublicKey(uncompressed)
    setRecoveredPubKey(compressed)

    return { nonce, signature, pubKey: compressed }
  }


  const handleVerify = async () => {
    if (!walletAddress) return
    setVerifyState("verifying")
    setErrorMsg("")
    setStatusMsg("Signing verification message…")

    try {
      const { nonce, signature } = await signVerification()

      setStatusMsg("Verifying proof on backend…")
      const verifyRes = await fetch(`${BACKEND_URL}/verify`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ walletAddress, attestationUrl, hederaAccountId, nonce, signature }),
      })
      const data = await verifyRes.json().catch(() => ({ status: "error", message: "Invalid response" })) as {
        status: string; message?: string
      }

      if (data.status === "minted") {
        setVerifyState("minted")
        setStatusMsg("")
        return
      }

      if (data.status === "needs_association") {
        // Account has no auto-association slots — ask user to associate manually
        throw new Error(
          data.message ??
          `Your Hedera account needs to associate the WHALE token first. ` +
          `Go to hashscan.io/testnet/token/${process.env.NEXT_PUBLIC_HEDERA_TOKEN_ID ?? "0.0.8515448"} ` +
          `and associate it via HashPack, then try again.`
        )
      }

      throw new Error(data.message ?? "Verification failed")
    } catch (err) {
      setErrorMsg(err instanceof Error ? err.message : "Unknown error")
      setVerifyState("error")
      setStatusMsg("")
    }
  }

  return (
    <div className="w-full max-w-lg mx-auto space-y-6">
      {/* Header */}
      <div className="text-center mb-8">
        <div className="w-14 h-14 rounded-full bg-primary/20 flex items-center justify-center mx-auto mb-4">
          <Crown className="w-7 h-7 text-primary" />
        </div>
        <h2 className="text-2xl font-bold text-foreground mb-2">Join the Whale Group</h2>
        <p className="text-muted-foreground">
          Prove your Binance PNL inside a Flare TEE enclave. Once your attestation is verified, claim your WHALE token — minted on Hedera and gated by real trading performance.
        </p>
      </div>

      {/* Step 1: Connect wallet */}
      <div className="p-5 rounded-xl bg-secondary/50 border border-border space-y-3">
        <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wide">Step 1 — Connect Wallet</p>
        {walletAddress ? (
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <CheckCircle2 className="w-4 h-4 text-primary" />
              <span className="text-sm font-mono text-foreground">
                {walletAddress.slice(0, 6)}…{walletAddress.slice(-4)}
              </span>
              {isSolana && (
                <span className="text-xs px-2 py-0.5 rounded bg-secondary text-muted-foreground">Solana</span>
              )}
            </div>
            <appkit-button balance="hide" size="sm" />
          </div>
        ) : (
          <div className="flex flex-col items-center gap-2">
            <p className="text-sm text-muted-foreground">Connect the EVM wallet you used to generate your proof.</p>
            <appkit-button balance="hide" />
          </div>
        )}
      </div>

      {/* Solana warning */}
      <AnimatePresence>
        {isSolana && (
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: 8 }}
            className="p-5 rounded-xl border border-yellow-500/30 bg-yellow-500/10 space-y-3"
          >
            <div className="flex items-center gap-2 text-yellow-400">
              <AlertTriangle className="w-5 h-5" />
              <span className="font-semibold">EVM Wallet Required</span>
            </div>
            <p className="text-sm text-muted-foreground">
              The WHALE token is a Hedera HTS token. You need a Hedera account to receive it — an EVM wallet alone is not enough.
            </p>
            <ul className="text-xs text-muted-foreground space-y-1 list-disc list-inside">
              <li>Switch to an EVM wallet using the button above</li>
              <li>Or create a Hedera account at{" "}
                <a href="https://portal.hedera.com" target="_blank" rel="noopener noreferrer" className="text-primary hover:underline">
                  portal.hedera.com
                </a>
              </li>
            </ul>
          </motion.div>
        )}
      </AnimatePresence>

      {/* No wallet nudges */}
      <AnimatePresence>
        {!walletAddress && (
          <>
            <motion.div
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: 8 }}
              transition={{ delay: 0.3, duration: 0.4 }}
              className="relative overflow-hidden rounded-2xl border border-primary/20 bg-gradient-to-br from-primary/10 via-card to-accent/10 p-6 text-center space-y-4"
            >
              {[0, 1, 2].map(i => (
                <motion.div
                  key={i}
                  className="absolute text-primary/30"
                  style={{ top: `${15 + i * 25}%`, left: `${10 + i * 30}%` }}
                  animate={{ y: [-4, 4, -4], opacity: [0.3, 0.7, 0.3] }}
                  transition={{ duration: 2.5 + i * 0.5, repeat: Infinity, ease: "easeInOut" }}
                >
                  <Sparkles className="w-4 h-4" />
                </motion.div>
              ))}
              <div className="relative z-10 space-y-1">
                <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wide">First time here?</p>
                <h3 className="text-lg font-bold text-foreground">Generate your proof first</h3>
                <p className="text-sm text-muted-foreground">
                  Generate a Flare TEE proof of your Binance PNL first. Once verified on-chain, you can claim your WHALE token here.
                </p>
              </div>
              <div className="relative z-10 flex flex-col items-center gap-2">
                <motion.div animate={{ x: [-4, 0, -4] }} transition={{ duration: 1.2, repeat: Infinity }}>
                  <ArrowLeft className="w-5 h-5 text-primary" />
                </motion.div>
                <Button onClick={onNavigateToCreate} className="bg-primary text-primary-foreground hover:bg-primary/90 px-6">
                  <Sparkles className="w-4 h-4 mr-2" />
                  Go to Create
                </Button>
              </div>
            </motion.div>

            <motion.div
              initial={{ opacity: 0, y: 12 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: 8 }}
              transition={{ delay: 0.5, duration: 0.4 }}
              className="flex items-start gap-3 p-4 rounded-xl border border-border bg-secondary/30"
            >
              <motion.div animate={{ scale: [1, 1.15, 1] }} transition={{ duration: 2, repeat: Infinity }}>
                <Wallet className="w-5 h-5 text-muted-foreground mt-0.5 shrink-0" />
              </motion.div>
              <div className="space-y-0.5">
                <p className="text-sm font-medium text-foreground">Already have a proof?</p>
                <p className="text-xs text-muted-foreground">
                  Connect your EVM wallet — your Hedera account is detected automatically.
                </p>
              </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>

      {/* Main form */}
      {isEvm && verifyState !== "minted" && (
        <div className="space-y-4">
          {/* Hedera account status */}
          <div className="p-4 rounded-xl bg-secondary/50 border border-border">
            <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wide mb-2">Hedera Account</p>
            {hederaLookup === "loading" && (
              <div className="flex items-center gap-2 text-muted-foreground">
                <Loader2 className="w-4 h-4 animate-spin" />
                <span className="text-sm">Looking up your Hedera account…</span>
              </div>
            )}
            {hederaLookup === "found" && (
              <div className="flex items-center gap-2 text-primary">
                <CheckCircle2 className="w-4 h-4" />
                <span className="text-sm font-mono">{hederaAccountId}</span>
              </div>
            )}
            {hederaLookup === "not_found" && (
              <div className="space-y-3">
                <div className="flex items-center gap-2 text-yellow-400">
                  <AlertTriangle className="w-4 h-4" />
                  <span className="text-sm">No Hedera account found for this wallet</span>
                </div>
                <p className="text-xs text-muted-foreground">
                  Create one at{" "}
                  <a href="https://portal.hedera.com" target="_blank" rel="noopener noreferrer" className="text-primary hover:underline">
                    portal.hedera.com
                  </a>{" "}
                  then enter your account ID below.
                </p>
                <div className="space-y-1">
                  <Label htmlFor="hedera-id-manual" className="text-xs text-muted-foreground">Hedera Account ID</Label>
                  <Input
                    id="hedera-id-manual"
                    placeholder="0.0.12345"
                    value={hederaAccountId}
                    onChange={e => setHederaAccountId(e.target.value)}
                    className="bg-secondary border-border text-foreground placeholder:text-muted-foreground h-10 text-sm"
                  />
                </div>
                <Button
                  size="sm" variant="ghost"
                  onClick={() => walletAddress && runLookup(walletAddress)}
                  className="text-muted-foreground hover:text-foreground"
                >
                  <RefreshCw className="w-3 h-3 mr-1" /> Retry lookup
                </Button>
              </div>
            )}
          </div>

          {/* Proof URL */}
          <div className="space-y-2">
            <Label htmlFor="attestation-url" className="text-sm text-foreground">
              Flare Proof URL
            </Label>
            <Input
              id="attestation-url"
              placeholder="https://coston2-explorer.flare.network/tx/0x..."
              value={attestationUrl}
              onChange={e => setAttestationUrl(e.target.value)}
              className="bg-secondary border-border text-foreground placeholder:text-muted-foreground h-12"
            />
            <p className="text-xs text-muted-foreground flex items-center gap-1">
              <Link2 className="w-3 h-3" />
              The Coston2 explorer URL from your proof card
            </p>
          </div>

          {/* Status message during flow */}
          {statusMsg && (
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <Loader2 className="w-4 h-4 animate-spin text-primary" />
              {statusMsg}
            </div>
          )}

          <Button
            onClick={handleVerify}
            disabled={!canSubmit}
            className="w-full h-12 bg-primary text-primary-foreground hover:bg-primary/90"
          >
            {verifyState === "verifying" || verifyState === "associating" ? (
              <>
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                {verifyState === "associating" ? "Associating token…" : "Verifying…"}
              </>
            ) : (
              <>
                <Shield className="w-4 h-4 mr-2" />
                Verify PNL &amp; Get WHALE Token
              </>
            )}
          </Button>
        </div>
      )}

      {/* Result panels */}
      <AnimatePresence>
        {verifyState === "minted" && (
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            className="p-5 rounded-xl bg-primary/10 border border-primary/30 space-y-3"
          >
            <div className="flex items-center gap-2 text-primary">
              <CheckCircle2 className="w-5 h-5" />
              <span className="font-semibold">WHALE Token Minted!</span>
            </div>
            <p className="text-sm text-muted-foreground">
              1 WHALE token has been sent to{" "}
              <span className="font-mono">{hederaAccountId}</span>.
              You can now join the private trading group.
            </p>
          </motion.div>
        )}

        {verifyState === "error" && (
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            className="p-4 rounded-xl bg-destructive/10 border border-destructive/30"
          >
            <div className="flex items-start gap-2">
              <AlertTriangle className="w-4 h-4 text-destructive mt-0.5" />
              <div>
                <p className="text-sm font-medium text-destructive">Verification Failed</p>
                <p className="text-xs text-muted-foreground mt-1">{errorMsg}</p>
              </div>
            </div>
            <Button
              variant="ghost"
              onClick={() => { setVerifyState("idle"); setStatusMsg("") }}
              className="w-full mt-3 text-muted-foreground hover:text-foreground"
            >
              Try Again
            </Button>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}

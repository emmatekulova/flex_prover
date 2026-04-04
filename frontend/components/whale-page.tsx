"use client"

import { useState } from "react"
import { motion, AnimatePresence } from "framer-motion"
import { Crown, Shield, CheckCircle2, AlertTriangle, Loader2, ExternalLink, Wallet, Sparkles, ArrowLeft } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { useSignMessage } from "wagmi"

const BACKEND_URL = process.env.NEXT_PUBLIC_BACKEND_URL ?? "http://localhost:8080"

// Exported for tests
export function buildWhaleMessage(nonce: string, attestationUrl: string, hederaAccountId: string): string {
  return `FlexProver Whale Verification\n\nAttestation: ${attestationUrl}\nHedera Account: ${hederaAccountId}\nNonce: ${nonce}`
}

// Exported for tests
export function parseHederaAccountId(id: string): boolean {
  return /^\d+\.\d+\.\d+$/.test(id) && id.split('.').every(p => p.length > 0)
}

type VerifyState = "idle" | "verifying" | "needs_association" | "minted" | "error"

interface WhalePageProps {
  walletAddress: string | null
  onNavigateToCreate?: () => void
}

export function WhalePage({ walletAddress, onNavigateToCreate }: WhalePageProps) {
  const [attestationUrl, setAttestationUrl] = useState("")
  const [hederaAccountId, setHederaAccountId] = useState("")
  const [verifyState, setVerifyState] = useState<VerifyState>("idle")
  const [errorMsg, setErrorMsg] = useState("")
  const [associationTokenId, setAssociationTokenId] = useState("")

  const { signMessageAsync } = useSignMessage()

  const canSubmit =
    !!walletAddress &&
    attestationUrl.startsWith("http") &&
    parseHederaAccountId(hederaAccountId) &&
    verifyState !== "verifying"

  const handleVerify = async () => {
    if (!walletAddress) return
    setVerifyState("verifying")
    setErrorMsg("")

    try {
      // 1. Fetch nonce
      const nonceRes = await fetch(`${BACKEND_URL}/nonce`)
      if (!nonceRes.ok) throw new Error("Failed to fetch nonce")
      const { nonce } = await nonceRes.json() as { nonce: string }

      // 2. Sign canonical message with EVM wallet
      const message = buildWhaleMessage(nonce, attestationUrl, hederaAccountId)
      const signature = await signMessageAsync({ message })

      // 3. Post to backend
      const verifyRes = await fetch(`${BACKEND_URL}/verify`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ walletAddress, attestationUrl, hederaAccountId, nonce, signature }),
      })
      const data = await verifyRes.json() as { status: string; message?: string }

      if (data.status === "minted") {
        setVerifyState("minted")
      } else if (data.status === "needs_association") {
        setAssociationTokenId(data.message ?? "")
        setVerifyState("needs_association")
      } else {
        setErrorMsg(data.message ?? "Verification failed")
        setVerifyState("error")
      }
    } catch (err) {
      setErrorMsg(err instanceof Error ? err.message : "Unknown error")
      setVerifyState("error")
    }
  }

  return (
    <div className="w-full max-w-lg mx-auto space-y-6">
      <div className="text-center mb-8">
        <div className="w-14 h-14 rounded-full bg-primary/20 flex items-center justify-center mx-auto mb-4">
          <Crown className="w-7 h-7 text-primary" />
        </div>
        <h2 className="text-2xl font-bold text-foreground mb-2">Join the Whale Group</h2>
        <p className="text-muted-foreground">
          Prove your PNL is above 20% and receive your Whale token to unlock private trading groups.
        </p>
      </div>

      {/* Step 1: Connect wallet — always visible */}
      <div className="p-5 rounded-xl bg-secondary/50 border border-border space-y-3">
        <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wide">Step 1 — Connect Wallet</p>
        {walletAddress ? (
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <CheckCircle2 className="w-4 h-4 text-primary" />
              <span className="text-sm font-mono text-foreground">
                {walletAddress.slice(0, 6)}…{walletAddress.slice(-4)}
              </span>
            </div>
            <appkit-button balance="hide" size="sm" />
          </div>
        ) : (
          <div className="flex flex-col items-center gap-2">
            <p className="text-sm text-muted-foreground">Connect the wallet you used to generate your proof.</p>
            <appkit-button balance="hide" />
          </div>
        )}
      </div>

      {/* New here? Animated nudge toward Create */}
      <AnimatePresence>
        {!walletAddress && (
          <motion.div
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: 8 }}
            transition={{ delay: 0.3, duration: 0.4 }}
            className="relative overflow-hidden rounded-2xl border border-primary/20 bg-gradient-to-br from-primary/10 via-card to-accent/10 p-6 text-center space-y-4"
          >
            {/* floating sparkles */}
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
                You need a verified Flare TEE proof before claiming your Whale token.
              </p>
            </div>

            {/* animated arrow + button */}
            <div className="relative z-10 flex flex-col items-center gap-2">
              <motion.div
                animate={{ x: [-4, 0, -4] }}
                transition={{ duration: 1.2, repeat: Infinity, ease: "easeInOut" }}
              >
                <ArrowLeft className="w-5 h-5 text-primary" />
              </motion.div>
              <Button
                onClick={onNavigateToCreate}
                className="bg-primary text-primary-foreground hover:bg-primary/90 px-6"
              >
                <Sparkles className="w-4 h-4 mr-2" />
                Go to Create
              </Button>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Returning user nudge */}
      <AnimatePresence>
        {!walletAddress && (
          <motion.div
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: 8 }}
            transition={{ delay: 0.5, duration: 0.4 }}
            className="flex items-start gap-3 p-4 rounded-xl border border-border bg-secondary/30"
          >
            <motion.div
              animate={{ scale: [1, 1.15, 1] }}
              transition={{ duration: 2, repeat: Infinity, ease: "easeInOut" }}
            >
              <Wallet className="w-5 h-5 text-muted-foreground mt-0.5 shrink-0" />
            </motion.div>
            <div className="space-y-0.5">
              <p className="text-sm font-medium text-foreground">Already have a proof?</p>
              <p className="text-xs text-muted-foreground">
                Just connect your wallet above and come back here — your proof link is all you need.
              </p>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      {walletAddress && verifyState !== "minted" && (
        <div className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="attestation-url" className="text-sm text-foreground">
              Flare Proof URL
            </Label>
            <Input
              id="attestation-url"
              placeholder="https://..."
              value={attestationUrl}
              onChange={e => setAttestationUrl(e.target.value)}
              className="bg-secondary border-border text-foreground placeholder:text-muted-foreground h-12"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="hedera-id" className="text-sm text-foreground">
              Hedera Account ID
            </Label>
            <Input
              id="hedera-id"
              placeholder="0.0.12345"
              value={hederaAccountId}
              onChange={e => setHederaAccountId(e.target.value)}
              className="bg-secondary border-border text-foreground placeholder:text-muted-foreground h-12"
            />
          </div>

          <Button
            onClick={handleVerify}
            disabled={!canSubmit}
            className="w-full h-12 bg-primary text-primary-foreground hover:bg-primary/90"
          >
            {verifyState === "verifying" ? (
              <>
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                Verifying...
              </>
            ) : (
              <>
                <Shield className="w-4 h-4 mr-2" />
                Verify PNL &amp; Get Token
              </>
            )}
          </Button>
        </div>
      )}

      <AnimatePresence>
        {verifyState === "needs_association" && (
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            className="p-5 rounded-xl bg-card border border-border space-y-3"
          >
            <div className="flex items-center gap-2 text-yellow-400">
              <AlertTriangle className="w-5 h-5" />
              <span className="font-semibold">Association Required</span>
            </div>
            <p className="text-sm text-muted-foreground">
              You need to associate your Hedera account with the Whale token before receiving it.
            </p>
            <a
              href={`https://hashscan.io/testnet/token/${associationTokenId.split(' ').pop()}`}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-1 text-sm text-primary hover:underline"
            >
              Associate via HashScan <ExternalLink className="w-3 h-3" />
            </a>
            <Button
              onClick={handleVerify}
              variant="outline"
              className="w-full"
            >
              I have associated — try again
            </Button>
          </motion.div>
        )}

        {verifyState === "minted" && (
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            className="p-5 rounded-xl bg-primary/10 border border-primary/30 space-y-3"
          >
            <div className="flex items-center gap-2 text-primary">
              <CheckCircle2 className="w-5 h-5" />
              <span className="font-semibold">Whale Token Minted!</span>
            </div>
            <p className="text-sm text-muted-foreground">
              1 WHALE token has been sent to your Hedera account <span className="font-mono">{hederaAccountId}</span>.
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
              onClick={() => setVerifyState("idle")}
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

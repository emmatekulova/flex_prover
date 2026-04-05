# Proof Card & Step 5 Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Embed the attestation result as Step 5 of the wizard and add a shareable Proof Card with a real QR code, Flarescan link, Share to X, Copy Link, and Download as PNG.

**Architecture:** The `celebration` appState is replaced: after attestation completes, the flow stays in `wizard` state at Step 5. `ProofCard` is rewritten to accept live `AttestationResult` data, render a real QR code pointing to the Flarescan TX, and expose three share actions.

**Tech Stack:** Next.js 15 App Router, React, Framer Motion, `qrcode.react` (QRCodeSVG), `html2canvas`, Tailwind CSS, Lucide icons

---

## File Map

| File | Action | Responsibility |
|------|--------|---------------|
| `frontend/components/proof-card.tsx` | Rewrite | Renders the visual card + share/download actions |
| `frontend/app/page.tsx` | Modify | Adds Step 5, redirects post-attestation to wizard step 5, removes celebration appState branch |

---

### Task 1: Rewrite `ProofCard` with live data, QR code, and share actions

**Files:**
- Modify: `frontend/components/proof-card.tsx`

The Flarescan TX URL pattern: `https://flarescan.com/tx/<txHash>`

- [ ] **Step 1: Replace `proof-card.tsx` with the following**

```tsx
"use client"

import { useRef, useState } from "react"
import { motion } from "framer-motion"
import { Shield, ExternalLink, Twitter, QrCode, TrendingUp, Copy, Download, CheckCircle2 } from "lucide-react"
import { QRCodeSVG } from "qrcode.react"
import { Button } from "@/components/ui/button"

interface ProofCardProps {
  walletAddress: string
  profitPercent: string
  txHash: string
  startDate: string
  endDate: string
}

export function ProofCard({ walletAddress, profitPercent, txHash, startDate, endDate }: ProofCardProps) {
  const cardRef = useRef<HTMLDivElement>(null)
  const [copiedLink, setCopiedLink] = useState(false)

  const txUrl = `https://flarescan.com/tx/${txHash}`
  const shortWallet = `${walletAddress.slice(0, 6)}...${walletAddress.slice(-4)}`
  const profitValue = profitPercent !== "" ? `${profitPercent}%` : "N/A"

  const handleShareToX = () => {
    const text = `Just verified my trading performance on-chain! ${profitValue} PNL proven with @FlexProver on Flare Network. #FlexProver #Flare #DeFi`
    window.open(
      `https://twitter.com/intent/tweet?text=${encodeURIComponent(text)}&url=${encodeURIComponent(txUrl)}`,
      "_blank",
    )
  }

  const handleCopyLink = async () => {
    try {
      await navigator.clipboard.writeText(txUrl)
    } catch {
      const el = document.createElement("textarea")
      el.value = txUrl
      el.style.cssText = "position:fixed;opacity:0"
      document.body.appendChild(el)
      el.select()
      document.execCommand("copy")
      document.body.removeChild(el)
    }
    setCopiedLink(true)
    setTimeout(() => setCopiedLink(false), 1500)
  }

  const handleDownloadPng = async () => {
    if (!cardRef.current) return
    const html2canvas = (await import("html2canvas")).default
    const canvas = await html2canvas(cardRef.current, {
      backgroundColor: "#0a0a0f",
      scale: 2,
      useCORS: true,
    })
    const link = document.createElement("a")
    link.download = `proof-${shortWallet}.png`
    link.href = canvas.toDataURL("image/png")
    link.click()
  }

  return (
    <div className="space-y-4">
      {/* The Proof Card — captured for PNG download */}
      <motion.div
        ref={cardRef}
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ duration: 0.4 }}
        className="relative overflow-hidden rounded-2xl p-[1px] bg-gradient-to-br from-primary/40 via-border to-accent/30"
      >
        <div className="rounded-2xl bg-[#0a0a0f] p-6 sm:p-8 space-y-6">
          {/* Header */}
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 rounded-lg bg-primary/20 flex items-center justify-center">
                <span className="text-primary font-bold text-sm">FP</span>
              </div>
              <span className="font-semibold text-sm text-muted-foreground">FlexProver</span>
            </div>
            <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-accent/10 border border-accent/30">
              <Shield className="w-4 h-4 text-accent" />
              <span className="text-xs font-medium text-accent">TEE Verified</span>
            </div>
          </div>

          {/* Wallet */}
          <div>
            <p className="text-xs text-muted-foreground mb-1">Wallet Address</p>
            <p className="text-lg sm:text-xl font-bold text-foreground font-mono tracking-tight break-all">
              {walletAddress}
            </p>
          </div>

          {/* PNL */}
          <div className="p-4 rounded-xl bg-primary/10 border border-primary/20">
            <div className="flex items-center gap-2 mb-1">
              <TrendingUp className="w-4 h-4 text-primary" />
              <span className="text-xs text-muted-foreground">Verified Trading Performance</span>
            </div>
            <p className="text-4xl font-bold text-primary">{profitValue}</p>
            <p className="text-xs text-muted-foreground mt-1">
              {startDate} — {endDate}
            </p>
          </div>

          {/* Footer: TX + QR */}
          <div className="flex items-end justify-between pt-2 border-t border-border/30">
            <div className="min-w-0 mr-4">
              <p className="text-xs text-muted-foreground mb-1">On-Chain Proof</p>
              <p className="text-xs font-mono text-foreground break-all">{txHash}</p>
            </div>
            <div className="w-16 h-16 shrink-0 rounded-lg bg-white flex items-center justify-center overflow-hidden p-1">
              <QRCodeSVG
                value={txUrl}
                size={52}
                bgColor="#ffffff"
                fgColor="#000000"
                level="M"
              />
            </div>
          </div>
        </div>
      </motion.div>

      {/* Action Buttons */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
        <Button
          onClick={handleShareToX}
          className="bg-foreground text-background hover:bg-foreground/90"
        >
          <Twitter className="w-4 h-4 mr-2" />
          Share to X
        </Button>
        <Button
          variant="outline"
          onClick={handleCopyLink}
          className="border-border text-foreground hover:bg-secondary"
        >
          {copiedLink ? (
            <>
              <CheckCircle2 className="w-4 h-4 mr-2 text-primary" />
              Copied!
            </>
          ) : (
            <>
              <Copy className="w-4 h-4 mr-2" />
              Copy Link
            </>
          )}
        </Button>
        <Button
          variant="outline"
          onClick={handleDownloadPng}
          className="border-border text-foreground hover:bg-secondary"
        >
          <Download className="w-4 h-4 mr-2" />
          Download PNG
        </Button>
      </div>

      {/* View on Flarescan */}
      <Button
        variant="ghost"
        className="w-full text-muted-foreground hover:text-foreground"
        onClick={() => window.open(txUrl, "_blank")}
      >
        <ExternalLink className="w-4 h-4 mr-2" />
        View Transaction on Flarescan
      </Button>
    </div>
  )
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/components/proof-card.tsx
git commit -m "feat: rewrite ProofCard with live attestation data, QR code, and share actions"
```

---

### Task 2: Add Step 5 to the wizard in `page.tsx` and remove the separate `celebration` state

**Files:**
- Modify: `frontend/app/page.tsx`

Changes needed:
1. Add `"Proof"` as step 5 in the `steps` array
2. In `runAttestationFlow`: replace `setAppState("celebration")` with `setCurrentStep(5)` (appState stays `"wizard"`)
3. Remove the `celebration` appState from the `AppState` type
4. Add step 5 content block inside the wizard's `AnimatePresence`
5. On step 5, hide the Back/Continue buttons and show a "Create Another Proof" reset link
6. Remove the entire `celebration` appState JSX section and the `copiedTx` / `copyTxHash` helpers (replaced by ProofCard)

- [ ] **Step 1: Update the `steps` array and `AppState` type**

Find and replace in `page.tsx`:

```tsx
// BEFORE
const steps = [
  { id: 1, title: "Wallet", description: "Connect your wallet" },
  { id: 2, title: "Exchange", description: "Link Binance API" },
  { id: 3, title: "Trade Part 3", description: "Coming soon" },
  { id: 4, title: "Generate", description: "Publish and fetch" },
]

type AppState = "landing" | "wizard" | "calculating" | "celebration"
```

```tsx
// AFTER
const steps = [
  { id: 1, title: "Wallet", description: "Connect your wallet" },
  { id: 2, title: "Exchange", description: "Link Binance API" },
  { id: 3, title: "Trade Part 3", description: "Coming soon" },
  { id: 4, title: "Generate", description: "Publish and fetch" },
  { id: 5, title: "Proof", description: "Your proof card" },
]

type AppState = "landing" | "wizard" | "calculating"
```

- [ ] **Step 2: Update `runAttestationFlow` to go to step 5 instead of `celebration`**

Find in `page.tsx`:

```tsx
      setAttestationResult(data.result)
      setAppState("celebration")
```

Replace with:

```tsx
      setAttestationResult(data.result)
      setCurrentStep(5)
```

- [ ] **Step 3: Remove `copiedTx` state and `copyTxHash` function**

Remove these lines (no longer needed, ProofCard handles copy):

```tsx
  const [copiedTx, setCopiedTx] = useState(false)
```

```tsx
  const copyTxHash = async () => {
    if (!attestationResult?.txHash) return

    try {
      await navigator.clipboard.writeText(attestationResult.txHash)
    } catch {
      const el = document.createElement("textarea")
      el.value = attestationResult.txHash
      el.style.position = "fixed"
      el.style.opacity = "0"
      document.body.appendChild(el)
      el.select()
      document.execCommand("copy")
      document.body.removeChild(el)
    }

    setCopiedTx(true)
    setTimeout(() => setCopiedTx(false), 1500)
  }
```

Also remove unused imports: `Link2`, `Copy` (if only used by copyTxHash), `Sparkles`.

- [ ] **Step 4: Add `ProofCard` import at the top of `page.tsx`**

Add alongside existing component imports:

```tsx
import { ProofCard } from "@/components/proof-card"
```

- [ ] **Step 5: Add step 5 content block inside the wizard's `AnimatePresence`**

After the `currentStep === 4` block (closing `</motion.div>` and before `</AnimatePresence>`), add:

```tsx
                    {currentStep === 5 && attestationResult && (
                      <motion.div
                        key="step5"
                        initial={{ opacity: 0, x: 20 }}
                        animate={{ opacity: 1, x: 0 }}
                        exit={{ opacity: 0, x: -20 }}
                        className="max-w-xl mx-auto space-y-6"
                      >
                        <div>
                          <h2 className="text-2xl font-bold text-foreground mb-2">Your Proof is Ready</h2>
                          <p className="text-muted-foreground">Published on-chain and verified. Share your proof below.</p>
                        </div>

                        <ProofCard
                          walletAddress={attestationResult.attestedWallet || attestationResult.providedWallet}
                          profitPercent={attestationResult.profitPercent}
                          txHash={attestationResult.txHash}
                          startDate={attestationResult.startDate}
                          endDate={attestationResult.endDate}
                        />
                      </motion.div>
                    )}
```

- [ ] **Step 6: Replace the Back/Continue nav buttons so they hide on step 5**

Find in `page.tsx`:

```tsx
                  <div className="flex gap-3 mt-8 max-w-xl mx-auto">
                    {currentStep > 1 && (
                      <Button
                        variant="outline"
                        onClick={prevStep}
                        className="flex-1 border-border text-foreground hover:bg-secondary"
                      >
                        Back
                      </Button>
                    )}
                    <Button
                      onClick={nextStep}
                      disabled={(currentStep === 1 && !walletConnected) || (currentStep === 2 && !apiVerified)}
                      className="flex-1 bg-primary text-primary-foreground hover:bg-primary/90"
                    >
                      {currentStep === 4 ? "Publish Attestation" : "Continue"}
                      <ArrowRight className="w-4 h-4 ml-2" />
                    </Button>
                  </div>
```

Replace with:

```tsx
                  {currentStep < 5 && (
                    <div className="flex gap-3 mt-8 max-w-xl mx-auto">
                      {currentStep > 1 && (
                        <Button
                          variant="outline"
                          onClick={prevStep}
                          className="flex-1 border-border text-foreground hover:bg-secondary"
                        >
                          Back
                        </Button>
                      )}
                      <Button
                        onClick={nextStep}
                        disabled={(currentStep === 1 && !walletConnected) || (currentStep === 2 && !apiVerified)}
                        className="flex-1 bg-primary text-primary-foreground hover:bg-primary/90"
                      >
                        {currentStep === 4 ? "Publish Attestation" : "Continue"}
                        <ArrowRight className="w-4 h-4 ml-2" />
                      </Button>
                    </div>
                  )}
                  {currentStep === 5 && (
                    <div className="flex justify-center mt-8">
                      <Button variant="ghost" onClick={resetApp} className="text-muted-foreground hover:text-foreground">
                        Create Another Proof
                      </Button>
                    </div>
                  )}
```

- [ ] **Step 7: Remove the entire `celebration` appState JSX block**

Delete this entire section from the `AnimatePresence` in the return:

```tsx
        {appState === "celebration" && attestationResult && (
          <motion.div
            key="celebration"
            ...
          >
            ...
          </motion.div>
        )}
```

(This is the block starting `{appState === "celebration" && attestationResult && (` and ending with its closing `)}`)

- [ ] **Step 8: Clean up unused imports from `page.tsx`**

Remove from the lucide import list any icons that are now unused. After the above changes, `Link2`, `Sparkles` are no longer used. Also `Copy` if only used in `copyTxHash`. The import line should become:

```tsx
import {
  ArrowRight,
  CheckCircle2,
  FileText,
  MoreVertical,
  Shield,
  Terminal,
  Wallet,
} from "lucide-react"
```

- [ ] **Step 9: Commit**

```bash
git add frontend/app/page.tsx
git commit -m "feat: integrate proof confirmation as step 5, remove separate celebration page"
```

---

### Task 3: Verify the integration runs

- [ ] **Step 1: Start the dev server and check for TypeScript/build errors**

```bash
cd /home/emma/Documents/flex_prover/frontend && npm run dev 2>&1 | head -40
```

Expected: server starts on port 3000, no fatal errors.

- [ ] **Step 2: Verify no unused-import linting errors (optional)**

```bash
cd /home/emma/Documents/flex_prover/frontend && npx tsc --noEmit 2>&1 | head -30
```

Expected: no output (or only pre-existing errors unrelated to the new code, since `ignoreBuildErrors: true`).

- [ ] **Step 3: Smoke test the full flow manually**

1. Open `http://localhost:3000`
2. Hover the logo → click Get Started
3. Connect wallet (step 1) → Continue
4. Enter Binance API keys (step 2) → Continue
5. Step 3 → Continue
6. Step 4 → Publish Attestation
7. Observe the calculating terminal screen
8. Stepper should now show Step 5 "Proof" as active
9. ProofCard renders with real wallet, PNL%, TX hash, and QR code
10. Click "View Transaction on Flarescan" → opens `https://flarescan.com/tx/<txHash>` in new tab
11. Click "Copy Link" → pastes the Flarescan URL
12. Click "Share to X" → opens Twitter intent with proof text
13. Click "Download PNG" → downloads `proof-0x1234...abcd.png`
14. Click "Create Another Proof" → returns to landing

- [ ] **Step 4: Final commit if any fixes were needed**

```bash
git add -p
git commit -m "fix: proof card integration cleanup"
```

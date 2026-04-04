"use client"

import { useEffect, useRef, useState } from "react"
import { AnimatePresence, motion } from "framer-motion"
import {
  ArrowRight,
  CheckCircle2,
  Copy,
  FileText,
  Link2,
  MoreVertical,
  Shield,
  Sparkles,
  Terminal,
  Wallet,
} from "lucide-react"

import { ApiKeyManager, type ApiKeyManagerHandle, type ApiKeySaveResult } from "@/components/api-key-manager"
import { DocsSheet } from "@/components/docs-sheet"
import { StepWizard } from "@/components/step-wizard"
import { Verifier } from "@/components/verifier"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import type { AttestationApiResponse, AttestationResult } from "@/lib/attestation"

const steps = [
  { id: 1, title: "Wallet", description: "Connect your wallet" },
  { id: 2, title: "Exchange", description: "Link Binance API" },
  { id: 3, title: "Trade Part 3", description: "Coming soon" },
  { id: 4, title: "Generate", description: "Publish and fetch" },
]

type AppState = "landing" | "wizard" | "calculating" | "celebration"
type Tab = "create" | "verify"

export default function FlexProver() {
  const [appState, setAppState] = useState<AppState>("landing")
  const [activeTab, setActiveTab] = useState<Tab>("create")
  const [currentStep, setCurrentStep] = useState(1)
  const [mousePos, setMousePos] = useState({ x: 0, y: 0 })
  const [logoRevealed, setLogoRevealed] = useState(false)
  const [docsOpen, setDocsOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  const [walletConnected, setWalletConnected] = useState(false)
  const [walletAddress, setWalletAddress] = useState<string | null>(null)

  const [apiVerified, setApiVerified] = useState(false)
  const [apiCredentials, setApiCredentials] = useState<ApiKeySaveResult | null>(null)
  const apiKeyRef = useRef<ApiKeyManagerHandle>(null)

  const [logs, setLogs] = useState<string[]>([])
  const [submitError, setSubmitError] = useState<string | null>(null)
  const [attestationResult, setAttestationResult] = useState<AttestationResult | null>(null)
  const [copiedTx, setCopiedTx] = useState(false)

  useEffect(() => {
    let cancelled = false
    const check = async () => {
      try {
        const res = await fetch("/api/auth/siwx?action=session")
        if (!res.ok || cancelled) return
        const data = (await res.json()) as { address?: string | null }
        if (!cancelled) {
          setWalletAddress(data.address ?? null)
          setWalletConnected(!!data.address)
        }
      } catch {
        // leave current state as-is
      }
    }

    void check()
    const interval = setInterval(() => void check(), 2000)

    return () => {
      cancelled = true
      clearInterval(interval)
    }
  }, [])

  useEffect(() => {
    const handleMouseMove = (e: MouseEvent) => {
      if (appState !== "landing" || logoRevealed) return

      const rect = containerRef.current?.getBoundingClientRect()
      if (!rect) return

      const x = e.clientX - rect.left
      const y = e.clientY - rect.top
      setMousePos({ x, y })

      const centerX = rect.width / 2
      const centerY = rect.height / 2
      const distance = Math.sqrt(Math.pow(x - centerX, 2) + Math.pow(y - centerY, 2))
      if (distance < 150) setLogoRevealed(true)
    }

    window.addEventListener("mousemove", handleMouseMove)
    return () => window.removeEventListener("mousemove", handleMouseMove)
  }, [appState, logoRevealed])

  const addLog = async (message: string, delayMs = 350) => {
    setLogs((prev) => [...prev, message])
    await new Promise((resolve) => setTimeout(resolve, delayMs))
  }

  const runAttestationFlow = async () => {
    if (!walletAddress || !apiCredentials) {
      setSubmitError("Wallet and Binance API credentials are required.")
      return
    }

    const apiKey = apiCredentials.keys.apiKey?.trim() ?? ""
    const secretKey = apiCredentials.keys.secretKey?.trim() ?? ""
    if (!apiKey || !secretKey) {
      setSubmitError("Both Binance API key and secret key are required.")
      return
    }

    setSubmitError(null)
    setLogs([])
    setAppState("calculating")

    try {
      await addLog("Preparing attestation request...")
      await addLog("Sending Binance credentials + wallet to backend...")

      const res = await fetch("/api/attestation", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          apiKey,
          secretKey,
          wallet: walletAddress,
          windowDays: 7,
        }),
      })

      if (!res.ok) {
        const errorJson = (await res.json()) as { error?: string }
        throw new Error(errorJson.error ?? "Attestation request failed")
      }

      await addLog("Publishing attestation on-chain...")
      const data = (await res.json()) as AttestationApiResponse
      await addLog(`Reading attestation by TX hash: ${data.result.txHash.slice(0, 12)}...`)
      await addLog("Attestation fetched successfully.")

      setAttestationResult(data.result)
      setAppState("celebration")
    } catch (error) {
      const message = error instanceof Error ? error.message : "Unknown error"
      setSubmitError(message)
      setAppState("wizard")
      setCurrentStep(4)
    }
  }

  const nextStep = async () => {
    setSubmitError(null)

    if (currentStep === 2) {
      const saved = await apiKeyRef.current?.save()
      if (!saved) {
        setSubmitError("Please provide valid exchange credentials.")
        return
      }
      if (saved.exchange !== "binance") {
        setSubmitError("Only Binance is supported in this flow.")
        return
      }
      setApiCredentials(saved)
    }

    if (currentStep < 4) {
      setCurrentStep((prev) => prev + 1)
      return
    }

    await runAttestationFlow()
  }

  const prevStep = () => {
    setSubmitError(null)
    if (currentStep > 1) setCurrentStep((prev) => prev - 1)
  }

  const resetApp = () => {
    setAppState("landing")
    setCurrentStep(1)
    setApiVerified(false)
    setApiCredentials(null)
    setAttestationResult(null)
    setSubmitError(null)
    setLogs([])
    setLogoRevealed(false)
  }

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

  const Header = ({ showTabs = false }: { showTabs?: boolean }) => (
    <header className="flex items-center justify-between p-4 sm:p-6">
      <div className="flex items-center gap-4">
        <button onClick={resetApp} className="flex items-center gap-2 hover:opacity-80 transition-opacity">
          <div className="w-8 h-8 rounded-lg bg-primary/20 flex items-center justify-center">
            <span className="text-primary font-bold text-sm">FP</span>
          </div>
          <span className="font-semibold text-foreground">FlexProver</span>
        </button>

        {showTabs && (
          <div className="hidden sm:flex items-center gap-1 p-1 rounded-lg bg-secondary/50 border border-border">
            <button
              onClick={() => setActiveTab("create")}
              className={`px-4 py-1.5 rounded-md text-sm font-medium transition-all ${
                activeTab === "create"
                  ? "bg-primary text-primary-foreground"
                  : "text-muted-foreground hover:text-foreground"
              }`}
            >
              Create
            </button>
            <button
              onClick={() => setActiveTab("verify")}
              className={`px-4 py-1.5 rounded-md text-sm font-medium transition-all ${
                activeTab === "verify"
                  ? "bg-primary text-primary-foreground"
                  : "text-muted-foreground hover:text-foreground"
              }`}
            >
              Verify
            </button>
          </div>
        )}
      </div>

      <div className="flex items-center gap-2">
        {showTabs && (
          <Button variant="ghost" onClick={resetApp} className="text-muted-foreground hover:text-foreground">
            Back
          </Button>
        )}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon" className="text-muted-foreground hover:text-foreground">
              <MoreVertical className="w-5 h-5" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="bg-card border-border">
            <DropdownMenuItem
              onClick={() => setDocsOpen(true)}
              className="text-foreground focus:bg-secondary cursor-pointer"
            >
              <FileText className="w-4 h-4 mr-2" />
              Documentation
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  )

  return (
    <main className="min-h-screen bg-background">
      <AnimatePresence mode="wait">
        {appState === "landing" && (
          <motion.div
            key="landing"
            ref={containerRef}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="min-h-screen flex flex-col"
          >
            <Header />
            <div className="flex-1 flex flex-col items-center justify-center p-6 relative overflow-hidden">
              <div
                className="absolute inset-0 opacity-[0.03]"
                style={{
                  backgroundImage: `
                    linear-gradient(rgba(255,255,255,0.1) 1px, transparent 1px),
                    linear-gradient(90deg, rgba(255,255,255,0.1) 1px, transparent 1px)
                  `,
                  backgroundSize: "64px 64px",
                }}
              />

              {!logoRevealed && (
                <motion.div
                  className="absolute w-96 h-96 rounded-full pointer-events-none"
                  style={{
                    background: "radial-gradient(circle, rgba(168,85,247,0.15) 0%, transparent 70%)",
                    left: mousePos.x - 192,
                    top: mousePos.y - 192,
                  }}
                />
              )}

              <motion.div
                initial={{ opacity: 1, y: 0 }}
                animate={{
                  opacity: logoRevealed ? 0.3 : 1,
                  y: logoRevealed ? -100 : 0,
                  scale: logoRevealed ? 0.8 : 1,
                }}
                transition={{ duration: 0.6 }}
                className="relative z-10 mb-8"
              >
                <div className="flex items-center gap-3">
                  <div className="w-14 h-14 rounded-xl bg-primary/20 border border-primary/30 flex items-center justify-center">
                    <span className="text-2xl font-bold text-primary">FP</span>
                  </div>
                  <span className="text-3xl font-bold text-foreground">FlexProver</span>
                </div>
              </motion.div>

              <AnimatePresence>
                {logoRevealed && (
                  <motion.div
                    initial={{ opacity: 0, y: 50, scale: 0.95 }}
                    animate={{ opacity: 1, y: 0, scale: 1 }}
                    transition={{ duration: 0.5, delay: 0.2 }}
                    className="relative z-10 max-w-lg mx-auto"
                  >
                    <div className="rounded-2xl border border-border/50 bg-card/30 backdrop-blur-xl p-8 shadow-2xl">
                      <div className="relative space-y-6">
                        <div className="flex items-center gap-2 text-primary text-sm font-medium">
                          <Shield className="w-4 h-4" />
                          <span>Proof of Flex</span>
                        </div>
                        <p className="text-lg text-foreground/80 leading-relaxed">
                          Submit Binance attestation data, publish it on-chain, and immediately fetch it back.
                        </p>
                        <Button
                          onClick={() => setAppState("wizard")}
                          className="w-full bg-primary text-primary-foreground hover:bg-primary/90 h-12 text-base"
                        >
                          Get Started
                          <ArrowRight className="w-4 h-4 ml-2" />
                        </Button>
                      </div>
                    </div>
                  </motion.div>
                )}
              </AnimatePresence>
            </div>
          </motion.div>
        )}

        {appState === "wizard" && (
          <motion.div
            key="wizard"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="min-h-screen flex flex-col"
          >
            <Header showTabs />

            <div className="flex-1 p-6">
              {activeTab === "create" ? (
                <div className="max-w-5xl mx-auto">
                  <StepWizard steps={steps} currentStep={currentStep} />

                  <AnimatePresence mode="wait">
                    {currentStep === 1 && (
                      <motion.div
                        key="step1"
                        initial={{ opacity: 0, x: 20 }}
                        animate={{ opacity: 1, x: 0 }}
                        exit={{ opacity: 0, x: -20 }}
                        className="max-w-xl mx-auto space-y-6"
                      >
                        <div>
                          <h2 className="text-2xl font-bold text-foreground mb-2">Connect Your Wallet</h2>
                          <p className="text-muted-foreground">Connect your wallet to get started.</p>
                        </div>

                        <div className="flex justify-center">
                          <appkit-button balance="hide" />
                        </div>

                        {walletConnected && walletAddress && (
                          <div className="p-4 rounded-xl bg-secondary/50 border border-border">
                            <div className="flex items-center justify-between">
                              <div className="flex items-center gap-3">
                                <div className="w-10 h-10 rounded-full bg-primary/20 flex items-center justify-center">
                                  <Wallet className="w-5 h-5 text-primary" />
                                </div>
                                <div>
                                  <p className="font-medium text-foreground font-mono text-sm">
                                    {walletAddress.slice(0, 6)}...{walletAddress.slice(-4)}
                                  </p>
                                  <p className="text-sm text-muted-foreground">Verified via SIWX</p>
                                </div>
                              </div>
                              <CheckCircle2 className="w-5 h-5 text-primary" />
                            </div>
                          </div>
                        )}
                      </motion.div>
                    )}

                    {currentStep === 2 && (
                      <motion.div
                        key="step2"
                        initial={{ opacity: 0, x: 20 }}
                        animate={{ opacity: 1, x: 0 }}
                        exit={{ opacity: 0, x: -20 }}
                        className="max-w-xl mx-auto space-y-6"
                      >
                        <div>
                          <h2 className="text-2xl font-bold text-foreground mb-2">Link Your Exchange Account</h2>
                          <p className="text-muted-foreground">Provide your Binance read-only API credentials.</p>
                        </div>

                        <ApiKeyManager ref={apiKeyRef} onValidChange={setApiVerified} />
                      </motion.div>
                    )}

                    {currentStep === 3 && (
                      <motion.div
                        key="step3"
                        initial={{ opacity: 0, x: 20 }}
                        animate={{ opacity: 1, x: 0 }}
                        exit={{ opacity: 0, x: -20 }}
                        className="max-w-xl mx-auto"
                      >
                        <div className="rounded-xl border border-border bg-card p-8 text-center">
                          <h2 className="text-2xl font-bold text-foreground">Trade Part 3</h2>
                          <p className="text-muted-foreground mt-3">Coming soon</p>
                        </div>
                      </motion.div>
                    )}

                    {currentStep === 4 && (
                      <motion.div
                        key="step4"
                        initial={{ opacity: 0, x: 20 }}
                        animate={{ opacity: 1, x: 0 }}
                        exit={{ opacity: 0, x: -20 }}
                        className="max-w-xl mx-auto space-y-6"
                      >
                        <div>
                          <h2 className="text-2xl font-bold text-foreground mb-2">Review & Publish</h2>
                          <p className="text-muted-foreground">Submit to publish attestation and fetch it by TX hash.</p>
                        </div>

                        <div className="space-y-3">
                          <div className="p-4 rounded-xl bg-secondary/50 border border-border">
                            <p className="text-xs text-muted-foreground mb-1">Provided wallet address</p>
                            <p className="font-semibold text-foreground break-all">{walletAddress ?? "-"}</p>
                          </div>
                          <div className="p-4 rounded-xl bg-secondary/50 border border-border">
                            <p className="text-xs text-muted-foreground mb-1">Binance credentials</p>
                            <p className="font-semibold text-foreground">{apiCredentials ? "Ready" : "Not captured yet"}</p>
                          </div>
                        </div>
                      </motion.div>
                    )}
                  </AnimatePresence>

                  {submitError && (
                    <div className="max-w-xl mx-auto mt-6 p-3 rounded-lg border border-red-500/30 bg-red-500/10 text-red-300 text-sm">
                      {submitError}
                    </div>
                  )}

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
                </div>
              ) : (
                <Verifier />
              )}
            </div>
          </motion.div>
        )}

        {appState === "calculating" && (
          <motion.div
            key="calculating"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="min-h-screen flex flex-col items-center justify-center p-6"
          >
            <div className="w-full max-w-lg">
              <div className="flex items-center gap-3 mb-6">
                <Terminal className="w-6 h-6 text-primary" />
                <h2 className="text-xl font-bold text-foreground">Running backend flow</h2>
              </div>

              <div className="rounded-xl bg-card border border-border p-4 font-mono text-sm">
                <div className="space-y-2">
                  {logs.map((log, index) => (
                    <motion.div
                      key={`${log}-${index}`}
                      initial={{ opacity: 0, x: -10 }}
                      animate={{ opacity: 1, x: 0 }}
                      className="flex items-start gap-2"
                    >
                      <span className="text-primary">&gt;</span>
                      <span className={index === logs.length - 1 ? "text-primary" : "text-muted-foreground"}>{log}</span>
                    </motion.div>
                  ))}
                </div>
              </div>
            </div>
          </motion.div>
        )}

        {appState === "celebration" && attestationResult && (
          <motion.div
            key="celebration"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="min-h-screen flex flex-col"
          >
            <Header />

            <div className="flex-1 p-6">
              <div className="max-w-xl mx-auto">
                <motion.div
                  initial={{ scale: 0.8, opacity: 0 }}
                  animate={{ scale: 1, opacity: 1 }}
                  transition={{ type: "spring", damping: 10, delay: 0.2 }}
                  className="text-center mb-8"
                >
                  <div className="relative inline-block mb-4">
                    <div className="w-20 h-20 rounded-full bg-primary/20 flex items-center justify-center">
                      <Sparkles className="w-10 h-10 text-primary" />
                    </div>
                    <div className="absolute -top-1 -right-1 w-8 h-8 rounded-full bg-accent flex items-center justify-center">
                      <CheckCircle2 className="w-5 h-5 text-accent-foreground" />
                    </div>
                  </div>
                  <h1 className="text-3xl font-bold text-foreground mb-2">Attestation Ready</h1>
                  <p className="text-muted-foreground">Published on-chain and read back successfully.</p>
                </motion.div>

                <div className="rounded-xl border border-border bg-card p-6 space-y-4">
                  <div>
                    <p className="text-xs text-muted-foreground">Wallet (from attestation)</p>
                    <p className="text-foreground break-all font-medium">{attestationResult.attestedWallet || "-"}</p>
                  </div>

                  <div>
                    <p className="text-xs text-muted-foreground">Provided wallet address</p>
                    <p className="text-foreground break-all font-medium">{attestationResult.providedWallet || "-"}</p>
                  </div>

                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    <div>
                      <p className="text-xs text-muted-foreground">Start date / timestamp</p>
                      <p className="text-foreground font-medium">
                        {attestationResult.startDate || "-"} / {attestationResult.startTimestamp || 0}
                      </p>
                    </div>
                    <div>
                      <p className="text-xs text-muted-foreground">End date / timestamp</p>
                      <p className="text-foreground font-medium">
                        {attestationResult.endDate || "-"} / {attestationResult.endTimestamp || 0}
                      </p>
                    </div>
                  </div>

                  <div>
                    <p className="text-xs text-muted-foreground">Profit percentage</p>
                    <p className="text-foreground font-semibold text-lg">
                      {attestationResult.profitPercent === "" ? "-" : `${attestationResult.profitPercent}%`}
                    </p>
                  </div>

                  <div className="pt-2 border-t border-border">
                    <p className="text-xs text-muted-foreground">Transaction hash</p>
                    <div className="flex gap-2 items-center">
                      <p className="text-foreground break-all font-mono text-sm">{attestationResult.txHash}</p>
                      <Button variant="ghost" size="icon" onClick={copyTxHash} className="shrink-0">
                        {copiedTx ? <CheckCircle2 className="w-4 h-4 text-primary" /> : <Copy className="w-4 h-4" />}
                      </Button>
                    </div>
                  </div>
                </div>

                <div className="mt-6 text-center">
                  <Button variant="ghost" onClick={resetApp} className="text-muted-foreground hover:text-foreground">
                    <Link2 className="w-4 h-4 mr-2" />
                    Create Another Proof
                  </Button>
                </div>
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      <AnimatePresence>
        {docsOpen && <DocsSheet open={docsOpen} onClose={() => setDocsOpen(false)} />}
      </AnimatePresence>
    </main>
  )
}

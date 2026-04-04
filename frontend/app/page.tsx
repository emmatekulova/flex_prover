"use client"

import { useState, useEffect, useRef, memo } from "react"
import { motion, AnimatePresence } from "framer-motion"
import {
  Wallet,
  Eye,
  Shield,
  CheckCircle2,
  Terminal,
  ArrowRight,
  MoreVertical,
  FileText,
  Download,
  Twitter,
  Sparkles,
  Link2,
  TrendingUp,
  TrendingDown,
  ChevronRight,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { StepWizard } from "@/components/step-wizard"
import { FlexCard } from "@/components/flex-card"
import { Verifier } from "@/components/verifier"
import { DocsSheet } from "@/components/docs-sheet"
import { ApiKeyManager, type ApiKeyManagerHandle } from "@/components/api-key-manager"

// Trade data type
interface TradeData {
  id: string
  pair: string
  direction: "long" | "short"
  entryPrice: number
  exitPrice?: number
  currentPrice?: number
  liquidationPrice?: number
  pnlPercent: number
  pnlAbsolute: number
  volume: number
  duration: string
  timeframe: string
  isOpen: boolean
  openedAt: string
  closedAt?: string
}

// Mock trade data
const mockOpenPositions: TradeData[] = [
  { id: "1", pair: "BTC/USDT", direction: "long", entryPrice: 68500, currentPrice: 95200, liquidationPrice: 52000, pnlPercent: 39, pnlAbsolute: 26700, volume: 100000, duration: "12 days", timeframe: "4H", isOpen: true, openedAt: "Mar 22, 2026" },
  { id: "2", pair: "ETH/USDT", direction: "short", entryPrice: 4200, currentPrice: 3850, liquidationPrice: 5100, pnlPercent: 8.3, pnlAbsolute: 3500, volume: 42000, duration: "3 days", timeframe: "1H", isOpen: true, openedAt: "Mar 31, 2026" },
  { id: "3", pair: "SOL/USDT", direction: "long", entryPrice: 145, currentPrice: 138, liquidationPrice: 110, pnlPercent: -4.8, pnlAbsolute: -696, volume: 14500, duration: "5 days", timeframe: "15m", isOpen: true, openedAt: "Mar 29, 2026" },
]

const mockClosedHistory: TradeData[] = [
  { id: "4", pair: "BTC/USDT", direction: "long", entryPrice: 42000, exitPrice: 68500, pnlPercent: 63, pnlAbsolute: 26500, volume: 42000, duration: "89 days", timeframe: "1D", isOpen: false, openedAt: "Dec 15, 2025", closedAt: "Mar 14, 2026" },
  { id: "5", pair: "ETH/USDT", direction: "long", entryPrice: 2200, exitPrice: 4100, pnlPercent: 86.4, pnlAbsolute: 19000, volume: 22000, duration: "120 days", timeframe: "4H", isOpen: false, openedAt: "Oct 1, 2025", closedAt: "Jan 29, 2026" },
  { id: "6", pair: "DOGE/USDT", direction: "short", entryPrice: 0.38, exitPrice: 0.22, pnlPercent: 42.1, pnlAbsolute: 6316, volume: 15000, duration: "14 days", timeframe: "1H", isOpen: false, openedAt: "Feb 1, 2026", closedAt: "Feb 15, 2026" },
  { id: "7", pair: "AVAX/USDT", direction: "long", entryPrice: 45, exitPrice: 38, pnlPercent: -15.6, pnlAbsolute: -3120, volume: 20000, duration: "7 days", timeframe: "4H", isOpen: false, openedAt: "Jan 10, 2026", closedAt: "Jan 17, 2026" },
  { id: "8", pair: "LINK/USDT", direction: "long", entryPrice: 12, exitPrice: 24, pnlPercent: 100, pnlAbsolute: 12000, volume: 12000, duration: "60 days", timeframe: "1D", isOpen: false, openedAt: "Nov 1, 2025", closedAt: "Dec 31, 2025" },
  { id: "9", pair: "ARB/USDT", direction: "short", entryPrice: 1.8, exitPrice: 1.2, pnlPercent: 33.3, pnlAbsolute: 5994, volume: 18000, duration: "21 days", timeframe: "4H", isOpen: false, openedAt: "Feb 20, 2026", closedAt: "Mar 13, 2026" },
]

const steps = [
  { id: 1, title: "Wallet", description: "Connect your wallet" },
  { id: 2, title: "Exchange", description: "Link Binance API" },
  { id: 3, title: "Select Trade", description: "Choose & customize" },
  { id: 4, title: "Generate", description: "Create proof" },
]

type AppState = "landing" | "wizard" | "calculating" | "celebration"
type Tab = "create" | "verify"
type TradeTab = "open" | "closed"

// ─── Memoized sub-component ───────────────────────────────────────────────────
// Extracted to flatten the nesting depth around the ChevronRight icon and
// prevent the entire list from re-rendering on every parent state change.

const TradeListItem = memo(function TradeListItem({
  trade,
  onSelect,
}: {
  trade: TradeData
  onSelect: (t: TradeData) => void
}) {
  const isPositive = trade.pnlPercent >= 0
  const pnlColor = isPositive ? "text-emerald-400" : "text-red-400"
  const iconBg = isPositive ? "bg-emerald-500/20" : "bg-red-500/20"
  const directionStyle = trade.direction === "long"
    ? "bg-emerald-500/20 text-emerald-400"
    : "bg-red-500/20 text-red-400"
  const dateLabel = trade.isOpen
    ? `Opened ${trade.openedAt}`
    : `Closed ${trade.closedAt ?? ""}`

  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      whileHover={{ scale: 1.01 }}
      onClick={() => onSelect(trade)}
      className="p-4 rounded-xl bg-secondary/50 border border-border hover:border-primary/50 cursor-pointer transition-all"
    >
      <div className="flex items-center justify-between">
        {/* Left: icon + pair info */}
        <div className="flex items-center gap-3">
          <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${iconBg}`}>
            {isPositive
              ? <TrendingUp className="w-5 h-5 text-emerald-400" />
              : <TrendingDown className="w-5 h-5 text-red-400" />}
          </div>
          <div>
            <div className="flex items-center gap-2">
              <span className="font-semibold text-foreground">{trade.pair}</span>
              <span className={`text-xs px-2 py-0.5 rounded ${directionStyle}`}>
                {trade.direction.toUpperCase()}
              </span>
            </div>
            <p className="text-xs text-muted-foreground">{dateLabel}</p>
          </div>
        </div>

        {/* Right: PNL + chevron */}
        <div className="flex items-center gap-3">
          <div className="text-right">
            <p className={`font-bold ${pnlColor}`}>
              {isPositive ? "+" : ""}{trade.pnlPercent}%
            </p>
            <p className="text-xs text-muted-foreground">
              {trade.pnlAbsolute >= 0 ? "+" : ""}${trade.pnlAbsolute.toLocaleString()}
            </p>
          </div>
          <ChevronRight className="w-4 h-4 text-muted-foreground" />
        </div>
      </div>
    </motion.div>
  )
})

export default function FlexProver() {
  const [appState, setAppState] = useState<AppState>("landing")
  const [activeTab, setActiveTab] = useState<Tab>("create")
  const [currentStep, setCurrentStep] = useState(1)
  const [mousePos, setMousePos] = useState({ x: 0, y: 0 })
  const [logoRevealed, setLogoRevealed] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)
  const [docsOpen, setDocsOpen] = useState(false)
  const [copiedLink, setCopiedLink] = useState(false)

  // Wallet state — driven by iron-session (populated after SIWX verification)
  const [walletConnected, setWalletConnected] = useState(false)
  const [walletAddress, setWalletAddress] = useState<string | null>(null)

  // Poll the server session so the UI reflects iron-session state
  useEffect(() => {
    let cancelled = false
    const check = async () => {
      try {
        const res = await fetch('/api/auth/siwx?action=session')
        if (!res.ok || cancelled) return
        const data = (await res.json()) as { address?: string | null }
        if (!cancelled) {
          setWalletAddress(data.address ?? null)
          setWalletConnected(!!data.address)
        }
      } catch {
        // network error — leave state unchanged
      }
    }
    void check()
    // Re-check every 2 s so the UI picks up the session after AppKit SIWX completes
    const interval = setInterval(() => void check(), 2000)
    return () => {
      cancelled = true
      clearInterval(interval)
    }
  }, [])

  // API Key state
  const [apiVerified, setApiVerified] = useState(false)
  const apiKeyRef = useRef<ApiKeyManagerHandle>(null)

  // Trade selection state
  const [tradeTab, setTradeTab] = useState<TradeTab>("closed")
  const [selectedTrade, setSelectedTrade] = useState<TradeData | null>(null)
  const [tradeListCollapsed, setTradeListCollapsed] = useState(false)

  // Flex card customization
  const [flexOptions, setFlexOptions] = useState({
    showPnl: true,
    showEntryExit: true,
    showVolume: false,
    showWalletAge: false,
    showWhaleBadge: true,
    showDuration: true,
    showTimeframe: false,
    showCurrentPrice: false,
    showLiquidationPrice: false,
  })

  // Calculation logs
  const [logs, setLogs] = useState<string[]>([])

  // Generated proof hash
  const [proofHash] = useState("0x8f3a...e2d1")

  // Handle mouse movement for logo reveal
  useEffect(() => {
    const handleMouseMove = (e: MouseEvent) => {
      if (appState !== "landing" || logoRevealed) return

      const rect = containerRef.current?.getBoundingClientRect()
      if (rect) {
        const x = e.clientX - rect.left
        const y = e.clientY - rect.top
        setMousePos({ x, y })

        const centerX = rect.width / 2
        const centerY = rect.height / 2
        const distance = Math.sqrt(Math.pow(x - centerX, 2) + Math.pow(y - centerY, 2))

        if (distance < 150) {
          setLogoRevealed(true)
        }
      }
    }

    window.addEventListener("mousemove", handleMouseMove)
    return () => window.removeEventListener("mousemove", handleMouseMove)
  }, [appState, logoRevealed])

  // Handle calculation simulation
  const runCalculation = async () => {
    setAppState("calculating")
    setLogs([])

    const logMessages = [
      "Initializing secure connection...",
      "Connecting to Flare TEE enclave...",
      "TEE handshake complete",
      "Fetching Binance UID...",
      `Retrieving trade data for ${selectedTrade?.pair}...`,
      "Calculating PNL metrics...",

      "Generating cryptographic proof...",
      "Proof signature verified",
      "Binding identity to proof...",
      "Minting Flex Card...",
      "Complete! Proof generated successfully",
    ]

    for (let i = 0; i < logMessages.length; i++) {
      await new Promise((resolve) => setTimeout(resolve, 400 + Math.random() * 400))
      setLogs((prev) => [...prev, logMessages[i]])
    }

    await new Promise((resolve) => setTimeout(resolve, 800))
    setAppState("celebration")
  }

  const selectTrade = (trade: TradeData) => {
    setSelectedTrade(trade)
    setTradeListCollapsed(true)
    // Auto-enable liquidation price for open positions
    if (trade.isOpen) {
      setFlexOptions(prev => ({ ...prev, showLiquidationPrice: true, showCurrentPrice: true }))
    } else {
      setFlexOptions(prev => ({ ...prev, showLiquidationPrice: false, showCurrentPrice: false }))
    }
  }

  const deselectTrade = () => {
    setSelectedTrade(null)
    setTradeListCollapsed(false)
  }

  const nextStep = async () => {
    if (currentStep === 2) {
      await apiKeyRef.current?.save()
    }
    if (currentStep < 4) {
      setCurrentStep(currentStep + 1)
    } else {
      runCalculation()
    }
  }

  const prevStep = () => {
    if (currentStep > 1) {
      setCurrentStep(currentStep - 1)
    }
  }

  const handleDownloadImage = () => {
    alert("Downloading Flex Card image...")
  }

  const handleShareToX = () => {
    const pnlText = selectedTrade ? `${selectedTrade.pnlPercent >= 0 ? "+" : ""}${selectedTrade.pnlPercent}%` : "+420%"
    const text = `Just verified my ${selectedTrade?.pair || "BTC/USDT"} trade with @FlexProver! ${pnlText} PNL - Proof of Flex certified. #Web3 #DeFi`
    window.open(`https://twitter.com/intent/tweet?text=${encodeURIComponent(text)}`, "_blank")
  }

  const handleCopyProofLink = () => {
    navigator.clipboard.writeText(`https://flexprover.io/verify/${proofHash}`)
    setCopiedLink(true)
    setTimeout(() => setCopiedLink(false), 2000)
  }

  const resetApp = () => {
    setAppState("landing")
    setCurrentStep(1)
    // Wallet state is managed by AppKit + iron-session; just reset UI
    setApiVerified(false)
    setLogoRevealed(false)
    setLogs([])
    setSelectedTrade(null)
    setTradeListCollapsed(false)
  }

  const currentTrades = tradeTab === "open" ? mockOpenPositions : mockClosedHistory

  // Header component
  const Header = ({ showTabs = false }: { showTabs?: boolean }) => (
    <header className="flex items-center justify-between p-4 sm:p-6">
      <div className="flex items-center gap-4">
        <button 
          onClick={resetApp}
          className="flex items-center gap-2 hover:opacity-80 transition-opacity"
        >
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
          <Button
            variant="ghost"
            onClick={resetApp}
            className="text-muted-foreground hover:text-foreground"
          >
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
        {/* Landing State */}
        {appState === "landing" && (
          <motion.div
            ref={containerRef}
            key="landing"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="min-h-screen flex flex-col"
          >
            <Header />
            <div className="flex-1 flex flex-col items-center justify-center p-6 relative overflow-hidden">
              {/* Subtle grid background */}
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

              {/* Mouse follow glow */}
              {!logoRevealed && (
                <motion.div
                  className="absolute w-96 h-96 rounded-full pointer-events-none"
                  style={{
                    background: "radial-gradient(circle, rgba(168,85,247,0.15) 0%, transparent 70%)",
                    left: mousePos.x - 192,
                    top: mousePos.y - 192,
                  }}
                  animate={{
                    left: mousePos.x - 192,
                    top: mousePos.y - 192,
                  }}
                  transition={{ type: "spring", damping: 30, stiffness: 200 }}
                />
              )}

              {/* Logo */}
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

              {/* Glassmorphism Card */}
              <AnimatePresence>
                {logoRevealed && (
                  <motion.div
                    initial={{ opacity: 0, y: 50, scale: 0.95 }}
                    animate={{ opacity: 1, y: 0, scale: 1 }}
                    transition={{ duration: 0.5, delay: 0.2 }}
                    className="relative z-10 max-w-lg mx-auto"
                  >
                    <div className="rounded-2xl border border-border/50 bg-card/30 backdrop-blur-xl p-8 shadow-2xl">
                      <div className="absolute inset-0 rounded-2xl bg-gradient-to-br from-primary/5 via-transparent to-accent/5" />
                      <div className="relative space-y-6">
                        <div className="flex items-center gap-2 text-primary text-sm font-medium">
                          <Shield className="w-4 h-4" />
                          <span>Proof of Flex</span>
                        </div>
                        <p className="text-lg text-foreground/80 leading-relaxed">
                          Verified performance meets Web3 identity. Bind your Binance trades to your ENS
                          name using secure Flare TEE enclaves.
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

              {/* Hint text */}
              {!logoRevealed && (
                <motion.p
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 0.5 }}
                  transition={{ delay: 2 }}
                  className="absolute bottom-8 text-sm text-muted-foreground"
                >
                  Move your cursor to the logo
                </motion.p>
              )}
            </div>
          </motion.div>
        )}

        {/* Wizard State */}
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
                    {/* Step 1: Wallet Connection */}
                    {currentStep === 1 && (
                      <motion.div
                        key="step1"
                        initial={{ opacity: 0, x: 20 }}
                        animate={{ opacity: 1, x: 0 }}
                        exit={{ opacity: 0, x: -20 }}
                        className="max-w-xl mx-auto space-y-6"
                      >
                        <div>
                          <h2 className="text-2xl font-bold text-foreground mb-2">
                            Connect Your Wallet
                          </h2>
                          <p className="text-muted-foreground">
                            Connect your wallet to get started.
                          </p>
                        </div>

                        {/* AppKit handles wallet connection + SIWX signing in one flow */}
                        <div className="flex justify-center">
                          <appkit-button balance="hide" />
                        </div>

                        {walletConnected && walletAddress && (
                          <div className="space-y-4">
                            <div className="p-4 rounded-xl bg-secondary/50 border border-border">
                              <div className="flex items-center justify-between">
                                <div className="flex items-center gap-3">
                                  <div className="w-10 h-10 rounded-full bg-primary/20 flex items-center justify-center">
                                    <Wallet className="w-5 h-5 text-primary" />
                                  </div>
                                  <div>
                                    <p className="font-medium text-foreground font-mono text-sm">
                                      {walletAddress.slice(0, 6)}…{walletAddress.slice(-4)}
                                    </p>
                                    <p className="text-sm text-muted-foreground">Verified via SIWX</p>
                                  </div>
                                </div>
                                <CheckCircle2 className="w-5 h-5 text-primary" />
                              </div>
                            </div>


                          </div>
                        )}
                      </motion.div>
                    )}

                    {/* Step 2: API Key Manager */}
                    {currentStep === 2 && (
                      <motion.div
                        key="step2"
                        initial={{ opacity: 0, x: 20 }}
                        animate={{ opacity: 1, x: 0 }}
                        exit={{ opacity: 0, x: -20 }}
                        className="max-w-xl mx-auto space-y-6"
                      >
                        <div>
                          <h2 className="text-2xl font-bold text-foreground mb-2">
                            Link Your Exchange Account
                          </h2>
                          <p className="text-muted-foreground">
                            Connect a read-only API key to verify your trading history.
                          </p>
                        </div>

                        <ApiKeyManager ref={apiKeyRef} onValidChange={setApiVerified} />

                        {/* TEE Security Badge */}
                        <div className="p-4 rounded-xl bg-accent/10 border border-accent/20">
                          <div className="flex items-start gap-3">
                            <Shield className="w-5 h-5 text-accent mt-0.5" />
                            <div>
                              <p className="text-sm font-medium text-foreground">
                                Shielded by Flare TEE
                              </p>
                              <p className="text-xs text-muted-foreground mt-1">
                                <span className="text-accent font-medium">Secure Processing:</span> Your keys are parsed inside a Trusted Execution Environment on Flare Network; they are never stored or seen by us.
                              </p>
                            </div>
                          </div>
                        </div>
                      </motion.div>
                    )}

                    {/* Step 3: Trade Selector */}
                    {currentStep === 3 && (
                      <motion.div
                        key="step3"
                        initial={{ opacity: 0, x: 20 }}
                        animate={{ opacity: 1, x: 0 }}
                        exit={{ opacity: 0, x: -20 }}
                        className="space-y-6"
                      >
                        <div className="text-center mb-6">
                          <h2 className="text-2xl font-bold text-foreground mb-2">
                            Select & Design Your Flex
                          </h2>
                          <p className="text-muted-foreground">
                            Choose a trade and customize your Flex Card.
                          </p>
                        </div>

                        <div className="grid lg:grid-cols-2 gap-8">
                          {/* Left: Trade List or Customization */}
                          <div className="space-y-4">
                            {/* Segmented Control */}
                            <div className="flex items-center gap-1 p-1 rounded-lg bg-secondary/50 border border-border">
                              <button
                                onClick={() => {
                                  setTradeTab("open")
                                  if (selectedTrade && !selectedTrade.isOpen) {
                                    deselectTrade()
                                  }
                                }}
                                className={`flex-1 px-4 py-2 rounded-md text-sm font-medium transition-all ${
                                  tradeTab === "open"
                                    ? "bg-primary text-primary-foreground"
                                    : "text-muted-foreground hover:text-foreground"
                                }`}
                              >
                                Open Positions
                              </button>
                              <button
                                onClick={() => {
                                  setTradeTab("closed")
                                  if (selectedTrade && selectedTrade.isOpen) {
                                    deselectTrade()
                                  }
                                }}
                                className={`flex-1 px-4 py-2 rounded-md text-sm font-medium transition-all ${
                                  tradeTab === "closed"
                                    ? "bg-primary text-primary-foreground"
                                    : "text-muted-foreground hover:text-foreground"
                                }`}
                              >
                                Closed History
                              </button>
                            </div>

                            {/* Trade List */}
                            <AnimatePresence mode="wait">
                              {!tradeListCollapsed ? (
                                <motion.div
                                  key="trade-list"
                                  initial={{ opacity: 0, height: 0 }}
                                  animate={{ opacity: 1, height: "auto" }}
                                  exit={{ opacity: 0, height: 0 }}
                                  className="space-y-2 max-h-[400px] overflow-y-auto pr-2"
                                >
                                  {currentTrades.map((trade) => (
                                    <TradeListItem
                                      key={trade.id}
                                      trade={trade}
                                      onSelect={selectTrade}
                                    />
                                  ))}
                                </motion.div>
                              ) : (
                                <motion.div
                                  key="customization"
                                  initial={{ opacity: 0 }}
                                  animate={{ opacity: 1 }}
                                  exit={{ opacity: 0 }}
                                  className="space-y-4"
                                >
                                  {/* Selected Trade Summary */}
                                  {selectedTrade && (
                                    <div className="p-4 rounded-xl bg-primary/10 border border-primary/30">
                                      <div className="flex items-center justify-between mb-2">
                                        <div className="flex items-center gap-2">
                                          <span className="font-semibold text-foreground">{selectedTrade.pair}</span>
                                          <span className={`text-xs px-2 py-0.5 rounded ${selectedTrade.direction === "long" ? "bg-emerald-500/20 text-emerald-400" : "bg-red-500/20 text-red-400"}`}>
                                            {selectedTrade.direction.toUpperCase()}
                                          </span>
                                          {selectedTrade.isOpen && (
                                            <span className="text-xs px-2 py-0.5 rounded bg-purple-500/20 text-purple-400">OPEN</span>
                                          )}
                                        </div>
                                        <button
                                          onClick={deselectTrade}
                                          className="text-xs text-muted-foreground hover:text-foreground transition-colors"
                                        >
                                          Change
                                        </button>
                                      </div>
                                      <p className={`text-2xl font-bold ${selectedTrade.pnlPercent >= 0 ? "text-emerald-400" : "text-red-400"}`}>
                                        {selectedTrade.pnlPercent >= 0 ? "+" : ""}{selectedTrade.pnlPercent}%
                                      </p>
                                    </div>
                                  )}

                                  {/* Customization Toggles */}
                                  <div className="space-y-2">
                                    <h3 className="text-sm font-semibold text-foreground flex items-center gap-2">
                                      <Key className="w-4 h-4 text-primary" />
                                      Core Proof Fields
                                    </h3>
                                    <div className="p-4 rounded-xl bg-secondary/50 border border-border space-y-4">
                                      {[
                                        { key: "showPnl", label: "Relative Gain/Loss (%)" },
                                        { key: "showEntryExit", label: "Entry & Exit Price" },
                                        { key: "showDuration", label: "Coverage (Duration)" },
                                        { key: "showTimeframe", label: "Timeframe" },
                                        { key: "showVolume", label: "Absolute Gain/Loss ($)" },
                                        ...(selectedTrade?.isOpen ? [
                                          { key: "showCurrentPrice", label: "Current Asset Price" },
                                          { key: "showLiquidationPrice", label: "Liquidation Price" },
                                        ] : []),
                                        { key: "showWalletAge", label: "Wallet Age" },
                                        { key: "showWhaleBadge", label: "Whale Status Badge" },
                                      ].map((option) => (
                                        <div
                                          key={option.key}
                                          className="flex items-center justify-between"
                                        >
                                          <Label
                                            htmlFor={option.key}
                                            className="text-sm text-foreground cursor-pointer"
                                          >
                                            {option.label}
                                          </Label>
                                          <Switch
                                            id={option.key}
                                            checked={flexOptions[option.key as keyof typeof flexOptions]}
                                            onCheckedChange={(checked) =>
                                              setFlexOptions((prev) => ({
                                                ...prev,
                                                [option.key]: checked,
                                              }))
                                            }
                                          />
                                        </div>
                                      ))}
                                    </div>
                                  </div>
                                </motion.div>
                              )}
                            </AnimatePresence>
                          </div>

                          {/* Right: Live Preview */}
                          <div>
                            <h3 className="text-sm font-semibold text-foreground mb-4 flex items-center gap-2">
                              <Eye className="w-4 h-4 text-accent" />
                              Live Preview
                            </h3>
                            <FlexCard
                              ensName={walletAddress?.slice(0, 6) ?? 'unknown'}
                              trade={selectedTrade || undefined}
                              showPnl={flexOptions.showPnl}
                              showEntryExit={flexOptions.showEntryExit}
                              showVolume={flexOptions.showVolume}
                              showWalletAge={flexOptions.showWalletAge}
                              showWhaleBadge={flexOptions.showWhaleBadge}
                              showDuration={flexOptions.showDuration}
                              showTimeframe={flexOptions.showTimeframe}
                              showCurrentPrice={flexOptions.showCurrentPrice}
                              showLiquidationPrice={flexOptions.showLiquidationPrice}
                              showQrCode={false}
                              isPreview={true}
                            />
                          </div>
                        </div>
                      </motion.div>
                    )}

                    {/* Step 4: Confirm */}
                    {currentStep === 4 && (
                      <motion.div
                        key="step4"
                        initial={{ opacity: 0, x: 20 }}
                        animate={{ opacity: 1, x: 0 }}
                        exit={{ opacity: 0, x: -20 }}
                        className="max-w-xl mx-auto space-y-6"
                      >
                        <div>
                          <h2 className="text-2xl font-bold text-foreground mb-2">
                            Review & Finalize
                          </h2>
                          <p className="text-muted-foreground">
                            Confirm your details and sign to generate your proof.
                          </p>
                        </div>

                        <div className="space-y-4">
                          <div className="p-4 rounded-xl bg-secondary/50 border border-border">
                            <p className="text-xs text-muted-foreground mb-1">Wallet Address</p>
                            <p className="font-semibold text-foreground">{walletAddress}</p>
                          </div>

                          {selectedTrade && (
                            <div className="p-4 rounded-xl bg-secondary/50 border border-border">
                              <p className="text-xs text-muted-foreground mb-2">Selected Trade</p>
                              <div className="flex items-center justify-between">
                                <div className="flex items-center gap-2">
                                  <span className="font-semibold text-foreground">{selectedTrade.pair}</span>
                                  <span className={`text-xs px-2 py-0.5 rounded ${selectedTrade.direction === "long" ? "bg-emerald-500/20 text-emerald-400" : "bg-red-500/20 text-red-400"}`}>
                                    {selectedTrade.direction.toUpperCase()}
                                  </span>
                                </div>
                                <span className={`font-bold ${selectedTrade.pnlPercent >= 0 ? "text-emerald-400" : "text-red-400"}`}>
                                  {selectedTrade.pnlPercent >= 0 ? "+" : ""}{selectedTrade.pnlPercent}%
                                </span>
                              </div>
                            </div>
                          )}

                          <div className="p-4 rounded-xl bg-secondary/50 border border-border">
                            <p className="text-xs text-muted-foreground mb-2">Card Options</p>
                            <div className="flex flex-wrap gap-2">
                              {Object.entries(flexOptions)
                                .filter(([, value]) => value)
                                .map(([key]) => (
                                  <span
                                    key={key}
                                    className="px-3 py-1 rounded-full bg-primary/10 text-primary text-sm font-medium"
                                  >
                                    {key
                                      .replace("show", "")
                                      .replace(/([A-Z])/g, " $1")
                                      .trim()}
                                  </span>
                                ))}
                            </div>
                          </div>

                          <div className="p-4 rounded-xl bg-accent/10 border border-accent/20">
                            <div className="flex items-start gap-3">
                              <Shield className="w-5 h-5 text-accent mt-0.5" />
                              <div>
                                <p className="text-sm font-medium text-foreground">
                                  Ready to Sign
                                </p>
                                <p className="text-xs text-muted-foreground mt-1">
                                  Clicking &quot;Finalize Proof&quot; will prompt a wallet signature to
                                  authenticate your proof.
                                </p>
                              </div>
                            </div>
                          </div>
                        </div>
                      </motion.div>
                    )}
                  </AnimatePresence>

                  {/* Navigation */}
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
                      disabled={
                        (currentStep === 1 && !walletConnected) ||
                        (currentStep === 2 && !apiVerified) ||
                        (currentStep === 3 && !selectedTrade)
                      }
                      className="flex-1 bg-primary text-primary-foreground hover:bg-primary/90"
                    >
                      {currentStep === 4 ? "Finalize Proof" : "Continue"}
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

        {/* Calculating State */}
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
                <h2 className="text-xl font-bold text-foreground">Generating Proof</h2>
              </div>

              <div className="rounded-xl bg-card border border-border p-4 font-mono text-sm">
                <div className="space-y-2">
                  {logs.map((log, index) => (
                    <motion.div
                      key={index}
                      initial={{ opacity: 0, x: -10 }}
                      animate={{ opacity: 1, x: 0 }}
                      className="flex items-start gap-2"
                    >
                      <span className="text-primary">{">"}</span>
                      <span
                        className={
                          index === logs.length - 1 ? "text-primary" : "text-muted-foreground"
                        }
                      >
                        {log}
                      </span>
                    </motion.div>
                  ))}
                  {logs.length < 12 && (
                    <motion.span
                      animate={{ opacity: [1, 0] }}
                      transition={{ repeat: Infinity, duration: 0.8 }}
                      className="inline-block w-2 h-4 bg-primary ml-4"
                    />
                  )}
                </div>
              </div>
            </div>
          </motion.div>
        )}

        {/* Celebration State */}
        {appState === "celebration" && (
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
                {/* Celebration Header */}
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
                    <motion.div
                      initial={{ scale: 0 }}
                      animate={{ scale: [0, 1.2, 1] }}
                      transition={{ delay: 0.5, duration: 0.4 }}
                      className="absolute -top-1 -right-1 w-8 h-8 rounded-full bg-accent flex items-center justify-center"
                    >
                      <CheckCircle2 className="w-5 h-5 text-accent-foreground" />
                    </motion.div>
                  </div>
                  <h1 className="text-3xl font-bold text-foreground mb-2">Proof Generated!</h1>
                  <p className="text-muted-foreground">
                    Your Proof of Flex is ready to share with the world.
                  </p>
                </motion.div>

                {/* Final Flex Card with QR Code */}
                <motion.div
                  initial={{ y: 20, opacity: 0 }}
                  animate={{ y: 0, opacity: 1 }}
                  transition={{ delay: 0.4 }}
                >
                  <FlexCard
                    ensName={walletAddress?.slice(0, 6) ?? 'unknown'}
                    trade={selectedTrade || undefined}
                    showPnl={flexOptions.showPnl}
                    showEntryExit={flexOptions.showEntryExit}
                    showVolume={flexOptions.showVolume}
                    showWalletAge={flexOptions.showWalletAge}
                    showWhaleBadge={flexOptions.showWhaleBadge}
                    showDuration={flexOptions.showDuration}
                    showTimeframe={flexOptions.showTimeframe}
                    showCurrentPrice={flexOptions.showCurrentPrice}
                    showLiquidationPrice={flexOptions.showLiquidationPrice}
                    showQrCode={true}
                    isPreview={false}
                  />
                </motion.div>

                {/* Action Buttons */}
                <motion.div
                  initial={{ y: 20, opacity: 0 }}
                  animate={{ y: 0, opacity: 1 }}
                  transition={{ delay: 0.6 }}
                  className="flex flex-col sm:flex-row gap-3 mt-6"
                >
                  <Button
                    onClick={handleShareToX}
                    className="flex-1 bg-foreground text-background hover:bg-foreground/90 h-12"
                  >
                    <Twitter className="w-4 h-4 mr-2" />
                    Share to X
                  </Button>
                  <Button
                    onClick={handleCopyProofLink}
                    variant="outline"
                    className="flex-1 border-border text-foreground hover:bg-secondary h-12"
                  >
                    <Link2 className="w-4 h-4 mr-2" />
                    {copiedLink ? "Copied!" : "Copy Proof Link"}
                  </Button>
                  <Button
                    onClick={handleDownloadImage}
                    variant="outline"
                    className="flex-1 border-border text-foreground hover:bg-secondary h-12"
                  >
                    <Download className="w-4 h-4 mr-2" />
                    Download
                  </Button>
                </motion.div>

                {/* Start Over */}
                <motion.div
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  transition={{ delay: 0.8 }}
                  className="mt-6 text-center"
                >
                  <Button
                    variant="ghost"
                    onClick={resetApp}
                    className="text-muted-foreground hover:text-foreground"
                  >
                    Create Another Proof
                  </Button>
                </motion.div>
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Documentation Sheet */}
      <AnimatePresence>
        {docsOpen && <DocsSheet open={docsOpen} onClose={() => setDocsOpen(false)} />}
      </AnimatePresence>
    </main>
  )
}

"use client"

import { motion } from "framer-motion"
import { Shield, TrendingUp, TrendingDown, DollarSign, Clock, Award, Activity, QrCode, Target, Timer, Zap } from "lucide-react"

interface TradeData {
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
}

interface FlexCardProps {
  ensName: string
  trade?: TradeData
  showPnl: boolean
  showEntryExit: boolean
  showVolume: boolean
  showWalletAge: boolean
  showWhaleBadge: boolean
  showQrCode: boolean
  showDuration?: boolean
  showTimeframe?: boolean
  showCurrentPrice?: boolean
  showLiquidationPrice?: boolean
  isPreview?: boolean
}

export function FlexCard({
  ensName,
  trade,
  showPnl,
  showEntryExit,
  showVolume,
  showWalletAge,
  showWhaleBadge,
  showQrCode,
  showDuration = false,
  showTimeframe = false,
  showCurrentPrice = false,
  showLiquidationPrice = false,
  isPreview = true,
}: FlexCardProps) {
  const defaultTrade: TradeData = {
    pair: "BTC/USDT",
    direction: "long",
    entryPrice: 23456,
    exitPrice: 98765,
    currentPrice: 95000,
    liquidationPrice: 18500,
    pnlPercent: 420,
    pnlAbsolute: 42069,
    volume: 1200000,
    duration: "45 days",
    timeframe: "4H",
    isOpen: false,
  }

  const tradeData = trade || defaultTrade
  const isPositive = tradeData.pnlPercent >= 0

  const pnlTextClass = isPositive ? "text-emerald-400" : "text-red-400"
  const pnlBorderClass = isPositive ? "border-emerald-500/30" : "border-red-500/30"

  const formatCurrency = (val: number) => {
    if (val >= 1000000) return `$${(val / 1000000).toFixed(1)}M`
    if (val >= 1000) return `$${(val / 1000).toFixed(1)}K`
    return `$${val.toLocaleString()}`
  }

  const formatPrice = (val: number) => `$${val.toLocaleString()}`

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.98 }}
      animate={{ opacity: 1, scale: 1 }}
      transition={{ duration: 0.3 }}
      className="relative"
    >
      {/* Card container with subtle gradient border */}
      <div className="rounded-xl p-[1px] bg-gradient-to-br from-primary/40 via-border to-accent/30">
        <div className="rounded-xl bg-card p-4 backdrop-blur-sm">
          {/* Header row */}
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-1.5">
              <div className="w-5 h-5 rounded bg-primary/20 flex items-center justify-center">
                <span className="text-primary font-bold text-[10px]">FP</span>
              </div>
              <span className="font-medium text-[11px] text-muted-foreground">FlexProver</span>
            </div>
            <div className="flex items-center gap-1 px-2 py-0.5 rounded-full bg-accent/10 border border-accent/20">
              <Shield className="w-3 h-3 text-accent" />
              <span className="text-[10px] font-medium text-accent">Verified</span>
            </div>
          </div>

          {/* ENS + Pair inline */}
          <div className="flex items-end justify-between mb-3">
            <div>
              <p className="text-[10px] text-muted-foreground mb-0.5">Identity</p>
              <h2 className="text-lg font-bold text-foreground leading-none">{ensName}</h2>
            </div>
            <div className="flex items-center gap-1.5">
              <span className="text-xs font-semibold text-foreground bg-secondary px-2 py-0.5 rounded">{tradeData.pair}</span>
              <span className={`text-[10px] font-semibold px-1.5 py-0.5 rounded ${tradeData.direction === "long" ? "bg-emerald-500/15 text-emerald-400" : "bg-red-500/15 text-red-400"}`}>
                {tradeData.direction.toUpperCase()}
              </span>
              {tradeData.isOpen && (
                <span className="text-[10px] font-semibold px-1.5 py-0.5 rounded bg-purple-500/15 text-purple-400">LIVE</span>
              )}
            </div>
          </div>

          {/* Whale Badge - inline */}
          {showWhaleBadge && (
            <div className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-primary/10 border border-primary/20 mb-3">
              <Award className="w-3 h-3 text-primary" />
              <span className="text-[10px] font-semibold text-primary">Whale</span>
            </div>
          )}

          {/* PNL - compact hero section */}
          {showPnl && (
            <div className={`mb-3 p-2.5 rounded-lg border ${pnlBorderClass} bg-gradient-to-r ${isPositive ? "from-emerald-500/5 to-transparent" : "from-red-500/5 to-transparent"}`}>
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-1.5">
                  {isPositive ? <TrendingUp className={`w-4 h-4 ${pnlTextClass}`} /> : <TrendingDown className={`w-4 h-4 ${pnlTextClass}`} />}
                  <span className="text-[10px] text-muted-foreground">{tradeData.isOpen ? "Unrealized" : "Realized"}</span>
                </div>
                <div className="flex items-baseline gap-1.5">
                  <span className={`text-xl font-bold ${pnlTextClass}`}>{isPositive ? "+" : ""}{tradeData.pnlPercent}%</span>
                  <span className={`text-xs ${pnlTextClass}`}>{isPositive ? "+" : ""}{formatCurrency(tradeData.pnlAbsolute)}</span>
                </div>
              </div>
            </div>
          )}

          {/* Metrics grid - 2 columns, compact */}
          <div className="grid grid-cols-2 gap-2 mb-3">
            {showEntryExit && (
              <>
                <div className="p-2 rounded-lg bg-secondary/40 border border-border/50">
                  <div className="flex items-center gap-1 mb-0.5">
                    <DollarSign className="w-3 h-3 text-muted-foreground" />
                    <span className="text-[9px] text-muted-foreground uppercase tracking-wide">Entry</span>
                  </div>
                  <span className="text-sm font-semibold text-foreground">{formatPrice(tradeData.entryPrice)}</span>
                </div>
                <div className="p-2 rounded-lg bg-secondary/40 border border-border/50">
                  <div className="flex items-center gap-1 mb-0.5">
                    <DollarSign className="w-3 h-3 text-muted-foreground" />
                    <span className="text-[9px] text-muted-foreground uppercase tracking-wide">{tradeData.isOpen ? "Current" : "Exit"}</span>
                  </div>
                  <span className="text-sm font-semibold text-foreground">
                    {formatPrice(tradeData.isOpen ? (tradeData.currentPrice || 0) : (tradeData.exitPrice || 0))}
                  </span>
                </div>
              </>
            )}
            
            {showCurrentPrice && tradeData.isOpen && tradeData.currentPrice && !showEntryExit && (
              <div className="p-2 rounded-lg bg-secondary/40 border border-border/50">
                <div className="flex items-center gap-1 mb-0.5">
                  <Target className="w-3 h-3 text-muted-foreground" />
                  <span className="text-[9px] text-muted-foreground uppercase tracking-wide">Current</span>
                </div>
                <span className="text-sm font-semibold text-foreground">{formatPrice(tradeData.currentPrice)}</span>
              </div>
            )}

            {showLiquidationPrice && tradeData.isOpen && tradeData.liquidationPrice && (
              <div className="p-2 rounded-lg bg-red-500/5 border border-red-500/20">
                <div className="flex items-center gap-1 mb-0.5">
                  <Zap className="w-3 h-3 text-red-400" />
                  <span className="text-[9px] text-red-400 uppercase tracking-wide">Liq Price</span>
                </div>
                <span className="text-sm font-semibold text-red-400">{formatPrice(tradeData.liquidationPrice)}</span>
              </div>
            )}

            {showDuration && (
              <div className="p-2 rounded-lg bg-secondary/40 border border-border/50">
                <div className="flex items-center gap-1 mb-0.5">
                  <Timer className="w-3 h-3 text-muted-foreground" />
                  <span className="text-[9px] text-muted-foreground uppercase tracking-wide">Duration</span>
                </div>
                <span className="text-sm font-semibold text-foreground">{tradeData.duration}</span>
              </div>
            )}

            {showTimeframe && (
              <div className="p-2 rounded-lg bg-secondary/40 border border-border/50">
                <div className="flex items-center gap-1 mb-0.5">
                  <Clock className="w-3 h-3 text-muted-foreground" />
                  <span className="text-[9px] text-muted-foreground uppercase tracking-wide">Timeframe</span>
                </div>
                <span className="text-sm font-semibold text-foreground">{tradeData.timeframe}</span>
              </div>
            )}

            {showVolume && (
              <div className="p-2 rounded-lg bg-secondary/40 border border-border/50">
                <div className="flex items-center gap-1 mb-0.5">
                  <Activity className="w-3 h-3 text-muted-foreground" />
                  <span className="text-[9px] text-muted-foreground uppercase tracking-wide">Volume</span>
                </div>
                <span className="text-sm font-semibold text-foreground">{formatCurrency(tradeData.volume)}</span>
              </div>
            )}

            {showWalletAge && (
              <div className="p-2 rounded-lg bg-secondary/40 border border-border/50">
                <div className="flex items-center gap-1 mb-0.5">
                  <Clock className="w-3 h-3 text-muted-foreground" />
                  <span className="text-[9px] text-muted-foreground uppercase tracking-wide">Wallet Age</span>
                </div>
                <span className="text-sm font-semibold text-foreground">2.5 years</span>
              </div>
            )}
          </div>

          {/* Footer row */}
          <div className="flex items-end justify-between pt-2 border-t border-border/30">
            <div>
              <p className="text-[9px] text-muted-foreground uppercase tracking-wide">Verified</p>
              <p className="text-xs font-medium text-foreground">Apr 3, 2026</p>
            </div>
            {showQrCode && (
              <div className="w-10 h-10 rounded bg-foreground flex items-center justify-center overflow-hidden">
                {isPreview ? (
                  <QrCode className="w-6 h-6 text-muted-foreground" />
                ) : (
                  <div className="w-full h-full bg-foreground p-1">
                    <div className="w-full h-full grid grid-cols-5 grid-rows-5 gap-px">
                      {Array.from({ length: 25 }).map((_, i) => (
                        <div key={i} className={[0,1,2,4,5,6,10,14,18,19,20,22,23,24].includes(i) ? "bg-background" : "bg-foreground"} />
                      ))}
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Preview badge */}
          {isPreview && (
            <div className="absolute top-2 right-2 px-1.5 py-0.5 rounded text-[9px] font-medium text-muted-foreground bg-muted/40">
              Preview
            </div>
          )}
        </div>
      </div>
    </motion.div>
  )
}

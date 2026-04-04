"use client"

import { useState } from "react"
import { motion } from "framer-motion"
import { Shield, ExternalLink, Twitter, QrCode, Eye, EyeOff, TrendingUp } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Switch } from "@/components/ui/switch"
import { Label } from "@/components/ui/label"

interface ProofCardProps {
  ensName: string
  pnl: string
  pnlPercentage: string
  verifiedAt: string
}

export function ProofCard({ ensName, pnl, pnlPercentage, verifiedAt }: ProofCardProps) {
  const [showPnl, setShowPnl] = useState(true)
  const [showQr, setShowQr] = useState(true)
  const [showTimestamp, setShowTimestamp] = useState(true)

  const handleShareToX = () => {
    const text = `Just verified my trading performance with @FlexProver! ${showPnl ? pnlPercentage + " PNL" : ""} - Proof of Whale certified. #Web3 #DeFi`
    window.open(`https://twitter.com/intent/tweet?text=${encodeURIComponent(text)}`, "_blank")
  }

  return (
    <div className="space-y-6">
      {/* The Proof Card */}
      <motion.div
        initial={{ opacity: 0, scale: 0.9 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ duration: 0.5 }}
        className="relative overflow-hidden rounded-2xl bg-gradient-to-br from-card via-card to-secondary p-1"
      >
        <div className="absolute inset-0 bg-gradient-to-br from-primary/20 via-transparent to-accent/20 opacity-50" />
        <div className="relative rounded-xl bg-card/90 backdrop-blur-xl p-6 sm:p-8">
          {/* Header */}
          <div className="flex items-center justify-between mb-6">
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

          {/* ENS Name */}
          <div className="mb-6">
            <p className="text-xs text-muted-foreground mb-1">Verified Identity</p>
            <h2 className="text-3xl sm:text-4xl font-bold text-foreground tracking-tight">
              {ensName}
            </h2>
          </div>

          {/* PNL Display */}
          {showPnl && (
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              className="mb-6 p-4 rounded-xl bg-primary/10 border border-primary/20"
            >
              <div className="flex items-center gap-2 mb-1">
                <TrendingUp className="w-4 h-4 text-primary" />
                <span className="text-xs text-muted-foreground">Trading Performance</span>
              </div>
              <div className="flex items-baseline gap-3">
                <span className="text-3xl sm:text-4xl font-bold text-primary">
                  {pnlPercentage}
                </span>
                <span className="text-lg text-muted-foreground">{pnl}</span>
              </div>
            </motion.div>
          )}

          {/* Footer */}
          <div className="flex items-end justify-between">
            <div>
              {showTimestamp && (
                <div>
                  <p className="text-xs text-muted-foreground">Verified on</p>
                  <p className="text-sm font-medium text-foreground">{verifiedAt}</p>
                </div>
              )}
            </div>
            {showQr && (
              <div className="w-16 h-16 rounded-lg bg-foreground/10 border border-border flex items-center justify-center">
                <QrCode className="w-10 h-10 text-muted-foreground" />
              </div>
            )}
          </div>
        </div>
      </motion.div>

      {/* Card Controls */}
      <div className="p-4 rounded-xl bg-secondary/50 border border-border space-y-4">
        <h3 className="text-sm font-semibold text-foreground">Customize Card</h3>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <div className="flex items-center justify-between">
            <Label htmlFor="show-pnl" className="text-sm text-muted-foreground flex items-center gap-2">
              {showPnl ? <Eye className="w-4 h-4" /> : <EyeOff className="w-4 h-4" />}
              Show PNL
            </Label>
            <Switch id="show-pnl" checked={showPnl} onCheckedChange={setShowPnl} />
          </div>
          <div className="flex items-center justify-between">
            <Label htmlFor="show-qr" className="text-sm text-muted-foreground flex items-center gap-2">
              <QrCode className="w-4 h-4" />
              QR Code
            </Label>
            <Switch id="show-qr" checked={showQr} onCheckedChange={setShowQr} />
          </div>
          <div className="flex items-center justify-between">
            <Label htmlFor="show-timestamp" className="text-sm text-muted-foreground">
              Timestamp
            </Label>
            <Switch id="show-timestamp" checked={showTimestamp} onCheckedChange={setShowTimestamp} />
          </div>
        </div>
      </div>

      {/* Action Buttons */}
      <div className="flex flex-col sm:flex-row gap-3">
        <Button
          onClick={handleShareToX}
          className="flex-1 bg-foreground text-background hover:bg-foreground/90"
        >
          <Twitter className="w-4 h-4 mr-2" />
          Share to X
        </Button>
        <Button
          variant="outline"
          className="flex-1 border-border text-foreground hover:bg-secondary"
        >
          <ExternalLink className="w-4 h-4 mr-2" />
          View On-Chain Proof
        </Button>
      </div>
    </div>
  )
}

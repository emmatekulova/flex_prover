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

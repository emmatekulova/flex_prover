"use client"

import { useRef, useState } from "react"
import { motion } from "framer-motion"
import { Shield, ExternalLink, Twitter, BarChart2, Copy, Download, CheckCircle2 } from "lucide-react"
import { QRCodeCanvas } from "qrcode.react"
import { Button } from "@/components/ui/button"
import type { TradePosition } from "@/lib/attestation"

interface TradesProofCardProps {
  walletAddress: string
  positions: TradePosition[]
  totalUsdt: string
  txHash: string
  fetchedAt: number
  exchange?: string
}

export function TradesProofCard({
  walletAddress,
  positions,
  totalUsdt,
  txHash,
  fetchedAt,
  exchange,
}: TradesProofCardProps) {
  const cardRef = useRef<HTMLDivElement>(null)
  const [copiedLink, setCopiedLink] = useState(false)

  const txUrl = `https://coston2-explorer.flare.network/tx/${txHash}`
  const shortWallet = `${walletAddress.slice(0, 6)}...${walletAddress.slice(-4)}`
  const exchangeLabel =
    exchange === "bitget" ? "Bitget" : exchange === "binance" ? "Binance" : exchange ?? "Unknown"
  const fetchedDate = fetchedAt
    ? new Date(fetchedAt * 1000).toISOString().split("T")[0]
    : ""
  const formattedTotal = Number(totalUsdt).toLocaleString("en-US", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })

  const handleShareToX = () => {
    const text = `Just attested my open positions on-chain! $${formattedTotal} USDT proven with @FlexProver on Flare Network. #FlexProver #Flare #DeFi`
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
    try {
      const { toPng } = await import("html-to-image")

      const sourceNode = cardRef.current
      const exportQr = sourceNode.querySelector("[data-proof-qr]") as HTMLDivElement | null
      const exportTeeLabel = sourceNode.querySelector("[data-proof-tee-label]") as HTMLSpanElement | null
      const previousQrStyles = exportQr ? { width: exportQr.style.width, height: exportQr.style.height, padding: exportQr.style.padding } : null
      const previousTeeStyles = exportTeeLabel ? { fontSize: exportTeeLabel.style.fontSize, whiteSpace: exportTeeLabel.style.whiteSpace } : null

      if (exportQr) {
        exportQr.style.width = "132px"
        exportQr.style.height = "132px"
        exportQr.style.padding = "10px"
      }
      if (exportTeeLabel) {
        exportTeeLabel.style.fontSize = "13px"
        exportTeeLabel.style.whiteSpace = "nowrap"
      }

      await new Promise<void>((resolve) => requestAnimationFrame(() => resolve()))

      const baseOptions = { backgroundColor: "#0a0a0f", pixelRatio: 3, cacheBust: true }
      let dataUrl: string
      try {
        dataUrl = await toPng(sourceNode, baseOptions)
      } catch {
        dataUrl = await toPng(sourceNode, { ...baseOptions, fontEmbedCSS: "" })
      } finally {
        if (exportQr && previousQrStyles) {
          exportQr.style.width = previousQrStyles.width
          exportQr.style.height = previousQrStyles.height
          exportQr.style.padding = previousQrStyles.padding
        }
        if (exportTeeLabel && previousTeeStyles) {
          exportTeeLabel.style.fontSize = previousTeeStyles.fontSize
          exportTeeLabel.style.whiteSpace = previousTeeStyles.whiteSpace
        }
      }

      const link = document.createElement("a")
      link.download = `trades-proof-${shortWallet}.png`
      link.href = dataUrl
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
    } catch (error) {
      console.error("Failed to download PNG", error)
      alert("Could not generate PNG right now. Please try again.")
    }
  }

  return (
    <div className="space-y-4">
      {/* The Proof Card — captured for PNG download */}
      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ duration: 0.4 }}
        className="relative overflow-hidden rounded-2xl p-[1px] bg-gradient-to-br from-primary/40 via-border to-accent/30"
      >
        <div
          ref={cardRef}
          className="rounded-2xl bg-[#0a0a0f] p-6 sm:p-8 md:p-10 flex flex-col gap-8 w-full max-w-5xl mx-auto"
        >
          {/* Header */}
          <div className="flex items-center justify-between gap-4">
            <div className="flex items-center gap-2">
              <div className="w-10 h-10 rounded-xl bg-primary/20 flex items-center justify-center">
                <span className="text-primary font-bold text-sm">FP</span>
              </div>
              <span className="font-semibold text-lg text-muted-foreground">FlexProver</span>
            </div>
            <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-accent/10 border border-accent/30">
              <Shield className="w-4 h-4 text-accent" />
              <span data-proof-tee-label className="text-sm font-medium text-accent">TEE Verified</span>
            </div>
          </div>

          {/* Wallet */}
          <div>
            <p className="text-sm text-muted-foreground mb-2">Wallet Address</p>
            <p className="text-xl sm:text-2xl font-bold text-foreground font-mono tracking-tight break-all leading-tight">
              {walletAddress}
            </p>
          </div>

          {/* Positions table */}
          <div className="p-5 rounded-xl bg-primary/10 border border-primary/25">
            <div className="flex items-center gap-2 mb-4">
              <BarChart2 className="w-5 h-5 text-primary" />
              <span className="text-sm text-muted-foreground">Verified Open Positions</span>
            </div>

            {/* Table header */}
            <div className="grid grid-cols-4 gap-2 mb-2 text-xs text-muted-foreground font-medium uppercase tracking-wide">
              <span>Asset</span>
              <span className="text-right">Quantity</span>
              <span className="text-right">Price (USDT)</span>
              <span className="text-right">Value (USDT)</span>
            </div>

            {/* Table rows */}
            <div className="space-y-1 max-h-48 overflow-y-auto">
              {positions.map((pos) => (
                <div
                  key={pos.asset}
                  className="grid grid-cols-4 gap-2 text-sm py-1 border-t border-border/20"
                >
                  <span className="font-semibold text-foreground">{pos.asset}</span>
                  <span className="text-right font-mono text-muted-foreground">{pos.quantity}</span>
                  <span className="text-right font-mono text-muted-foreground">
                    {Number(pos.priceUsdt).toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                  </span>
                  <span className="text-right font-mono text-foreground">
                    {Number(pos.valueUsdt).toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                  </span>
                </div>
              ))}
            </div>

            {/* Total row */}
            <div className="grid grid-cols-4 gap-2 mt-3 pt-3 border-t border-primary/30">
              <span className="col-span-3 text-sm font-semibold text-foreground">Total</span>
              <span className="text-right font-mono font-bold text-primary text-base">
                ${formattedTotal}
              </span>
            </div>

            <p className="text-sm text-muted-foreground mt-3">
              Attested at: <span className="text-foreground">{fetchedDate}</span>
            </p>
            <p className="text-sm text-muted-foreground mt-1">
              Source CEX: <span className="text-foreground font-semibold">{exchangeLabel}</span>
            </p>
          </div>

          {/* Footer: TX + QR */}
          <div className="flex items-end justify-between pt-4 border-t border-border/30 gap-4">
            <div className="min-w-0 mr-4">
              <p className="text-sm text-muted-foreground mb-2">On-Chain Proof</p>
              <p className="text-sm font-mono text-foreground break-all leading-relaxed">{txHash}</p>
            </div>
            <div
              data-proof-qr
              className="w-28 h-28 shrink-0 rounded-xl bg-white flex items-center justify-center overflow-hidden p-2 shadow-sm"
            >
              <QRCodeCanvas
                value={txUrl}
                size={92}
                bgColor="#ffffff"
                fgColor="#000000"
                level="H"
                includeMargin
              />
            </div>
          </div>
        </div>
      </motion.div>

      {/* Action Buttons */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
        <Button onClick={handleShareToX} className="bg-foreground text-background hover:bg-foreground/90">
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

      {/* View on Explorer */}
      <Button
        variant="ghost"
        className="w-full text-muted-foreground hover:text-foreground"
        onClick={() => window.open(txUrl, "_blank")}
      >
        <ExternalLink className="w-4 h-4 mr-2" />
        View Transaction on Coston2 Explorer
      </Button>
    </div>
  )
}

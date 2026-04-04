"use client"

import { useState, useEffect, memo, useCallback, forwardRef, useImperativeHandle, useRef } from "react"
import { motion, AnimatePresence } from "framer-motion"
import {
  KeyRound,
  Eye,
  EyeOff,
  AlertCircle,
  CheckCircle2,
  BookOpen,
  X,
  Maximize2,
  Minimize2,
} from "lucide-react"
import { cn } from "@/lib/utils"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

// ─── Types ────────────────────────────────────────────────────────────────────

type Exchange = "binance" | "bitget"

interface FieldConfig {
  id: string
  label: string
  placeholder: string
  masked: boolean
  pattern: RegExp
  errorMsg: string
}

// ─── Field definitions ────────────────────────────────────────────────────────

const EXCHANGE_FIELDS: Record<Exchange, FieldConfig[]> = {
  binance: [
    {
      id: "apiKey",
      label: "API Key",
      placeholder: "64-character alphanumeric key",
      masked: false,
      pattern: /^[a-zA-Z0-9]{64}$/,
      errorMsg: "Must be exactly 64 alphanumeric characters.",
    },
    {
      id: "secretKey",
      label: "Secret Key",
      placeholder: "64-character alphanumeric secret",
      masked: true,
      pattern: /^[a-zA-Z0-9]{64}$/,
      errorMsg: "Must be exactly 64 alphanumeric characters.",
    },
  ],
  bitget: [
    {
      id: "apiKey",
      label: "API Key",
      placeholder: "32–64 character key",
      masked: false,
      pattern: /^[a-zA-Z0-9_-]{32,64}$/,
      errorMsg: "Must be 32–64 alphanumeric characters (- and _ allowed).",
    },
    {
      id: "secretKey",
      label: "Secret Key",
      placeholder: "32–64 character secret",
      masked: true,
      pattern: /^[a-zA-Z0-9_-]{32,64}$/,
      errorMsg: "Must be 32–64 alphanumeric characters (- and _ allowed).",
    },
    {
      id: "passphrase",
      label: "Passphrase",
      placeholder: "Min 8 characters",
      masked: true,
      pattern: /^.{8,}$/,
      errorMsg: "Passphrase must be at least 8 characters.",
    },
  ],
}

const EXCHANGE_LABELS: Record<Exchange, string> = {
  binance: "Binance",
  bitget: "Bitget",
}

const TUTORIAL_GIF: Record<Exchange, string> = {
  binance: "/binance.gif",
  bitget: "/bitget.gif",
}

const TUTORIAL_STEPS: Record<Exchange, string[]> = {
  binance: [
    "Log in to Binance and go to Profile → API Management.",
    "Click Create API, choose System Generated, and give it a label.",
    "Under restrictions, enable Read Only — do NOT enable trading or withdrawals.",
    "Complete 2FA verification, then copy your API Key and Secret Key.",
  ],
  bitget: [
    "Log in to Bitget and go to Profile → API Management.",
    "Click Create API, set a label, and choose Read Only permissions.",
    "Set an IP whitelist if prompted, then complete verification.",
    "Copy your API Key, Secret Key, and Passphrase.",
  ],
}

// ─── Tutorial Lightbox ────────────────────────────────────────────────────────

function TutorialLightbox({
  exchange,
  onClose,
}: {
  exchange: Exchange
  onClose: () => void
}) {
  const [fullscreen, setFullscreen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  // Close on Escape
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose()
    }
    window.addEventListener("keydown", handler)
    return () => window.removeEventListener("keydown", handler)
  }, [onClose])

  const toggleFullscreen = useCallback(async () => {
    if (!fullscreen) {
      try {
        await containerRef.current?.requestFullscreen()
        setFullscreen(true)
      } catch {
        // Browser blocked fullscreen — fall back to CSS-only expanded view
        setFullscreen(true)
      }
    } else {
      if (document.fullscreenElement) {
        await document.exitFullscreen()
      }
      setFullscreen(false)
    }
  }, [fullscreen])

  // Sync state when user exits fullscreen with browser controls
  useEffect(() => {
    const handler = () => {
      if (!document.fullscreenElement) setFullscreen(false)
    }
    document.addEventListener("fullscreenchange", handler)
    return () => document.removeEventListener("fullscreenchange", handler)
  }, [])

  return (
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm"
        onClick={(e) => { if (e.target === e.currentTarget) onClose() }}
      >
        <motion.div
          ref={containerRef}
          initial={{ scale: 0.95, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          exit={{ scale: 0.95, opacity: 0 }}
          transition={{ duration: 0.2 }}
          className={cn(
            "relative flex flex-col rounded-2xl border border-border bg-card shadow-2xl overflow-hidden",
            fullscreen
              ? "w-screen h-screen rounded-none"
              : "w-full max-w-2xl max-h-[90vh]",
          )}
        >
          {/* Header */}
          <div className="flex items-center justify-between px-5 py-3 border-b border-border shrink-0">
            <div className="flex items-center gap-2">
              <BookOpen className="size-4 text-primary" />
              <span className="text-sm font-semibold text-foreground">
                How to get your {EXCHANGE_LABELS[exchange]} API keys
              </span>
            </div>
            <div className="flex items-center gap-1">
              <button
                type="button"
                onClick={toggleFullscreen}
                className="p-1.5 rounded-md text-muted-foreground hover:text-foreground hover:bg-secondary transition-colors"
                aria-label={fullscreen ? "Exit fullscreen" : "Fullscreen"}
              >
                {fullscreen ? <Minimize2 className="size-4" /> : <Maximize2 className="size-4" />}
              </button>
              <button
                type="button"
                onClick={onClose}
                className="p-1.5 rounded-md text-muted-foreground hover:text-foreground hover:bg-secondary transition-colors"
                aria-label="Close tutorial"
              >
                <X className="size-4" />
              </button>
            </div>
          </div>

          {/* GIF */}
          <div className="relative flex-1 min-h-0 bg-black flex items-center justify-center overflow-hidden">
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src={TUTORIAL_GIF[exchange]}
              alt={`${EXCHANGE_LABELS[exchange]} API key tutorial`}
              className="w-full h-full object-contain"
            />
          </div>

          {/* Steps */}
          <div className="px-5 py-4 border-t border-border shrink-0 bg-card">
            <ol className="space-y-1.5">
              {TUTORIAL_STEPS[exchange].map((step, i) => (
                <li key={i} className="flex items-start gap-2.5 text-xs text-muted-foreground">
                  <span className="shrink-0 w-4 h-4 rounded-full bg-primary/20 text-primary flex items-center justify-center text-[10px] font-bold mt-0.5">
                    {i + 1}
                  </span>
                  {step}
                </li>
              ))}
            </ol>
          </div>
        </motion.div>
      </motion.div>
    </AnimatePresence>
  )
}

// ─── FieldRow (memoized) ──────────────────────────────────────────────────────

interface FieldRowProps {
  field: FieldConfig
  value: string
  onChange: (val: string) => void
}

const FieldRow = memo(function FieldRow({ field, value, onChange }: FieldRowProps) {
  const [visible, setVisible] = useState(false)
  const touched = value.length > 0
  const valid = field.pattern.test(value)
  const showError = touched && !valid

  const toggleVisible = useCallback(() => setVisible((v) => !v), [])

  return (
    <div className="space-y-1.5">
      <Label
        htmlFor={field.id}
        className="text-xs font-medium text-muted-foreground uppercase tracking-wide"
      >
        {field.label}
      </Label>

      <div className="relative">
        <Input
          id={field.id}
          type={field.masked && !visible ? "password" : "text"}
          placeholder={field.placeholder}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          aria-invalid={showError}
          autoComplete="off"
          className={cn(
            "pr-10 font-mono text-sm",
            touched && valid && "border-emerald-500/50 focus-visible:border-emerald-500",
          )}
        />

        {/* Eye toggle for masked fields */}
        {field.masked && (
          <button
            type="button"
            onClick={toggleVisible}
            className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
            aria-label={visible ? "Hide value" : "Show value"}
          >
            {visible ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
          </button>
        )}

        {/* Inline validity icon for non-masked fields */}
        {!field.masked && touched && (
          <span className="absolute right-2.5 top-1/2 -translate-y-1/2 pointer-events-none">
            {valid
              ? <CheckCircle2 className="size-4 text-emerald-500" />
              : <AlertCircle className="size-4 text-destructive" />}
          </span>
        )}
      </div>

      {/* Animated error message */}
      <AnimatePresence initial={false}>
        {showError && (
          <motion.p
            key="error"
            initial={{ opacity: 0, height: 0, marginTop: 0 }}
            animate={{ opacity: 1, height: "auto", marginTop: 4 }}
            exit={{ opacity: 0, height: 0, marginTop: 0 }}
            transition={{ duration: 0.18 }}
            className="flex items-center gap-1 text-xs text-destructive overflow-hidden"
          >
            <AlertCircle className="size-3 shrink-0" />
            {field.errorMsg}
          </motion.p>
        )}
      </AnimatePresence>
    </div>
  )
})

// ─── ApiKeyManager ────────────────────────────────────────────────────────────

export interface ApiKeyManagerHandle {
  save: () => Promise<ApiKeySaveResult | null>
}

export interface ApiKeySaveResult {
  exchange: Exchange
  keys: Record<string, string>
}

interface ApiKeyManagerProps {
  onValidChange?: (valid: boolean) => void
}

export const ApiKeyManager = forwardRef<ApiKeyManagerHandle, ApiKeyManagerProps>(
function ApiKeyManager({ onValidChange } = {}, ref) {
  const [exchange, setExchange] = useState<Exchange | "">("")
  const [values, setValues] = useState<Record<string, string>>({})
  const [tutorialOpen, setTutorialOpen] = useState(false)

  const fields = exchange ? EXCHANGE_FIELDS[exchange] : []
  const allValid =
    exchange !== "" &&
    fields.every((f) => f.pattern.test(values[f.id] ?? ""))

  // Notify parent whenever validity changes
  useEffect(() => {
    onValidChange?.(allValid)
  }, [allValid, onValidChange])

  // Expose save() so the parent's Continue button can trigger it
  useImperativeHandle(ref, () => ({
    save: async () => {
      if (!allValid || exchange === "") return null
      return {
        exchange,
        keys: { ...values },
      }
    },
  }), [allValid, exchange, values])

  const handleExchangeChange = useCallback((val: string) => {
    setExchange(val as Exchange)
    setValues({})
    setTutorialOpen(false)
  }, [])

  const handleFieldChange = useCallback((id: string, val: string) => {
    setValues((prev) => ({ ...prev, [id]: val }))
  }, [])

  const closeTutorial = useCallback(() => setTutorialOpen(false), [])

  return (
    <>
      <div className="rounded-xl p-[1px] bg-gradient-to-br from-primary/40 via-border to-accent/30">
        <div className="rounded-xl bg-card p-5 space-y-5 backdrop-blur-sm">

          {/* Header */}
          <div className="flex items-center gap-2">
            <div className="w-7 h-7 rounded-lg bg-primary/15 flex items-center justify-center">
              <KeyRound className="size-4 text-primary" />
            </div>
            <div>
              <p className="text-sm font-semibold text-foreground leading-none">API Key Manager</p>
              <p className="text-[10px] text-muted-foreground mt-0.5">Connect your CEX account securely</p>
            </div>
          </div>

          {/* Exchange selector */}
          <div className="space-y-1.5">
            <Label className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
              Exchange
            </Label>
            <Select value={exchange} onValueChange={handleExchangeChange}>
              <SelectTrigger className="w-full">
                <SelectValue placeholder="Select an exchange…" />
              </SelectTrigger>
              <SelectContent>
                {(Object.keys(EXCHANGE_LABELS) as Exchange[]).map((key) => (
                  <SelectItem key={key} value={key}>
                    {EXCHANGE_LABELS[key]}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/*
            Fixed-height container prevents layout jump when switching between
            the 2-field (Binance) and 3-field (Bitget) layouts.
            min-h covers 3 fields × ~72px each + gaps.
          */}
          <div className="min-h-[220px]">
            <AnimatePresence mode="wait">
              {exchange && (
                <motion.div
                  key={exchange}
                  initial={{ opacity: 0, y: 6 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -6 }}
                  transition={{ duration: 0.18 }}
                  className="space-y-4"
                >
                  {fields.map((field) => (
                    <FieldRow
                      key={field.id}
                      field={field}
                      value={values[field.id] ?? ""}
                      onChange={(val) => handleFieldChange(field.id, val)}
                    />
                  ))}

                  {/* Tutorial link — appears once an exchange is selected */}
                  <button
                    type="button"
                    onClick={() => setTutorialOpen(true)}
                    className="flex items-center gap-1.5 text-xs text-primary/80 hover:text-primary transition-colors mt-1"
                  >
                    <BookOpen className="size-3.5" />
                    How to get your {EXCHANGE_LABELS[exchange as Exchange]} API keys
                  </button>
                </motion.div>
              )}
            </AnimatePresence>
          </div>

          {/* Disclaimer */}
          <p className="text-[10px] text-muted-foreground text-center leading-relaxed">
            Keys are used read-only to fetch trade data. Never share keys with withdrawal permissions.
          </p>
        </div>
      </div>

      {/* Tutorial Lightbox */}
      {tutorialOpen && exchange && (
        <TutorialLightbox exchange={exchange as Exchange} onClose={closeTutorial} />
      )}
    </>
  )
})

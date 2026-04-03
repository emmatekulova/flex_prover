"use client"

import { X, Shield, ArrowRight, Lock, Cpu, Link2 } from "lucide-react"
import { motion } from "framer-motion"
import { Button } from "@/components/ui/button"

interface DocsSheetProps {
  open: boolean
  onClose: () => void
}

export function DocsSheet({ open, onClose }: DocsSheetProps) {
  if (!open) return null

  return (
    <>
      {/* Backdrop */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        onClick={onClose}
        className="fixed inset-0 bg-background/80 backdrop-blur-sm z-50"
      />

      {/* Sheet */}
      <motion.div
        initial={{ x: "100%" }}
        animate={{ x: 0 }}
        exit={{ x: "100%" }}
        transition={{ type: "spring", damping: 25, stiffness: 200 }}
        className="fixed right-0 top-0 bottom-0 w-full max-w-md bg-card border-l border-border z-50 overflow-y-auto"
      >
        <div className="p-6">
          {/* Header */}
          <div className="flex items-center justify-between mb-8">
            <h2 className="text-xl font-bold text-foreground">Documentation</h2>
            <Button
              variant="ghost"
              size="icon"
              onClick={onClose}
              className="text-muted-foreground hover:text-foreground"
            >
              <X className="w-5 h-5" />
            </Button>
          </div>

          {/* How It Works */}
          <section className="mb-8">
            <h3 className="text-sm font-semibold text-foreground mb-4 flex items-center gap-2">
              <Cpu className="w-4 h-4 text-primary" />
              How It Works
            </h3>
            <div className="space-y-3">
              <Step number={1} title="Connect Binance API" desc="Provide your read-only API key to access trading data" />
              <Arrow />
              <Step number={2} title="Flare TEE Processing" desc="Data is processed inside a secure hardware enclave" />
              <Arrow />
              <Step number={3} title="ENS Binding" desc="Your verified metrics are cryptographically bound to your ENS" />
              <Arrow />
              <Step number={4} title="On-Chain Proof" desc="A verifiable proof is published to the blockchain" />
            </div>
          </section>

          {/* Privacy Section */}
          <section className="mb-8">
            <h3 className="text-sm font-semibold text-foreground mb-4 flex items-center gap-2">
              <Lock className="w-4 h-4 text-accent" />
              Privacy & Security
            </h3>
            <div className="p-4 rounded-xl bg-secondary/50 border border-border space-y-3">
              <div className="flex items-start gap-3">
                <Shield className="w-5 h-5 text-accent mt-0.5" />
                <div>
                  <p className="text-sm font-medium text-foreground">Secure Enclave Processing</p>
                  <p className="text-xs text-muted-foreground mt-1">
                    Your API keys never leave the secure TEE enclave. We cannot see, store, 
                    or access your credentials at any point.
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <Link2 className="w-5 h-5 text-accent mt-0.5" />
                <div>
                  <p className="text-sm font-medium text-foreground">Zero-Knowledge Design</p>
                  <p className="text-xs text-muted-foreground mt-1">
                    Only the metrics you choose to prove are included in the output. 
                    Your trading history and positions remain private.
                  </p>
                </div>
              </div>
            </div>
          </section>

          {/* What is a TEE */}
          <section className="mb-8">
            <h3 className="text-sm font-semibold text-foreground mb-4">
              What is a TEE?
            </h3>
            <p className="text-sm text-muted-foreground leading-relaxed">
              A Trusted Execution Environment (TEE) is a secure area of a processor that 
              guarantees code and data loaded inside is protected with respect to confidentiality 
              and integrity. FlexProver uses Flare&apos;s TEE infrastructure to ensure your sensitive 
              data is never exposed.
            </p>
          </section>

          {/* FAQ */}
          <section>
            <h3 className="text-sm font-semibold text-foreground mb-4">
              FAQ
            </h3>
            <div className="space-y-4">
              <FaqItem 
                q="Do you store my API keys?" 
                a="No. Your keys are processed entirely within the TEE and destroyed after verification." 
              />
              <FaqItem 
                q="Can I revoke a proof?" 
                a="Proofs are immutable once published. However, if you sell your ENS, the identity binding will show a mismatch warning." 
              />
              <FaqItem 
                q="What if I want to update my stats?" 
                a="Simply generate a new proof. Your Flex Card will always reflect the most recent verification." 
              />
            </div>
          </section>
        </div>
      </motion.div>
    </>
  )
}

function Step({ number, title, desc }: { number: number; title: string; desc: string }) {
  return (
    <div className="flex items-start gap-3">
      <div className="w-6 h-6 rounded-full bg-primary/20 flex items-center justify-center flex-shrink-0">
        <span className="text-xs font-bold text-primary">{number}</span>
      </div>
      <div>
        <p className="text-sm font-medium text-foreground">{title}</p>
        <p className="text-xs text-muted-foreground">{desc}</p>
      </div>
    </div>
  )
}

function Arrow() {
  return (
    <div className="flex justify-center py-1">
      <ArrowRight className="w-4 h-4 text-muted-foreground rotate-90" />
    </div>
  )
}

function FaqItem({ q, a }: { q: string; a: string }) {
  return (
    <div className="p-3 rounded-lg bg-secondary/30 border border-border">
      <p className="text-sm font-medium text-foreground mb-1">{q}</p>
      <p className="text-xs text-muted-foreground">{a}</p>
    </div>
  )
}

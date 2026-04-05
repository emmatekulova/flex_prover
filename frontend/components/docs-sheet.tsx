"use client"

import { Shield, Lock, Cpu, Link2, Github, ExternalLink } from "lucide-react"

export function DocsContent() {
  return (
    <div className="max-w-2xl mx-auto space-y-10">
      <div>
        <h2 className="text-2xl font-bold text-foreground mb-1">Documentation</h2>
        <p className="text-muted-foreground text-sm">Learn how FlexProver works and how your data is protected.</p>
      </div>

      {/* How It Works */}
      <section>
        <h3 className="text-sm font-semibold text-foreground mb-4 flex items-center gap-2">
          <Cpu className="w-4 h-4 text-primary" />
          How It Works
        </h3>
        <div className="space-y-3">
          <Step number={1} title="Connect Binance API" desc="Provide your read-only API key to access trading data" />
          <Step number={2} title="Flare TEE Processing" desc="Data is processed inside a secure hardware enclave" />
          <Step number={3} title="ENS Binding" desc="Your verified metrics are cryptographically bound to your ENS" />
          <Step number={4} title="On-Chain Proof" desc="A verifiable proof is published to the blockchain" />
        </div>
      </section>

      {/* Privacy Section */}
      <section>
        <h3 className="text-sm font-semibold text-foreground mb-4 flex items-center gap-2">
          <Lock className="w-4 h-4 text-accent" />
          Privacy &amp; Security
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
      <section>
        <h3 className="text-sm font-semibold text-foreground mb-4">What is a TEE?</h3>
        <p className="text-sm text-muted-foreground leading-relaxed mb-4">
          A Trusted Execution Environment (TEE) is a secure area of a processor that
          guarantees code and data loaded inside is protected with respect to confidentiality
          and integrity. FlexProver uses Flare&apos;s TEE infrastructure to ensure your sensitive
          data is never exposed.
        </p>
        <div className="flex flex-col gap-2">
          <a
            href="https://dev.flare.network/fcc/guides/sign-extension#architecture"
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-2 text-xs text-primary hover:underline"
          >
            <ExternalLink className="w-3.5 h-3.5" />
            Adapted from Flare TEE Sign Extension guide
          </a>
          <a
            href="https://github.com/emmatekulova/flex_prover/tree/main/fce-sign"
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-2 text-xs text-primary hover:underline"
          >
            <Github className="w-3.5 h-3.5" />
            Verify the TEE extension source code on GitHub
          </a>
        </div>
      </section>

      {/* FAQ */}
      <section>
        <h3 className="text-sm font-semibold text-foreground mb-4">FAQ</h3>
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


function FaqItem({ q, a }: { q: string; a: string }) {
  return (
    <div className="p-3 rounded-lg bg-secondary/30 border border-border">
      <p className="text-sm font-medium text-foreground mb-1">{q}</p>
      <p className="text-xs text-muted-foreground">{a}</p>
    </div>
  )
}

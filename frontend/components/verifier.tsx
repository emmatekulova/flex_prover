"use client"

import { useState, useEffect } from "react"
import { motion, AnimatePresence } from "framer-motion"
import { Search, Upload, Shield, CheckCircle2, AlertTriangle, Cpu, FileCheck, User } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"

export function Verifier({ initialHash = "" }: { initialHash?: string }) {
  const [proofHash, setProofHash] = useState(initialHash)
  const [verificationState, setVerificationState] = useState<"idle" | "verifying" | "verified" | "mismatch">("idle")
  const [showReport, setShowReport] = useState(false)
  const [reportDate, setReportDate] = useState<string>("")

  useEffect(() => {
    if (showReport) setReportDate(new Date().toLocaleDateString())
  }, [showReport])

  const handleVerify = async () => {
    if (!proofHash.trim()) return
    
    setVerificationState("verifying")
    setShowReport(false)
    
    // Simulate verification
    await new Promise((resolve) => setTimeout(resolve, 2000))
    
    // Randomly determine if identity matches (for demo)
    const matches = Math.random() > 0.3
    setVerificationState(matches ? "verified" : "mismatch")
    setShowReport(true)
  }

  const handleUpload = () => {
    // Mock file upload
    setProofHash("0x7f8a9b3c...proof_hash_from_image")
  }

  const resetVerifier = () => {
    setProofHash("")
    setVerificationState("idle")
    setShowReport(false)
  }

  return (
    <div className="w-full max-w-lg mx-auto space-y-6">
      <div className="text-center mb-8">
        <div className="w-14 h-14 rounded-full bg-accent/20 flex items-center justify-center mx-auto mb-4">
          <Shield className="w-7 h-7 text-accent" />
        </div>
        <h2 className="text-2xl font-bold text-foreground mb-2">Verify a Proof</h2>
        <p className="text-muted-foreground">
          Enter a proof hash or upload a Flex Card image to verify its authenticity.
        </p>
      </div>

      {/* Input Section */}
      <div className="space-y-4">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input
            placeholder="Paste proof hash (0x...)"
            value={proofHash}
            onChange={(e) => setProofHash(e.target.value)}
            className="pl-10 bg-secondary border-border text-foreground placeholder:text-muted-foreground h-12"
          />
        </div>

        <div className="flex items-center gap-3">
          <div className="flex-1 h-px bg-border" />
          <span className="text-xs text-muted-foreground">or</span>
          <div className="flex-1 h-px bg-border" />
        </div>

        <Button
          variant="outline"
          onClick={handleUpload}
          className="w-full h-12 border-dashed border-2 border-border text-muted-foreground hover:text-foreground hover:bg-secondary"
        >
          <Upload className="w-4 h-4 mr-2" />
          Upload Flex Image
        </Button>

        <Button
          onClick={handleVerify}
          disabled={!proofHash.trim() || verificationState === "verifying"}
          className="w-full h-12 bg-accent text-accent-foreground hover:bg-accent/90"
        >
          {verificationState === "verifying" ? (
            <>
              <motion.div
                animate={{ rotate: 360 }}
                transition={{ repeat: Infinity, duration: 1, ease: "linear" }}
                className="w-4 h-4 mr-2 border-2 border-accent-foreground/30 border-t-accent-foreground rounded-full"
              />
              Verifying...
            </>
          ) : (
            <>
              <Shield className="w-4 h-4 mr-2" />
              Verify Proof
            </>
          )}
        </Button>
      </div>

      {/* Attestation Report */}
      <AnimatePresence>
        {showReport && (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            className="space-y-4"
          >
            <div className="p-5 rounded-xl bg-card border border-border space-y-4">
              <div className="flex items-center justify-between">
                <h3 className="font-semibold text-foreground">TEE Attestation Report</h3>
                <span className="text-xs text-muted-foreground">
                  {reportDate}
                </span>
              </div>

              {/* Hardware Verified */}
              <div className="flex items-center justify-between p-3 rounded-lg bg-secondary/50 border border-border">
                <div className="flex items-center gap-3">
                  <Cpu className="w-5 h-5 text-muted-foreground" />
                  <span className="text-sm text-foreground">Hardware Verified</span>
                </div>
                <div className="flex items-center gap-1.5 text-primary">
                  <CheckCircle2 className="w-5 h-5" />
                  <span className="text-sm font-medium">Verified</span>
                </div>
              </div>

              {/* Proof Integrity */}
              <div className="flex items-center justify-between p-3 rounded-lg bg-secondary/50 border border-border">
                <div className="flex items-center gap-3">
                  <FileCheck className="w-5 h-5 text-muted-foreground" />
                  <span className="text-sm text-foreground">Proof Integrity</span>
                </div>
                <div className="flex items-center gap-1.5 text-primary">
                  <CheckCircle2 className="w-5 h-5" />
                  <span className="text-sm font-medium">Valid</span>
                </div>
              </div>

              {/* ENS Identity Match */}
              <div
                className={`flex items-center justify-between p-3 rounded-lg border ${
                  verificationState === "verified"
                    ? "bg-primary/10 border-primary/30"
                    : "bg-destructive/10 border-destructive/30"
                }`}
              >
                <div className="flex items-center gap-3">
                  <User className="w-5 h-5 text-muted-foreground" />
                  <span className="text-sm text-foreground">ENS Identity Match</span>
                </div>
                {verificationState === "verified" ? (
                  <div className="flex items-center gap-1.5 text-primary">
                    <CheckCircle2 className="w-5 h-5" />
                    <span className="text-sm font-medium">Verified</span>
                  </div>
                ) : (
                  <div className="flex items-center gap-1.5 text-destructive">
                    <AlertTriangle className="w-5 h-5" />
                    <span className="text-sm font-medium">Mismatch</span>
                  </div>
                )}
              </div>

              {/* Warning for mismatch */}
              {verificationState === "mismatch" && (
                <motion.div
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  className="p-3 rounded-lg bg-destructive/10 border border-destructive/30"
                >
                  <div className="flex items-start gap-2">
                    <AlertTriangle className="w-4 h-4 text-destructive mt-0.5" />
                    <div>
                      <p className="text-sm font-medium text-destructive">Identity Mismatch Warning</p>
                      <p className="text-xs text-muted-foreground mt-1">
                        The ENS name has been transferred since this proof was generated. 
                        The current owner does not match the original prover.
                      </p>
                    </div>
                  </div>
                </motion.div>
              )}
            </div>

            <Button
              variant="ghost"
              onClick={resetVerifier}
              className="w-full text-muted-foreground hover:text-foreground"
            >
              Verify Another Proof
            </Button>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}

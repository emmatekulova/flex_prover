#!/usr/bin/env bash
# run-attest.sh — convenience wrapper for the attest tool.
#
# Usage:
#   ./run-attest.sh -mode growth -lookbackDays 7 -apiKey YOUR_KEY -secretKey YOUR_SECRET
#   ./run-attest.sh -mode ticker -symbol BTCUSDT
#   ./run-attest.sh -mode profile -apiKey YOUR_KEY -secretKey YOUR_SECRET
#
# All flags are forwarded to cmd/attest. If INSTRUCTION_SENDER, PRIVATE_KEY,
# TUNNEL_URL, etc. are set in .env they are picked up automatically.
#
# The script automatically ensures the InstructionSender contract has its
# extension ID set before attempting attestation (idempotent).

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TOOLS_DIR="$SCRIPT_DIR/fce-sign/go/tools"

if [[ ! -d "$TOOLS_DIR" ]]; then
  echo "ERROR: expected tools directory not found: $TOOLS_DIR" >&2
  echo "Run this script from the repository root, or from the directory that contains run-attest.sh." >&2
  exit 1
fi

cd "$TOOLS_DIR"
exec go run ./cmd/attest "$@"

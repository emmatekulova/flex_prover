export interface AttestationSubmitRequest {
  exchange?: "binance" | "bitget"
  apiKey: string
  secretKey: string
  passphrase?: string
  wallet: string
  windowDays?: number
}

export interface AttestationResult {
  txHash: string
  attestedWallet: string
  providedWallet: string
  startDate: string
  startTimestamp: number
  endDate: string
  endTimestamp: number
  profitPercent: string
}

export interface AttestationApiResponse {
  result: AttestationResult
}

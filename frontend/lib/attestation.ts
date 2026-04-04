export interface AttestationSubmitRequest {
  apiKey: string
  secretKey: string
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

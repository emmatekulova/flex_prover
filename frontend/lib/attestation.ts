export interface AttestationSubmitRequest {
  exchange?: "binance" | "bitget"
  apiKey: string
  secretKey: string
  passphrase?: string
  wallet: string
  windowDays?: number
  attestationType?: "portfolio-growth" | "individual-trades"
  selectedAssets?: string[]
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

export interface TradePosition {
  asset: string
  quantity: string
  priceUsdt: string
  valueUsdt: string
}

export interface IndividualTradesResult {
  txHash: string
  attestedWallet: string
  providedWallet: string
  positions: TradePosition[]
  totalUsdt: string
  fetchedAt: number
}

export interface PositionsFetchResponse {
  exchange: string
  positions: TradePosition[]
  fetchedAt: number
}

export interface AttestationApiResponse {
  result?: AttestationResult
  tradesResult?: IndividualTradesResult
}

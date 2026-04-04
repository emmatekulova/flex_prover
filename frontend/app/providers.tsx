'use client'

import '@/lib/reown' // side-effect: initialises AppKit + registers <appkit-button>
import { cookieToInitialState, WagmiProvider, type Config } from 'wagmi'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { wagmiConfig } from '@/lib/reown'
import { type ReactNode, useState } from 'react'

export function Providers({
  children,
  cookies,
}: {
  children: ReactNode
  cookies: string | null
}) {
  const [queryClient] = useState(() => new QueryClient())
  const initialState = cookieToInitialState(wagmiConfig as Config, cookies)

  return (
    <WagmiProvider config={wagmiConfig as Config} initialState={initialState}>
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    </WagmiProvider>
  )
}

'use client'

import '@/lib/reown' // side-effect: initialises AppKit + registers <appkit-button>
import { WagmiProvider } from 'wagmi'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { wagmiConfig } from '@/lib/reown'
import { type ReactNode, useState } from 'react'

export function Providers({ children }: { children: ReactNode }) {
  // QueryClient must be created per-component to avoid sharing across requests
  const [queryClient] = useState(() => new QueryClient())

  return (
    <WagmiProvider config={wagmiConfig}>
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    </WagmiProvider>
  )
}

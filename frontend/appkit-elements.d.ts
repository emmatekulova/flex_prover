// Type declarations for Reown AppKit web components registered via createAppKit()
import 'react'

declare module 'react' {
  namespace JSX {
    interface IntrinsicElements {
      'appkit-button': React.DetailedHTMLProps<React.HTMLAttributes<HTMLElement>, HTMLElement> & {
        label?: string
        loadingLabel?: string
        disabled?: boolean
        balance?: 'show' | 'hide'
        size?: 'md' | 'sm'
      }
    }
  }
}

import { describe, it, expect } from 'vitest'
import { buildWhaleMessage, parseHederaAccountId } from '@/components/whale-page'

describe('buildWhaleMessage', () => {
  it('builds a canonical message for signing', () => {
    const msg = buildWhaleMessage(
      'abc123',
      'https://example.com/proof/0x1234',
      '0.0.12345'
    )
    expect(msg).toContain('FlexProver Whale Verification')
    expect(msg).toContain('abc123')
    expect(msg).toContain('https://example.com/proof/0x1234')
    expect(msg).toContain('0.0.12345')
  })
})

describe('parseHederaAccountId', () => {
  it('accepts valid Hedera account IDs', () => {
    expect(parseHederaAccountId('0.0.12345')).toBe(true)
    expect(parseHederaAccountId('0.0.1')).toBe(true)
  })
  it('rejects invalid formats', () => {
    expect(parseHederaAccountId('')).toBe(false)
    expect(parseHederaAccountId('12345')).toBe(false)
    expect(parseHederaAccountId('0.0.')).toBe(false)
  })
})

import { describe, expect, it } from 'vitest'
import { resultCountText } from './result-count'

describe('resultCountText', () => {
  it('counts every hit when the server returned fewer than its cap', () => {
    expect(resultCountText(4, 'status code')).toBe('4 results for “status code”')
  })

  it('says only the first hits are shown when the server stopped at its cap', () => {
    expect(resultCountText(50, 'http')).toBe('First 50 results for “http”')
  })
})

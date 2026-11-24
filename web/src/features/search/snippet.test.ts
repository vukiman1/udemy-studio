import { describe, expect, it } from 'vitest'
import { splitSnippet } from './snippet'

describe('splitSnippet', () => {
  it('separates highlighted matches from the text around them', () => {
    expect(splitSnippet('Mỗi \ue000kết\ue001 \ue000nối\ue001 TCP')).toEqual([
      { text: 'Mỗi ', isMatch: false },
      { text: 'kết', isMatch: true },
      { text: ' ', isMatch: false },
      { text: 'nối', isMatch: true },
      { text: ' TCP', isMatch: false },
    ])
  })

  it('keeps markup in the text as text', () => {
    expect(splitSnippet('<b>\ue000x\ue001</b>')).toEqual([
      { text: '<b>', isMatch: false },
      { text: 'x', isMatch: true },
      { text: '</b>', isMatch: false },
    ])
  })
})

import { useEffect, useState } from 'react'

export function useDelayedFlag(isRaised: boolean, delayMs: number): boolean {
  const [isShown, setIsShown] = useState(false)
  useEffect(() => {
    if (!isRaised) return undefined
    const timer = setTimeout(() => {
      setIsShown(true)
    }, delayMs)
    return () => {
      clearTimeout(timer)
      setIsShown(false)
    }
  }, [isRaised, delayMs])
  return isRaised && isShown
}

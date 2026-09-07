import { useEffect, useState } from 'react'

// Re-renders every `intervalMs` so callers can derive live "Xs ago"-style
// labels from a fixed timestamp without polling the API again.
export function useNow(intervalMs = 1000): Date {
  const [now, setNow] = useState(() => new Date())

  useEffect(() => {
    const id = setInterval(() => setNow(new Date()), intervalMs)
    return () => clearInterval(id)
  }, [intervalMs])

  return now
}

import { useEffect, useState } from 'react'
import type { HealthStatus, Price } from './types'

interface PriceFeedState {
  price: Price | null
  health: HealthStatus
  error: string | null
  lastFetchedAt: Date | null
}

export function usePriceFeed(pollMs: number): PriceFeedState {
  const [price, setPrice] = useState<Price | null>(null)
  const [health, setHealth] = useState<HealthStatus>('unknown')
  const [error, setError] = useState<string | null>(null)
  const [lastFetchedAt, setLastFetchedAt] = useState<Date | null>(null)

  useEffect(() => {
    // Scoped to this effect instance (not a component-level ref) so React's
    // dev-mode StrictMode double-invoke (mount -> cleanup -> mount) can't
    // let a stale in-flight request block the new instance's own poll.
    const controller = new AbortController()
    let inFlight = false

    async function poll() {
      if (inFlight) return
      inFlight = true

      try {
        const [priceRes, healthRes] = await Promise.all([
          fetch('/price', { signal: controller.signal }),
          fetch('/health', { signal: controller.signal }),
        ])

        if (!priceRes.ok) {
          throw new Error(`/price returned ${priceRes.status}`)
        }

        const priceData: Price = await priceRes.json()
        setPrice(priceData)
        setHealth(healthRes.ok ? 'healthy' : 'unhealthy')
        setLastFetchedAt(new Date())
        setError(null)
      } catch (err) {
        if (controller.signal.aborted) return
        setError(err instanceof Error ? err.message : 'Failed to reach the API')
        setHealth('unknown')
      } finally {
        inFlight = false
      }
    }

    poll()
    const id = setInterval(poll, pollMs)

    return () => {
      controller.abort()
      clearInterval(id)
    }
  }, [pollMs])

  return { price, health, error, lastFetchedAt }
}

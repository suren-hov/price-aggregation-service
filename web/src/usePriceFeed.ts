import { useEffect, useRef, useState } from 'react'
import type { HealthStatus, Price } from './types'

const POLL_MS = 5000

interface PriceFeedState {
  price: Price | null
  health: HealthStatus
  error: string | null
  lastFetchedAt: Date | null
}

export function usePriceFeed(): PriceFeedState {
  const [price, setPrice] = useState<Price | null>(null)
  const [health, setHealth] = useState<HealthStatus>('unknown')
  const [error, setError] = useState<string | null>(null)
  const [lastFetchedAt, setLastFetchedAt] = useState<Date | null>(null)
  const inFlight = useRef(false)

  useEffect(() => {
    let cancelled = false

    async function poll() {
      if (inFlight.current) return
      inFlight.current = true

      try {
        const [priceRes, healthRes] = await Promise.all([
          fetch('/price'),
          fetch('/health'),
        ])

        if (cancelled) return

        if (!priceRes.ok) {
          throw new Error(`/price returned ${priceRes.status}`)
        }

        const priceData: Price = await priceRes.json()
        setPrice(priceData)
        setHealth(healthRes.ok ? 'healthy' : 'unhealthy')
        setLastFetchedAt(new Date())
        setError(null)
      } catch (err) {
        if (cancelled) return
        setError(err instanceof Error ? err.message : 'Failed to reach the API')
        setHealth('unknown')
      } finally {
        inFlight.current = false
      }
    }

    poll()
    const id = setInterval(poll, POLL_MS)

    return () => {
      cancelled = true
      clearInterval(id)
    }
  }, [])

  return { price, health, error, lastFetchedAt }
}

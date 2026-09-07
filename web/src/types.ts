export interface Price {
  price: number
  currency: string
  sources_used: number
  last_updated: string
  stale: boolean
}

export type HealthStatus = 'healthy' | 'unhealthy' | 'unknown'

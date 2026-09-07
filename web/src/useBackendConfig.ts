import { useEffect, useState } from 'react'
import type { BackendConfig } from './types'

// Fetched once - the backend's poll interval is fixed at startup (from its
// own env config), so there's nothing to keep polling for here.
export function useBackendConfig(): BackendConfig | null {
  const [config, setConfig] = useState<BackendConfig | null>(null)

  useEffect(() => {
    const controller = new AbortController()

    fetch('/config', { signal: controller.signal })
      .then((res) => (res.ok ? res.json() : null))
      .then((data: BackendConfig | null) => {
        if (data) setConfig(data)
      })
      .catch(() => {
        // Non-critical: the dashboard works fine without this label.
      })

    return () => controller.abort()
  }, [])

  return config
}

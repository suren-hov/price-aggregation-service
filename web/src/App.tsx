import { usePriceFeed } from './usePriceFeed'

const currencyFormatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  minimumFractionDigits: 2,
})

function StatusBadge({ health }: { health: 'healthy' | 'unhealthy' | 'unknown' }) {
  const styles: Record<typeof health, string> = {
    healthy: 'bg-emerald-500/15 text-emerald-400 ring-emerald-500/30',
    unhealthy: 'bg-rose-500/15 text-rose-400 ring-rose-500/30',
    unknown: 'bg-slate-500/15 text-slate-400 ring-slate-500/30',
  }
  const label: Record<typeof health, string> = {
    healthy: 'Healthy',
    unhealthy: 'Unhealthy',
    unknown: 'Unknown',
  }

  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-sm font-medium ring-1 ring-inset ${styles[health]}`}
    >
      <span className="h-1.5 w-1.5 rounded-full bg-current" />
      {label[health]}
    </span>
  )
}

function Stat({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-xl bg-slate-800/50 p-4 ring-1 ring-inset ring-slate-700/50">
      <dt className="text-sm text-slate-400">{label}</dt>
      <dd className="mt-1 text-lg font-semibold text-slate-100">{value}</dd>
    </div>
  )
}

export default function App() {
  const { price, health, error, lastFetchedAt } = usePriceFeed()

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100">
      <div className="mx-auto max-w-2xl px-6 py-16">
        <header className="mb-8 flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold tracking-tight">BTC Price Aggregator</h1>
            <p className="mt-1 text-sm text-slate-400">
              Live average across Coinbase, Kraken, and Bitstamp
            </p>
          </div>
          <StatusBadge health={health} />
        </header>

        {error && (
          <div className="mb-6 rounded-xl bg-rose-500/10 px-4 py-3 text-sm text-rose-300 ring-1 ring-inset ring-rose-500/30">
            {error}
          </div>
        )}

        <div className="rounded-2xl bg-slate-900 p-8 ring-1 ring-inset ring-slate-800">
          <p className="text-sm text-slate-400">BTC / {price?.currency ?? 'USD'}</p>
          <p className="mt-2 text-5xl font-bold tabular-nums">
            {price ? currencyFormatter.format(price.price) : '—'}
          </p>
          {price?.stale && (
            <p className="mt-3 inline-flex items-center gap-1.5 rounded-full bg-amber-500/15 px-3 py-1 text-sm font-medium text-amber-400 ring-1 ring-inset ring-amber-500/30">
              Stale price — sources may be unreachable
            </p>
          )}
        </div>

        <dl className="mt-6 grid grid-cols-2 gap-4">
          <Stat label="Sources used" value={price ? String(price.sources_used) : '—'} />
          <Stat
            label="Last updated"
            value={
              price?.last_updated
                ? new Date(price.last_updated).toLocaleTimeString()
                : '—'
            }
          />
        </dl>

        <p className="mt-8 text-center text-xs text-slate-500">
          {lastFetchedAt
            ? `Dashboard refreshed at ${lastFetchedAt.toLocaleTimeString()}`
            : 'Connecting to the price service…'}
        </p>
      </div>
    </div>
  )
}

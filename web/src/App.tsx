import { usePriceFeed } from './usePriceFeed'
import { useNow } from './useNow'

const currencyFormatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  minimumFractionDigits: 2,
})

function secondsAgoLabel(lastUpdated: string, now: Date): string {
  const secondsAgo = Math.max(0, Math.round((now.getTime() - new Date(lastUpdated).getTime()) / 1000))
  if (secondsAgo < 2) return 'just now'
  return `${secondsAgo}s ago`
}

type DisplayHealth = 'healthy' | 'unhealthy' | 'unknown' | 'starting'

function StatusBadge({ health }: { health: DisplayHealth }) {
  const styles: Record<DisplayHealth, string> = {
    healthy: 'bg-emerald-500/15 text-emerald-400 ring-emerald-500/30',
    unhealthy: 'bg-rose-500/15 text-rose-400 ring-rose-500/30',
    unknown: 'bg-slate-500/15 text-slate-400 ring-slate-500/30',
    starting: 'bg-slate-500/15 text-slate-400 ring-slate-500/30',
  }
  const label: Record<DisplayHealth, string> = {
    healthy: 'Healthy',
    unhealthy: 'Unhealthy',
    unknown: 'Unknown',
    starting: 'Starting…',
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
  const now = useNow()

  // The backend returns HTTP 200 with a zero-value price before its very
  // first successful poll (and while `stale`), rather than an error -
  // sources_used only becomes >0 once a real aggregation has happened.
  const hasPrice = price !== null && price.sources_used > 0
  const displayHealth: DisplayHealth = !hasPrice && health !== 'healthy' ? 'starting' : health

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
          <StatusBadge health={displayHealth} />
        </header>

        {error && (
          <div className="mb-6 rounded-xl bg-rose-500/10 px-4 py-3 text-sm text-rose-300 ring-1 ring-inset ring-rose-500/30">
            {error}
          </div>
        )}

        <div className="rounded-2xl bg-slate-900 p-8 ring-1 ring-inset ring-slate-800">
          <p className="text-sm text-slate-400">BTC / {hasPrice ? price.currency : 'USD'}</p>
          <p className="mt-2 text-5xl font-bold tabular-nums">
            {hasPrice ? currencyFormatter.format(price.price) : '—'}
          </p>
          {!hasPrice && (
            <p className="mt-3 text-sm text-slate-500">
              Waiting for the first price fetch from the backend…
            </p>
          )}
          {hasPrice && price.stale && (
            <p className="mt-3 inline-flex items-center gap-1.5 rounded-full bg-amber-500/15 px-3 py-1 text-sm font-medium text-amber-400 ring-1 ring-inset ring-amber-500/30">
              Stale price — sources may be unreachable
            </p>
          )}
        </div>

        <dl className="mt-6 grid grid-cols-2 gap-4">
          <Stat label="Sources used" value={hasPrice ? String(price.sources_used) : '—'} />
          <Stat
            label="Last updated"
            value={hasPrice ? secondsAgoLabel(price.last_updated, now) : '—'}
          />
        </dl>

        <p className="mt-8 text-center text-xs text-slate-500">
          {lastFetchedAt
            ? `Dashboard checked at ${lastFetchedAt.toLocaleTimeString()}`
            : 'Connecting to the price service…'}
        </p>
        <p className="mt-1 text-center text-xs text-slate-600">
          Prices refresh from exchanges automatically in the background — refreshing
          this page won't fetch a newer price faster than that.
        </p>
      </div>
    </div>
  )
}

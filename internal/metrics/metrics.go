package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	FetchSuccess = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fetch_success_total",
			Help: "Total successful fetches",
		},
		[]string{"source"},
	)

	FetchFailure = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fetch_failure_total",
			Help: "Total failed fetches",
		},
		[]string{"source"},
	)

	CurrentPrice = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "current_price",
			Help: "Current aggregated BTC price",
		},
	)

	SourceStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "source_status",
			Help: "Source health status (1=healthy,0=unhealthy)",
		},
		[]string{"source"},
	)

	AggregatorExcluded = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "aggregator_excluded_total",
			Help: "Total times a source's price was excluded from aggregation as an outlier",
		},
		[]string{"source"},
	)

	CircuitOpen = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "circuit_open",
			Help: "Whether a source's circuit breaker is currently open (1=open,0=closed)",
		},
		[]string{"source"},
	)
)

func Register() {
	prometheus.MustRegister(
		FetchSuccess,
		FetchFailure,
		CurrentPrice,
		SourceStatus,
		AggregatorExcluded,
		CircuitOpen,
	)
}

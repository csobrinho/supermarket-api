package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

var (
	// Counter for total runs.
	runsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "supermarket_runs_total",
		Help: "Total number of supermarket runs",
	})

	// Counter for successful runs.
	runsSuccessTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "supermarket_runs_success_total",
		Help: "Total number of successful supermarket runs",
	})

	// Counter for errors by category.
	errorsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "supermarket_errors_total",
		Help: "Total number of errors by category",
	}, []string{"category"})

	// Gauge for last successful run timestamp.
	lastSuccessTimestamp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "supermarket_last_success_timestamp",
		Help: "Timestamp of the last successful run",
	})

	// Gauge for last run timestamp (success or failure).
	lastRunTimestamp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "supermarket_last_run_timestamp",
		Help: "Timestamp of the last run (success or failure)",
	})

	// Gauge for consecutive failures.
	consecutiveFailures = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "supermarket_consecutive_failures",
		Help: "Number of consecutive failures since last success",
	})

	// Gauge for last run success status.
	lastRunSuccess = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "supermarket_last_run_success",
		Help: "Whether the last run was successful (1) or failed (0)",
	})

	// Gauge for last error timestamp by category (use changes() to count occurrences).
	lastErrorTimestamp = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "supermarket_last_error_timestamp",
		Help: "Timestamp of the last error by category",
	}, []string{"category"})

	// Gauge for total promotions count.
	promotionsCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "supermarket_promotions_count",
		Help: "Total number of promotions found",
	})

	// Histogram for execution duration.
	executionDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "supermarket_execution_duration_seconds",
		Help:    "Execution duration in seconds",
		Buckets: prometheus.DefBuckets,
	})

	// Histogram for token refresh duration.
	tokenRefreshDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "supermarket_token_refresh_duration_seconds",
		Help:    "Token refresh operation duration in seconds",
		Buckets: prometheus.DefBuckets,
	})

	// Histogram for promotions fetch duration.
	promotionsFetchDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "supermarket_promotions_fetch_duration_seconds",
		Help:    "Promotions data fetch operation duration in seconds",
		Buckets: prometheus.DefBuckets,
	})

	// Histogram for clip deal duration.
	clipDealDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "supermarket_clip_deal_duration_seconds",
		Help:    "Clip deal operation duration in seconds",
		Buckets: prometheus.DefBuckets,
	})

	// Counter for retries by host, method, and status code.
	retriesTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "supermarket_retries_total",
		Help: "Total number of retries by host, method, and status code",
	}, []string{"host", "method", "status_code"})

	// Gauge for build info.
	buildInfo = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "supermarket_build_info",
		Help: "Build information (version, go_version)",
	}, []string{"version", "go_version"})

	// Gauges for clip statistics.
	clipStatsAlreadyClipped = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "supermarket_clip_stats_already_clipped",
		Help: "Number of promotions that were already clipped",
	})

	clipStatsNewlyClipped = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "supermarket_clip_stats_newly_clipped",
		Help: "Number of promotions newly clipped in this run",
	})

	clipStatsDeleted = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "supermarket_clip_stats_deleted",
		Help: "Number of deleted promotions encountered",
	})

	clipStatsIgnored = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "supermarket_clip_stats_ignored",
		Help: "Number of promotions ignored (not clippable)",
	})

	clipStatsErrors = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "supermarket_clip_stats_errors",
		Help: "Number of errors encountered while clipping",
	})

	metricsRegistry = prometheus.NewRegistry()
)

func init() {
	// Register all metrics with the custom registry.
	metricsRegistry.MustRegister(runsTotal)
	metricsRegistry.MustRegister(runsSuccessTotal)
	metricsRegistry.MustRegister(errorsTotal)
	metricsRegistry.MustRegister(lastSuccessTimestamp)
	metricsRegistry.MustRegister(lastRunTimestamp)
	metricsRegistry.MustRegister(consecutiveFailures)
	metricsRegistry.MustRegister(lastRunSuccess)
	metricsRegistry.MustRegister(lastErrorTimestamp)
	metricsRegistry.MustRegister(promotionsCount)
	metricsRegistry.MustRegister(executionDuration)
	metricsRegistry.MustRegister(tokenRefreshDuration)
	metricsRegistry.MustRegister(promotionsFetchDuration)
	metricsRegistry.MustRegister(clipDealDuration)
	metricsRegistry.MustRegister(retriesTotal)
	metricsRegistry.MustRegister(buildInfo)
	metricsRegistry.MustRegister(clipStatsAlreadyClipped)
	metricsRegistry.MustRegister(clipStatsNewlyClipped)
	metricsRegistry.MustRegister(clipStatsDeleted)
	metricsRegistry.MustRegister(clipStatsIgnored)
	metricsRegistry.MustRegister(clipStatsErrors)
}

// errorCategory represents an error category for metrics.
type errorCategory string

// Error categories.
const (
	ErrorCategoryConfigValidation errorCategory = "config_validation"
	ErrorCategoryTokenRefresh     errorCategory = "token_refresh"
	ErrorCategoryPromotionsFetch  errorCategory = "promotions_fetch"
	ErrorCategoryPromotionsParse  errorCategory = "promotions_parse"
	ErrorCategoryClipDeal         errorCategory = "clip_deal"
	ErrorCategoryMetricsPush      errorCategory = "metrics_push"
)

// RecordError increments the error counter and updates the last error timestamp for a specific category.
func RecordError(category errorCategory) {
	errorsTotal.WithLabelValues(string(category)).Inc()
	lastErrorTimestamp.WithLabelValues(string(category)).Set(float64(time.Now().Unix()))
}

// RecordSuccess records a successful run.
func RecordSuccess() {
	runsSuccessTotal.Inc()
	lastSuccessTimestamp.Set(float64(time.Now().Unix()))
	lastRunSuccess.Set(1)
	consecutiveFailures.Set(0)
}

// RecordFailure records a failed run.
func RecordFailure() {
	lastRunSuccess.Set(0)
	consecutiveFailures.Inc()
}

// RecordRunStart records the start of a run.
func RecordRunStart() {
	runsTotal.Inc()
	lastRunTimestamp.Set(float64(time.Now().Unix()))
}

// SetBuildInfo sets the build information metric.
func SetBuildInfo(version, goVersion string) {
	buildInfo.WithLabelValues(version, goVersion).Set(1)
}

// RecordRetry increments the retry counter for a specific host, method, and status code.
func RecordRetry(host, method string, statusCode int) {
	retriesTotal.WithLabelValues(host, method, fmt.Sprintf("%d", statusCode)).Inc()
}

// RecordTokenRefreshDuration records the duration of a token refresh operation.
func RecordTokenRefreshDuration(duration time.Duration) {
	tokenRefreshDuration.Observe(duration.Seconds())
}

// RecordPromotionsFetchDuration records the duration of a promotions fetch operation.
func RecordPromotionsFetchDuration(duration time.Duration) {
	promotionsFetchDuration.Observe(duration.Seconds())
}

// RecordClipDealDuration records the duration of a clip deal operation.
func RecordClipDealDuration(duration time.Duration) {
	clipDealDuration.Observe(duration.Seconds())
}

// RecordExecutionDuration records the total execution duration.
func RecordExecutionDuration(duration time.Duration) {
	executionDuration.Observe(duration.Seconds())
}

// RecordPromotionsCount sets the total promotions count.
func RecordPromotionsCount(count int) {
	promotionsCount.Set(float64(count))
}

// RecordClipStats sets the clip statistics gauges.
func RecordClipStats(alreadyClipped, newlyClipped, deleted, ignored, errors int) {
	clipStatsAlreadyClipped.Set(float64(alreadyClipped))
	clipStatsNewlyClipped.Set(float64(newlyClipped))
	clipStatsDeleted.Set(float64(deleted))
	clipStatsIgnored.Set(float64(ignored))
	clipStatsErrors.Set(float64(errors))
}

// PushMetrics pushes all metrics to the Prometheus Pushgateway.
func PushMetrics(ctx context.Context, endpoint string, job string) error {
	if endpoint == "" {
		// If no endpoint is configured, skip pushing metrics.
		return nil
	}

	pusher := push.New(endpoint, job).Gatherer(metricsRegistry)
	if err := pusher.PushContext(ctx); err != nil {
		return fmt.Errorf("failed to push metrics: %w", err)
	}
	return nil
}

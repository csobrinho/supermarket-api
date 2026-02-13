package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/csobrinho/supermarket-api/internal/metrics"
	"github.com/csobrinho/supermarket-api/internal/promotion"
	"github.com/csobrinho/supermarket-api/pkg/supermarket"
	"github.com/csobrinho/supermarket-api/providers/safeway"
	"github.com/google/logger"
)

var (
	// version is set at build time via -ldflags.
	version = "dev"

	refreshToken       = flag.String("refresh_token", supermarket.LookupEnv("REFRESH_TOKEN", ""), "Refresh token for authentication. Can also be provided via 'REFRESH_TOKEN' env.")
	clientId           = flag.String("client_id_token", supermarket.LookupEnv("CLIENT_ID", ""), "Client ID for authentication. Can also be provided via 'CLIENT_ID' env.")
	userAgent          = flag.String("user_agent", supermarket.LookupEnv("USER_AGENT", "okhttp/4.12.0"), "User agent for authentication. Can also be provided via 'USER_AGENT' env.")
	apiKey             = flag.String("api_key", supermarket.LookupEnv("API_KEY", ""), "API key for authentication. Can also be provided via 'API_KEY' env.")
	store              = flag.String("store_id", supermarket.LookupEnv("STORE_ID", ""), "Store ID to search for promotions. Can also be provided via 'STORE_ID' env.")
	appVersion         = flag.String("app_version", supermarket.LookupEnv("APP_VERSION", ""), "App version to emulate. Can also be provided via 'APP_VERSION' env.")
	clipAll            = flag.Bool("clip_all", supermarket.LookupEnvBool("CLIP_ALL", false), "If true, also clip all coupons. Can also be provided via 'CLIP_ALL' env.")
	delayMs            = flag.Int("delay_ms", supermarket.LookupEnvInt("DELAY_MS", 1000), "If provided, delay in milliseconds between requests. This value will be randomized +/-50%. Can also be provided via 'DELAY_MS' env.")
	verbose            = flag.Int("verbose", supermarket.LookupEnvInt("VERBOSE", 0), "Log verbosity level [0-4]. Can also be provided via 'VERBOSE' env.")
	prometheusEndpoint = flag.String("prometheus_endpoint", supermarket.LookupEnv("PROMETHEUS_ENDPOINT", ""), "Prometheus Pushgateway endpoint (e.g., http://localhost:9091). Can also be provided via 'PROMETHEUS_ENDPOINT' env.")
	prometheusJob      = flag.String("prometheus_job", supermarket.LookupEnv("PROMETHEUS_JOB", "supermarket"), "Prometheus job name for pushing metrics. Can also be provided via 'PROMETHEUS_JOB' env.")
)

func run(ctx context.Context) error {
	logger.SetLevel(logger.Level(*verbose))

	// Validate configuration.
	if *refreshToken == "" || *clientId == "" || *apiKey == "" || *store == "" {
		metrics.RecordError(metrics.ErrorCategoryConfigValidation)
		return fmt.Errorf("missing required configuration: refresh_token, client_id, api_key, and store_id are required")
	}

	factory := supermarket.NewFactory()
	factory.Register("safeway", safeway.Creator)

	rateLimiter := supermarket.NewRateLimiter((time.Duration(*delayMs))*time.Millisecond, 0.5)
	sm, err := factory.Create(ctx, "safeway",
		supermarket.WithUserAgent(*userAgent),
		supermarket.WithAppVersion(*appVersion),
		supermarket.WithCredentials(*clientId, *refreshToken),
		supermarket.WithApiKey(*apiKey),
		supermarket.WithDebug(*verbose > 0),
		supermarket.WithStoreID(*store),
	)
	if err != nil {
		metrics.RecordError(metrics.ErrorCategoryConfigValidation)
		return fmt.Errorf("creating client, %w", err)
	}
	a, err := sm.Authenticator()
	if err != nil {
		metrics.RecordError(metrics.ErrorCategoryConfigValidation)
		return fmt.Errorf("creating authenticator, %w", err)
	}

	logger.Info("main: getting an access token...")
	start := time.Now()
	t, err := a.RefreshToken(ctx)
	if err != nil {
		metrics.RecordError(metrics.ErrorCategoryTokenRefresh)
		return fmt.Errorf("refreshing token, %w", err)
	}
	metrics.RecordTokenRefreshDuration(time.Since(start))
	logger.V(1).Infof("main: access token: %+v", t.AccessToken)

	logger.Infof("main: getting all promotions...")
	ps, err := sm.Promotion()
	if err != nil {
		metrics.RecordError(metrics.ErrorCategoryPromotionsParse)
		return fmt.Errorf("creating promotion service, %w", err)
	}
	start = time.Now()
	cds, err := ps.GetClipDeals(ctx, promotion.PromotionSearchOptions{})
	if err != nil {
		metrics.RecordError(metrics.ErrorCategoryPromotionsFetch)
		return fmt.Errorf("getting promotions, %w", err)
	}
	metrics.RecordPromotionsFetchDuration(time.Since(start))
	metrics.RecordPromotionsCount(len(cds))
	if !*clipAll {
		logger.Infof("main: not clipping any promotions...")
		return nil
	}

	stats := struct {
		prev         int
		new          int
		deleted      int
		notclippable int
		err          int
		log          bool
	}{}
	defer func() {
		logger.Infof(`main: clip stats:
    - already: %d
    - newly:   %d
    - deleted: %d
    - ignored: %d
    - errors:  %d`, stats.prev, stats.new, stats.deleted, stats.notclippable, stats.err)
		// Set clip stats metrics.
		metrics.RecordClipStats(stats.prev, stats.new, stats.deleted, stats.notclippable, stats.err)
	}()

	for _, cd := range cds {
		if cd.IsClipped {
			stats.prev++
			continue
		}
		if !cd.IsClippable {
			stats.notclippable++
			continue
		}
		if cd.IsDeleted {
			stats.deleted++
			continue
		}
		if stats.log {
			logger.Infof("main: clipping all promotions...")
			stats.log = false
		}
		start := time.Now()
		if err := ps.ClipDeal(ctx, cd); err != nil {
			metrics.RecordError(metrics.ErrorCategoryClipDeal)
			stats.err++
			logger.Errorf("main: error clipping deal %v, %v", cd, err)
			continue
		}
		metrics.RecordClipDealDuration(time.Since(start))
		stats.new++
		rateLimiter.Wait()
	}
	return nil
}

func main() {
	logger.Init("supermarket", true, false, io.Discard)
	flag.Parse()

	// Set build info.
	metrics.SetBuildInfo(version, runtime.Version())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals for graceful shutdown.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-ch
		logger.Infof("main: received signal %v, shutting down...", sig)
		cancel()
	}()

	metrics.RecordRunStart()
	start := time.Now()
	err := run(ctx)
	metrics.RecordExecutionDuration(time.Since(start))

	// Record success or failure.
	if err != nil {
		logger.Errorf("main: error, %v", err)
		metrics.RecordFailure()
	} else {
		logger.Infof("main: all done ✅")
		metrics.RecordSuccess()
	}

	// Push metrics to Prometheus Pushgateway if configured.
	if *prometheusEndpoint != "" {
		logger.Infof("main: pushing metrics to %s...", *prometheusEndpoint)
		if pushErr := metrics.PushMetrics(ctx, *prometheusEndpoint, *prometheusJob); pushErr != nil {
			logger.Errorf("main: failed to push metrics: %v", pushErr)
		}
	}

	if err != nil {
		os.Exit(1)
	}
}

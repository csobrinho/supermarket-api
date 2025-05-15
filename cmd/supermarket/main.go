package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/csobrinho/supermarket-api/internal/promotion"
	"github.com/csobrinho/supermarket-api/pkg/supermarket"
	"github.com/csobrinho/supermarket-api/providers/safeway"
	"github.com/google/logger"
)

var (
	refreshToken = flag.String(
		"refresh_token",
		supermarket.LookupEnv("REFRESH_TOKEN", ""),
		"Refresh token for authentication. Can also be provided via 'REFRESH_TOKEN' env.")
	clientId = flag.String(
		"client_id_token",
		supermarket.LookupEnv("CLIENT_ID", ""),
		"Client ID for authentication. Can also be provided via 'CLIENT_ID' env.")
	userAgent = flag.String(
		"user_agent",
		supermarket.LookupEnv("USER_AGENT", "okhttp/4.12.0"),
		"User agent for authentication. Can also be provided via 'USER_AGENT' env.")
	apiKey = flag.String(
		"api_key",
		supermarket.LookupEnv("API_KEY", ""),
		"API key for authentication. Can also be provided via 'API_KEY' env.")
	store = flag.String(
		"store_id",
		supermarket.LookupEnv("STORE_ID", ""),
		"Store ID to search for promotions. Can also be provided via 'STORE_ID' env.")
	appVersion = flag.String(
		"app_version",
		supermarket.LookupEnv("APP_VERSION", ""),
		"App version to emulate. Can also be provided via 'APP_VERSION' env.")
	clipAll = flag.Bool(
		"clip_all",
		supermarket.LookupEnvBool("CLIP_ALL", false),
		"If true, also clip all coupons. Can also be provided via 'CLIP_ALL' env.")
	delayMs = flag.Int(
		"delay_ms",
		supermarket.LookupEnvInt("DELAY_MS", 1000),
		"If provided, delay in milliseconds between requests. This value will be randomized +/-50%. "+
			"Can also be provided via 'DELAY_MS' env.")
	verbose = flag.Int(
		"verbose",
		supermarket.LookupEnvInt("VERBOSE", 0),
		"Log verbosity level [0-4]. Can also be provided via 'VERBOSE' env.")
)

func run(ctx context.Context) error {
	logger.Infof("main: registering safeway...")
	logger.SetLevel(logger.Level(*verbose))

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
		return fmt.Errorf("creating client, %w", err)
	}
	a, err := sm.Authenticator()
	if err != nil {
		return fmt.Errorf("creating authenticator, %w", err)
	}

	logger.Info("main: getting an access token...")
	t, err := a.RefreshToken(ctx)
	if err != nil {
		return fmt.Errorf("refreshing token, %w", err)
	}
	logger.V(1).Infof("main: access token: %+v", t.AccessToken)

	logger.Infof("main: getting all promotions...")
	ps, err := sm.Promotion()
	if err != nil {
		return fmt.Errorf("creating promotion service, %w", err)
	}
	cds, err := ps.GetClipDeals(ctx, promotion.PromotionSearchOptions{})
	if err != nil {
		return fmt.Errorf("getting promotions, %w", err)
	}
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
		if !cd.IsDeleted {
			stats.deleted++
			continue
		}
		if !cd.IsDeleted {
			stats.deleted++
			continue
		}
		if stats.log {
			logger.Infof("main: clipping all promotions...")
			stats.log = false
		}
		if err := ps.ClipDeal(ctx, cd); err != nil {
			return fmt.Errorf("clipping deal, %w", err)
		}
		stats.new++
		rateLimiter.Wait()
	}
	return nil
}

func main() {
	logger.Init("supermarket", true, false, io.Discard)
	flag.Parse()

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

	if err := run(ctx); err != nil {
		logger.Errorf("main: error, %v", err)
		os.Exit(1)
	}

	logger.Infof("main: all done ✅")
}

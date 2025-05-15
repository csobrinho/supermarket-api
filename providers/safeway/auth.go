package safeway

import (
	"context"
	"flag"
	"maps"
	"net/http"

	"github.com/csobrinho/supermarket-api/internal/auth"
	ihttp "github.com/csobrinho/supermarket-api/internal/http"
	"github.com/csobrinho/supermarket-api/pkg/supermarket"
	"golang.org/x/oauth2"
)

var _ auth.Service = (*authenticatorService)(nil)

type authenticatorService struct {
	client        *http.Client
	ts            oauth2.TokenSource
	authenticated bool
}

var (
	tokenUrl = flag.String(
		"safeway_token_url",
		supermarket.LookupEnv("SAFEWAY_TOKEN_URL", "https://albertsons.okta.com/oauth2/ausp6soxrIyPrm8rS2p6/v1/token"),
		"Safeway token url. Can also be provided via 'SAFEWAY_TOKEN_URL' env.")

	oauthScopes       = []string{"openid", "profile", "offline_access", "partner"}
	oauthExtraHeaders = map[string]string{
		"banner":                 "safeway",
		"platform":               "Android",
		"x-swy-application-type": "native-mobile",
		"x-swy_api_key":          "appandroid",
		"x-swy_version":          "1.1",
		"x-swy_banner":           "safeway",
		// "appversion":          "2025.\d+.\d+",
		// "storeid":             "\d+",
	}
)

func NewAuthenticator(ctx context.Context, cfg *supermarket.Config) *authenticatorService {
	config := &oauth2.Config{
		ClientID: cfg.ClientID,
		Endpoint: oauth2.Endpoint{
			TokenURL:  *tokenUrl,
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: oauthScopes,
	}
	token := &oauth2.Token{
		RefreshToken: cfg.RefreshToken,
		TokenType:    "Bearer",
	}
	// Note: The original request also sends the scopes in the refresh token request.
	ts := config.TokenSource(ctx, token)

	headers := maps.Clone(oauthExtraHeaders)
	headers["storeid"] = cfg.StoreID
	headers["appversion"] = cfg.AppVersion

	t := http.DefaultTransport
	if cfg.Debug {
		t = &ihttp.LoggingTransport{Next: http.DefaultTransport}
	}
	t = &ihttp.CustomTransport{
		Next:         t,
		UserAgent:    cfg.UserAgent,
		ExtraHeaders: headers,
	}
	base := &http.Client{
		Timeout:   cfg.Timeout,
		Transport: t,
	}

	return &authenticatorService{
		client: oauth2.NewClient(context.WithValue(ctx, oauth2.HTTPClient, base), ts),
		ts:     ts,
	}
}

func (as *authenticatorService) RefreshToken(ctx context.Context) (*oauth2.Token, error) {
	t, err := as.ts.Token()
	as.authenticated = err == nil
	return t, err
}
func (as *authenticatorService) HttpClient() *http.Client { return as.client }
func (as *authenticatorService) IsAuthenticated(ctx context.Context) bool {
	if !as.authenticated {
		return false
	}
	_, _ = as.RefreshToken(ctx)
	return as.authenticated
}

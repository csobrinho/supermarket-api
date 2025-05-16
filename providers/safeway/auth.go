package safeway

import (
	"context"
	"flag"
	"fmt"
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

	oauthScopes = []string{"openid", "profile", "offline_access", "partner"}
)

func NewAuthenticator(ctx context.Context, cfg *supermarket.Config) (*authenticatorService, error) {
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

	client, err := ihttp.New(true, cfg.Debug, cfg.UserAgent, nil, cfg.Timeout, nil)
	if err != nil {
		return nil, fmt.Errorf("authenticator: new http client, error %w", err)
	}
	return &authenticatorService{
		client: oauth2.NewClient(context.WithValue(ctx, oauth2.HTTPClient, client), ts),
		ts:     ts,
	}, nil
}

func (as *authenticatorService) RefreshToken(ctx context.Context) (*oauth2.Token, error) {
	t, err := as.ts.Token()
	as.authenticated = err == nil
	return t, err
}
func (as *authenticatorService) TokenSource() oauth2.TokenSource { return as.ts }
func (as *authenticatorService) IsAuthenticated(ctx context.Context) bool {
	if !as.authenticated {
		return false
	}
	_, _ = as.RefreshToken(ctx)
	return as.authenticated
}

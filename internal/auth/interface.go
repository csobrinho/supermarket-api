package auth

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"
)

// Service provides methods for authentication and user management.
type Service interface {
	HttpClient() *http.Client

	// RefreshToken refreshes the access token using the refresh token.
	RefreshToken(ctx context.Context) (*oauth2.Token, error)

	// IsAuthenticated checks if there is a valid authenticated session.
	IsAuthenticated(ctx context.Context) bool
}

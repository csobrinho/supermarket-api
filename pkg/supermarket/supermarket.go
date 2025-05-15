package supermarket

import (
	"github.com/csobrinho/supermarket-api/internal/auth"
	"github.com/csobrinho/supermarket-api/internal/promotion"
)

type Supermarket interface {
	// Authenticator retrieves the authenticator service.
	Authenticator() (auth.Service, error)
	// Promotion retrieves the promotion service.
	Promotion() (promotion.Service, error)
}

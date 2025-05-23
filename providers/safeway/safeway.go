package safeway

import (
	"context"

	"github.com/csobrinho/supermarket-api/internal/auth"
	"github.com/csobrinho/supermarket-api/internal/promotion"
	"github.com/csobrinho/supermarket-api/pkg/supermarket"
)

var _ supermarket.Supermarket = (*safeway)(nil)
var _ promotion.Service = (*promotionService)(nil)

func Creator(ctx context.Context, cfg *supermarket.Config) (supermarket.Supermarket, error) {
	a, err := NewAuthenticator(ctx, cfg)
	if err != nil {
		return nil, err
	}
	ps, err := NewPromotion(ctx, cfg, a.ts)
	if err != nil {
		return nil, err
	}
	return &safeway{as: a, ps: ps}, nil
}

type safeway struct {
	as *authenticatorService
	ps *promotionService
}

func (s *safeway) Authenticator() (auth.Service, error)  { return s.as, nil }
func (s *safeway) Promotion() (promotion.Service, error) { return s.ps, nil }

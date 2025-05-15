package promotion

import (
	"context"
	"time"
)

// PromotionType represents the type of promotion.
type PromotionType string

const (
	PromotionTypeClipDeal      PromotionType = "clip_deal"
	PromotionTypeCoupon        PromotionType = "coupon"
	PromotionTypeWeeklySale    PromotionType = "weekly_sale"
	PromotionTypeClearance     PromotionType = "clearance"
	PromotionTypeBOGO          PromotionType = "bogo" // Buy one get one
	PromotionTypeMixAndMatch   PromotionType = "mix_and_match"
	PromotionTypeLoyaltyReward PromotionType = "loyalty_reward"
)

// Promotion represents a promotion or deal.
type Promotion struct {
	Brand               string        `json:"brand"`
	Categories          []string      `json:"categories,omitempty"`
	ID                  string        `json:"id"`
	Description         string        `json:"description"`
	Disclaimer          string        `json:"disclaimer,omitempty"`
	Type                PromotionType `json:"type"`
	ImageID             string        `json:"image_id,omitempty"`
	Upcs                []string      `json:"upcs"` // Product IDs
	MinPurchaseQuantity *float64      `json:"min_purchase_quantity,omitempty"`
	MaxPurchaseQuantity *float64      `json:"max_purchase_quantity,omitempty"`
	Price               *float64      `json:"price,omitempty"`
	PromoCode           *string       `json:"promo_code,omitempty"`
	PromoType           *string       `json:"promo_type,omitempty"`
	ProgramType         *string       `json:"program_type,omitempty"`
	Status              string        `json:"status"`
	UsageType           string        `json:"usage_type"`
	StartDate           time.Time     `json:"start_date"`
	EndDate             time.Time     `json:"end_date"`
	IsDeleted           bool          `json:"is_deleted"`
	IsClippable         bool          `json:"is_clippable"`   // For clip deals
	IsDisplayable       bool          `json:"is_displayable"` // For clip deals
	IsClipped           bool          `json:"is_clipped"`     // For clip deals
	Item                any           // Original item from provider.
}

// ClipDeal represents a clippable coupon or deal.
type ClipDeal struct {
	Promotion
	ExpiresAfterClip bool       `json:"expires_after_clip"`
	ClippedAt        *time.Time `json:"clipped_at,omitempty"`
}

// PromotionSearchOptions provides filtering for promotions.
type PromotionSearchOptions struct {
	Type        *PromotionType `json:"type,omitempty"`
	Category    *string        `json:"category,omitempty"`
	ProductID   *string        `json:"product_id,omitempty"`
	ClippedOnly *bool          `json:"clipped_only"`
}

// Service provides methods to work with promotions.
type Service interface {
	// GetClipDeals retrieves available clip deals.
	GetClipDeals(ctx context.Context, opts PromotionSearchOptions) ([]ClipDeal, error)

	// ClipDeal clips a deal for the current user.
	ClipDeal(ctx context.Context, clipDeal ClipDeal) error
}

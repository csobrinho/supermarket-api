package safeway

import (
	"fmt"
	"strconv"
	"time"

	"github.com/csobrinho/supermarket-api/internal/promotion"
)

// EpochMillisTime is a custom type to handle Unix milliseconds timestamps.
type EpochMillisTime time.Time

// UnmarshalJSON implements the json.Unmarshaler interface and parses a JSON string representing Unix time in milliseconds.
func (emt *EpochMillisTime) UnmarshalJSON(b []byte) error {
	// Remove quotes from the JSON string.
	s := string(b)
	if s == "null" || s == `""` { // Handle null or empty string if necessary
		return nil
	}
	// JSON strings are quoted, remove them.
	// Example: "1747157766952" -> 1747157766952
	unquotedStr, err := strconv.Unquote(s)
	if err != nil {
		// If it's already a number (not a string in JSON), try parsing directly.
		// This path is less likely given the original JSON, but good for robustness.
		unquotedStr = s
	}

	// Convert string to int64 (milliseconds).
	ms, err := strconv.ParseInt(unquotedStr, 10, 64)
	if err != nil {
		return fmt.Errorf("error parsing timestamp string '%s' to int64: %w", unquotedStr, err)
	}

	// Convert milliseconds to time.Time (seconds and nanoseconds).
	*emt = EpochMillisTime(time.Unix(ms/1000, (ms%1000)*1000000))
	return nil
}

// String returns the time in a human-readable format.
func (emt EpochMillisTime) String() string {
	return time.Time(emt).Format(time.RFC3339)
}

// GetClipDealsResponse is the top-level structure for the deals response.
type GetClipDealsResponse struct {
	Coupons           []Promotion `json:"cc,omitempty"`
	PersonalizedDeals []Promotion `json:"pd,omitempty"`
}

// Promotion represents an individual promotion item.
type Promotion struct {
	AllocationID        string             `json:"allocationId,omitempty"` //
	Brand               string             `json:"brand"`                  // Maps to: Brand
	Category            string             `json:"category"`               // Maps to: Category
	ClipID              string             `json:"clipId"`                 // Maps to: ID
	Description         string             `json:"description"`            // Maps to: Description
	Disclaimer          string             `json:"disclaimer"`             // Maps to: Disclaimer
	EcomDescription     string             `json:"ecomDescription"`        // Maps to: Shorter description
	ExternalOfferID     string             `json:"extlOfferId"`            //
	ForUDescription     string             `json:"forUDescription"`        //
	Hierarchies         PromotionHierarchy `json:"hierarchies"`            // Maps to: Categories
	ImageID             string             `json:"imageId"`                // Maps to: ImageID
	Upcs                []string           `json:"upcs"`                   // Maps to: ApplicableItems
	MinPurchaseQuantity int                `json:"minPurchaseQty"`         // Maps to: MinPurchaseQuantity
	MaxPurchaseQuantity int                `json:"maxPurchaseQty"`         // Maps to: MaxPurchaseQuantity
	Price               float64            `json:"price"`                  // Maps to: Price
	RegularPrice        string             `json:"regularPrice,omitempty"` //
	OfferID             string             `json:"offerId"`                // Maps to: PromoCode
	OfferPgm            string             `json:"offerPgm"`               // Maps to: PromoType
	OfferProgramType    string             `json:"offerProgramType"`       // Maps to: ProgramType
	ProgramType         string             `json:"programType,omitempty"`  //
	OfferProtoType      string             `json:"offerProtoType"`         //
	OfferSubPgm         string             `json:"offerSubPgm,omitempty"`  //
	PurchaseIndex       string             `json:"purchaseInd"`            //
	PurchaseRank        string             `json:"purchaseRank"`           //
	Status              clipStatusType     `json:"status"`                 // Maps to: Status
	UsageType           string             `json:"usageType"`              // Maps to: UsageType
	StartDate           EpochMillisTime    `json:"startDate"`              // Maps to: StartDate
	EndDate             EpochMillisTime    `json:"endDate"`                // Maps to: EndDate
	OfferTimestamp      EpochMillisTime    `json:"offerTs"`                //
	ClippedAt           EpochMillisTime    `json:"clipTs"`                 // Maps to: ClippedAt
	IsDeleted           bool               `json:"deleted"`                // Maps to: IsDeleted
	IsClippable         bool               `json:"isClippable"`            // Maps to: IsClippable
	IsDisplayable       bool               `json:"isDisplayable"`          // Maps to: IsDisplayable
	VendorBannerCd      string             `json:"vndrBannerCd,omitempty"` //
}

func (p Promotion) convert() promotion.ClipDeal {
	pf := func(f int) *float64 {
		if f == 0 {
			return nil
		}
		ff := float64(f)
		return &ff
	}
	pt := func(t EpochMillisTime) *time.Time {
		tt := time.Time(t)
		if tt.IsZero() {
			return nil
		}
		return &tt
	}
	id := p.ClipID
	if id == "" {
		id = p.ExternalOfferID
	}
	return promotion.ClipDeal{
		ClippedAt: pt(p.ClippedAt),
		Promotion: promotion.Promotion{
			Brand:       p.Brand,
			Categories:  p.Hierarchies.Categories,
			ID:          id,
			Description: p.Description,
			Disclaimer:  p.Disclaimer,
			// Type
			ImageID:             p.ImageID,
			Upcs:                p.Upcs,
			MinPurchaseQuantity: pf(p.MinPurchaseQuantity),
			MaxPurchaseQuantity: pf(p.MaxPurchaseQuantity),
			Price:               &p.Price,
			PromoCode:           &p.OfferID,
			PromoType:           &p.OfferPgm,
			ProgramType:         &p.OfferProgramType,
			Status:              string(p.Status),
			UsageType:           p.UsageType,
			StartDate:           time.Time(p.StartDate),
			EndDate:             time.Time(p.EndDate),
			IsDeleted:           p.IsDeleted,
			IsClippable:         p.IsClippable,
			IsClipped:           p.Status == "C",
			IsDisplayable:       p.IsDisplayable,
			Item:                p,
		},
	}
}

// PromotionHierarchy represents the nested hierarchy data.
type PromotionHierarchy struct {
	Categories []string `json:"categories"`
	Events     []string `json:"events"`
}

func newClipDeal(offerID string, offerPgm string) *ClipDealRoot {
	return &ClipDealRoot{Items: []ClipDeal{
		{ClipType: "C", ItemID: offerID, ItemType: offerPgm},
		{ClipType: "L", ItemID: offerID, ItemType: offerPgm},
	}}
}

type ClipDealRoot struct {
	Items []ClipDeal `json:"items"`
}

type ClipDeal struct {
	ClipType string `json:"clipType"`
	ItemID   string `json:"itemId"`
	ItemType string `json:"itemType"`
	Status   int    `json:"status,omitempty"`
	ClipID   string `json:"clipId,omitempty"`
	ClipTs   string `json:"clipTs,omitempty"`
	Checked  bool   `json:"checked,omitempty"`
}

type couponType string

const (
	COUPON_TYPE_PERSONALIZED_DEAL couponType = "PD"
	COUPON_TYPE_COUPON_MF         couponType = "MF"
	COUPON_TYPE_COUPON_SC         couponType = "SC"
	COUPON_TYPE_COUPON_CC         couponType = "CC"
	COUPON_TYPE_GROCERY_REWARD    couponType = "GR"
	COUPON_TYPE_MONOPOLY_PRIZE    couponType = "TR"
	COUPON_TYPE_WEEKLY_AD         couponType = "WS"
)

type filterCouponType string

const (
	FILTER_COUPON_TYPE_PERSONALIZED_DEAL filterCouponType = "PD"
	FILTER_COUPON_TYPE_COUPON_MF         filterCouponType = "MF"
	FILTER_COUPON_TYPE_COUPON_SC         filterCouponType = "SC"
	FILTER_COUPON_TYPE_COUPON_CC         filterCouponType = "CC"
	FILTER_COUPON_TYPE_COUPON_MC         filterCouponType = "manufacturerCoupons"
)

type couponActionType string

const (
	COUPON_ACTION_TYPE_CLIP                couponActionType = "CL"
	COUPON_ACTION_TYPE_CLIPPED             couponActionType = "CLD"
	COUPON_ACTION_TYPE_ADD                 couponActionType = "AC"
	COUPON_ACTION_TYPE_ADDED               couponActionType = "ADC"
	COUPON_ACTION_TYPE_CLIP_COUPON         couponActionType = "CLC"
	COUPON_ACTION_TYPE_CLIP_OFFER          couponActionType = "CLCUDC"
	COUPON_ACTION_TYPE_ACTIVATE            couponActionType = "SCC"
	COUPON_ACTION_TYPE_BONUS_PATH_COMPLETE couponActionType = "CC"
)

type offerType string

const (
	OFFER_TYPE_UNLIMITED_USE  offerType = "U"
	OFFER_TYPE_ONE_TIME_USE_O offerType = "O"
	OFFER_TYPE_ONE_TIME_USE_M offerType = "M"
)

type clipStatusType string

const (
	CLIP_STATUS_TYPE_CLIPPED   clipStatusType = "C"
	CLIP_STATUS_TYPE_UNCLIPPED clipStatusType = "U"
)

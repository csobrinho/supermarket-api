package safeway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	"github.com/csobrinho/supermarket-api/internal/promotion"
	"github.com/google/logger"
	"golang.org/x/exp/maps"
)

const (
	PROMOTIONS_GET_CLIP_DEALS_URL = "https://www.safeway.com/abs/pub/mobile/j4u/api/ecomgallery?storeId=%s&offerPgm=PD-CC&includeRedeemedBonusOffers=y"
	PROMOTIONS_CLIP_DEALS_URL     = "https://www.safeway.com/abs/pub/mobile/j4u/api/offers/clip?storeId=%s"
)

var promotionsExtraHeaders = map[string]string{
	"content-type":  "application/json",
	"accept":        "application/json",
	"platform":      "android",
	"x-swy_version": "2.1",
	"x-swy_banner":  "safeway",
	// "appversion":    "2025.\d+.\d+",
	// "storeid":       "\d+",
	// "x-swy_api_key": "...",
}

type promotionService struct {
	client     *http.Client
	apiKey     string
	appVersion string
	storeID    string
}

// GetClipDeals retrieves available clip deals.
func (ps *promotionService) GetClipDeals(ctx context.Context, opts promotion.PromotionSearchOptions) ([]promotion.ClipDeal, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(PROMOTIONS_GET_CLIP_DEALS_URL, ps.storeID), nil)
	if err != nil {
		return nil, fmt.Errorf("promotion: get clip deals request, error %w", err)
	}

	for k, v := range promotionsExtraHeaders {
		req.Header.Set(k, v)
	}
	req.Header.Set("storeid", ps.storeID)
	req.Header.Set("x-swy_api_key", ps.apiKey)
	req.Header.Set("appversion", ps.appVersion)

	res, err := ps.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("promotion: get clip deals response, error %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("promotion: get clip deals response, error status %s", res.Status)
	}

	root := GetClipDealsResponse{}
	if err := json.NewDecoder(res.Body).Decode(&root); err != nil {
		return nil, fmt.Errorf("promotion: get clip deals response, error decoding %w", err)
	}

	logger.Infof("promotion: found %d clip deals", len(root.All))
	ret := make([]promotion.ClipDeal, 0, len(root.All))

	status := map[string]int{
		"C": 0,
		"U": 0,
	}
	for _, offer := range root.All {
		status[offer.Status]++
		ret = append(ret, offer.convert())
	}
	keys := maps.Keys(status)
	slices.Sort(keys)
	for _, key := range keys {
		logger.Infof("promotion:   - %s: %d", key, status[key])
	}
	return ret, nil
}

// ClipDeal clips a deal for the current user.
func (ps *promotionService) ClipDeal(ctx context.Context, cd promotion.ClipDeal) error {
	if cd.ID == "" {
		return fmt.Errorf("promotion[%s]: clip deal missing id", cd.ID)
	}
	if cd.PromoCode == nil || cd.PromoType == nil {
		return fmt.Errorf("promotion[%s]: clip deal missing promo code or type", cd.ID)
	}
	if cd.IsClipped {
		return fmt.Errorf("promotion[%s]: clip deal already clipped", cd.ID)
	}
	if cd.IsClippable {
		return fmt.Errorf("promotion[%s]: clip deal is not clippable", cd.ID)
	}
	if cd.IsDeleted {
		return fmt.Errorf("promotion[%s]: clip deal is deleted", cd.ID)
	}

	body, err := json.Marshal(newClipDeal(*cd.PromoCode, *cd.PromoType))
	if err != nil {
		return fmt.Errorf("promotion[%s]: clip deal failed to marshal, error %w", cd.ID, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf(PROMOTIONS_CLIP_DEALS_URL, ps.storeID), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("promotion[%s]: clip deal request, error %w", cd.ID, err)
	}

	for k, v := range promotionsExtraHeaders {
		req.Header.Set(k, v)
	}
	req.Header.Set("storeid", ps.storeID)
	req.Header.Set("x-swy_api_key", "appandroid")

	res, err := ps.client.Do(req)
	if err != nil {
		return fmt.Errorf("promotion[%s]: clip deal response, error %w", cd.ID, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("promotion[%s]: clip deal response, error status %s", cd.ID, res.Status)
	}

	cdres := ClipDealRoot{}
	if err := json.NewDecoder(res.Body).Decode(&cdres); err != nil {
		return fmt.Errorf("promotion[%s]: clip deal response, error decoding %w", cd.ID, err)
	}

	logger.Infof("promotion[%s]: clipped deal", cd.ID)
	logger.V(1).Infof("promotion[%s]: clipped deal response %+v", cd.ID, cdres)
	return nil
}

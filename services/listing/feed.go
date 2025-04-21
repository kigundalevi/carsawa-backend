package listing

import (
	"carsawa/models"
	"context"
	"sort"
	"strings"
	"time"
)

const (
	defaultListingLimit = 20
	promotionLimit      = 3
	bannerLimit         = 2
)

func (s *listingService) GetFeed(ctx context.Context, filter models.ListingFilter, pagination models.Pagination) (*models.FeedResponse, error) {
	if pagination.Limit == 0 {
		pagination.Limit = defaultListingLimit
	}

	// Fetch content concurrently
	var (
		listings   []models.Listing
		promotions []models.Promotion
		banners    []models.Banner
		errs       = make(chan error, 3)
	)

	go func() {
		var err error
		listings, err = s.repo.GetActiveListings(ctx, filter, pagination)
		errs <- err
	}()

	go func() {
		var err error
		promotions, err = s.repo.GetPromotions(ctx)
		errs <- err
	}()

	go func() {
		var err error
		banners, err = s.repo.GetBanners(ctx)
		errs <- err
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		if err := <-errs; err != nil {
			return nil, err
		}
	}

	// Apply business logic to prioritize content
	prioritizedListings := prioritizeListings(listings)
	currentPromotions := filterActivePromotions(promotions)
	positionedBanners := positionBanners(banners)

	return &models.FeedResponse{
		Listings:   prioritizedListings,
		Promotions: currentPromotions,
		Banners:    positionedBanners,
	}, nil
}

func prioritizeListings(listings []models.Listing) []models.Listing {
	// Sort by views and freshness
	sort.Slice(listings, func(i, j int) bool {
		if listings[i].Views != listings[j].Views {
			return listings[i].Views > listings[j].Views
		}
		return listings[i].CreatedAt.After(listings[j].CreatedAt)
	})
	return listings
}

func filterActivePromotions(promotions []models.Promotion) []models.Promotion {
	var active []models.Promotion
	now := time.Now()
	for _, p := range promotions {
		if now.After(p.DisplayFrom) && now.Before(p.DisplayTo) {
			active = append(active, p)
		}
	}
	if len(active) > promotionLimit {
		return active[:promotionLimit]
	}
	return active
}

func positionBanners(banners []models.Banner) []models.Banner {
	var result []models.Banner
	for _, b := range banners {
		if b.Position == "top" || b.Position == "bottom" {
			result = append(result, b)
		}
	}
	if len(result) > bannerLimit {
		return result[:bannerLimit]
	}
	return result
}

func (s *listingService) Search(ctx context.Context, query string, filter models.ListingFilter, pagination models.Pagination) (*models.SearchResult, error) {
	// Record search for analytics
	go func() {
		_ = s.repo.RecordSearchQuery(context.Background(), query, filter)
	}()

	var (
		listings []models.Listing
		err      error
	)

	if strings.TrimSpace(query) != "" {
		listings, err = s.repo.TextSearch(ctx, query, pagination)
	} else {
		listings, err = s.repo.GetActiveListings(ctx, filter, pagination)
	}

	if err != nil {
		return nil, err
	}

	suggestions, _ := s.repo.GetSearchSuggestions(ctx, query)

	return &models.SearchResult{
		Listings:    listings,
		Suggestions: suggestions,
	}, nil
}

func (s *listingService) GetSearchSuggestions(ctx context.Context, query string) ([]models.SearchSuggestion, error) {
	return s.repo.GetSearchSuggestions(ctx, query)
}

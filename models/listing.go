// models/listing.go
package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ListingType string

const (
	ListingTypeDealer  ListingType = "dealer"
	ListingTypeUserBid ListingType = "user_bid"
)

type ListingStatus string

const (
	ListingStatusDraft    ListingStatus = "draft"
	ListingStatusActive   ListingStatus = "active"
	ListingStatusOpen     ListingStatus = "open"
	ListingStatusAccepted ListingStatus = "accepted"
	ListingStatusClosed   ListingStatus = "closed"
	ListingStatusSold     ListingStatus = "sold"
)

type Listing struct {
	ID            primitive.ObjectID `bson:"_id" json:"id"`
	Type          ListingType        `bson:"type" json:"type"`
	CarDetails    CarDetails         `bson:"carDetails" json:"carDetails" binding:"required"`
	Status        ListingStatus      `bson:"status" json:"status"`
	Views         int64              `bson:"views" json:"views"`
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt" json:"updatedAt"`
	UserListing   UserListing        `bson:"userListing" json:"userListing,omitzero"`
	DealerListing DealerListing      `bson:"dealerListing" json:"dealerListing,omitzero"`
}

type CarDetails struct {
	VIN   string  `bson:"vin" json:"vin"`
	Make  string  `bson:"make" json:"make"`
	Model string  `bson:"model" json:"model"`
	Year  int     `bson:"year" json:"year"`
	Price float64 `bson:"price,omitempty" json:"price,omitempty"`
}

type DealerListing struct {
	DealerID  primitive.ObjectID `bson:"dealerId,omitempty" json:"dealerId,omitempty"`
	Interests []InterestedUser   `bson:"interestedUser" json:"interestedUser,omitzero"`
}

type InterestedUser struct {
	UserId    string    `bson:"userId" json:"userId" binding:"required"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}

type UserListing struct {
	UserID      primitive.ObjectID  `bson:"userId,omitempty" json:"userId,omitempty"`
	Bids        []Bid               `bson:"bids,omitempty" json:"bids,omitempty"`
	AcceptedBid *primitive.ObjectID `bson:"acceptedBid,omitempty" json:"acceptedBid,omitempty"`
}

type Bid struct {
	ID        primitive.ObjectID `bson:"_id" json:"id"`
	DealerID  primitive.ObjectID `bson:"dealerId" json:"dealerId"`
	Offer     float64            `bson:"offer" json:"offer"`
	Message   string             `bson:"message" json:"message"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type ListingFilter struct {
	Type     ListingType `json:"type"`
	Make     string      `json:"make"`
	Model    string      `json:"model"`
	MinYear  int         `json:"minYear"`
	MaxYear  int         `json:"maxYear"`
	MinPrice float64     `json:"minPrice"`
	MaxPrice float64     `json:"maxPrice"`
}

type Pagination struct {
	Limit  int `json:"limit"`  // No of items per page
	Offset int `json:"offset"` // No of items to skip
}

type ListingPerformance struct {
	ListingID     string `bson:"listingId" json:"listingId"`
	Views         int    `bson:"views" json:"views"`
	ContactClicks int    `bson:"contactClicks" json:"contactClicks"`
}

type FeedItem struct {
	Type        string      `json:"type"` // "listing", "promotion", "banner"
	Content     interface{} `json:"content"`
	Priority    int         `json:"priority"`
	DisplayFrom time.Time   `json:"displayFrom"`
	DisplayTo   time.Time   `json:"displayTo"`
}

type FeedResponse struct {
	Listings   []Listing   `json:"listings"`
	Promotions []Promotion `json:"promotions,omitempty"`
	Banners    []Banner    `json:"banners,omitempty"`
}

type Promotion struct {
	ID                   primitive.ObjectID `bson:"_id" json:"id"`
	Title                string             `bson:"title" json:"title"`
	Description          string             `bson:"description" json:"description"`
	ImageURL             string             `bson:"imageUrl" json:"imageUrl"`
	TargetURL            string             `bson:"targetUrl" json:"targetUrl"`
	DisplayFrom          time.Time          `bson:"displayFrom" json:"displayFrom"`
	DisplayTo            time.Time          `bson:"displayTo" json:"displayTo"`
	DiscountType         string             `bson:"discountType" json:"discountType"` // "percentage", "fixed", "trade-in"
	DiscountValue        float64            `bson:"discountValue" json:"discountValue"`
	ApplicableToListings []string           `bson:"applicableToListings" json:"applicableToListings"` // Listing IDs
	ValidUntil           time.Time          `bson:"validUntil" json:"validUntil"`
}

type Banner struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	ImageURL    string             `bson:"imageUrl" json:"imageUrl"`
	TargetURL   string             `bson:"targetUrl" json:"targetUrl"`
	Position    string             `bson:"position" json:"position"` // "top", "middle", "bottom"
	Title       string             `bson:"title" json:"title,omitempty"`
	Description string             `bson:"description" json:"description,omitempty"`
	CTA         string             `bson:"cta" json:"cta,omitempty"` // Call-to-action link
	IsActive    bool               `bson:"isActive" json:"isActive"`
	StartDate   time.Time          `bson:"startDate" json:"startDate"`
	EndDate     time.Time          `bson:"endDate" json:"endDate"`
}

type SearchResult struct {
	Listings    []Listing          `json:"listings"`
	Promotions  []Promotion        `json:"promotions,omitempty"`
	Suggestions []SearchSuggestion `json:"suggestions,omitempty"`
}

type SearchSuggestion struct {
	Text  string  `json:"text"`
	Type  string  `json:"type"` // "make", "model", "keyword"
	Score float64 `json:"score"`
}

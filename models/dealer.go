package models

import (
	"time"
)

type Dealer struct {
	ID           string        `bson:"_id" json:"id"`
	Profile      DealerProfile `bson:"profile" json:"profile"`
	Store        Store         `bson:"store" json:"store"`
	Security     Security      `bson:"security" json:"-"`
	Verification Verification  `bson:"verification" json:"verification"`
	CreatedAt    time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time     `bson:"updatedAt" json:"updatedAt"`
	Devices      []Device      `bson:"devices" json:"devices"`
}

type Store struct {
	Banners        []Banner         `bson:"banners" json:"banners"`
	ServiceCatalog ServiceCatalogue `bson:"serviceCatalog" json:"serviceCatalog"`
	Promotions     []Promotion      `bson:"promotions" json:"promotions"`
	ListingIDs     []string         `bson:"listingIds" json:"listingIds"` // Reference to listings
	Branding       StoreBranding    `bson:"branding" json:"branding"`
}

// DealerProfile contains public-facing dealer metadata
type DealerProfile struct {
	DealerName string   `bson:"dealerName" json:"dealerName" binding:"required"`
	Slug       string   `bson:"slug" json:"slug"` // Unique URL identifier (e.g., "prestige-motors-nairobi")
	Contact    Contact  `bson:"contact" json:"contact"`
	Location   Location `bson:"location" json:"location"`
	Rating     float64  `bson:"rating" json:"rating,omitempty"`
}

// ServiceCatalog lists dealer services (e.g., financing, repairs)
type ServiceCatalogue struct {
	Services []Service `bson:"services" json:"services"`
}

type Service struct {
	Name        string `bson:"name" json:"name" binding:"required"` // e.g., "7-Year Warranty"
	Description string `bson:"description" json:"description"`
	IconURL     string `bson:"iconUrl" json:"iconUrl,omitempty"` // Optional service icon
	IsFeatured  bool   `bson:"isFeatured" json:"isFeatured"`     // Highlight on storefront
}

// Analytics tracks storefront performance over time
type Analytics struct {
	PeriodStart time.Time            `bson:"periodStart" json:"periodStart"` // Weekly/Monthly aggregation
	PeriodEnd   time.Time            `bson:"periodEnd" json:"periodEnd"`
	Metrics     AnalyticsMetrics     `bson:"metrics" json:"metrics"`
	TopListings []ListingPerformance `bson:"topListings" json:"topListings"`   // Best-performing listings
	Reviews     []Review             `bson:"reviews" json:"reviews,omitempty"` // Customer feedback
}

type AnalyticsMetrics struct {
	TotalImpressions int     `bson:"totalImpressions" json:"totalImpressions"` // Storefront views
	UniqueVisitors   int     `bson:"uniqueVisitors" json:"uniqueVisitors"`
	ContactClicks    int     `bson:"contactClicks" json:"contactClicks"`   // "Contact Dealer" actions
	ListingViews     int     `bson:"listingViews" json:"listingViews"`     // Individual car views
	EngagementRate   float64 `bson:"engagementRate" json:"engagementRate"` // (Clicks + Views)/Impressions
	AvgResponseTime  int     `bson:"avgResponseTime" json:"avgResponseTime"`
}

// Review captures customer feedback tied to a period
type Review struct {
	Rating     int       `bson:"rating" json:"rating" binding:"required,min=1,max=5"`
	Comment    string    `bson:"comment" json:"comment"`
	ListingID  string    `bson:"listingId" json:"listingId,omitempty"` // Optional association
	ReviewedBy string    `bson:"reviewedBy" json:"reviewedBy"`         // User ID
	CreatedAt  time.Time `bson:"createdAt" json:"createdAt"`
}

// Security handles authentication credentials
type Security struct {
	PasswordHash string `bson:"passwordHash" json:"-"`
	TokenHash    string `bson:"tokenHash" json:"-"`
}

// Verification tracks KYC status
type Verification struct {
	Level      string    `bson:"level" json:"level"`                   // "basic", "advanced"
	Documents  []string  `bson:"documents" json:"documents,omitempty"` // PDF/image URLs
	VerifiedAt time.Time `bson:"verifiedAt" json:"verifiedAt,omitempty"`
	Status     string    `bson:"status" json:"status"` // "pending", "approved"
}

// Location with geospatial data
type Location struct {
	Address  string   `bson:"address" json:"address" binding:"required"`
	City     string   `bson:"city" json:"city" binding:"required"`
	GeoPoint GeoPoint `bson:"geoPoint" json:"geoPoint"` // For location-based feeds
}

// GeoPoint for MongoDB geospatial queries
type GeoPoint struct {
	Type        string    `bson:"type" json:"type" default:"Point"`
	Coordinates []float64 `bson:"coordinates" json:"coordinates"` // [longitude, latitude]
}

// Contact information
type Contact struct {
	Email    string `bson:"email" json:"email" binding:"required,email"`
	Phone    string `bson:"phone" json:"phone" binding:"required"`
	WhatsApp string `bson:"whatsapp" json:"whatsapp,omitempty"` // Common in Kenya
}

// StoreBranding defines visual identity
type StoreBranding struct {
	PrimaryColor   string `bson:"primaryColor" json:"primaryColor"` // Hex code
	SecondaryColor string `bson:"secondaryColor" json:"secondaryColor"`
	FontFamily     string `bson:"fontFamily" json:"fontFamily,omitempty"`
}

type ContactAttempt struct {
	ID        string    `bson:"_id" json:"id"`
	DealerID  string    `bson:"dealerId" json:"dealerId"`
	ListingID string    `bson:"listingId" json:"listingId"`
	UserID    string    `bson:"userId" json:"userId"`
	Channel   string    `bson:"channel" json:"channel"` // "whatsapp", "call", "email"
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}

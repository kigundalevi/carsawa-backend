package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationTarget string
type NotificationType string

const (
	// Target is WHO the notification is sent to (audience)
	NotificationTargetUser   NotificationTarget = "user"
	NotificationTargetDealer NotificationTarget = "dealer"

	// Type is WHAT the notification is about (action)
	NotificationTypeBidPlaced        NotificationType = "bid_placed"
	NotificationTypeBidAccepted      NotificationType = "bid_accepted"
	NotificationTypeListingCreated   NotificationType = "listing_created"
	NotificationTypeListingUpdated   NotificationType = "listing_updated"
	NotificationTypeListingPublished NotificationType = "listing_published"
	NotificationTypeListingClosed    NotificationType = "listing_closed"
)

type Notification struct {
	ID        primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Recipient primitive.ObjectID     `bson:"recipient" json:"recipient"` // user or dealer ID
	Target    NotificationTarget     `bson:"target" json:"target"`       // "user" or "dealer"
	Type      NotificationType       `bson:"type" json:"type"`           // bid_placed, bid_accepted, etc.
	Title     string                 `bson:"title" json:"title"`
	Body      string                 `bson:"body" json:"body"`
	Data      map[string]interface{} `bson:"data,omitempty" json:"data,omitempty"` // listing, user, dealer, bid, etc.
	Read      bool                   `bson:"read" json:"read"`
	CreatedAt time.Time              `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time              `bson:"updatedAt" json:"updatedAt"`
}

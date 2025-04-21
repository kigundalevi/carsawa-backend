package listing

import (
	"carsawa/models"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *listingService) notifyUser(
	ctx context.Context,
	userID primitive.ObjectID,
	ntype models.NotificationType,
	title, body string,
	data map[string]interface{},
) {
	if err := s.notifier.CreateUserNotification(
		ctx,
		userID.Hex(),
		ntype,
		title,
		body,
		data,
	); err != nil {
		// log error, but don’t fail business logic
		fmt.Printf("notifyUser error: %v\n", err)
	}
}

func (s *listingService) notifyDealer(
	ctx context.Context,
	dealerID primitive.ObjectID,
	ntype models.NotificationType,
	title, body string,
	data map[string]interface{},
) {
	if err := s.notifier.CreateDealerNotification(
		ctx,
		dealerID.Hex(),
		ntype,
		title,
		body,
		data,
	); err != nil {
		fmt.Printf("notifyDealer error: %v\n", err)
	}
}

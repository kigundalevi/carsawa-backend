package dealerRepo

import (
	"carsawa/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetDealerByIDWithProjection retrieves a dealer by its ID using a field projection.
func (r *mongoDealerRepo) GetDealerByIDWithProjection(id string, projection bson.M) (*models.Dealer, error) {
	var dealer models.Dealer
	err := r.collection.
		FindOne(r.ctx, bson.M{"_id": id}, options.FindOne().SetProjection(projection)).
		Decode(&dealer)
	return &dealer, err
}

// GetDealerByEmail retrieves a dealer by its email.
func (r *mongoDealerRepo) GetDealerByEmail(email string) (*models.Dealer, error) {
	var dealer models.Dealer
	err := r.collection.FindOne(r.ctx, bson.M{"email": email}).Decode(&dealer)
	return &dealer, err
}

// GetDealerByEmailWithProjection retrieves a dealer by its email using a field projection.
func (r *mongoDealerRepo) GetDealerByEmailWithProjection(email string, projection bson.M) (*models.Dealer, error) {
	var dealer models.Dealer
	err := r.collection.
		FindOne(r.ctx, bson.M{"email": email}, options.FindOne().SetProjection(projection)).
		Decode(&dealer)
	return &dealer, err
}

// GetAllDealersWithProjection retrieves all dealer documents using a specified projection.
func (r *mongoDealerRepo) GetAllDealersWithProjection(projection bson.M) ([]models.Dealer, error) {
	cursor, err := r.collection.Find(r.ctx, bson.M{}, options.Find().SetProjection(projection))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(r.ctx)

	var dealers []models.Dealer
	for cursor.Next(r.ctx) {
		var dealer models.Dealer
		if err := cursor.Decode(&dealer); err != nil {
			continue
		}
		dealers = append(dealers, dealer)
	}
	return dealers, nil
}

// GetDealerBySlug retrieves a dealer by its public slug.
func (r *mongoDealerRepo) GetDealerBySlug(slug string) (*models.Dealer, error) {
	var dealer models.Dealer
	err := r.collection.FindOne(r.ctx, bson.M{"slug": slug}).Decode(&dealer)
	return &dealer, err
}

// IsProviderAvailable checks if a dealer with the given email or username already exists.
func (r *mongoDealerRepo) IsDealerAvailable(basicReq models.DealerBasicRegistrationData) (bool, error) {
	ctx, cancel := newContext(5 * time.Second)
	defer cancel()

	// Adjusted filter: check within profile for email and dealerName.
	filter := bson.M{
		"$or": []bson.M{
			{"profile.email": basicReq.Email},
			{"profile.dealerName": basicReq.DealerName},
		},
	}

	var dealer models.Dealer
	err := r.collection.FindOne(ctx, filter).Decode(&dealer)
	if err != nil {
		// If no document is found, then dealer unAvailable.
		if err.Error() == "mongo: no documents in result" {
			return false, nil
		}
		return true, err
	}
	// Document found – dealer available
	return true, nil
}

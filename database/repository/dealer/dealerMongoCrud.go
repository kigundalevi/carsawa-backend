// File: dealerCrud.go
package dealerRepo

import (
	"carsawa/models"
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoDealerRepo struct {
	collection *mongo.Collection
	ctx        context.Context
}

// NewMongoDealerRepo creates a new instance of mongoDealerRepo.
func NewMongoDealerRepo(db *mongo.Database) DealerRepository {
	return &mongoDealerRepo{
		collection: db.Collection("dealers"),
		ctx:        context.Background(),
	}
}

// CreateDealer inserts a new dealer document.
func (r *mongoDealerRepo) CreateDealer(dealer *models.Dealer) error {
	_, err := r.collection.InsertOne(r.ctx, dealer)
	return err
}

// UpdateDealer performs a partial update on the dealer document identified by id.
// The updateDoc should be a bson.M containing the fields to update.
func (r *mongoDealerRepo) UpdateDealer(id string, updateDoc bson.M) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": updateDoc}
	_, err := r.collection.UpdateOne(r.ctx, filter, update)
	return err
}

// DeleteDealer removes a dealer document by its ID.
func (r *mongoDealerRepo) DeleteDealer(id string) error {
	res, err := r.collection.DeleteOne(r.ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("dealer not found")
	}
	return nil
}

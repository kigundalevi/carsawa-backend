package listingRepo

import (
	"carsawa/models"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (r *MongoListingsRepository) GetActiveListings(ctx context.Context, filter models.ListingFilter, pagination models.Pagination) ([]models.Listing, error) {
	query := bson.M{
		"status": bson.M{"$in": []models.ListingStatus{
			models.ListingStatusActive,
			models.ListingStatusOpen,
		}},
	}

	// Apply filters
	if filter.Make != "" {
		query["carDetails.make"] = filter.Make
	}
	if filter.Model != "" {
		query["carDetails.model"] = filter.Model
	}
	if filter.MinYear > 0 {
		query["carDetails.year"] = bson.M{"$gte": filter.MinYear}
	}
	if filter.MaxYear > 0 {
		if _, ok := query["carDetails.year"]; ok {
			query["carDetails.year"] = bson.M{"$gte": filter.MinYear, "$lte": filter.MaxYear}
		} else {
			query["carDetails.year"] = bson.M{"$lte": filter.MaxYear}
		}
	}

	opts := options.Find().
		SetLimit(int64(pagination.Limit)).
		SetSkip(int64(pagination.Offset)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := r.listings.Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}

	var results []models.Listing
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *MongoListingsRepository) TextSearch(ctx context.Context, query string, pagination models.Pagination) ([]models.Listing, error) {
	filter := bson.M{
		"$text": bson.M{"$search": query},
		"status": bson.M{"$in": []models.ListingStatus{
			models.ListingStatusActive,
			models.ListingStatusOpen,
		}},
	}

	opts := options.Find().
		SetLimit(int64(pagination.Limit)).
		SetSkip(int64(pagination.Offset)).
		SetSort(bson.D{{Key: "score", Value: bson.M{"$meta": "textScore"}}})

	cursor, err := r.listings.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}

	var results []models.Listing
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *MongoListingsRepository) GetSearchSuggestions(ctx context.Context, query string) ([]models.SearchSuggestion, error) {
	pipeline := mongo.Pipeline{
		{
			{Key: "$match", Value: bson.M{
				"$text": bson.M{"$search": query},
			}},
		},
		{
			{Key: "$project", Value: bson.M{
				"make":  "$carDetails.make",
				"model": "$carDetails.model",
			}},
		},
		{
			{Key: "$group", Value: bson.M{
				"_id":    nil,
				"makes":  bson.M{"$addToSet": "$make"},
				"models": bson.M{"$addToSet": "$model"},
			}},
		},
	}

	cursor, err := r.listings.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}

	var result struct {
		Makes  []string `bson:"makes"`
		Models []string `bson:"models"`
	}
	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
	}

	var suggestions []models.SearchSuggestion
	for _, make := range result.Makes {
		suggestions = append(suggestions, models.SearchSuggestion{
			Text:  make,
			Type:  "make",
			Score: 1.0,
		})
	}
	for _, model := range result.Models {
		suggestions = append(suggestions, models.SearchSuggestion{
			Text:  model,
			Type:  "model",
			Score: 0.8,
		})
	}

	return suggestions, nil
}

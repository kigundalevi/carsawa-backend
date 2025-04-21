package repository

import (
	dealerRepo "carsawa/database/repository/dealer"
	listingRepo "carsawa/database/repository/listing"
	userRepo "carsawa/database/repository/user"
)

// Re-export the DealerRepository interface and constructors.
type DealerRepository = dealerRepo.DealerRepository

var NewMongoDealerRepo = dealerRepo.NewMongoDealerRepo

// Re-export the UserRepository interface and constructor.
type UserRepository = userRepo.UserRepository

var NewMongoUserRepository = userRepo.NewMongoUserRepo

// Re-export the ListingsRepository interface and constructor.
type ListingsRepository = listingRepo.ListingRepository

var NewMongoListingsRepo = listingRepo.NewMongoListingsRepository

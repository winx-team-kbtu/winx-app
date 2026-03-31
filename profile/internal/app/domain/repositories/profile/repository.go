package profile

import (
	"context"
	"errors"
	"fmt"

	"winx-profile/internal/app/models/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	mongodriver "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var ErrNotFound = errors.New("profile not found")

type Repository interface {
	EnsureIndexes(ctx context.Context) error
	GetByUserID(ctx context.Context, userID int64) (models.Profile, error)
	Create(ctx context.Context, profile models.Profile) (models.Profile, error)
	Update(ctx context.Context, profile models.Profile) (models.Profile, error)
	DeleteByUserID(ctx context.Context, userID int64) error
}

type repository struct {
	collection *mongodriver.Collection
}

func NewRepository(db *mongodriver.Database) Repository {
	return &repository{collection: db.Collection("profiles")}
}

func (r *repository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongodriver.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("create profiles user_id index: %w", err)
	}

	return nil
}

func (r *repository) GetByUserID(ctx context.Context, userID int64) (models.Profile, error) {
	var profile models.Profile

	err := r.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&profile)
	if err != nil {
		if errors.Is(err, mongodriver.ErrNoDocuments) {
			return models.Profile{}, ErrNotFound
		}

		return models.Profile{}, fmt.Errorf("get profile by user id: %w", err)
	}

	return profile, nil
}

func (r *repository) Create(ctx context.Context, profile models.Profile) (models.Profile, error) {
	res, err := r.collection.InsertOne(ctx, profile)
	if err != nil {
		return models.Profile{}, fmt.Errorf("create profile: %w", err)
	}

	if oid, ok := res.InsertedID.(bson.ObjectID); ok {
		profile.ID = oid
	}

	return profile, nil
}

func (r *repository) Update(ctx context.Context, profile models.Profile) (models.Profile, error) {
	res, err := r.collection.ReplaceOne(ctx, bson.M{"user_id": profile.UserID}, profile)
	if err != nil {
		return models.Profile{}, fmt.Errorf("update profile: %w", err)
	}

	if res.MatchedCount == 0 {
		return models.Profile{}, ErrNotFound
	}

	return profile, nil
}

func (r *repository) DeleteByUserID(ctx context.Context, userID int64) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"user_id": userID})
	if err != nil {
		return fmt.Errorf("delete profile: %w", err)
	}

	return nil
}

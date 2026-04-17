package profilemeta

import (
	"context"
	"fmt"

	"winx-recommendation/internal/app/models/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	mongodriver "go.mongodb.org/mongo-driver/v2/mongo"
)

type Repository interface {
	// GetByUserID fetches one profile from MongoDB.
	GetByUserID(ctx context.Context, userID int64) (models.ProfileMeta, error)
	// GetByUserIDs batch-fetches profiles from MongoDB and returns a map keyed by user_id.
	GetByUserIDs(ctx context.Context, userIDs []int64) (map[int64]models.ProfileMeta, error)
}

type repository struct {
	collection *mongodriver.Collection
}

func NewRepository(db *mongodriver.Database) Repository {
	return &repository{collection: db.Collection("profiles")}
}

func (r *repository) GetByUserID(ctx context.Context, userID int64) (models.ProfileMeta, error) {
	var meta models.ProfileMeta
	err := r.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&meta)
	if err != nil {
		return models.ProfileMeta{}, fmt.Errorf("get profile meta by user_id=%d: %w", userID, err)
	}
	return meta, nil
}

func (r *repository) GetByUserIDs(ctx context.Context, userIDs []int64) (map[int64]models.ProfileMeta, error) {
	if len(userIDs) == 0 {
		return map[int64]models.ProfileMeta{}, nil
	}

	cursor, err := r.collection.Find(ctx, bson.M{
		"user_id": bson.M{"$in": userIDs},
	})
	if err != nil {
		return nil, fmt.Errorf("batch fetch profile metas: %w", err)
	}
	defer cursor.Close(ctx)

	result := make(map[int64]models.ProfileMeta, len(userIDs))
	for cursor.Next(ctx) {
		var meta models.ProfileMeta
		if err := cursor.Decode(&meta); err != nil {
			continue
		}
		result[meta.UserID] = meta
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("iterate profile metas cursor: %w", err)
	}

	return result, nil
}

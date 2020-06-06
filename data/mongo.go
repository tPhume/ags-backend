package data

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoRepo struct {
	Col *mongo.Collection
}

func (m *MongoRepo) GetData(ctx context.Context, entity *Entity) error {
	result := m.Col.FindOne(ctx, bson.M{
		"_id":     entity.ControllerId,
		"user_id": entity.UserId,
	})

	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return notFound
		}

		return result.Err()
	}

	if err := result.Decode(entity); err != nil {
		return err
	}

	return nil
}

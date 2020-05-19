package summary

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Mongo struct {
	Col *mongo.Collection
}

func (m *Mongo) ListSummary(ctx context.Context, userId string, controllerId string) ([]*Summary, error) {
	cursor, err := m.Col.Find(ctx, bson.M{"user_id": userId, "controller_id": controllerId})
	if err != nil {
		return nil, err
	}

	entities := make([]*Summary, 0)
	for cursor.Next(ctx) {
		result := &Summary{}
		if err := cursor.Decode(result); err != nil {
			return nil, err
		}

		entities = append(entities, result)
	}

	return entities, nil
}

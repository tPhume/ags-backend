package plan

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoRepo struct {
	Col *mongo.Collection
}

func (m MongoRepo) CreatePlan(ctx context.Context, entity *Entity) error {
	if _, err := m.Col.InsertOne(ctx, entity); err != nil {
		writeException, ok := err.(mongo.WriteException)
		if !ok {
			return err
		} else if len(writeException.WriteErrors) == 0 {
			return err
		} else if writeException.WriteErrors[0].Code == 11000 {
			return errPlanDuplicate
		}

		return err
	}

	return nil
}

func (m MongoRepo) ListPlans(ctx context.Context, userId string) ([]*Entity, error) {
	cursor, err := m.Col.Find(ctx, bson.M{"userId": userId})
	if err != nil {
		return nil, err
	}

	var entities []*Entity
	if err := cursor.All(ctx, &entities); err != nil {
		return nil, err
	}

	return entities, nil
}

func (m MongoRepo) GetPlan(ctx context.Context, entity *Entity) error {
	panic("implement me")
}

func (m MongoRepo) ReplacePlan(ctx context.Context, entity *Entity) error {
	panic("implement me")
}

func (m MongoRepo) DeletePlan(ctx context.Context, userId string, planId string) error {
	panic("implement me")
}

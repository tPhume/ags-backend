package plan

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoRepo struct {
	Col *mongo.Collection
}

func (m MongoRepo) CreatePlan(ctx context.Context, entity *Entity) error {
	panic("implement me")
}

func (m MongoRepo) ListPlans(ctx context.Context, userId string) ([]*Entity, error) {
	panic("implement me")
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

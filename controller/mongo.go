package controller

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoRepo struct {
	col *mongo.Collection
}

func (m MongoRepo) AddController(ctx context.Context, entity *Entity) error {
	if _, err := m.col.InsertOne(ctx, bson.M{
		"_id":    entity.ControllerId,
		"userId": entity.UserId,
		"name":   entity.Name,
		"desc":   entity.Desc,
		"plan":   entity.Plan,
	}); err != nil {
		writeException, ok := err.(mongo.WriteException)
		if !ok {
			return err
		}

		if len(writeException.WriteErrors) == 0 {
			return err
		}

		if writeException.WriteErrors[0].Code == 11000 {
			return duplicateName
		}
	}

	return nil
}

func (m MongoRepo) ListControllers(string) ([]*Entity, error) {
	panic("implement me")
}

func (m MongoRepo) GetController(*Entity) error {
	panic("implement me")
}

func (m MongoRepo) UpdateController(*Entity) error {
	panic("implement me")
}

func (m MongoRepo) RemoveController(string, string) error {
	panic("implement me")
}

func (m MongoRepo) GenerateToken(string, string, string) error {
	panic("implement me")
}

func (m MongoRepo) VerifyToken(string, string, string) error {
	panic("implement me")
}

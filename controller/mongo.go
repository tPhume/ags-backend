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

func (m MongoRepo) ListControllers(ctx context.Context, userId string) ([]*Entity, error) {
	cursor, err := m.col.Find(ctx, bson.M{"userId": userId})
	if err != nil {
		return nil, err
	}

	entities := make([]*Entity, 0)

	for cursor.Next(ctx) {
		result := &Result{}
		if err := cursor.Decode(result); err != nil {
			return nil, err
		}

		entities = append(entities, &Entity{
			ControllerId: result.ControllerId,
			Name:         result.Name,
			Desc:         result.Desc,
			Plan:         result.Plan,
		})
	}

	return entities, nil
}

func (m MongoRepo) GetController(ctx context.Context, entity *Entity) error {
	result := m.col.FindOne(ctx, bson.M{
		"_id":    entity.ControllerId,
		"userId": entity.UserId,
	})

	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return controllerNotFound
		}

		return result.Err()
	}

	resultBody := &Result{}
	if err := result.Decode(resultBody); err != nil {
		return err
	}

	entity.Name = resultBody.Name
	entity.Desc = resultBody.Desc
	entity.Plan = resultBody.Plan

	return nil
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

type Result struct {
	ControllerId string `json:"_id"`
	Name         string `json:"name"`
	Desc         string `json:"desc"`
	Plan         string `json:"plan"`
}

package session

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoRepo struct {
	userDb    *mongo.Collection
	sessionDb *mongo.Collection
}

func (m *MongoRepo) CreateSession(ctx context.Context, userEntity *UserEntity, sessionId string) error {
	return nil
}

func (m *MongoRepo) DeleteSession(ctx context.Context, sessionId string) error {
	return nil
}

func (m *MongoRepo) GetUser(ctx context.Context, sessionId string) (string, error) {
	return "", nil
}

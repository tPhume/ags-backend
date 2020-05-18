package session

import (
	"context"
	"github.com/go-redis/redis/v7"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type RedisMongo struct {
	UserDb    *mongo.Collection
	SessionDb *redis.Client
}

func (r *RedisMongo) CreateSession(ctx context.Context, userEntity *UserEntity, sessionId string) error {
	res := r.UserDb.FindOne(ctx, bson.M{"name": userEntity.Name, "password": userEntity.Password})
	if res.Err() != nil {
		if res.Err() == mongo.ErrNoDocuments {
			return errUserDoesNotExist
		}

		return res.Err()
	}

	// Create new session
	if err := r.SessionDb.Set(sessionId, userEntity.UserId, time.Hour*8).Err(); err != nil {
		return err
	}

	return nil
}

func (r *RedisMongo) DeleteSession(ctx context.Context, sessionId string) error {
	// Delete session
	if err := r.SessionDb.Del(sessionId).Err(); err != nil {
		return err
	}

	return nil
}

func (r *RedisMongo) GetUser(ctx context.Context, sessionId string) (string, error) {
	res := r.SessionDb.Get(sessionId)
	if res.Err() != nil {
		if res.Err() == redis.Nil {
			return "", errNotFound
		}

		return "", res.Err()
	}

	result, err := res.Result()
	if err != nil {
		return "", err
	}

	return result, nil
}

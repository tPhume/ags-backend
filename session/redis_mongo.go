package session

import (
	"context"
	"github.com/go-redis/redis/v7"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type RedisMongo struct {
	UserDb    *mongo.Collection
	SessionDb *redis.Client
}

func (r *RedisMongo) CreateSession(ctx context.Context, userEntity *UserEntity, sessionId string) error {
	// Upsert user entity
	opt := options.Replace()
	opt.SetUpsert(true)

	if _, err := r.UserDb.ReplaceOne(ctx, bson.M{"_id": userEntity.UserId}, bson.M{
		"_id":            userEntity.UserId,
		"email":          userEntity.Email,
		"email_verified": userEntity.EmailVerified,
		"name":           userEntity.Name,
		"picture":        userEntity.Picture,
	}, opt); err != nil {
		return err
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

package session

import (
	"context"
	"github.com/go-redis/redis/v7"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type RedisMongo struct {
	userDb    *mongo.Collection
	sessionDb *redis.Client
}

func (r *RedisMongo) CreateSession(ctx context.Context, userEntity *UserEntity, sessionId string) error {
	// Upsert user entity
	opt := options.Replace()
	opt.SetUpsert(true)

	if _, err := r.userDb.ReplaceOne(ctx, mapping{"_id": userEntity.UserId}, mapping{
		"email":          userEntity.Email,
		"email_verified": userEntity.EmailVerified,
		"name":           userEntity.Name,
		"picture":        userEntity.Picture,
	}, opt); err != nil {
		return err
	}

	// Create new session
	if err := r.sessionDb.Set(sessionId, userEntity.UserId, time.Hour*10).Err(); err != nil {
		return err
	}

	return nil
}

func (r *RedisMongo) DeleteSession(ctx context.Context, sessionId string) error {
	// Delete session
	if err := r.sessionDb.Del(sessionId).Err(); err != nil {
		return err
	}

	return nil
}

func (r *RedisMongo) GetUser(ctx context.Context, sessionId string) (string, error) {
	res := r.sessionDb.Get(sessionId)
	if res.Err() != nil {
		if res.Err() == redis.Nil {
			return "", errNotFound
		}

		return "", res.Err()
	}

	return res.String(), nil
}

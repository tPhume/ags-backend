package plan

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepo struct {
	Col           *mongo.Collection
	ControllerCol *mongo.Collection
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
	cursor, err := m.Col.Find(ctx, bson.M{"user_id": userId})
	if err != nil {
		return nil, err
	}

	entities := make([]*Entity, 0)
	if err := cursor.All(ctx, &entities); err != nil {
		return nil, err
	}

	return entities, nil
}

func (m MongoRepo) GetPlan(ctx context.Context, entity *Entity) error {
	result := m.Col.FindOne(ctx, bson.M{"_id": entity.PlanId, "user_id": entity.UserId})
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return errPlanNotFound
		}

		return result.Err()
	}

	if err := result.Decode(entity); err != nil {
		return err
	}

	return nil
}

func (m MongoRepo) ReplacePlan(ctx context.Context, entity *Entity) error {
	projection := options.FindOneAndReplace().SetProjection(bson.M{"_id": 1})
	result := m.Col.FindOneAndReplace(ctx, bson.M{"_id": entity.PlanId, "user_id": entity.UserId}, entity, projection)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return errPlanNotFound
		}

		if err, ok := result.Err().(mongo.CommandError); ok {
			if err.Code == 11000 {
				return errPlanDuplicate
			}
		}

		return result.Err()
	}

	return nil
}

func (m MongoRepo) DeletePlan(ctx context.Context, userId string, planId string) error {
	result, err := m.Col.DeleteOne(ctx, bson.M{"_id": planId, "user_id": userId})
	if err != nil {
		return err
	} else if result.DeletedCount == 0 {
		return errPlanNotFound
	} else if result.DeletedCount != 1 {
		return errors.New("should be 1")
	}

	return nil
}

func (m *MongoRepo) GetPlanId(ctx context.Context, token string) (*Entity, error) {
	res := m.ControllerCol.FindOne(ctx, bson.M{"token": token})
	if res.Err() != nil {
		if res.Err() == mongo.ErrNoDocuments {
			return nil, errTokenNotFound
		}

		return nil, res.Err()
	}

	temp := make(map[string]interface{})
	if err := res.Decode(&temp); err != nil {
		return nil, err
	}

	entity := &Entity{
		PlanId: temp["plan_id"].(string),
		UserId: temp["user_id"].(string),
	}

	return entity, nil
}

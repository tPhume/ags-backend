package controller

import "go.mongodb.org/mongo-driver/mongo"

type MongoRepo struct {
	col *mongo.Collection
}

func (m MongoRepo) AddController(*Entity) error {
	panic("implement me")
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




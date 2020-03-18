package controller

import (
	"gopkg.in/go-playground/validator.v9"
	"testing"
)

var controller = Entity{
	ControllerId: "f1d67e51-4ca4-4b25-a4b7-6c8f06822075",
	UserId:       "76de6d55-e457-4070-8aef-5633726d498f",
	Name:         "Controller",
	Desc:         "",
	Plan:         "",
}

// Repo struct for happy path testing
type repoHappy struct{}

func (t repoHappy) AddController(entity *Entity) error {
	return nil
}

func (t repoHappy) ListControllers(string) ([]*Entity, error) {
	return []*Entity{&Entity{Name: "Happy"}}, nil
}

func (t repoHappy) GetController(string) (*Entity, error) {
	return &Entity{Name: "Happy"}, nil
}

func (t repoHappy) UpdateController(*Entity) error {
	return nil
}

func (t *repoHappy) RemoveController(string) error {
	return nil
}

// Test Entity Struct validation
func TestStructValidation(t *testing.T) {
	v := validator.New()
	v.RegisterStructValidation(StructValidation, Entity{})

	if err := v.Struct(controller); err != nil {
		t.Fatal(err)
	}
}

// Test AddControllers handler
func TestHandler_AddController(t *testing.T) {
	
}

// Test ListController handler
func TestHandler_ListControllers(t *testing.T) {

}

// Test GetController handler
func TestHandler_GetController(t *testing.T) {

}

// Test UpdateController handler
func TestHandler_UpdateController(t *testing.T) {

}

// Test RemoveController handler
func TestHandler_RemoveController(t *testing.T) {

}

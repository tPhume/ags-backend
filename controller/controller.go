package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"gopkg.in/go-playground/validator.v9"
	"strings"
)

// Controller Entity type represent edge device
type Entity struct {
	ControllerId string `json:"controller_id" validate:"required"`
	UserId       string `json:"user_id" validate:"required"`
	Name         string `json:"name" validate:"required"`
	Desc         string `json:"desc"`
	Plan         string `json:"plan"`
}

// Custom Controller type struct validation function
func StructValidation(sl validator.StructLevel) {
	entity := sl.Current().Interface().(Entity)

	// checks ControllerId
	if _, err := uuid.Parse(entity.ControllerId); err != nil {
		sl.ReportError(entity.ControllerId, "ControllerId", "ControllerId", "", "")
	}

	// checks UserId
	if _, err := uuid.Parse(entity.UserId); err != nil {
		sl.ReportError(entity.UserId, "UserId", "UserId", "", "")
	}

	// checks name
	if len(strings.TrimSpace(entity.Name)) < 1 {
		sl.ReportError(entity.Name, "name", "name", "", "")
	}
}

// addStructValidation register StructValidation function to Gin's default validator Engine
func addStructValidation(engine *gin.Engine) {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterStructValidation(StructValidation, Entity{})
	}
}

// Controller Repo - interface to communicate with data source
type Repo interface {
	// AddController creates new controller at data source given *Entity type
	// Duplicated Controller entity will result in an error
	AddController(entity *Entity) error

	// ListControllers fetches all controller under the given UserId
	// Return of empty slice does not imply error
	ListControllers(string) ([]*Entity, error)

	// GetController fetches specific controller by given ControllerId
	// Return of nil value for *Entity indicates error
	GetController(string) (*Entity, error)

	// UpdateController does partial update given *Controller type
	// No new controller is created if controller is not found
	UpdateController(*Entity) error

	// RemoveController deletes data from data source given ControllerId
	// Cascade deletion is done asynchronously
	// Missing controller will result in an error
	RemoveController(string) error
}

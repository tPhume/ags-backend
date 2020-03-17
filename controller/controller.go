// Package controller deals with Controller resource in our data source
// Usage outside of this package should only be to register routes for Gin Engine
package controller

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
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

// Contains errors that implementation of Repo should use
var (
	duplicateName      = errors.New("duplicate name")
	controllerNotFound = errors.New("controller not found")
)

// Handler for controller REST API
type Handler struct {
	repo Repo
}

var (
	keyNotFound = errors.New("key not found")
	castingFail = errors.New("casting fail")
)

func (h *Handler) AddController(ctx *gin.Context) {
	userId, err := getUserId(ctx)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	entity := &Entity{
		ControllerId: uuid.New().String(),
		UserId:       userId,
	}

	if err = ctx.ShouldBindJSON(entity); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid values"})
		return
	}

	if err = h.repo.AddController(entity); err != nil {
		if err == duplicateName {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "duplicate name"})
		} else {
			ctx.Status(http.StatusInternalServerError)
		}

		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "controller added", "controller_id": entity.ControllerId})
}

func (h *Handler) ListControllers(ctx *gin.Context) {
	userId, err := getUserId(ctx)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	entityList, err := h.repo.ListControllers(userId)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"list": entityList})
}

// Helper function that returns userId from context
func getUserId(ctx *gin.Context) (string, error) {
	// get userId
	v, exist := ctx.Get("userId")
	if !exist {
		return "", keyNotFound
	}

	userId, ok := v.(string)
	if !ok {
		return "", castingFail
	}

	return userId, nil
}

// Package controller deals with Controller resource in our data source
// Usage outside of this package should only be to register routes for Gin Engine
package controller

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/tPhume/ags-backend/session"
	"net/http"
	"strings"
)

type mapping map[string]interface{}

func RegisterRoutes(handler *Handler, engine *gin.Engine, sessionHandler *session.Handler) {
	addValidation()
	group := engine.Group("api/v1/controller")
	group.Use(sessionHandler.GetUser)

	group.POST("", handler.AddController)
	group.GET("", handler.ListControllers)
	group.GET("/:controllerId", handler.GetController)
	group.PUT("/:controllerId", handler.UpdateController)
	group.DELETE("/:controllerId", handler.RemoveController)

	group.POST("/:controllerId/token/generate", handler.GenerateToken)
}

// Controller Entity type represent edge device
type Entity struct {
	ControllerId string `json:"controller_id"`
	UserId       string `json:"-"`
	Name         string `json:"name" binding:"required,name"`
	Desc         string `json:"desc"`
	Plan         string `json:"plan" binding:"omitempty,uuid4"`
	Token        string `json:"token,omitempty"`
}

// addStructValidation register StructValidation function to Gin's default validator Engine
func addValidation() {
	v := binding.Validator.Engine().(*validator.Validate)
	_ = v.RegisterValidation("name", NameValidation)
}

// Field level validation
func NameValidation(fl validator.FieldLevel) bool {
	field := fl.Field()

	value := field.String()
	if strings.TrimSpace(value) == "" {
		return false
	}

	return true
}

// Controller Repo - interface to communicate with data source
type Repo interface {
	// AddController creates new controller at data source given *Entity type
	// Duplicated Controller entity will result in an error
	AddController(context.Context, *Entity) (error, error)

	// ListControllers fetches all controller under the given UserId
	// Return of empty slice does not imply error
	ListControllers(context.Context, string) ([]*Entity, error)

	// GetController fetches specific controller by given Entity with UserId and ControllerId
	// Return of nil value for *Entity indicates error
	GetController(context.Context, *Entity) error

	// UpdateController replaces the controller given Entity object
	UpdateController(context.Context, *Entity) error

	// RemoveController deletes data from data source given ControllerId
	// Cascade deletion is done asynchronously
	// Missing controller will result in an error
	RemoveController(context.Context, string, string) error

	// GenerateToken replaces the token (must be hashed prior) given the userId, controllerId and tokenId
	// Missing controller will result in an error
	GenerateToken(context.Context, string, string, string) error
}

// Contains errors that implementation of Repo should use
var (
	duplicateName      = errors.New("duplicate name")
	controllerNotFound = errors.New("controller not found")
	tokenIncorrect     = errors.New("token incorrect")
)

// PlanRepo
type PlanRepo interface {
	PlanExist(context.Context, string, string) error
}

var planNotFound = errors.New("plan not found")

// Handler for controller REST API
type Handler struct {
	Repo     Repo
	PlanRepo PlanRepo
}

var (
	// error messages in general
	keyNotFound = errors.New("key not found")
	castingFail = errors.New("casting fail")
	badFormat   = errors.New("")

	// ok message responses for handler
	resAdded    = "controller added"
	resList     = "list of controllers retrieved"
	resGet      = "controller retrieved"
	resUpdate   = "controller updated"
	resRemove   = "controller removed"
	resGenerate = "controller's token generated"
	resVerifyOk = "token is correct"

	// error message responses for handler
	resInternal        = "not your fault, don't worry"
	resInvalid         = "invalid values"
	resDup             = "duplicate name"
	resNotFound        = "not found"
	resVerifyIncorrect = "token incorrect"
	resPlanNotFound    = "plan not found"
)

func (h *Handler) AddController(ctx *gin.Context) {
	userId := ctx.GetString("userId")
	if userId == "" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}

	entity := &Entity{
		ControllerId: uuid.New().String(),
		UserId:       userId,
		Token:        uuid.New().String(),
	}

	if err := ctx.ShouldBindJSON(entity); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": resInvalid})
		return
	}

	entity.Name = strings.TrimSpace(entity.Name)

	if entity.Plan != "" {
		if err := h.PlanRepo.PlanExist(ctx, entity.UserId, entity.Plan); err != nil {
			if err == planNotFound {
				ctx.JSON(http.StatusNotFound, gin.H{"message": resPlanNotFound})
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
			}

			return
		}
	}

	if err1, err2 := h.Repo.AddController(ctx, entity); err1 != nil {
		if err1 == duplicateName {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": resDup, "err": err2.Error(), "raw": err2})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		}

		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": resAdded, "controller": entity})
}

func (h *Handler) ListControllers(ctx *gin.Context) {
	userId := ctx.GetString("userId")
	if userId == "" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}

	entityList, err := h.Repo.ListControllers(ctx, userId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": resList, "controller_list": entityList})
}

func (h *Handler) GetController(ctx *gin.Context) {
	userId := ctx.GetString("userId")
	if userId == "" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}

	// check controllerId
	controllerId := ctx.Param("controllerId")
	if _, err := uuid.Parse(controllerId); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": resInvalid})
		return
	}

	entity := &Entity{ControllerId: controllerId, UserId: userId}
	if err := h.Repo.GetController(ctx, entity); err != nil {
		if err == controllerNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"message": resNotFound})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		}

		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": resGet, "controller": entity})
}

func (h *Handler) UpdateController(ctx *gin.Context) {
	userId := ctx.GetString("userId")
	if userId == "" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}

	// check controllerId
	controllerId := ctx.Param("controllerId")
	if _, err := uuid.Parse(controllerId); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": resInvalid})
		return
	}

	// Bind body to Entity object
	entity := &Entity{
		ControllerId: controllerId,
		UserId:       userId,
	}

	if err := ctx.ShouldBindJSON(entity); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": resInvalid})
		return
	}

	entity.Name = strings.TrimSpace(entity.Name)

	if entity.Plan != "" {
		if err := h.PlanRepo.PlanExist(ctx, entity.UserId, entity.Plan); err != nil {
			if err == planNotFound {
				ctx.JSON(http.StatusNotFound, gin.H{"message": resPlanNotFound})
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
			}

			return
		}
	}

	// use repo to call external data source
	if err := h.Repo.UpdateController(ctx, entity); err != nil {
		if err == controllerNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"message": resNotFound})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		}

		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": resUpdate, "controller": entity})
}

func (h *Handler) RemoveController(ctx *gin.Context) {
	userId := ctx.GetString("userId")
	if userId == "" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}

	// check controllerId
	controllerId := ctx.Param("controllerId")
	if _, err := uuid.Parse(controllerId); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": resInvalid})
		return
	}

	if err := h.Repo.RemoveController(ctx, userId, controllerId); err != nil {
		if err == controllerNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"message": resNotFound})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": resRemove})
}

func (h *Handler) GenerateToken(ctx *gin.Context) {
	userId := ctx.GetString("userId")
	if userId == "" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}

	// check controllerId
	controllerId := ctx.Param("controllerId")
	if _, err := uuid.Parse(controllerId); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": resInvalid})
		return
	}

	// generate token
	token := uuid.New().String()
	if err := h.Repo.GenerateToken(ctx, userId, controllerId, token); err != nil {
		if err == controllerNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"message": resNotFound})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": resGenerate, "token": token})
}

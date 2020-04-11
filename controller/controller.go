// Package controller deals with Controller resource in our data source
// Usage outside of this package should only be to register routes for Gin Engine
package controller

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/tPhume/ags-backend/session"
	"io"
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

	group.POST("/:controllerId/token", handler.GenerateToken)
	group.GET("/:controllerId/token", handler.VerifyToken)
}

// Controller Entity type represent edge device
type Entity struct {
	ControllerId string `json:"controller_id"`
	UserId       string `json:"-"`
	Name         string `json:"name" binding:"required,name"`
	Desc         string `json:"desc"`
	Plan         string `json:"plan" binding:"omitempty,uuid4"`
}

// VerifyToken request body
type VerifyToken struct {
	Token string `json:"token" binding:"required,uuid4"`
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
	AddController(context.Context, *Entity) error

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

	// GenerateToken replaces the token (must be hashed prior) given the userId, controllerId and hashed token
	// Missing controller will result in an error
	GenerateToken(context.Context, string, string, string) error

	// VerifyToken checks the hashed token against the one it has in store
	// If token does not match the hash it result in an error
	// Missing controller will result in an error
	VerifyToken(context.Context, string, string, string) error
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
	Key      string
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
	userId, err := getUserId(ctx)
	if err != nil {
		if err == badFormat {
			ctx.JSON(http.StatusBadRequest, resInvalid)
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}

	entity := &Entity{
		ControllerId: uuid.New().String(),
		UserId:       userId,
	}

	if err = ctx.ShouldBindJSON(entity); err != nil {
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

	if err = h.Repo.AddController(ctx, entity); err != nil {
		if err == duplicateName {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": resDup})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		}

		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": resAdded, "controller_id": entity.ControllerId})
}

func (h *Handler) ListControllers(ctx *gin.Context) {
	userId, err := getUserId(ctx)
	if err != nil {
		if err == badFormat {
			ctx.JSON(http.StatusBadRequest, resInvalid)
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}

	entityList, err := h.Repo.ListControllers(ctx, userId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": resList, "list": entityList})
}

func (h *Handler) GetController(ctx *gin.Context) {
	userId, err := getUserId(ctx)
	if err != nil {
		if err == badFormat {
			ctx.JSON(http.StatusBadRequest, resInvalid)
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}

	// check controllerId
	controllerId := ctx.Param("controllerId")
	if _, err = uuid.Parse(controllerId); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": resInvalid})
		return
	}

	entity := &Entity{ControllerId: controllerId, UserId: userId}
	if err = h.Repo.GetController(ctx, entity); err != nil {
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
	userId, err := getUserId(ctx)
	if err != nil {
		if err == badFormat {
			ctx.JSON(http.StatusBadRequest, resInvalid)
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}

	// check controllerId
	controllerId := ctx.Param("controllerId")
	if _, err = uuid.Parse(controllerId); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": resInvalid})
		return
	}

	// Bind body to Entity object
	entity := &Entity{
		ControllerId: controllerId,
		UserId:       userId,
	}

	if err = ctx.ShouldBindJSON(entity); err != nil {
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
	if err = h.Repo.UpdateController(ctx, entity); err != nil {
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
	userId, err := getUserId(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}

	// check controllerId
	controllerId := ctx.Param("controllerId")
	if _, err = uuid.Parse(controllerId); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": resInvalid})
		return
	}

	if err = h.Repo.RemoveController(ctx, userId, controllerId); err != nil {
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
	userId, err := getUserId(ctx)
	if err != nil {
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
	tokenId := uuid.New().String()
	hash := hmac.New(sha256.New, []byte(h.Key))

	if _, err := io.WriteString(hash, tokenId); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}
	hashedToken := hex.EncodeToString(hash.Sum(nil))

	if err := h.Repo.GenerateToken(ctx, userId, controllerId, hashedToken); err != nil {
		if err == controllerNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"message": resNotFound})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": resGenerate, "token_id": tokenId})
}

func (h *Handler) VerifyToken(ctx *gin.Context) {
	userId, err := getUserId(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}

	// check controllerId
	controllerId := ctx.Param("controllerId")
	if _, err := uuid.Parse(controllerId); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": resInvalid})
		return
	}

	// check controllerToken
	body := &VerifyToken{}
	if err := ctx.ShouldBindJSON(body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": resInvalid})
		return
	}

	// find hash of controllerToken
	hasher := hmac.New(sha256.New, []byte(h.Key))

	if _, err := io.WriteString(hasher, body.Token); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}
	hashedToken := hex.EncodeToString(hasher.Sum(nil))

	if err := h.Repo.VerifyToken(ctx, userId, controllerId, hashedToken); err != nil {
		if err == controllerNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"message": resNotFound})
			return
		}

		if err == tokenIncorrect {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": resVerifyIncorrect})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": resVerifyOk})
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

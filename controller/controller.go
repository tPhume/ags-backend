// Package controller deals with Controller resource in our data source
// Usage outside of this package should only be to register routes for Gin Engine
package controller

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"gopkg.in/go-playground/validator.v9"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type mapping map[string]interface{}

// Controller Entity type represent edge device
type Entity struct {
	ControllerId string `json:"controller_id"`
	UserId       string
	Name         string `json:"name"`
	Desc         string `json:"desc"`
	Plan         string `json:"plan"`
}

// Custom Controller type struct validation function
func StructValidation(sl validator.StructLevel) {
	entity := sl.Current().Interface().(Entity)

	// checks Name
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
	AddController(*Entity) error

	// ListControllers fetches all controller under the given UserId
	// Return of empty slice does not imply error
	ListControllers(string) ([]*Entity, error)

	// GetController fetches specific controller by given Entity with UserId and ControllerId
	// Return of nil value for *Entity indicates error
	GetController(*Entity) error

	// UpdateController does partial update given *Controller type and some value to update (mapping)
	// No new controller is created if controller is not found
	UpdateController(mapping) (*Entity, error)

	// RemoveController deletes data from data source given ControllerId
	// Cascade deletion is done asynchronously
	// Missing controller will result in an error
	RemoveController(string, string) error

	// GenerateToken replaces the token (must be hashed prior) given the userId, controllerId and hashed token
	// Missing controller will result in an error
	GenerateToken(string, string, string) error

	// VerifyToken checks the hashed token against the one it has in store
	// If token does not match the hash it result in an error
	// Missing controller will result in an error
	VerifyToken(string, string, string) error
}

// Contains errors that implementation of Repo should use
var (
	duplicateName      = errors.New("duplicate name")
	controllerNotFound = errors.New("controller not found")
	tokenIncorrect     = errors.New("token incorrect")
)

// Handler for controller REST API
type Handler struct {
	repo Repo
	key  string
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

	if err = h.repo.AddController(entity); err != nil {
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

	entityList, err := h.repo.ListControllers(userId)
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
	if err = h.repo.GetController(entity); err != nil {
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

	// extract request body from context
	reqBody, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}

	// unmarshal to map
	updateMap := make(mapping)
	if err = json.Unmarshal(reqBody, &updateMap); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}

	// addUserId and controllerId
	updateMap["userId"] = userId
	updateMap["controllerId"] = controllerId

	if err = checkUpdateMap(updateMap); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": resInvalid})
		return
	}

	// use repo to call external data source
	entity, err := h.repo.UpdateController(updateMap)
	if err != nil {
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

	if err = h.repo.RemoveController(userId, controllerId); err != nil {
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
		if err == badFormat {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": resInvalid})
			return
		}

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
	hash := hmac.New(sha256.New, []byte(h.key))

	if _, err := io.WriteString(hash, tokenId); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}
	hashedToken := hex.EncodeToString(hash.Sum(nil))

	if err := h.repo.GenerateToken(userId, controllerId, hashedToken); err != nil {
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
		if err == badFormat {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": resInvalid})
			return
		}

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
	controllerToken := ctx.Param("controllerToken")
	if _, err := uuid.Parse(controllerToken); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": resInvalid})
		return
	}

	// find hash of controllerToken
	hash := hmac.New(sha256.New, []byte(h.key))

	if _, err := io.WriteString(hash, controllerToken); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}
	hashedToken := hex.EncodeToString(hash.Sum(nil))

	if err := h.repo.VerifyToken(userId, controllerId, hashedToken); err != nil {
		if err == controllerNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"message": resNotFound})
			return
		}

		if err == tokenIncorrect {
			ctx.JSON(http.StatusNotFound, gin.H{"message": resVerifyIncorrect})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": resVerifyOk})
}

// Utility functions that makes life somewhat easier
// Helper function to check map for update
func checkUpdateMap(updateMap mapping) error {
	if v, exist := updateMap["name"]; exist {
		s, ok := v.(string)
		if !ok {
			return badFormat
		}

		if len(strings.TrimSpace(s)) < 1 {
			return badFormat
		}
	}

	if v, exist := updateMap["desc"]; exist {
		_, ok := v.(string)
		if !ok {
			return badFormat
		}
	}

	if v, exist := updateMap["name"]; exist {
		s, ok := v.(string)
		if !ok {
			return badFormat
		}

		if _, err := uuid.Parse(s); err != nil {
			return err
		}
	}

	return nil
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

	if _, err := uuid.Parse(userId); err != nil {
		return "", badFormat
	}

	return userId, nil
}

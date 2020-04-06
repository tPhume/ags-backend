package session

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"net/http"
	"strings"
)

type mapping map[string]interface{}

func RegisterRoutes(handler *Handler, engine *gin.Engine) {
	AddValidation()

	group := engine.Group("api/v1/session")
	group.POST("", handler.CreateSession)
	group.DELETE("", handler.DeleteSession)
}

// Represent a user
type UserEntity struct {
	UserId        string `json:"user_id"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// To be bind for create request
type CreateRequest struct {
	AccessCode string `json:"access_code" binding:"accessCode"`
}

// field level validation
func AccessCodeValidation(fl validator.FieldLevel) bool {
	if strings.TrimSpace(fl.Field().String()) == "" {
		return false
	}

	return true
}

func AddValidation() {
	validate := binding.Validator.Engine().(*validator.Validate)
	_ = validate.RegisterValidation("accessCode", AccessCodeValidation)
}

// Repo type interacts with data source that has session database
type Repo interface {
	CreateSession(context.Context, *UserEntity, string) error

	DeleteSession(context.Context, string) error

	GetUser(context.Context, string) (string, error)
}

var errNotFound = errors.New("session not found")

// GoogleRepo interacts with google api
type GoogleRepo interface {
	GetIdToken(context.Context, string, *UserEntity) error
}

var errBadCode = errors.New("bad access_code")

// Handler message responses
const (
	resCreate = "session created"
	resDelete = "session deleted"

	resInvalid  = "bad format"
	resInternal = "not your fault, internal error"
	resNotAuth  = "not authorized"
)

// Handler stores Repo type that interacts with data source
type Handler struct {
	Domain     string
	Repo       Repo
	GoogleRepo GoogleRepo
}

// CreateSession takes an exchange token and set cookie
// Return body includes user information
func (h *Handler) CreateSession(ctx *gin.Context) {
	request := &CreateRequest{}
	if err := ctx.ShouldBindJSON(request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": resInvalid})
		return
	}

	userEntity := &UserEntity{}
	if err := h.GoogleRepo.GetIdToken(ctx, request.AccessCode, userEntity); err != nil {
		if err == errBadCode {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": resInvalid})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal, "details": "google exchange"})
		}

		return
	}

	sessionId := uuid.New().String()
	if err := h.Repo.CreateSession(ctx, userEntity, sessionId); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal, "details": err})
		return
	}

	ctx.SetCookie("sessionId", sessionId, 28800, "/", h.Domain, false, true)
	ctx.JSON(http.StatusCreated, gin.H{"message": resCreate})
}

// DeleteSession will delete the session cookie
func (h *Handler) DeleteSession(ctx *gin.Context) {
	sessionId, err := ctx.Cookie("sessionId")
	if err != nil || strings.TrimSpace(sessionId) == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": resInvalid})
		return
	}

	if err = h.Repo.DeleteSession(ctx, sessionId); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}

	ctx.SetCookie("sessionId", "", 28800, "/", h.Domain, false, true)
	ctx.JSON(http.StatusOK, gin.H{"message": resDelete})
}

// GetSession is the middleware that will check the session cookie from request
// It then sets the userId in context
func (h *Handler) GetUser(ctx *gin.Context) {
	sessionId, err := ctx.Cookie("sessionId")
	if err != nil || strings.TrimSpace(sessionId) == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": resNotAuth})
		return
	}

	userId, err := h.Repo.GetUser(ctx, sessionId)
	if err != nil {
		if err == errNotFound {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": resNotAuth})
		} else {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		}

		return
	}

	ctx.Set("userId", userId)
}

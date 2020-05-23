package session

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"strings"
)

func RegisterRoutes(handler *Handler, engine *gin.Engine) {

	engine.POST("api/v1/user", handler.CreateUser)

	group := engine.Group("api/v1/session")
	group.POST("", handler.CreateSession)
	group.DELETE("", handler.DeleteSession)
}

// Represent a user
type UserEntity struct {
	UserId   string `json:"user_id" bson:"_id"`
	Name     string `json:"name" bson:"name" binding:"required"`
	Password string `json:"password" bson:"password" binding:"required"`
}

// Repo type interacts with data source that has session database
type Repo interface {
	CreateSession(context.Context, *UserEntity, string) error

	DeleteSession(context.Context, string) error

	CreateUser(context.Context, *UserEntity) error

	GetUser(context.Context, string) (string, error)
}

var (
	errNotFound         = errors.New("session not found")
	errUserDoesNotExist = errors.New("user does not exist")
	errConflict         = errors.New("conflict")
)

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
	Repo Repo
}

// CreateSession takes an exchange token and set cookie
// Return body includes user information
func (h *Handler) CreateSession(ctx *gin.Context) {
	userEntity := &UserEntity{}
	if err := ctx.ShouldBindJSON(userEntity); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": resInvalid})
		return
	}

	session := uuid.New().String()
	if err := h.Repo.CreateSession(ctx, userEntity, session); err != nil {
		if err == errUserDoesNotExist {
			ctx.JSON(http.StatusNotFound, gin.H{"message": "credentials not match or user does not exist"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal, "details": err})
		}

		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": resCreate, "user": userEntity.Name, "session": session, "user_id": userEntity.UserId})
}

// DeleteSession will delete the session cookie
func (h *Handler) DeleteSession(ctx *gin.Context) {
	sessionId := ctx.GetHeader("session")
	if strings.TrimSpace(sessionId) == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": resNotAuth})
		return
	}

	if err := h.Repo.DeleteSession(ctx, sessionId); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": resDelete})
}

func (h *Handler) CreateUser(ctx *gin.Context) {
	userEntity := &UserEntity{}
	if err := ctx.ShouldBindJSON(userEntity); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": resInvalid})
		return
	}

	userEntity.UserId = uuid.New().String()
	if err := h.Repo.CreateUser(ctx, userEntity); err != nil {
		if err == errConflict {
			ctx.JSON(http.StatusConflict, gin.H{"message": "could not create user"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal, "details": err})
		}

		return
	}

	ctx.Status(http.StatusCreated)
}

// GetSession is the middleware that will check the session cookie from request
// It then sets the userId in context
func (h *Handler) GetUser(ctx *gin.Context) {
	session := ctx.GetHeader("session")
	if strings.TrimSpace(session) == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": resNotAuth})
		return
	}

	userId, err := h.Repo.GetUser(ctx, session)
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

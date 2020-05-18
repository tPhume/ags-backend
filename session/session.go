package session

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"strings"
)

type mapping map[string]interface{}

func RegisterRoutes(handler *Handler, engine *gin.Engine) {

	group := engine.Group("api/v1/session")
	group.POST("", handler.CreateSession)
	group.DELETE("", handler.DeleteSession)
}

// Represent a user
type UserEntity struct {
	UserId   string `json:"user_id"`
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Repo type interacts with data source that has session database
type Repo interface {
	CreateSession(context.Context, *UserEntity, string) error

	DeleteSession(context.Context, string) error

	GetUser(context.Context, string) (string, error)
}

var (
	errNotFound         = errors.New("session not found")
	errUserDoesNotExist = errors.New("user does not exist")
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

	ctx.JSON(http.StatusCreated, gin.H{"message": resCreate, "user": userEntity.Name, "session": session})
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

	ctx.JSON(http.StatusOK, gin.H{"message": resDelete})
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

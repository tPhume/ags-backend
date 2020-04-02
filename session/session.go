package session

import (
	"context"
	"github.com/gin-gonic/gin"
)

// Represent a user
type UserEntity struct {
	UserId        string `json:"user_id"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// Repo type interacts with data source that has session database
type Repo interface {
	CreateSession(context.Context, *UserEntity, string) error

	DeleteSession(context.Context, string) error

	GetSession(context.Context, string) (string, error)
}

// Handler stores Repo type that interacts with data source
type Handler struct {
	repo Repo
}

// CreateSession takes an exchange token and set cookie
// Return body includes user information
func (h *Handler) CreateSession(ctx *gin.Context) {

}

// DeleteSession will delete the session cookie
func (h *Handler) DeleteSession(ctx *gin.Context) {

}

// GetSession is the middleware that will check the session cookie from request
// It then sets the userId in context
func (h *Handler) GetSession(ctx *gin.Context) {

}

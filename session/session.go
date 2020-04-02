package session

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"strings"
)

type mapping map[string]interface{}

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

	GetSession(context.Context, string) (string, error)
}

// GoogleRepo interacts with google api
type GoogleRepo interface {
	GetIdToken(string) error
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
	repo       Repo
	googleRepo GoogleRepo
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

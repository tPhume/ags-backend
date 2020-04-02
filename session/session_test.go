package session

import (
	"context"
	"github.com/gin-gonic/gin"
	"testing"
)

type repoStruct struct{}

func (r *repoStruct) CreateSession(ctx context.Context, userEntity *UserEntity, sessionId string) error {
	return nil
}

func (r *repoStruct) DeleteSession(ctx context.Context, sessionId string) error {
	return nil
}

func (r *repoStruct) GetSession(ctx context.Context, sessionId string) (string, error) {
	return "", nil
}

var handler = &Handler{repo: &repoStruct{}}

func setUp() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestHandler_CreateSession(t *testing.T) {

}

func TestHandler_DeleteSession(t *testing.T) {

}

func TestHandler_GetSession(t *testing.T) {

}

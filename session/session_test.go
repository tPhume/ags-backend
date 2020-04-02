package session

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	goodCode = "good code"
	badCode  = "bad code"
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

type googleRepoStruct struct{}

func (g *googleRepoStruct) GetIdToken(string) error {
	return nil
}

var handler = &Handler{repo: &repoStruct{}, googleRepo: &googleRepoStruct{}}

func setUp() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestHandler_CreateSession(t *testing.T) {
	engine := setUp()
	engine.POST("", handler.CreateSession)

	testCases := []struct {
		in      mapping
		message string
		code    int
	}{
		{
			in:      mapping{"access_code": goodCode},
			message: resCreate,
			code:    http.StatusCreated,
		}, {
			in:      mapping{"access_code": badCode},
			message: resInvalid,
			code:    http.StatusBadRequest,
		}, {
			in:      mapping{"access_code": ""},
			message: resInvalid,
			code:    http.StatusBadRequest,
		},
	}

	for i, c := range testCases {
		resp := httptest.NewRecorder()

		body, _ := json.Marshal(c.in)
		req, _ := http.NewRequest(http.MethodPost, "/", bytes.NewReader(body))

		engine.ServeHTTP(resp, req)

		respBody := mapping{}
		_ = json.Unmarshal(resp.Body.Bytes(), &respBody)

		if c.code != resp.Code {
			t.Fatalf("Case %d: expected [%v], got = [%v]", i, c.code, resp.Code)
		}

		if c.message != respBody["message"] {
			t.Fatalf("Case %d: expected [%v], got = [%v]", i, c.message, respBody["message"])
		}
	}
}

func TestHandler_DeleteSession(t *testing.T) {

}

func TestHandler_GetSession(t *testing.T) {

}

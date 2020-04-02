package session

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	goodCode = "good code"
	badCode  = "bad code"

	goodSessionId = "good sessionId"
	badSessionId  = "bad sessionId"
)

type repoStruct struct{}

func (r *repoStruct) CreateSession(ctx context.Context, userEntity *UserEntity, sessionId string) error {
	return nil
}

func (r *repoStruct) DeleteSession(ctx context.Context, sessionId string) error {
	if sessionId == goodSessionId {
		return nil
	} else if sessionId == badSessionId {
		return errNotFound
	}

	return errors.New("some internal error")
}

func (r *repoStruct) GetSession(ctx context.Context, sessionId string) (string, error) {
	return "", nil
}

type googleRepoStruct struct{}

func (g *googleRepoStruct) GetIdToken(code string, entity *UserEntity) error {
	if code == goodCode {
		return nil
	} else if code == badCode {
		return errBadCode
	}

	return errors.New("some internal error")
}

var handler = &Handler{repo: &repoStruct{}, googleRepo: &googleRepoStruct{}, domain: "testing"}

func setUp() *gin.Engine {
	gin.SetMode(gin.TestMode)

	AddValidation()
	engine := gin.New()

	return engine
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
		}, {
			in:      mapping{"access_code": "some internal error"},
			message: resInternal,
			code:    http.StatusInternalServerError,
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
	engine := setUp()
	engine.DELETE("", handler.DeleteSession)

	testCases := []struct {
		in      string
		message string
		code    int
	}{
		{
			in:      goodSessionId,
			message: resDelete,
			code:    http.StatusOK,
		}, {
			in:      badSessionId,
			message: resNotFound,
			code:    http.StatusNotFound,
		}, {
			in:      "",
			message: resInvalid,
			code:    http.StatusBadRequest,
		}, {
			in:      "some internal error",
			message: resInternal,
			code:    http.StatusInternalServerError,
		},
	}

	for i, c := range testCases {
		cookie := &http.Cookie{
			Name:   "sessionId",
			Value:  c.in,
			Path:   "/",
			Domain: handler.domain,
		}

		resp := httptest.NewRecorder()

		req, _ := http.NewRequest(http.MethodDelete, "/", nil)
		req.AddCookie(cookie)

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

func TestHandler_GetSession(t *testing.T) {

}

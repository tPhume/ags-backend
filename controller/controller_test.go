package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
)

// For out test controller.ControllerId uuid indicate existing controller
// Any other uuid will indicate missing controller - for 404 cases

const goodHashedToken string = "f911ec3e7c28625729dd99f1ff27704e3f64f4aaaa4118fedd1e19c9df5f4c1a"

var controller = Entity{
	ControllerId: "f1d67e51-4ca4-4b25-a4b7-6c8f06822075",
	UserId:       "76de6d55-e457-4070-8aef-5633726d498f",
	Name:         "Controller",
	Desc:         "",
	Plan:         "",
}

// Repo struct for testing
type repoStruct struct{}

func (t *repoStruct) AddController(ctx context.Context, entity *Entity) error {
	if entity.Name == "DuplicateName" {
		return duplicateName
	} else if entity.Name == "InternalName" {
		return errors.New("some error")
	}

	return nil
}

func (t *repoStruct) ListControllers(context.Context, string) ([]*Entity, error) {
	return []*Entity{&controller}, nil
}

func (t *repoStruct) GetController(ctx context.Context, entity *Entity) error {
	if entity.ControllerId == controller.ControllerId {
		return nil
	}

	return controllerNotFound
}

func (t *repoStruct) UpdateController(ctx context.Context, entity *Entity) error {
	if entity.ControllerId == controller.ControllerId {
		return nil
	}

	return controllerNotFound
}

func (t *repoStruct) RemoveController(ctx context.Context, userId string, controllerId string) error {
	if controllerId == controller.ControllerId {
		return nil
	} else if controllerId == controller.UserId {
		return controllerNotFound
	}

	return nil
}

func (t *repoStruct) GenerateToken(ctx context.Context, userId string, controllerId string, token string) error {
	if controllerId == controller.ControllerId {
		return nil
	} else if controllerId == controller.UserId {
		return controllerNotFound
	}

	return nil
}

func (t *repoStruct) VerifyToken(ctx context.Context, userId string, controllerId string, hashedToken string) error {
	if controllerId != controller.ControllerId {
		return controllerNotFound
	}

	if hashedToken != goodHashedToken {
		return tokenIncorrect
	}

	return nil
}

// PlanRepo struct for testing
type planRepoStruct struct {}

func (p *planRepoStruct) PlanExist(context.Context, string, string) error {
	return nil
}

// Handler struct for testing
var handler = &Handler{Repo: &repoStruct{}, PlanRepo:&planRepoStruct{}, Key: "fake"}

// Setup func for handler testing
func setUp() *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	addValidation()
	engine.Use(func(context *gin.Context) {
		context.Set("userId", controller.UserId)
	})

	return engine
}

// Test AddControllers handler
func TestHandler_AddController(t *testing.T) {
	engine := setUp()
	engine.POST("", handler.AddController)

	testCases := []struct {
		in      mapping
		message string
		code    int
	}{
		{
			in:      mapping{"Name": "GoodName", "Desc": "GoodDesc"},
			message: resAdded,
			code:    http.StatusCreated,
		}, {
			in:      mapping{"Name": "GoodName", "Desc": ""},
			message: resAdded,
			code:    http.StatusCreated,
		}, {
			in:      mapping{"Name": "", "Desc": "GoodDesc"},
			message: resInvalid,
			code:    http.StatusBadRequest,
		}, {
			in:      mapping{"Name": "    ", "Desc": "GoodDesc"},
			message: resInvalid,
			code:    http.StatusBadRequest,
		}, {
			in:      mapping{"Name": "DuplicateName", "Desc": "GoodDesc"},
			message: resDup,
			code:    http.StatusBadRequest,
		}, {
			in:      mapping{"Name": "InternalName", "Desc": "GoodDesc"},
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

// Test ListController handler
func TestHandler_ListControllers(t *testing.T) {
	engine := setUp()
	engine.GET("", handler.ListControllers)

	testCases := []struct {
		message string
		code    int
	}{
		{
			message: resList,
			code:    http.StatusOK,
		},
	}

	for _, c := range testCases {
		resp := httptest.NewRecorder()

		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		engine.ServeHTTP(resp, req)

		respBody := mapping{}
		_ = json.Unmarshal(resp.Body.Bytes(), &respBody)

		if c.code != resp.Code {
			t.Fatalf("expected [%v], got = [%v]", c.code, resp.Code)
		}

		if c.message != respBody["message"] {
			t.Fatalf("expected [%v], got = [%v]", c.message, respBody["message"])
		}
	}
}

// Test GetController handler
func TestHandler_GetController(t *testing.T) {
	engine := setUp()
	engine.GET(":controllerId", handler.GetController)

	testCases := []struct {
		in           mapping
		controllerId string
		message      string
		code         int
	}{
		{
			in:           mapping{},
			controllerId: controller.ControllerId,
			message:      resGet,
			code:         http.StatusOK,
		}, {
			in:           mapping{},
			controllerId: "fmkdjsnlfk",
			message:      resInvalid,
			code:         http.StatusBadRequest,
		}, {
			in:           mapping{},
			controllerId: controller.UserId,
			message:      resNotFound,
			code:         http.StatusNotFound,
		},
	}

	for _, c := range testCases {
		resp := httptest.NewRecorder()

		body, _ := json.Marshal(c.in)
		req, _ := http.NewRequest(http.MethodGet, "/"+c.controllerId, bytes.NewReader(body))
		engine.ServeHTTP(resp, req)

		respBody := mapping{}
		_ = json.Unmarshal(resp.Body.Bytes(), &respBody)

		if c.code != resp.Code {
			t.Fatalf("expected [%v], got = [%v]", c.code, resp.Code)
		}

		if c.message != respBody["message"] {
			t.Fatalf("expected [%v], got = [%v]", c.message, respBody["message"])
		}
	}
}

// Test UpdateController handler
func TestHandler_UpdateController(t *testing.T) {
	engine := setUp()
	engine.PATCH(":controllerId", handler.UpdateController)

	testCases := []struct {
		in      string
		body    mapping
		message string
		code    int
	}{
		{
			in:      controller.ControllerId,
			body:    mapping{"Name": "GoodName", "Desc": "GoodDesc"},
			message: resUpdate,
			code:    http.StatusOK,
		},
		{
			in:
			"lkmwklfmd",
			body:    mapping{},
			message: resInvalid,
			code:    http.StatusBadRequest,
		}, {
			in:      controller.ControllerId,
			body:    mapping{"Name": "", "Desc": "GoodDesc"},
			message: resInvalid,
			code:    http.StatusBadRequest,
		},
	}

	for _, c := range testCases {
		resp := httptest.NewRecorder()

		body, _ := json.Marshal(c.body)
		req, _ := http.NewRequest(http.MethodPatch, "/"+c.in, bytes.NewReader(body))
		engine.ServeHTTP(resp, req)

		respBody := mapping{}
		_ = json.Unmarshal(resp.Body.Bytes(), &respBody)

		if c.code != resp.Code {
			t.Fatalf("expected [%v], got = [%v]", c.code, resp.Code)
		}

		if c.message != respBody["message"] {
			t.Fatalf("expected [%v], got = [%v]", c.message, respBody["message"])
		}
	}
}

// Test RemoveController handler
func TestHandler_RemoveController(t *testing.T) {
	engine := setUp()
	engine.DELETE("/:controllerId", handler.RemoveController)

	testCases := []struct {
		in      string
		message string
		code    int
	}{
		{
			in:      controller.ControllerId,
			message: resRemove,
			code:    http.StatusOK,
		}, {
			in:      controller.UserId,
			message: resNotFound,
			code:    http.StatusNotFound,
		}, {
			in:      "fenwklfmke",
			message: resInvalid,
			code:    http.StatusBadRequest,
		},
	}

	for _, c := range testCases {
		resp := httptest.NewRecorder()

		req, _ := http.NewRequest(http.MethodDelete, "/"+c.in, nil)
		engine.ServeHTTP(resp, req)

		respBody := mapping{}
		_ = json.Unmarshal(resp.Body.Bytes(), &respBody)

		if c.code != resp.Code {
			t.Fatalf("expected [%v], got = [%v]", c.code, resp.Code)
		}

		if c.message != respBody["message"] {
			t.Fatalf("expected [%v], got = [%v]", c.message, respBody["message"])
		}
	}
}

// Test GenerateToken
func TestGenerateToken(t *testing.T) {
	engine := setUp()
	engine.POST("/:controllerId/token", handler.GenerateToken)

	testCases := []struct {
		in      string
		message string
		code    int
	}{
		{
			in:      controller.ControllerId,
			message: resGenerate,
			code:    http.StatusOK,
		}, {
			in:      controller.UserId,
			message: resNotFound,
			code:    http.StatusNotFound,
		}, {
			in:      "fewfe",
			message: resInvalid,
			code:    http.StatusBadRequest,
		},
	}

	for _, c := range testCases {
		resp := httptest.NewRecorder()

		req, _ := http.NewRequest(http.MethodPost, "/"+c.in+"/token", nil)
		engine.ServeHTTP(resp, req)

		respBody := mapping{}
		_ = json.Unmarshal(resp.Body.Bytes(), &respBody)

		if c.code != resp.Code {
			t.Fatalf("expected [%v], got = [%v]", c.code, resp.Code)
		}

		if c.message != respBody["message"] {
			t.Fatalf("expected [%v], got = [%v]", c.message, respBody["message"])
		}
	}
}

// Test VerifyToken
func TestVerifyToken(t *testing.T) {
	engine := setUp()
	engine.GET(":controllerId", handler.VerifyToken)

	testCases := []struct {
		in      string
		body    mapping
		message string
		code    int
	}{
		{
			in:      fmt.Sprintf("/%s", controller.ControllerId),
			body:    mapping{"token": controller.ControllerId},
			message: resVerifyOk,
			code:    http.StatusOK,
		}, {
			in:      fmt.Sprintf("/%s", controller.UserId),
			body:    mapping{"token": controller.ControllerId},
			message: resNotFound,
			code:    http.StatusNotFound,
		}, {
			in:      fmt.Sprintf("/%s", controller.ControllerId),
			body:    mapping{"token": "fnjdslfnlk"},
			message: resInvalid,
			code:    http.StatusBadRequest,
		}, {
			in:      fmt.Sprintf("/%s", "dfwqfe"),
			body:    mapping{"token": controller.ControllerId},
			message: resInvalid,
			code:    http.StatusBadRequest,
		}, {
			in:      fmt.Sprintf("/%s", controller.ControllerId),
			body:    mapping{"token": controller.UserId},
			message: resVerifyIncorrect,
			code:    http.StatusBadRequest,
		},
	}

	for _, c := range testCases {
		resp := httptest.NewRecorder()

		body, _ := json.Marshal(c.body)

		req, _ := http.NewRequest(http.MethodGet, c.in, bytes.NewReader(body))
		engine.ServeHTTP(resp, req)

		respBody := mapping{}
		_ = json.Unmarshal(resp.Body.Bytes(), &respBody)

		if c.code != resp.Code {
			t.Fatalf("expected [%v], got = [%v]", c.code, resp.Code)
		}

		if c.message != respBody["message"] {
			t.Fatalf("expected [%v], got = [%v]", c.message, respBody["message"])
		}
	}
}

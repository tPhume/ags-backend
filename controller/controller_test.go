package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mapping = map[string]interface{}

var controller = Entity{
	ControllerId: "f1d67e51-4ca4-4b25-a4b7-6c8f06822075",
	UserId:       "76de6d55-e457-4070-8aef-5633726d498f",
	Name:         "Controller",
	Desc:         "",
	Plan:         "",
}

// Repo struct for testing
type repoStruct struct{}

func (t *repoStruct) AddController(entity *Entity) error {
	if entity.Name == "DuplicateName" {
		return duplicateName
	} else if entity.Name == "InternalName" {
		return errors.New("some error")
	}

	return nil
}

func (t *repoStruct) ListControllers(string) ([]*Entity, error) {
	return []*Entity{&controller}, nil
}

func (t *repoStruct) GetController(entity *Entity) error {
	if entity.ControllerId == controller.ControllerId {
		return nil
	}

	return controllerNotFound
}

func (t *repoStruct) UpdateController(entity *Entity) error {
	if entity.ControllerId == controller.ControllerId {
		return nil
	}

	return controllerNotFound
}

func (t *repoStruct) RemoveController(string) error {
	return nil
}

// Handler struct for testing
var handler = &Handler{repo: &repoStruct{}}

// Setup func for handler testing
func setUp() *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	addStructValidation(engine)
	engine.Use(func(context *gin.Context) {
		context.Set("userId", "76de6d55-e457-4070-8aef-5633726d498f")
	})

	return engine
}

// Test Entity Struct validation
func TestStructValidation(t *testing.T) {
	v := validator.New()
	v.RegisterStructValidation(StructValidation, Entity{})

	if err := v.Struct(controller); err != nil {
		t.Fatal(err)
	}
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

	for _, c := range testCases {
		resp := httptest.NewRecorder()

		body, _ := json.Marshal(c.in)
		req, _ := http.NewRequest(http.MethodPost, "", bytes.NewReader(body))
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

		req, _ := http.NewRequest(http.MethodGet, "", nil)
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
		req, _ := http.NewRequest(http.MethodGet, c.controllerId, bytes.NewReader(body))
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
	engine.PATCH(":controllerId", handler.UpdateHandler)

	testCases := []struct {
		in      string
		message string
		code    int
	}{
		{
			in:      controller.ControllerId,
			message: resUpdate,
			code:    http.StatusOK,
		}, {
			in:      "lkmwklfmd",
			message: resInvalid,
			code:    http.StatusBadRequest,
		}, {
			in:      controller.UserId,
			message: resNotFound,
			code:    http.StatusNotFound,
		},
	}

	for _, c := range testCases {
		resp := httptest.NewRecorder()

		req, _ := http.NewRequest(http.MethodPatch, c.in, nil)
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

}

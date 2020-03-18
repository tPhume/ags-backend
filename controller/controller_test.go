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
	return []*Entity{{Name: "Happy"}}, nil
}

func (t *repoStruct) GetController(string) (*Entity, error) {
	return &Entity{Name: "Happy"}, nil
}

func (t *repoStruct) UpdateController(*Entity) error {
	return nil
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

	// response messages expected
	badResponse := "invalid values"
	duplicateResponse := "duplicate name"
	goodResponse := "controller added"

	testCases := []struct {
		in      mapping
		message string
		code    int
	}{
		{
			in:      mapping{"Name": "GoodName", "Desc": "GoodDesc"},
			message: goodResponse,
			code:    http.StatusCreated,
		}, {
			in:      mapping{"Name": "GoodName", "Desc": ""},
			message: goodResponse,
			code:    http.StatusCreated,
		}, {
			in:      mapping{"Name": "", "Desc": "GoodDesc"},
			message: badResponse,
			code:    http.StatusBadRequest,
		}, {
			in:      mapping{"Name": "    ", "Desc": "GoodDesc"},
			message: badResponse,
			code:    http.StatusBadRequest,
		}, {
			in:      mapping{"Name": "DuplicateName", "Desc": "GoodDesc"},
			message: duplicateResponse,
			code:    http.StatusBadRequest,
		}, {
			in:      mapping{"Name": "InternalName", "Desc": "GoodDesc"},
			message: "",
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

}

// Test GetController handler
func TestHandler_GetController(t *testing.T) {

}

// Test UpdateController handler
func TestHandler_UpdateController(t *testing.T) {

}

// Test RemoveController handler
func TestHandler_RemoveController(t *testing.T) {

}

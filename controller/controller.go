package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"gopkg.in/go-playground/validator.v9"
	"strings"
)

// Controller represent a Raspberry Pi unit
type Controller struct {
	ID   string `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
	Desc string `json:"desc"`
	Plan string `json:"plan"`
}

// Custom Controller type struct validation method
func StructValidation(sl validator.StructLevel) {
	controller := sl.Current().Interface().(Controller)

	if _, err := uuid.Parse(controller.ID); err != nil {
		sl.ReportError(controller.ID, "id", "id", "", "")
	}

	if len(strings.TrimSpace(controller.Name)) < 1 {
		sl.ReportError(controller.Name, "name", "name", "", "")
	}
}

// Function to add ControllerStructValidation to Gin's default Validator engine
func addStructValidation(engine *gin.Engine) {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterStructValidation(StructValidation)
	}
}

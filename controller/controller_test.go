package controller

import (
	"gopkg.in/go-playground/validator.v9"
	"testing"
)

var controller = Controller{
	ID:   "f1d67e51-4ca4-4b25-a4b7-6c8f06822075",
	Name: "Controller",
	Desc: "",
	Plan: "",
}

func TestStructValidation(t *testing.T) {
	v := validator.New()
	v.RegisterStructValidation(StructValidation, Controller{})

	if err := v.Struct(controller); err != nil {
		t.Fatal(err)
	}
}

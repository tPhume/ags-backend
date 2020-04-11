package plan

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
)

// Represent a Plan object
type Entity struct {
	PlanId        string     `json:"plan_id" bson:"_id"`
	Name          string     `json:"name" bson:"name"`
	HumidityState int        `json:"humidity_state"`
	TempState     float32    `json:"temp_state"`
	Daily         []Daily    `json:"daily"`
	Weekly        []Weekly   `json:"weekly"`
	Monthly       []Monthly  `json:"monthly"`
	Interval      []Interval `json:"interval"`
}

// Different type of routine
type Daily struct {
	DailyTime string   `json:"daily_time"`
	Action    []Action `json:"action"`
}

type Weekly struct {
	WeeklyTime string   `json:"weekly_time"`
	Action     []Action `json:"action"`
}

type Monthly struct {
	MonthlyTime string   `json:"monthly_time"`
	Action      []Action `json:"action"`
}

type Interval struct {
	IntervalTime string   `json:"interval_time"`
	Action       []Action `json:"action"`
}

// Action type
type Action struct {
	Type      string `json:"type"`
	Intensity int    `json:"intensity"`
	Duration  int    `json:"duration"`
}

// Repo interface for data source
// Errors that should be used with Repo interface
var (
	errPlanNotFound  = errors.New("plan not found")
	errPlanDuplicate = errors.New("plan with that name already exist")
)

type Repo interface {
	CreatePlan(ctx context.Context, entity *Entity) error

	ListPlans(ctx context.Context, userId string) ([]*Entity, error)

	GetPlan(ctx context.Context, entity *Entity) error

	ReplacePlan(ctx context.Context, entity *Entity) error

	DeletePlan(ctx context.Context, userId string, planId string) error
}

// Handler for Plan endpoint
// Response messages to use
const (
	// Success responses
	resCreatePlan  = "plan created"
	resListPlans   = "list of plans retrieved"
	resGetPlan     = "plan retrieved"
	resReplacePlan = "plan replaced"
	resDeletePlan  = "plan deleted"

	// Error responses
	resInvalid      = "invalid format"
	resInternal     = "internal error"
	resPlanConflict = "plan with same name already exist"
	resPlanNotFound = "plan not found"
)

type Handler struct {
	Repo Repo
}

func (h *Handler) CreatePlan(ctx *gin.Context) {

}

func (h *Handler) ListPlans(ctx *gin.Context) {

}

func (h *Handler) GetPlan(ctx *gin.Context) {

}

func (h *Handler) ReplacePlan(ctx *gin.Context) {

}

func (h *Handler) DeletePlan(ctx *gin.Context) {

}

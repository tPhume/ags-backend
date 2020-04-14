package plan

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/tPhume/ags-backend/session"
	"net/http"
	"strconv"
	"strings"
)

func RegisterRoutes(handler *Handler, engine *gin.Engine, sessionHandler *session.Handler) {
	if err := addValidation(); err != nil {
		panic("can't register Plan endpoint routes")
	}

	group := engine.Group("api/v1/plan")
	group.Use(sessionHandler.GetUser)

	group.POST("", handler.CreatePlan)
	group.GET("", handler.ListPlans)
	group.GET(":planId", handler.GetPlan)
	group.PUT(":planId", handler.ReplacePlan)
	group.DELETE(":planId", handler.DeletePlan)
}

func addValidation() error {
	validate := binding.Validator.Engine().(*validator.Validate)

	if err := validate.RegisterValidation("plan_name", planName); err != nil {
		return err
	}

	if err := validate.RegisterValidation("daily_time", dailyTime); err != nil {
		return err
	}

	if err := validate.RegisterValidation("weekly_time", weeklyTime); err != nil {
		return err
	}

	if err := validate.RegisterValidation("monthly_time", monthlyTime); err != nil {
		return err
	}

	if err := validate.RegisterValidation("action_type", actionType); err != nil {
		return err
	}

	return nil
}

// Represent a Plan object
type Entity struct {
	PlanId        string    `json:"plan_id" bson:"_id" binding:"omitempty,uuid4"`
	UserId        string    `json:"-" bson:"user_id" binding:"omitempty"`
	Name          string    `json:"name" bson:"name" binding:"plan_name"`
	HumidityState int       `json:"humidity_state" bson:"humidity_state" binding:"gte=0,lte=100"`
	TempState     float32   `json:"temp_state" bson:"temp_state" binding:"gte=0,lte=50"`
	Daily         []Daily   `json:"daily" bson:"daily" binding:"dive"`
	Weekly        []Weekly  `json:"weekly" bson:"weekly" binding:"dive"`
	Monthly       []Monthly `json:"monthly" bson:"monthly" binding:"dive"`
}

// Different type of routine
type Daily struct {
	DailyTime string   `json:"daily_time" bson:"daily_time" binding:"daily_time"`
	Action    Action `json:"action" bson:"action"`
}

type Weekly struct {
	WeeklyTime string   `json:"weekly_time" bson:"weekly_time" binding:"weekly_time"`
	Action     Action `json:"action" bson:"action"`
}

type Monthly struct {
	MonthlyTime string   `json:"monthly_time" bson:"monthly_time" binding:"monthly_time"`
	Action      Action `json:"action" bson:"action"`
}

// Action type
type Action struct {
	Type     string `json:"type" bson:"type" binding:"action_type"`
	Level    int    `json:"level" bson:"level" binding:"gte=0,lte=100"`
	Duration int    `json:"duration" bson:"duration" binding:"gte=0"`
}

const (
	waterAction = "water"
	lightAction = "light"
)

// Custom field validation
func planName(fl validator.FieldLevel) bool {
	if strings.TrimSpace(fl.Field().String()) == "" {
		return false
	}

	return true
}

func dailyTime(fl validator.FieldLevel) bool {
	field := fl.Field().String()

	dt := strings.Split(field, ":")
	if len(dt) != 2 {
		return false
	}

	if hour, err := strconv.Atoi(dt[0]); err != nil {
		return false
	} else if hour < 0 || hour > 23 {
		return false
	}

	if minute, err := strconv.Atoi(dt[1]); err != nil {
		return false
	} else if minute < 0 || minute > 59 {
		return false
	}

	return true
}

func weeklyTime(fl validator.FieldLevel) bool {
	field := fl.Field().String()

	wt := strings.Split(field, ":")
	if len(wt) != 3 {
		return false
	}

	if day, err := strconv.Atoi(wt[0]); err != nil {
		return false
	} else if day < 0 || day > 6 {
		return false
	}

	if hour, err := strconv.Atoi(wt[1]); err != nil {
		return false
	} else if hour < 0 || hour > 23 {
		return false
	}

	if minute, err := strconv.Atoi(wt[2]); err != nil {
		return false
	} else if minute < 0 || minute > 59 {
		return false
	}

	return true
}

func monthlyTime(fl validator.FieldLevel) bool {
	field := fl.Field().String()

	mt := strings.Split(field, ":")
	if len(mt) != 3 {
		return false
	}

	if date, err := strconv.Atoi(mt[0]); err != nil {
		return false
	} else if date < 0 || date > 31 {
		return false
	}

	if hour, err := strconv.Atoi(mt[1]); err != nil {
		return false
	} else if hour < 0 || hour > 23 {
		return false
	}

	if minute, err := strconv.Atoi(mt[2]); err != nil {
		return false
	} else if minute < 0 || minute > 59 {
		return false
	}

	return true
}

func actionType(fl validator.FieldLevel) bool {
	field := fl.Field().String()
	if field == waterAction || field == lightAction {
		return true
	}

	return false
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
	userId := ctx.GetString("userId")
	if userId == "" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}

	entity := &Entity{PlanId: uuid.New().String(), UserId: userId}
	if err := ctx.ShouldBindJSON(entity); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": resInvalid, "err": err.Error()})
		return
	}

	if err := h.Repo.CreatePlan(ctx, entity); err != nil {
		if err == errPlanDuplicate {
			ctx.JSON(http.StatusConflict, gin.H{"message": resPlanConflict})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		}

		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": resCreatePlan, "result": entity})
}

func (h *Handler) ListPlans(ctx *gin.Context) {
	userId := ctx.GetString("userId")
	if userId == "" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}

	entities, err := h.Repo.ListPlans(ctx, userId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": resListPlans, "result": entities})
}

func (h *Handler) GetPlan(ctx *gin.Context) {
	userId := ctx.GetString("userId")
	if userId == "" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}

	planId := ctx.Param("planId")
	if _, err := uuid.Parse(planId); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": resInvalid})
		return
	}

	entity := &Entity{
		PlanId: planId,
		UserId: userId,
	}

	if err := h.Repo.GetPlan(ctx, entity); err != nil {
		if err == errPlanNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"message": resPlanNotFound})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		}

		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": resGetPlan, "result": entity})
}

func (h *Handler) ReplacePlan(ctx *gin.Context) {
	userId := ctx.GetString("userId")
	if userId == "" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}

	entity := &Entity{PlanId: ctx.Param("planId"), UserId: userId}
	if err := ctx.ShouldBindJSON(entity); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": resInvalid})
		return
	}

	if err := h.Repo.ReplacePlan(ctx, entity); err != nil {
		if err == errPlanDuplicate {
			ctx.JSON(http.StatusConflict, gin.H{"message": resPlanConflict})
		} else if err == errPlanNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"message": resPlanNotFound})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		}

		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": resReplacePlan, "result": entity})
}

func (h *Handler) DeletePlan(ctx *gin.Context) {
	userId := ctx.GetString("userId")
	if userId == "" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}

	planId := ctx.Param("planId")
	if _, err := uuid.Parse(planId); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": resInvalid})
		return
	}

	if err := h.Repo.DeletePlan(ctx, userId, planId); err != nil {
		if err == errPlanNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"message": resPlanNotFound})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		}

		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": resDeletePlan})
}

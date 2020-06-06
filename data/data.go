package data

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tPhume/ags-backend/session"
	"net/http"
)

func RegisterRoutes(handler *Handler, engine *gin.Engine, sessionHandler *session.Handler) {
	group := engine.Group("api/v1/data")
	group.Use(sessionHandler.GetUser)

	group.GET("/:controllerId", handler.GetData)
}

// Controller Entity type represent edge device
type Entity struct {
	ControllerId string  `bson:"_id"`
	UserId       string  `bson:"user_id"`
	Temperature  float64 `bson:"temperature"`
	Humidity     float64 `bson:"humidity" validate:"gte=0,lte=100"`
	Light        float64 `bson:"light" validate:"gte=0,lte=65535"`
	SoilMoisture int     `bson:"soil_moisture" validate:"gte=0,lte=1000"`
	WaterLevel   int     `bson:"water_level" validate:"gte=0"`
}

// Repo
type Repo interface {
	GetData(ctx context.Context, entity *Entity) error
}

var (
	notFound = errors.New("not found")

	// ok message responses for handler
	resGet = "data retrieved"

	// error message responses for handler
	resInternal = "not your fault, don't worry"
	resInvalid  = "invalid values"
	resNotFound = "not found"
)

type Handler struct {
	Repo Repo
}

func (h *Handler) GetData(ctx *gin.Context) {
	userId := ctx.GetString("userId")
	if userId == "" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		return
	}

	// check controllerId
	controllerId := ctx.Param("controllerId")
	if _, err := uuid.Parse(controllerId); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": resInvalid})
		return
	}

	entity := &Entity{ControllerId: controllerId, UserId: userId}
	if err := h.Repo.GetData(ctx, entity); err != nil {
		if err == notFound {
			ctx.JSON(http.StatusNotFound, gin.H{"message": resNotFound})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": resInternal})
		}

		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": resGet, "data": entity})
}

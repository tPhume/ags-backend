package summary

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/tPhume/ags-backend/session"
	"net/http"
)

func RegisterRoutes(handler *Handler, engine *gin.Engine, sessionHandler *session.Handler) {
	group := engine.Group("api/v1/summary")
	group.Use(sessionHandler.GetUser)

	group.GET(":controllerId", handler.ListSummary)

}

type Summary struct {
	UserId             string  `json:"user_id" bson:"user_id"`
	ControllerId       string  `json:"controller_id" bson:"controller_id"`
	Date               string  `json:"date" bson:"date" bson:"date"`
	MeanHumidity       float64 `json:"mean_humidity" bson:"mean_humidity"`
	MeanLight          float64 `json:"mean_light" bson:"mean_light"`
	MeanSoilMoisture   float64 `json:"mean_soil_moisture" bson:"mean_soil_moisture"`
	MeanTemperature    float64 `json:"mean_temperature" bson:"mean_temperature"`
	MeanWaterLevel     float64 `json:"mean_water_level" bson:"mean_water_level"`
	MedianHumidity     float64 `json:"median_humidity" bson:"median_humidity"`
	MedianLight        float64 `json:"median_light" bson:"median_light"`
	MedianSoilMoisture float64 `json:"median_soil_moisture" bson:"median_soil_moisture"`
	MedianTemperature  float64 `json:"median_temperature" bson:"median_temperature"`
	MedianWaterLevel   float64 `json:"median_water_level" bson:"median_water_level"`
}

type Repo interface {
	ListSummary(ctx context.Context, userId string, controllerId string) ([]*Summary, error)
}

type Handler struct {
	Repo Repo
}

func (h *Handler) ListSummary(ctx *gin.Context) {
	// Get values
	userId := ctx.GetString("userId")
	if userId == "" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error from middleware"})
		return
	}

	controllerId := ctx.Param("controllerId")

	// Get List
	entities, err := h.Repo.ListSummary(ctx, userId, controllerId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error on retrieval"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"summary_list": entities})
}

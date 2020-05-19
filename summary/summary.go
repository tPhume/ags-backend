package summary

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

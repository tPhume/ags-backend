package plan

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

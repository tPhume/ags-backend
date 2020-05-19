package main

import (
	"context"
	"errors"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"github.com/spf13/viper"
	"github.com/tPhume/ags-backend/controller"
	"github.com/tPhume/ags-backend/plan"
	"github.com/tPhume/ags-backend/session"
	"github.com/tPhume/ags-backend/summary"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	// Read the config to Viper firs
	readConfig()

	// Get the config
	mongoUri := viper.GetString("MONGO_URI")
	mongoDb := viper.GetString("MONGO_DB")

	redisAddr := viper.GetString("REDIS_ADDR")
	redisDb := viper.GetInt("REDIS_DB")

	clientId := viper.GetString("CLIENT_ID")
	clientSecret := viper.GetString("CLIENT_SECRET")
	redirectUri := viper.GetString("REDIRECT_URI")

	failOnEmpty(mongoUri, mongoDb, redisAddr, clientId, clientSecret, redirectUri)

	// Setup Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   redisDb,
	})

	// Setup Mongo
	mongoClient, err := mongo.NewClient(options.Client().ApplyURI(mongoUri))
	failOnError("could not create mongo client", err)

	timeout, _ := context.WithTimeout(context.Background(), time.Second*10)
	err = mongoClient.Connect(timeout)

	failOnError("could not start mongo connection", err)
	timeout.Done()

	mongoDatabase := mongoClient.Database(mongoDb)

	// Setup session
	userCol := mongoDatabase.Collection("user")

	sessionRepo := &session.RedisMongo{
		UserDb:    userCol,
		SessionDb: redisClient,
	}

	sessionHandler := &session.Handler{
		Repo: sessionRepo,
	}

	// Setup controller
	controllerCol := mongoDatabase.Collection("controller")
	controllerPlanCol := mongoDatabase.Collection("plan")

	controllerPlanRepo := &controller.MongoPlanRepo{Col: controllerPlanCol}
	controllerRepo := &controller.MongoRepo{Col: controllerCol}

	controllerHandler := &controller.Handler{
		Repo:     controllerRepo,
		PlanRepo: controllerPlanRepo,
	}

	// Setup plan
	planCol := mongoDatabase.Collection("plan")
	planRepo := &plan.MongoRepo{Col: planCol, ControllerCol: controllerCol}

	planHandler := &plan.Handler{Repo: planRepo}

	// Setup summary
	summaryCol := mongoDatabase.Collection("summary")
	summaryRepo := &summary.Mongo{Col: summaryCol}

	summaryHandler := &summary.Handler{Repo: summaryRepo}

	// Setup gin
	corsConfig := cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}

	engine := gin.New()
	engine.Use(cors.New(corsConfig))

	session.RegisterRoutes(sessionHandler, engine)
	controller.RegisterRoutes(controllerHandler, engine, sessionHandler)
	plan.RegisterRoutes(planHandler, engine, sessionHandler)
	summary.RegisterRoutes(summaryHandler, engine, sessionHandler)

	log.Fatal(engine.Run("0.0.0.0:9700"))
}

func readConfig() {
	// Set and read configurations
	viper.SetConfigFile(os.Args[1])
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	failOnError("could not read config", err)
}

func failOnEmpty(values ...string) {
	for _, v := range values {
		if strings.TrimSpace(v) == "" {
			failOnError("some values are empty", errors.New("improper config file"))
		}
	}
}

func failOnError(msg string, err error) {
	if err != nil {
		log.Fatalf("%s:%s", msg, err)
	}
}

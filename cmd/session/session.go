package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"github.com/spf13/viper"
	"github.com/tPhume/ags-backend/session"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
)

func main() {
	// Set env config
	viper.SetConfigFile("session.env")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	// Create GoogleApi
	googleApi := &session.GoogleApi{
		ClientId:     viper.GetString("CLIENT_ID"),
		ClientSecret: viper.GetString("CLIENT_SECRET"),
		RedirectUri:  viper.GetString("REDIRECT_URI"),
	}

	// Create redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     viper.GetString("REDIS_ADDR"),
		Password: viper.GetString("REDIS_PASSWORD"),
		DB:       viper.GetInt("REDIS_DB"),
	})

	// Create mongo client and collection
	mongoClient, err := mongo.NewClient(options.Client().ApplyURI(viper.GetString("MONGO_URI")))
	failOnError("could not create mongodb client", err)

	err = mongoClient.Connect(context.Background())
	failOnError("could not create connection to mongodb", err)

	db := viper.GetString("MONGO_DATABASE")
	col := viper.GetString("MONGO_COLLECTION")

	if db == "" || col == "" {
		log.Fatal("missing env")
	}

	mongoCollection := mongoClient.Database(db).Collection(col)

	// Create RedisMongo
	redisMongo := &session.RedisMongo{UserDb: mongoCollection, SessionDb: redisClient}

	// Create handler
	handler := &session.Handler{
		Domain:     "localhost",
		Repo:       redisMongo,
		GoogleRepo: googleApi,
	}

	// Init Gin Engine
	engine := gin.New()
	session.RegisterRoutes(handler, engine)

	engine.GET("/ping", func(ctx *gin.Context) {
		if err := mongoClient.Ping(ctx, nil); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "database not connected"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"message": "ready"})
	})

	ping := engine.Group("/api/v1/session/ping")
	ping.Use(handler.GetUser)
	ping.GET("", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": ctx.GetString("userId")})
	})

	log.Fatal(engine.Run("0.0.0.0:9700"))
}

func failOnError(msg string, err error) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

package main

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"github.com/tPhume/ags-backend/session"
	"log"
	"os"
)

func main() {
	// Set environment configurations
	viper.SetConfigFile("google.env")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	failOnError("could not read in env", err)

	// Get environment variable
	googleApi := session.GoogleApi{
		ClientId:     viper.GetString("CLIENT_ID"),
		ClientSecret: viper.GetString("CLIENT_SECRET"),
		RedirectUri:  viper.GetString("REDIRECT_URI"),
	}

	userEntity := session.UserEntity{}
	err = googleApi.GetIdToken(context.Background(), os.Args[1], &userEntity)
	failOnError("something went wrong", err)

	fmt.Println(userEntity)
}

func failOnError(msg string, err error) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

package main

import (
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
	"log"
	"time"
)

func main() {
	// Set environment variables
	viper.SetConfigFile("rabbitmq.env")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	failOnError("could not read in env", err)

	// Get environment variables
	amqpUri := viper.Get("amqp_uri").(string)

	// Connect to RabbitMQ
	conn, err := amqp.Dial(amqpUri)
	failOnError("fail to connect to rabbitmq", err)

	ch, err := conn.Channel()
	failOnError("fail to open a channel", err)

	// we will assume that the queue have already been declared explciitly before
	currentTime, err := time.Now().MarshalBinary()
	failOnError("could not marshal time", err)

	err = ch.Publish(
		"backend.direct.test",
		"backend.direct.test",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        currentTime,
		},
	)
	failOnError("could not publish message", err)

	log.Println("Direct message published")
}

func failOnError(msg string, err error) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

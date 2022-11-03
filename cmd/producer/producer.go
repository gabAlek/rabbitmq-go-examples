package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"rabbitmq-go/pkg/rabbit"

	"github.com/gofiber/fiber/v2"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	log.Println("Trying to connect...")
	address := os.Getenv("RABBITMQ_ADDRESS")
	username := os.Getenv("RABBITMQ_USERNAME")
	password := os.Getenv("RABBITMQ_PASSWORD")

	rabbitMqSvc := rabbit.Connect(username, password, address)
	channel := rabbitMqSvc.Channel
	defer channel.Close()

	//default exchange
	queue := rabbitMqSvc.CreateSimpleQueueDefaultExchange("Queue1", false)
	workerQueue := rabbitMqSvc.CreateSimpleQueueDefaultExchange("WorkerQueueDurable", true)

	rabbitMqSvc.ExchangeDeclare("logs", "fanout", true)
	rabbitMqSvc.ExchangeDeclare("logs_direct", "direct", true)
	rabbitMqSvc.ExchangeDeclare("logs_topic", "topic", true)

	app := fiber.New()

	app.Get("/send", func(c *fiber.Ctx) error {
		message := c.Query("msg") //messages can be sent by "?msg=" query param
		publish(message, queue.Name, channel)
		return nil
	})

	app.Get("/sendToWorker", func(c *fiber.Ctx) error {
		message := strconv.Itoa(rand.Intn(1000))
		publishPersistent(message, workerQueue.Name, channel)
		return nil
	})

	app.Get("/sendFanout", func(c *fiber.Ctx) error {
		message := strconv.Itoa(rand.Intn(1000))
		publishFanout(message, "logs", channel)
		return nil
	})

	app.Get("/sendDirect", func(c *fiber.Ctx) error {
		message := strconv.Itoa(rand.Intn(1000))
		severity := c.Query("severity")
		if severity != "info" && severity != "crit" {
			log.Println("Severity must be 'info' or 'crit'")
			return nil
		}
		publishDirectOrTopic(message, "sev."+severity, "logs_direct", channel)
		return nil
	})

	app.Get("/sendTopic", func(c *fiber.Ctx) error {
		message := strconv.Itoa(rand.Intn(1000))
		severity := c.Query("severity")
		if severity != "info" && severity != "crit" {
			log.Println("Severity must be 'info' or 'crit'")
			return nil
		}
		publishDirectOrTopic(message, "sev."+severity, "logs_topic", channel)
		return nil
	})
	log.Fatal(app.Listen(":3000"))
}

func publish(message, queueName string, channel *amqp.Channel) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := channel.PublishWithContext(ctx,
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
	failOnError(err, "Fail to publish message")
}

func publishPersistent(message, queueName string, channel *amqp.Channel) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := channel.PublishWithContext(ctx,
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(message),
		})
	failOnError(err, "Fail to publish message")
}

func publishFanout(message, exchangeName string, channel *amqp.Channel) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := channel.PublishWithContext(ctx,
		exchangeName,
		"",
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(message),
		})
	failOnError(err, "Fail to publish message")
}

func publishDirectOrTopic(message, routingKey, exchangeName string, channel *amqp.Channel) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := channel.PublishWithContext(ctx,
		exchangeName,
		routingKey,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(message),
		})
	failOnError(err, "Fail to publish message")
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

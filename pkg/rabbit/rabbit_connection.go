package rabbit

import (
	"fmt"
	"log"

	errHandler "rabbitmq-go/pkg/error"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitSvc struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
}

func Connect(username string, password string, address string) *RabbitSvc {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s/", username, password, address))
	errHandler.FailOnError(err, "Fail to connect to RabbitMQ")

	log.Println("Connected to RabbitMQ!")

	channel, err := conn.Channel()
	errHandler.FailOnError(err, "Fail to open connection to channel")
	return &RabbitSvc{
		Connection: conn,
		Channel:    channel,
	}
}

func (r *RabbitSvc) ExchangeDeclare(name string, exchangeType string, durable bool) {
	err := r.Channel.ExchangeDeclare(
		name,
		exchangeType,
		true,
		false,
		false,
		false,
		nil,
	)
	errHandler.FailOnError(err, "Fail to declare an exchange")
}

func (r *RabbitSvc) CreateSimpleQueueDefaultExchange(name string, durable bool) amqp.Queue {
	queue, err := r.Channel.QueueDeclare(
		name,
		durable,
		false,
		false,
		false,
		nil,
	)
	errHandler.FailOnError(err, "Fail to declare a queue")
	return queue
}

func (r *RabbitSvc) CreateUnnamedQueueBoundToFanoutExchange(durable bool, exchange string) amqp.Queue {
	queue, err := r.Channel.QueueDeclare(
		"",
		durable,
		false,
		false,
		false,
		nil,
	)
	errHandler.FailOnError(err, "Fail to declare a queue")

	err = r.Channel.QueueBind(
		queue.Name,
		"",
		"logs",
		false,
		nil,
	)
	errHandler.FailOnError(err, "Fail to bind queue to exchange")

	return queue
}

func (r *RabbitSvc) CreateSimpleQueueDirectOrTopicExchange(name string, durable bool, exchange string, routingKey string) amqp.Queue {
	queue, err := r.Channel.QueueDeclare(
		name,
		durable,
		false,
		false,
		false,
		nil,
	)
	errHandler.FailOnError(err, "Fail to declare a queue")

	err = r.Channel.QueueBind(
		queue.Name,
		routingKey,
		exchange,
		false,
		nil,
	)
	errHandler.FailOnError(err, "Fail to bind queue to exchange")
	return queue
}

func (r *RabbitSvc) ConsumeMessages(queue amqp.Queue, autoAck bool) <-chan amqp.Delivery {
	msgs, err := r.Channel.Consume(
		queue.Name,
		"",
		autoAck,
		false,
		false,
		false,
		nil,
	)
	errHandler.FailOnError(err, "Fail to register a consumer")
	return msgs
}

func (r *RabbitSvc) ConsumeFanoutMessages(queue amqp.Queue) <-chan amqp.Delivery {
	msgs, err := r.Channel.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	errHandler.FailOnError(err, "Fail to register a consumer")
	return msgs
}

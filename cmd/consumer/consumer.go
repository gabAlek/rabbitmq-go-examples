package main

import (
	"log"
	"math/rand"
	"os"
	errHandler "rabbitmq-go/pkg/error"
	"rabbitmq-go/pkg/rabbit"
	"time"

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

	err := channel.Qos(
		1, // server will deliver no more than one message before receiving acks
		0,
		false,
	)
	errHandler.FailOnError(err, "Fail to set QoS")

	forever := make(chan struct{})

	//testing simplest case - default exchange
	queue := rabbitMqSvc.CreateSimpleQueueDefaultExchange("Queue1", false)
	msgs := rabbitMqSvc.ConsumeMessages(queue, true)
	go receiveMessages(queue.Name, msgs)

	//testing with multiple workers - default exchange
	workerQueue := rabbitMqSvc.CreateSimpleQueueDefaultExchange("WorkerQueueDurable", true)
	workerMsgs := rabbitMqSvc.ConsumeMessages(workerQueue, false)
	go receiveWithSleep(workerQueue.Name, workerMsgs)

	//testing fanout exchange
	unnamedQueue1 := rabbitMqSvc.CreateUnnamedQueueBoundToFanoutExchange(true, "logs")
	fanoutMsgs1 := rabbitMqSvc.ConsumeFanoutMessages(unnamedQueue1)
	go receiveMessages(unnamedQueue1.Name, fanoutMsgs1)

	unnamedQueue2 := rabbitMqSvc.CreateUnnamedQueueBoundToFanoutExchange(true, "logs")
	fanoutMsgs2 := rabbitMqSvc.ConsumeFanoutMessages(unnamedQueue2)
	go receiveMessages(unnamedQueue2.Name, fanoutMsgs2)

	//testing direct exchange
	infoLogsQueue := rabbitMqSvc.CreateSimpleQueueDirectOrTopicExchange("LOGS_INFO", false, "logs_direct", "sev.info")
	infoMsgs := rabbitMqSvc.ConsumeMessages(infoLogsQueue, true)
	go receiveMessages(infoLogsQueue.Name, infoMsgs)

	criticalLogsQueue := rabbitMqSvc.CreateSimpleQueueDirectOrTopicExchange("LOGS_CRIT", false, "logs_direct", "sev.crit")
	criticalMsgs := rabbitMqSvc.ConsumeMessages(criticalLogsQueue, true)
	go receiveMessages(criticalLogsQueue.Name, criticalMsgs)

	//testing topic exchange
	allLogsQueue := rabbitMqSvc.CreateSimpleQueueDirectOrTopicExchange("LOGS_ALL", false, "logs_topic", "sev.*")
	allMsgs := rabbitMqSvc.ConsumeMessages(allLogsQueue, true)
	go receiveMessages(allLogsQueue.Name, allMsgs)

	log.Println("Waiting for messages. To exit, press Ctrl + C")
	<-forever
}

func receiveMessages(queueName string, msgs <-chan amqp.Delivery) {
	for d := range msgs {
		log.Printf("Received a message on queue %q: %s\n", queueName, d.Body)
	}
}

func receiveWithSleep(queueName string, msgs <-chan amqp.Delivery) {
	for d := range msgs {
		log.Printf("Received a message on queue %q: %s\n", queueName, d.Body)
		sleepTime := time.Duration(rand.Intn(6))
		log.Printf("Sleeping for %d seconds...", sleepTime)
		time.Sleep(sleepTime * time.Second)
		log.Println("Done!")
		d.Ack(false)
	}
}

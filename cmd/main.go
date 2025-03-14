package main

import (
	"context"
	"log"

	"github.com/matevz-ap/weather/internal/weather"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/valkey-io/valkey-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	client, err := valkey.NewClient(valkey.ClientOption{InitAddress: []string{"valkey:6379"}})
	if err != nil {
		panic(err)
	}
	defer client.Close()

	ctx := context.Background()

	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"weather", // name
		"fanout",  // type
		true,      // durable
		false,     // auto-deleted
		false,     // internal
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	request_queue, err := ch.QueueDeclare(
		"weather", // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		request_queue.Name, // queue name
		"",                 // routing key
		"weather",          // exchange
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")

	msgs, err := ch.Consume(
		request_queue.Name, // queue
		"",                 // consumer
		true,               // auto-ack
		false,              // exclusive
		false,              // no-local
		false,              // no-wait
		nil,                // args
	)
	failOnError(err, "Failed to register a consumer")
	var forever chan struct{}
	go func() {
		for d := range msgs {
			location := string(d.Body)

			data, err := client.Do(ctx, client.B().Get().Key(location).Build()).ToString()
			if err != nil {
				data, err = weather.Weather(location)
				if err != nil {
					log.Println("Problems with weather", err)
				}
				err = client.Do(ctx, client.B().Set().Key(location).Value(data).Build()).Error()
				if err != nil {
					log.Print("Problem with setting cache", err)
				}
			} else {
				log.Printf("Returning results from cache for: %s", location)
			}

			_, err = ch.QueueDeclare(
				d.CorrelationId, // name
				false,           // durable
				true,            // delete when unused
				false,           // exclusive
				false,           // no-wait
				nil,             // arguments
			)
			failOnError(err, "Failed to declare a queue")

			err = ch.PublishWithContext(ctx,
				"",
				d.CorrelationId,
				false,
				false,
				amqp.Publishing{
					ContentType: "application/json",
					Body:        []byte(data),
				},
			)

			failOnError(err, "Failed to publish results")
		}
	}()

	log.Printf("[*] Waiting for messages.")
	<-forever
}

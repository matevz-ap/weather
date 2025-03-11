package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Response struct {
	Results []Location
}

type Location struct {
	Latitude  float64
	Longitude float64
}

func geocode(location string) (float64, float64, error) {
	url := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1&language=en&format=json", location)
	res, err := http.Get(url)

	if err != nil {
		return 0, 0, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, 0, err
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, 0, err
	}

	if len(response.Results) == 0 {
		return 0, 0, err
	}

	return response.Results[0].Latitude, response.Results[0].Longitude, nil
}

func location_weather(lat float64, long float64) (string, error) {
	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&hourly=temperature_2m", lat, long)
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func weather(location string) (string, error) {
	long, lat, err := geocode(location)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	weather_data, err := location_weather(lat, long)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return weather_data, nil
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
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

	_, err = ch.QueueDeclare(
		"weather_responses", // name
		false,               // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
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
			log.Printf("Fetching weather for: %s", d.Body)
			data, err := weather(string(d.Body))
			if err != nil {
				log.Println("problems with weather")
			}
			err = ch.Publish(
				"",
				"weather_responses",
				false,
				false,
				amqp.Publishing{
					DeliveryMode: amqp.Persistent,
					ContentType:  "application/json",
					Body:         []byte(data),
				},
			)
			failOnError(err, "Failed to publish results")
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

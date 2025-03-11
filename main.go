package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

func handler(w http.ResponseWriter, r *http.Request) {
	res, err := http.Get("https://api.open-meteo.com/v1/forecast?latitude=46.0511&longitude=14.5051&hourly=temperature_2m")
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error reading response body"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(body)
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

	// We create a Queue to send the message to.
	q, err := ch.QueueDeclare(
		"golang-queue", // name
		false,          // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	failOnError(err, "Failed to declare a queue")

	// We set the payload for the message.
	body := "Golang is awesome - Keep Moving Forward!"
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	// If there is an error publishing the message, a log will be displayed in the terminal.
	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Congrats, sending message: %s", body)
	http.HandleFunc("/weather", handler)
	http.ListenAndServe(":8080", nil)
}

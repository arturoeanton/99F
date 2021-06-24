package event

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
)

func Emmit(eventName string, name string, id string, elem map[string]interface{}) error {
	var err error
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return err
	}

	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Println(err.Error())
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		name,    // name
		"topic", // type
		true,    // durable
		false,   // auto-deleted
		false,   // internal
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	jsonContentStr, _ := json.Marshal(elem)
	err = ch.Publish(
		name,                      // exchange
		name+"."+id+"."+eventName, // routing key
		false,                     // mandatory
		false,                     // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        jsonContentStr,
		})
	if err != nil {
		log.Println(err.Error())
		return err
	}

	return nil
}

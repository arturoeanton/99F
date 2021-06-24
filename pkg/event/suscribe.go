package event

import (
	"encoding/json"
	"log"

	"github.com/arturoeanton/gocommons/utils"
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"github.com/dop251/goja_nodejs/util"
	"github.com/streadway/amqp"
)

var (
	subcribe map[string]chan bool = make(map[string]chan bool)
	registry *require.Registry
)

func Suscribe(nameSrc string, nameTo string, routingKey string, id string, self map[string]interface{}) error {
	_, ok := subcribe[id]
	if !ok {
		subcribe[id] = make(chan bool)
		defer delete(subcribe, nameTo)

		conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
		if err != nil {
			log.Println(err.Error())
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
			nameSrc, // name
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

		q, err := ch.QueueDeclare(
			"",    // name
			false, // durable
			false, // delete when unused
			true,  // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			log.Println(err.Error())
			return err
		}

		err = ch.QueueBind(
			q.Name,     // queue name
			routingKey, // routing key
			nameSrc,    // exchange
			false,
			nil)
		if err != nil {
			log.Println(err.Error())
			return err
		}

		msgs, err := ch.Consume(
			q.Name, // queue
			"",     // consumer
			true,   // auto ack
			false,  // exclusive
			false,  // no local
			false,  // no wait
			nil,    // args
		)
		if err != nil {
			log.Println(err.Error())
			return err
		}

		vm := goja.New()
		if registry == nil {
			registry = new(require.Registry) // this can be shared by multiple runtimes
			registry.RegisterNativeModule("console", console.Require)
			registry.RegisterNativeModule("util", util.Require)
		}

		registry.Enable(vm)
		console.Enable(vm)
		go func() {
			for d := range msgs {
				//log.Printf(" [x] %s -> %s", d.Body, nameTo)
				if utils.Exists("entities/" + nameTo + "/notify.js") {

					code, err := utils.FileToString("entities/" + nameTo + "/notify.js")
					if err != nil {
						continue
					}
					code += "\n notify_" + nameSrc + "()"

					var elem map[string]interface{}
					err = json.Unmarshal(d.Body, &elem)
					if err != nil {
						elem = make(map[string]interface{})
					}

					vm.Set("schema_src", nameSrc)
					vm.Set("schema", nameTo)
					vm.Set("id", id)
					vm.Set("self", self)

					vm.Set("message", d)
					vm.Set("routingKey", d.RoutingKey)
					vm.Set("routingKeyFilter", routingKey)
					vm.Set("payload", elem)

					_, err = vm.RunString(code)
					if err != nil {
						log.Println(err.Error())
						continue
					}
				}
			}
		}()
		<-subcribe[id]
	}
	return nil
}

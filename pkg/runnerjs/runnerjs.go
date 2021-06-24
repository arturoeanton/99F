package runnerjs

import (
	"crypto/md5"
	"encoding/json"
	"log"

	"github.com/arturoeanton/99F/pkg/event"
	"github.com/arturoeanton/99F/pkg/jsonschema"
	"github.com/arturoeanton/gocommons/utils"
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"github.com/dop251/goja_nodejs/util"
	"github.com/gofiber/fiber/v2"
)

var (
	registry *require.Registry
	semVM        = make(chan int, 50) // 70 aguanto - 80 no aguanto
	isem     int = 0
)

func Run(c *fiber.Ctx, element map[string]interface{}, schema *jsonschema.JSONSchame, id string, nameSchema string, method string, fx string) error {

	code, err := utils.FileToString("entities/" + nameSchema + "/" + method + ".js")
	if err != nil {
		return err
	}
	code += "\n" + fx + "()"
	vm := goja.New()
	if registry == nil {
		registry = new(require.Registry) // this can be shared by multiple runtimes
		registry.RegisterNativeModule("console", console.Require)
		registry.RegisterNativeModule("util", util.Require)
	}

	registry.Enable(vm)
	console.Enable(vm)

	vm.Set("fiberCtx", c)
	vm.Set("self", element)
	vm.Set("schema", schema)
	vm.Set("name", nameSchema)
	vm.Set("id", id)

	vm.Set("eventSuscribe", event.Suscribe)

	jsoncontent, _ := json.Marshal(element)
	Md5Inst := md5.New()
	Md5Inst.Write([]byte(jsoncontent))
	md5B := string(Md5Inst.Sum([]byte("")))

	err = func() error {
		defer func() {
			err := recover()
			if err != nil {
				log.Println("runJs_00010 ****", err)
			}
		}()
		semVM <- 1
		isem++
		_, err := vm.RunString(code)
		if err != nil {
			log.Println("runJs_00011 ****", err)
		}
		isem--
		<-semVM
		return err
	}()

	jsoncontent, _ = json.Marshal(element)
	Md5Inst.Write([]byte(jsoncontent))
	md5A := string(Md5Inst.Sum([]byte("")))

	if md5A != md5B {
		element, err = jsonschema.Bind(nameSchema, jsoncontent)
		if err != nil {
			return err
		}
		_, err = jsonschema.Replace(element, id, nameSchema)
		if err != nil {
			return err
		}
	}

	return err
}

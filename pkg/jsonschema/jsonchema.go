package jsonschema

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"time"
)

type JSONSchame struct {
	Schema           *string                `json:"$schema,omitempty"`
	ID               *string                `json:"$id,omitempty"`
	Title            *string                `json:"title,omitempty"`
	Description      *string                `json:"description,omitempty"`
	Type             string                 `json:"type"`
	Properties       *map[string]JSONSchame `json:"properties,omitempty"`
	Required         *[]string              `json:"required,omitempty"`
	ExclusiveMinimum *float64               `json:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum *float64               `json:"exclusiveMaximum,omitempty"`
	Minimum          *float64               `json:"minimum,omitempty"`
	Maximum          *float64               `json:"maximum,omitempty"`
	MultipleOf       *float64               `json:"multipleOf,omitempty"`
	MinLength        *int                   `json:"minLength,omitempty"`
	MaxLength        *int                   `json:"maxLength,omitempty"`
	UniqueItems      *bool                  `json:"uniqueItems,omitempty"`
	Format           *string                `json:"format,omitempty"`
	Example          *string                `json:"example,omitempty"`
	Regex            *string                `json:"regex,omitempty"`
	Pattern          *string                `json:"pattern,omitempty"`
	Enum             *[]interface{}         `json:"enum,omitempty"`
	Const            *interface{}           `json:"const,omitempty"`
	Contains         *JSONSchame            `json:"contains,omitempty"`
	Items            *interface{}           `json:"items,omitempty"`
	AdditionalItems  *interface{}           `json:"additionalItems,omitempty"`
	MinItems         *int                   `json:"minItems,omitempty"`
	MaxItems         *int                   `json:"maxItems,omitempty"`
	MaxContains      *int                   `json:"maxContains,omitempty"`
	MinContains      *int                   `json:"minContains,omitempty"`
	MaxProperties    *int                   `json:"maxProperties,omitempty"`
	MinProperties    *int                   `json:"minProperties,omitempty"`

	//no standard
	Strict *bool `json:"_strict,omitempty"`
}

var (
	emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	uuidRegex  = regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
)

func GetSchema(name string) (*JSONSchame, error) {
	file, err := ioutil.ReadFile("entities/" + name + "/schema.json")
	if err != nil {
		return nil, err
	}
	data := JSONSchame{}
	err = json.Unmarshal([]byte(file), &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func Bind(name string, body []byte) (map[string]interface{}, error) {
	schema, err := GetSchema(name)
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	err = validate(data, *schema, "root")
	if err != nil {
		return nil, err
	}

	return data, nil
}

func validate(data interface{}, schema JSONSchame, fieldName string) error {

	if schema.Enum != nil {
		flag := false
		for _, d := range *schema.Enum {
			if d == data {
				flag = true
				break
			}
		}
		if !flag {
			return errors.New(fieldName + " is enum " + fmt.Sprint(*schema.Enum))
		}
	}

	if schema.Const != nil {
		if *schema.Const != data {
			return errors.New(fieldName + " is const " + fmt.Sprint(*schema.Const))
		}
	}

	if schema.Type == "object" {
		item := data.(map[string]interface{})
		for _, v := range *schema.Required {
			if _, ok := item[v]; !ok {
				return errors.New("not found " + fmt.Sprint(v))
			}
		}
		lenProperties := 0
		if schema.Properties != nil {
			for k, v := range *schema.Properties {
				if subItem, ok := item[k]; ok {
					err := validate(subItem, v, k)
					if err != nil {
						return err
					}
				}
			}
			for k := range item {
				lenProperties++
				if schema.Strict != nil {
					if *schema.Strict {
						if _, ok := (*schema.Properties)[k]; !ok {
							return errors.New("not found in schema " + fmt.Sprint(k))
						}
					}
				}

			}
		}

		if schema.MaxProperties != nil {
			if lenProperties > *schema.MaxProperties {
				return errors.New(fieldName + " has maxProperties " + fmt.Sprint(*schema.MaxProperties))
			}
		}
		if schema.MinProperties != nil {
			if lenProperties < *schema.MinProperties {
				return errors.New(fieldName + " has minProperties " + fmt.Sprint(*schema.MinProperties))
			}
		}

	}

	if schema.Type == "number" {
		item, ok := data.(float64)
		if !ok {
			return errors.New(fieldName + " is not number: " + fmt.Sprint(data))
		}
		if schema.ExclusiveMinimum != nil {
			if item <= *schema.ExclusiveMinimum {
				return errors.New(fieldName + " should to be greater than  " + fmt.Sprint(*schema.ExclusiveMinimum))
			}
		}
		if schema.ExclusiveMaximum != nil {
			if item >= *schema.ExclusiveMaximum {
				return errors.New(fieldName + " should to be less than  " + fmt.Sprint(*schema.ExclusiveMaximum))
			}
		}

		if schema.Minimum != nil {
			if item < *schema.Minimum {
				return errors.New(fieldName + " should to be greater or equals  than  " + fmt.Sprint(*schema.Minimum))
			}
		}
		if schema.Maximum != nil {
			if item > *schema.Maximum {
				return errors.New(fieldName + " should to be less or equals than  " + fmt.Sprint(*schema.Maximum))
			}
		}

		if schema.MultipleOf != nil {
			if math.Mod(item, *schema.MultipleOf) != 0 {
				return errors.New(fieldName + " should to be Multiple Of  " + fmt.Sprint(*schema.MultipleOf))
			}
		}
	}

	if schema.Type == "integer" {
		item, ok := data.(float64)
		if !ok {
			return errors.New(fieldName + " is not integer: " + fmt.Sprint(data))
		}
		if item != float64(int(item)) {
			return errors.New(fieldName + " is not integer: " + fmt.Sprint(data))
		}

		if schema.ExclusiveMinimum != nil {
			if item <= *schema.ExclusiveMinimum {
				return errors.New(fieldName + " should to be greater than  " + fmt.Sprint(*schema.ExclusiveMinimum))
			}
		}
		if schema.ExclusiveMaximum != nil {
			if item >= *schema.ExclusiveMaximum {
				return errors.New(fieldName + " should to be less than  " + fmt.Sprint(*schema.ExclusiveMaximum))
			}
		}

		if schema.Minimum != nil {
			if item < *schema.Minimum {
				return errors.New(fieldName + " should to be greater or equals  than  " + fmt.Sprint(*schema.Minimum))
			}
		}
		if schema.Maximum != nil {
			if item > *schema.Maximum {
				return errors.New(fieldName + " should to be less or equals than  " + fmt.Sprint(*schema.Maximum))
			}
		}
		if schema.MultipleOf != nil {
			if math.Mod(item, *schema.MultipleOf) != 0 {
				return errors.New(fieldName + " should to be Multiple Of  " + fmt.Sprint(*schema.MultipleOf))
			}
		}
	}

	if schema.Type == "boolean" {
		_, ok := data.(string)
		if !ok {
			return errors.New(fieldName + " is not boolean: " + fmt.Sprint(data))
		}
	}

	if schema.Type == "string" {
		item, ok := data.(string)
		if !ok {
			return errors.New(fieldName + " is not string: " + fmt.Sprint(data))
		}

		if schema.MaxLength != nil {
			if len(item) > *schema.MaxLength {
				return errors.New(fieldName + " has MaxLength " + fmt.Sprint(*schema.MaxLength))
			}
		}

		if schema.MinLength != nil {
			if len(item) < *schema.MinLength {
				return errors.New(fieldName + " has MinLength " + fmt.Sprint(*schema.MinLength))
			}
		}

		if schema.Format != nil {
			if *schema.Format == "email" {
				_, err := mail.ParseAddress(item)
				if err != nil {
					return errors.New(fieldName + ": Format should to be  " + *schema.Format)
				}
				if !emailRegex.MatchString(item) {
					return errors.New(fieldName + ": Format should to be  " + *schema.Format)
				}
			}

			if *schema.Format == "ipv4" {
				ip := net.ParseIP(item)
				if ip == nil || ip.To4() == nil {
					return errors.New(fieldName + ": Format should to be  " + *schema.Format)
				}
			}

			if *schema.Format == "ipv6" {
				ip := net.ParseIP(item)
				if ip == nil || ip.To4() != nil {
					return errors.New(fieldName + ": Format should to be  " + *schema.Format)
				}
			}

			if *schema.Format == "url" {
				_, err := url.ParseRequestURI(item)
				if err != nil {
					return errors.New(fieldName + ": Format should to be  " + *schema.Format)
				}
			}

			if *schema.Format == "uuid" {
				if !uuidRegex.MatchString(item) {
					return errors.New(fieldName + ": Format should to be  " + *schema.Format)
				}
			}

			if *schema.Format == "date" {
				layout := ""
				if schema.Example == nil {
					layout = "2006-01-02"
				} else {
					layout = *schema.Example
				}
				var _, err = time.Parse(layout, item)
				if err != nil {
					return errors.New(fieldName + ": Format should to be  " + *schema.Format + " - Example:" + layout)
				}
			}
		}

		if schema.Pattern != nil {
			rx := regexp.MustCompile(*schema.Pattern)
			if !rx.MatchString(item) {
				return errors.New(fieldName + ": Format should to be  " + *schema.Pattern)
			}
		}
		if schema.Regex != nil {
			rx := regexp.MustCompile(*schema.Regex)
			if !rx.MatchString(item) {
				return errors.New(fieldName + ": Format should to be  " + *schema.Regex)
			}
		}

	}

	if schema.Type == "array" {
		item, ok := data.([]interface{})
		if !ok {
			return errors.New(fieldName + " is not array: " + fmt.Sprint(data))
		}
		if schema.MinItems != nil {
			if len(item) < *schema.MinItems {
				return errors.New(fieldName + " should be min  " + fmt.Sprint(*schema.MinItems))
			}
		}

		if schema.MaxItems != nil {
			if len(item) > *schema.MaxItems {
				return errors.New(fieldName + " should be max  " + fmt.Sprint(*schema.MaxItems))

			}
		}
		if schema.UniqueItems != nil {
			if *schema.UniqueItems {
				collec := make(map[interface{}]int)
				for _, v := range item {
					if _, ok := collec[v]; !ok {
						collec[v] = 0
					}
					collec[v]++
					if collec[v] > 1 {
						return errors.New(fieldName + " should be unique repite:" + fmt.Sprint(v))
					}
				}
			}
		}
		if schema.Items != nil {
			schemaItemsArray, ok := (*schema.Items).([]interface{})
			if ok {
				for i, shmap := range schemaItemsArray {
					if i >= len(item) {
						break
					}
					jsonString, _ := json.Marshal(shmap)
					var sh JSONSchame
					json.Unmarshal(jsonString, &sh)
					err := validate(item[i], sh, fieldName)
					if err != nil {
						return err
					}
				}
				if schema.AdditionalItems != nil {
					flag, ok := (*schema.AdditionalItems).(bool)
					if ok {
						if !flag {
							if len(item) > len(schemaItemsArray) {
								return errors.New(fieldName + " has  additionalItems false")
							}
						}
					} else {
						jsonString, _ := json.Marshal(*schema.AdditionalItems)
						var sh JSONSchame
						json.Unmarshal(jsonString, &sh)
						for i := len(schemaItemsArray); i < len(item); i++ {
							err := validate(item[i], sh, fieldName)
							if err != nil {
								return err
							}
						}
					}
				}
			} else {
				jsonString, _ := json.Marshal(*schema.Items)
				var sh JSONSchame
				json.Unmarshal(jsonString, &sh)
				for _, v := range item {
					err := validate(v, sh, fieldName)
					if err != nil {
						return err
					}
				}
			}
		}

		if schema.Contains != nil {
			collec := 0
			flagContains := false
			for _, v := range item {
				err := validate(v, *schema.Contains, fieldName)
				if err == nil {
					collec++
					if !flagContains {
						flagContains = true
					}
				}
			}
			if !flagContains {
				jsonString, _ := json.Marshal(*schema.Contains)
				return errors.New(fieldName + " has contatins " + string(jsonString))
			}
			if collec < *schema.MinContains {
				return errors.New(fieldName + " has min contatins " + fmt.Sprint(*schema.MinContains))

			}
			if collec > *schema.MaxContains {
				return errors.New(fieldName + " has max contatins " + fmt.Sprint(*schema.MaxContains))
			}

		}
	}
	return nil
}

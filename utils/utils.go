package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

// ErrorResponse : This is error model.
type ErrorResponse struct {
	StatusCode   int    `json:"status"`
	ErrorMessage string `json:"message"`
}

// SuccessResponse : This is success model.
type SuccessResponse struct {
	StatusCode int         `json:"status"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
}

// GetError : This is helper function to prepare error model.
func GetError(err error, StatusCode int, w http.ResponseWriter) {
	var response = ErrorResponse{
		ErrorMessage: err.Error(),
		StatusCode:   StatusCode,
	}

	w.WriteHeader(response.StatusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error sending response: %v", err)
	}
}

// GetSuccess : This is helper function to prepare success model.
func GetSuccess(msg string, data interface{}, w http.ResponseWriter) {
	var response = SuccessResponse{
		Message:    msg,
		StatusCode: http.StatusOK,
		Data:       data,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error sending response: %v", err)
	}
}

// get env vars; return empty string if not found
func Env(key string) string {
	return os.Getenv(key)
}

// check if a file exists, useful in checking for .env
func FileExists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

// convert map to bson.M for mongoDB docs
func MapToBson(data map[string]interface{}) bson.M {
	return bson.M(data) // they have the same underlying type so type conversion is enough
}

// MapToStruct converts generic map[string]interface{} to a struct, passing a pointer to the struct
// This uses struct tags instead of field names since StructToMap also uses tags
// USE AT YOUR RISK
func MapToStruct(m map[string]interface{}, s interface{}, tag string) error {
	v := reflect.ValueOf(s)

	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("s should be a pointer to struct, got: %T", v)
	}

	v = v.Elem()

	// StructToMap uses tagValue as the map key for a struct field,
	// A map is used here to store key:value pair of tagValue:fieldName
	tagToFieldName := mapStructTagValueToFieldName(v, tag)

	for name, value := range m {
		structFieldValue := v.FieldByName(tagToFieldName[name])

		if !structFieldValue.IsValid() {
			return fmt.Errorf("No such field: %s in struct", name)
		}
		if !structFieldValue.CanSet() {
			return fmt.Errorf("Cannot set %s field value", name)
		}

		val := reflect.ValueOf(value)

		if structFieldValue.Type() != val.Type() {
			return fmt.Errorf("Provided value didn't match struct field type")

		}

		structFieldValue.Set(val)
	}
	return nil
}

func mapStructTagValueToFieldName(v reflect.Value, tag string) map[string]string {
	sType := v.Type()
	m := make(map[string]string)
	for i := 0; i < v.NumField(); i++ {
		field := sType.Field(i)
		if tagVal := field.Tag.Get(tag); !shouldOmitTag(tagVal) {
			m[tagVal] = field.Name
		}
	}
	return m
}

// StructToMap converts a struct of any type to a map[string]inteface{} based on struct tags
// The struct tag is used to decide which field is added to the map.
// This function is useful when you want to convert a model struct to a map[string]interface{}
// the tag name here should be `bson` and the value should be the name of the struct field
func StructToMap(inStruct interface{}, tag string) (map[string]interface{}, error) {
	out := make(map[string]interface{})

	v := reflect.ValueOf(inStruct)

	// if it is a pointer to a struct, extract the element
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// only accepts struct kind, any other kind is an error
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("StructToMap only accepts structs or pointer to structs: got %T", v)
	}

	typ := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := typ.Field(i)
		if tagVal := field.Tag.Get(tag); !shouldOmitTag(tagVal) {
			out[tagVal] = v.Field(i).Interface()
		}
	}
	return out, nil
}

func shouldOmitTag(tagVal string) bool {
	return tagVal == "" || tagVal == "-" || strings.Contains(tagVal, "omitempty")
}

func ParseJsonFromRequest(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

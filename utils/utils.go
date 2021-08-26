package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"

	"github.com/joho/godotenv"
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

	message, _ := json.Marshal(response)

	w.WriteHeader(response.StatusCode)
	w.Write(message)
}

// GetSuccess : This is helper function to prepare success model.
func GetSuccess(msg string, data interface{}, w http.ResponseWriter) {
	var response = SuccessResponse{
		Message:    msg,
		StatusCode: http.StatusOK,
		Data:       data,
	}

	message, _ := json.Marshal(response)

	//	w.WriteHeader(response.StatusCode) if status code is not set write automatically sets it to 200
	w.Write(message)
}

// get env vars; return empty string if not found
func Env(key string) string {
	if !FileExists(".env") {
		log.Fatal("error loading .env file")
	}

	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

	env, ok := os.LookupEnv(key)

	if ok {
		return env
	}

	return ""
}

// check if a file exists, usefull in checking for .env
func FileExists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

// convert map to bson.M for mongoDB docs
func MapToBson(data map[string]interface{}) bson.M {
	bsonM := bson.M{}

	for k, v := range data {
		bsonM[k] = v
	}

	return bsonM
}

// ParseJSONFromRequest unmarshals JSON from request into a Go native type
func ParseJSONFromRequest(r *http.Request, v interface{}) error {
	return parseJSON(r.Body, v)
}

func parseJSON(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

// StructToMap converts a struct of any type to a map[string]inteface{} based on struct tags
// The struct tag is used to decide which field is added to the map.
// This function is useful when you want to convert a model struct to a map[string]interface{}
// for use with the MapToBson() function.
// this intermediate Go stuff, uses reflection and struct annotations (tags)
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
	return tagVal == "" || tagVal == "-"
}

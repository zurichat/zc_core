package utils

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

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
	StatusCode   int    `json:"status"`
	Message string `json:"message"`
	Data	interface{} `json:"data"`
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
		Message: msg,
		StatusCode: http.StatusOK,
		Data: data,
	}

	message, _ := json.Marshal(response)

	w.WriteHeader(response.StatusCode)
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

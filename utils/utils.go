package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"os"

	"github.com/dgrijalva/jwt-go"
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
	w.Header().Set("Content-Type", "application/json")
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
	w.Header().Set("Content-Type", "application/json")
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
	return bson.M(data)
}

// StructToMap converts a struct of any type to a map[string]inteface{}
func StructToMap(inStruct interface{}) (map[string]interface{}, error) {
	out := make(map[string]interface{})
	inrec, _ := json.Marshal(inStruct)
	json.Unmarshal(inrec, &out)
	return out, nil
}

// ConvertStructure does map to struct conversion and vice versa.
// The input structure will be converted to the output
func ConvertStructure(input interface{}, output interface{}) error {
	data, err := json.Marshal(input)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, output)
}

func ParseJsonFromRequest(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func TokenIsVaid(utoken string, user_id string) (bool, string, error) {
	SECRET_KEY, _ := os.LookupEnv("AUTH_SECRET_KEY")

	var signKey = []byte(SECRET_KEY)
	token, err := jwt.Parse(utoken, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return signKey, nil
	})

	if err != nil {
		return false, "Token Expired", err
	}

	claims, _ := token.Claims.(jwt.MapClaims)
	fmt.Println(claims["user_id"])

	if user_id == claims["user_id"] {
		return true, user_id, nil
	} else {
		return false, "Unauthorized user", errors.New("Not Authorized.")
	}

}

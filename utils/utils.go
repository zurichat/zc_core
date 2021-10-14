package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/mail"
	"os"

	"math/rand"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	// "zuri.chat/zccore/auth"/.
)

type M map[string]interface{}

// ErrorResponse : This is error model.
type ErrorResponse struct {
	StatusCode   int    `json:"status"`
	ErrorMessage string `json:"message"`
}

// DetailedErrorResponse : This is success model.
type DetailedErrorResponse struct {
	StatusCode int         `json:"status"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
}

// SuccessResponse : This is success model.
type SuccessResponse struct {
	StatusCode int         `json:"status"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
}

type AuthUser struct {
	ID    primitive.ObjectID
	Email string
}

type MyCustomClaims struct {
	Authorized bool
	User       AuthUser
	jwt.StandardClaims
}

type Event struct {
	Identifier interface{}            `json:"identifier" validate:"required"`
	Type       string                 `json:"type" validate:"required"`
	Event      string                 `json:"event" validate:"required"`
	Channel    interface{}            `json:"channel" validate:"required"`
	Payload    map[string]interface{} `json:"payload" validate:"required"`
}

type CentrifugoRequestBody struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}

// GetError : This is helper function to prepare error model.
func GetError(err error, statusCode int, w http.ResponseWriter) {
	var response = ErrorResponse{
		ErrorMessage: err.Error(),
		StatusCode:   statusCode,
	}

	w.WriteHeader(response.StatusCode)
	w.Header().Set("Content-Type", "application/json<Left>")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error sending response: %v", err)
	}
}

// GetDetailedError: This function provides detailed error information.
func GetDetailedError(msg string, statusCode int, data interface{}, w http.ResponseWriter) {
	var response = DetailedErrorResponse{
		Message:    msg,
		StatusCode: statusCode,
		Data:       data,
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

// get env vars; return empty string if not found.
func Env(key string) string {
	return os.Getenv(key)
}

// check if a file exists, useful in checking for .env.
func FileExists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

// convert map to bson.M for mongoDB docs.
func MapToBson(data map[string]interface{}) bson.M {
	return bson.M(data)
}

// StructToMap converts a struct of any type to a map[string]inteface{}.
func StructToMap(inStruct interface{}) (map[string]interface{}, error) {
	out := make(map[string]interface{})
	inrec, _ := json.Marshal(inStruct)

	if err := json.Unmarshal(inrec, &out); err != nil {
		return nil, err
	}

	return out, nil
}

// ConvertStructure does map to struct conversion and vice versa.
// The input structure will be converted to the output.
func OldConvertStructure(input, output interface{}) error {
	data, err := json.Marshal(input)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, output)
}

func ConvertStructure(input, output interface{}) error {
	if err := mapstructure.Decode(input, output); err != nil {
		return err
	}
	return nil
}

func ParseJSONFromRequest(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)

	return err == nil
}

func TokenIsValid(utoken string) (result bool, id string, err error) {
	SecretKey, _ := os.LookupEnv("AUTH_SECRET_KEY")

	var cclaims MyCustomClaims

	token, err := jwt.ParseWithClaims(utoken, &cclaims, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})

	if _, ok := token.Claims.(*MyCustomClaims); ok && token.Valid {
		var iid interface{} = cclaims.User.ID
		return true, iid.(primitive.ObjectID).Hex(), nil
	}

	fmt.Print(err)

	return false, "not authorized", errors.New("not authorized")
}

func TokenAgainstUserID(utoken, userID string) (result bool, id string, err error) {
	SecretKey, _ := os.LookupEnv("AUTH_SECRET_KEY")

	var cclaims MyCustomClaims

	token, err := jwt.ParseWithClaims(utoken, &cclaims, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})

	var iiid string

	if _, ok := token.Claims.(*MyCustomClaims); ok && token.Valid {
		var iid interface{} = cclaims.User.ID
		iiid = iid.(primitive.ObjectID).Hex()

		return true, iiid, nil
	}

	fmt.Print(err)

	return false, "not authorized", errors.New("not authorized")
}

func RandomGen(n int, sType string) (status bool, str string) {
	rand.Seed(time.Now().UnixNano())

	var finalString = ""

	if sType == "l" {
		randgenS := `abcdefghijklmnopqrstuvwsyz`
		s := strings.Split(randgenS, "")

		for j := 1; j <= n; j++ {
			randIdx := rand.Intn(len(s))
			finalString += s[randIdx]
		}

		return true, finalString
	}

	if sType == "d" {
		randgenI := `0123456789`
		i := strings.Split(randgenI, "")

		for j := 1; j <= n; j++ {
			randIdx := rand.Intn(len(i))
			finalString += i[randIdx]
		}

		return true, finalString
	}

	return false, "wrong type"
}

func GenWorkspaceURL(orgName string) string {
	organizationsCollection := "organizations"
	orgNamestr := strings.ReplaceAll(strings.ToLower(orgName), " ", "")
	lenRandLetters, lenRandNumbers := 3, 4
	_, randLetters := RandomGen(lenRandLetters, "l")
	_, randNumbers := RandomGen(lenRandNumbers, "d")
	wksURL := orgNamestr + "-" + randLetters + randNumbers + ".zurichat.com"

	result, _ := GetMongoDBDoc(organizationsCollection, bson.M{"url": wksURL})
	if result != nil {
		GenWorkspaceURL(orgName)
	}

	return wksURL
}

func GenJwtToken(data, tokenType string) (string, error) {
	SecretKey, _ := os.LookupEnv("AUTH_SECRET_KEY")

	claims := struct {
		Data      string
		TokenType string
		jwt.StandardClaims
	}{
		data,
		tokenType,
		jwt.StandardClaims{
			ExpiresAt: 15000,
			Issuer:    "api.zuri.chat",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(SecretKey))

	if err != nil {
		return "", err
	}

	return ss, nil
}

func GenUUID() string {
	id := uuid.New()
	return id.String()
}

// Check the validaity of a UUID. Returns a valid UUID from a string input. Returns an error otherwise.
func ValidateUUID(s string) (uuid.UUID, error) {
	validUUIDLen := 36

	if len(s) != validUUIDLen {
		return uuid.Nil, errors.New("invalid uuid format")
	}

	b, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, err
	}

	return b, nil
}

func ConvertImageTo64(imgDirectory string) string {
	// Read the entire file into a byte slice
	_bytes, err := ioutil.ReadFile(imgDirectory)
	if err != nil {
		log.Fatal(err)
	}

	var base64Encoding string

	base64Encoding += base64.StdEncoding.EncodeToString(_bytes)

	// Print the full base64 representation of the image
	return base64Encoding
}

func CentrifugoConn(body map[string]interface{}) int {
	configs := NewConfigurations()
	jsonData, err := json.Marshal(body)

	errCode500 := 500

	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		return errCode500
	}

	requestBody := bytes.NewBuffer(jsonData)
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("POST", configs.CentrifugoEndpoint, requestBody)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		return errCode500
	}

	req.Header.Add("Authorization", "apikey "+configs.CentrifugoKey)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)

	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		return errCode500
	}

	if resp.StatusCode == 403 || resp.StatusCode == 401 {
		fmt.Println("Unauthorized: Invalid API key for Websocket Server")
	}

	resp.Body.Close()

	return resp.StatusCode
}

func Emitter(event Event) int {
	event.Payload["id"] = event.Identifier
	event.Payload["type"] = event.Type
	event.Payload["event"] = event.Event
	err400 := 400
	reqBody := CentrifugoRequestBody{Method: "publish",
		Params: map[string]interface{}{"channel": event.Channel, "data": event.Payload},
	}

	body, err := StructToMap(reqBody)
	if err != nil {
		fmt.Printf("There is an error")
		return err400
	}

	status := CentrifugoConn(body)

	return status
}

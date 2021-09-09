package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"os"

	"math/rand"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	// "zuri.chat/zccore/auth"/
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

type AuthUser struct {
	ID    primitive.ObjectID
	Email string
}

type MyCustomClaims struct {
	Authorized bool
	User       AuthUser
	jwt.StandardClaims
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

func TokenIsValid(utoken string) (bool, string, error) {

	SECRET_KEY, _ := os.LookupEnv("AUTH_SECRET_KEY")
	var cclaims MyCustomClaims

	token, err := jwt.ParseWithClaims(utoken, &cclaims, func(token *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})

	if _, ok := token.Claims.(*MyCustomClaims); ok && token.Valid {
		var iid interface{} = cclaims.User.ID
		return true, iid.(primitive.ObjectID).Hex(), nil
	} else {
		fmt.Print(err)
		return false, "Not Authorized", errors.New("Not Authorized.")
	}

}

func TokenAgainstUserId(utoken string, user_id string) (bool, string, error) {
	SECRET_KEY, _ := os.LookupEnv("AUTH_SECRET_KEY")
	var cclaims MyCustomClaims

	token, err := jwt.ParseWithClaims(utoken, &cclaims, func(token *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})
	var iiid string
	if _, ok := token.Claims.(*MyCustomClaims); ok && token.Valid {
		var iid interface{} = cclaims.User.ID
		iiid = iid.(primitive.ObjectID).Hex()
		return true, iiid, nil
	} else {
		fmt.Print(err)
		return false, "Not Authorized", errors.New("Not Authorized.")
	}
}

func RandomGen(n int, s_type string) (status bool, str string) {
	rand.Seed(time.Now().UnixNano())
	var final_string = ""
	if s_type == "l" {
		randgen_s := `abcdefghijklmnopqrstuvwsyz`
		s := strings.Split(randgen_s, "")

		for j := 1; j <= n; j++ {
			randIdx := rand.Intn(len(s))
			final_string = final_string + s[randIdx]
		}
		return true, final_string

	} else if s_type == "d" {
		randgen_i := `0123456789`
		i := strings.Split(randgen_i, "")
		for j := 1; j <= n; j++ {
			randIdx := rand.Intn(len(i))
			final_string = final_string + i[randIdx]
		}
		return true, final_string

	} else {
		return false, "wrong type"
	}

}

func GenWorkspaceUrl(orgName string) string {
	organisation_collection := "organizations"
	orgNamestr := strings.ReplaceAll(strings.ToLower(orgName), " ", "")
	_, randLetters := RandomGen(3, "l")
	_, randNumbers := RandomGen(4, "d")
	wksUrl := orgNamestr + "-" + randLetters + randNumbers + ".zurichat.com"
	result, _ := GetMongoDbDoc(organisation_collection, bson.M{"url": wksUrl})
	if result != nil {
		GenWorkspaceUrl(orgName)
	}
	return wksUrl
}

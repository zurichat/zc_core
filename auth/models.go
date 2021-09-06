package auth

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"zuri.chat/zccore/utils"
)

var (
	NoAuthToken = errors.New("No Authorization header provided.")
	TokenExp = errors.New("Token expired.")
	NotAuthorized = errors.New("Not Authorized.")
)

type Authentication struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type Token struct {
	Email			string					`json:"email"`
	TokenString 	string					`json:"token"`
	UserID			primitive.ObjectID		`json:"user_id"`
	OrganizationID 	primitive.ObjectID		`json:"org_id,omitempty"`
}

// Method to compare password
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Generate token
func GenerateJWT(email, org_id string) (string, error) {
	SECRET_KEY, _ := os.LookupEnv("SECRET_KEY")
	if SECRET_KEY == "" { SECRET_KEY = secretKey }
	
	var signKey = []byte(SECRET_KEY)

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["email"] = email
	// claims["org_id"] = org_id
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	tokenString, err := token.SignedString(signKey)

	if err != nil {
		fmt.Errorf("Something went wrong: %s", err.Error())
		return "", err
	}

	return tokenString, nil
}

// middleware to check if user is authorized
func IsAuthorized(nextHandler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header["Bearer"] == nil {
			utils.GetError(NoAuthToken, http.StatusForbidden, w)
			return
		}
				
		SECRET_KEY, _ := os.LookupEnv("SECRET_KEY")
		if SECRET_KEY == "" { SECRET_KEY = secretKey }

		var signKey = []byte(SECRET_KEY)
		token, err := jwt.Parse(r.Header["Bearer"][0], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return signKey, nil			
		})

		if err != nil {
			utils.GetError(TokenExp, http.StatusBadRequest, w)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// @TODO: work on this later.
			fmt.Println(claims["authorized"], claims["email"], claims["user_id"]) 
			nextHandler(w, r)
		} else {
			utils.GetError(NotAuthorized, http.StatusBadRequest, w)
			return
		}
	}
}
package auth

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"zuri.chat/zccore/utils"
)

type Authentication struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type Token struct {
	Email          string `json:"email"`
	TokenString    string `json:"token"`
	OrganizationID string `json:"org_id"`
}

// Method to compare password
func CheckPassword(password, hash string) bool {
	//err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	//return err == nil
	return true
}

// Generate JWT
func GenerateJWT(email, org_id string) (string, error) {
	var signKey = []byte(secretKey)
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["email"] = email
	// claims["org_id"] = org_id
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	tokenString, err := token.SignedString(signKey)

	if err != nil {
		log.Printf("Something went wrong: %s", err.Error())
		return "", err
	}

	return tokenString, nil
}

// MiddlewareValidateAccessToken allows us to secure authenticated routes
// Long name indeed.
func MiddlewareValidateAccessToken(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := IsTokenValid(r)
		if err != nil {
			utils.GetError(err, http.StatusUnauthorized, w)
			return
		}
		next(w, r)
	}
}

// IsTokenValid checks the validity of a token, whether it is still useful or it has expired
func IsTokenValid(r *http.Request) error {
	token, err := VerifyToken(r)
	if err != nil {
		return err
	}
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		return err
	}
	return nil
}

// parsing and validating a token using the HMAC signing method
func VerifyToken(r *http.Request) (*jwt.Token, error) {
	tokenString := RefineToken(r)
	token, err := jwt.Parse(tokenString, verifyAlgorithm)
	if err != nil {
		return nil, err
	}
	return token, nil
}

// validate the algorithm
func verifyAlgorithm(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("Unexpected Signing method: %v", token.Header["alg"])
	}
	hmacSecret := []byte(secretKey)
	return hmacSecret, nil
}

// RefineToken extract the token from the request header
func RefineToken(r *http.Request) string {
	token := r.Header.Get("Authorization")
	str := strings.Split(token, " ")
	if len(str) == 2 {
		return str[1]
	}
	return ""
}

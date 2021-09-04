package auth

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type Authentication struct {
	Email		string		`json:"email" validate:"required,email"`
	Password	string		`json:"password" validate:"required"`
}

type Token struct {
	Email		string		`json:"email"`
	TokenString string		`json:"token"`
	OrganizationID string 	`json:"org_id"`
}

// Method to compare password
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
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
		fmt.Errorf("Something went wrong: %s", err.Error())
		return "", err
	}

	return tokenString, nil
}
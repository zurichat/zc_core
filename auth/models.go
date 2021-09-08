package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"zuri.chat/zccore/user"
	"zuri.chat/zccore/utils"
)

type contextKey int
const authUserKey contextKey = 0

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
	TokenString 	string					`json:"token"`
	User			user.User				`json:"user"`
}

type AuthUser struct {
	ID                primitive.ObjectID      `json:"id"`
	Email             string                  `json:"email"`	
}

type MyCustomClaims struct {
	Authorized 			bool 		`json:"authorized"`
	User 				AuthUser
	jwt.StandardClaims
}

// Method to compare password
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func fetchUserByEmail(filter map[string]interface{}) (*user.User, error) {
	user := &user.User{}
	userCollection, err := utils.GetMongoDbCollection(os.Getenv("DB_NAME"), user_collection)
	if err != nil {
		return user, err
	}
	result := userCollection.FindOne(context.TODO(), filter)
	err = result.Decode(&user)
	return user, err
}

// Generate token
func GenerateJWT(userID, email string) (string, error) {
	SECRET_KEY, _ := os.LookupEnv("AUTH_SECRET_KEY")
	if SECRET_KEY == "" { SECRET_KEY = secretKey }
	
	var signKey = []byte(SECRET_KEY)

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil { return "", errors.New("Invalid user ObjectID") }	


	claims := MyCustomClaims{
		true,
		AuthUser{
			ID: objID,
			Email: email,
		},
		jwt.StandardClaims{ 
			ExpiresAt: time.Now().Add(time.Hour * 12).Unix(), // 12 hours
			Issuer: "api.zuri.chat",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(signKey)

	if err != nil {
		fmt.Errorf("Something went wrong: %s", err.Error())
		return "", err
	}

	return tokenString, nil
}

// middleware to check if user is authorized
func IsAuthenticated(nextHandler http.HandlerFunc) http.HandlerFunc {
	// token format "Authorization": "Bearer token"
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "application/json")

		if r.Header["Authorization"] == nil {
			utils.GetError(NoAuthToken, http.StatusUnauthorized, w)
			return
		}
				
		SECRET_KEY, _ := os.LookupEnv("AUTH_SECRET_KEY")
		if SECRET_KEY == "" { SECRET_KEY = secretKey }
		
		authToken := strings.Split(r.Header["Authorization"][0], " ")[1]
		
		token, err := jwt.ParseWithClaims(authToken, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		})

		if claims, ok := token.Claims.(*MyCustomClaims); ok && token.Valid {
			// Extract user information and add it to request context.
			ctx := context.WithValue(r.Context(), "user", claims.User)
			nextHandler.ServeHTTP(w, r.WithContext(ctx))
		} else {
			fmt.Print(err)
			utils.GetError(NotAuthorized, http.StatusBadRequest, w)
			return
		}
	}
}
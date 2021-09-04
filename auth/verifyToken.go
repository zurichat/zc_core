package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2/bson"
	"zuri.chat/zccore/utils"
)

// MiddlewareValidateAccessToken allows us to secure authenticated routes
// Long name indeed.
func MiddlewareValidateAccessToken(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := IsTokenValid(r)
		if err != nil {
			utils.GetError(err, http.StatusUnauthorized, w)
			return
		}
		token, err := ExtractTokenMetaData(r)
		if err != nil {
			utils.GetError(err, http.StatusUnauthorized, w)
			return
		}
		userID, err := FetchAuthSession(token)
		if userID == primitive.NilObjectID || err != nil {
			err := errors.New("Session Expired")
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
	hmacSecret := []byte(os.Getenv("HMAC_SECRET"))
	return hmacSecret, nil
}

// RefineToken extract the token from the request header
func RefineToken(r *http.Request) string {
	log.Println("in RefineToken")
	token := r.Header.Get("Authorization")
	str := strings.Split(token, " ")
	if len(str) == 2 {
		return str[1]
	}
	return ""
}

// ExtractTokenMetaData allows other route to get users token. this function could be
// called if token metadat is required for any reason.
func ExtractTokenMetaData(r *http.Request) (*AccessDetails, error) {
	token, err := VerifyToken(r)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		accessUuid, ok := claims["access_uuid"].(string)
		if !ok {
			return nil, err
		}
		userId, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)
		if err != nil {
			return nil, err
		}
		return &AccessDetails{
			AccessUuid: accessUuid,
			UserId:     userId,
		}, nil
	}
	return nil, err
}
func FetchAuthSession(authD *AccessDetails) (primitive.ObjectID, error) {
	DbName := os.Getenv("DB_NAME")
	ctx := context.TODO()
	session := &Session{}
	sessionCollection, err := utils.GetMongoDbCollection(DbName, "Session")
	if err != nil {
		return primitive.NilObjectID, err
	}
	filter := bson.M{"token_uuid": authD.AccessUuid}
	result := sessionCollection.FindOne(ctx, filter)
	err = result.Decode(&session)
	userid := session.UserID
	return userid, nil
}

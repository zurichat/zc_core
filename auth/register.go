package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/golang/gddo/httputil/header"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2/bson"
	"zuri.chat/zccore/utils"
)

func Register(w http.ResponseWriter, r *http.Request) {
	var user User
	// If the Content-Type header is present, check that it has the value
	// application/json. Note that we are using the gddo/httputil/header
	// package to parse and extract the value here, so the check works
	// even if the client includes additional charset or boundary
	// information in the header.
	if r.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
		if value != "application/json" {
			err := errors.New("Content-Type header is not application/json")
			utils.GetError(err, http.StatusUnsupportedMediaType, w)
			return
		}
	}

	// Use http.MaxBytesReader to enforce a maximum read of 1MB from the
	// response body. A request body larger than that will now result in
	// Decode() returning a "http: request body too large" error.
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	// Setup the decoder and call the DisallowUnknownFields() method on it.
	// This will cause Decode() to return a "json: unknown field ..." error
	// if it encounters any extra unexpected fields in the JSON. Strictly
	// speaking, it returns an error for "keys which do not match any
	// non-ignored, exported fields in the destination".
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&user)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		// Catch any syntax errors in the JSON and send an error message
		// which interpolates the location of the problem to make it
		// easier for the client to fix.
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			err = errors.New(msg)
			utils.GetError(err, http.StatusBadRequest, w)

		// In some circumstances Decode() may also return an
		// io.ErrUnexpectedEOF error for syntax errors in the JSON.
		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := fmt.Sprintf("Request body contains badly-formed JSON")
			err = errors.New(msg)
			utils.GetError(err, http.StatusBadRequest, w)

		// Catch any type errors, like trying to assign a string in the
		// JSON request body to an int field in our Payload struct. We can
		// interpolate the relevant field name and position into the error
		// message to make it easier for the client to fix.
		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			err = errors.New(msg)
			utils.GetError(err, http.StatusBadRequest, w)

		// Catch the error caused by extra unexpected fields in the request
		// body. We extract the field name from the error message and
		// interpolate it in our custom error message.
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			err = errors.New(msg)
			utils.GetError(err, http.StatusBadRequest, w)

		// An io.EOF error is returned by Decode() if the request body is
		// empty.
		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			log.Println(msg)
			utils.GetError(err, http.StatusBadRequest, w)

		// Catch the error caused by the request body being too large.
		case err.Error() == "http: request body too large":
			//msg := "Request body must not be larger than 1MB"
			utils.GetError(err, http.StatusRequestEntityTooLarge, w)

		// Otherwise default to logging the error and sending a 500 Internal
		// Server Error response.
		default:
			log.Println(err.Error())
			utils.GetError(err, http.StatusInternalServerError, w)
		}
		return
	}

	// Call decode again, using a pointer to an empty anonymous struct as
	// the destination. If the request body only contained a single JSON
	// object this will return an io.EOF error. So if we get anything else,
	// we know that there is additional data in the request body.
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		msg := "Request body must only contain a single JSON object"
		err = errors.New(msg)
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	if utils.IsValidEmail(user.Email) {
		err := saveUser(user)
		if err != nil {
			utils.GetError(err, http.StatusInternalServerError, w)
			return
		}
	}
	_, err = fetchUser(user.Email)
	if err != nil {
		err = errors.New("This email already exist")
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}
	tokenDetail, err := createToken(user.UserID)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}
	saveErr := saveSessions(user.UserID, *tokenDetail)
	if saveErr != nil {
		utils.GetError(saveErr, http.StatusUnprocessableEntity, w)
		return
	}

	tokens := map[string]interface{}{
		"access_token":  tokenDetail.AccessToken,
		"refresh_token": tokenDetail.RefreshToken,
		"expires_at":    tokenDetail.AtExpires,
	}

	utils.GetSuccess("okay", tokens, w)

}

// save a user in Authentication database
func saveUser(user User) error {
	user.UserID = primitive.NewObjectID()
	DbName := os.Getenv("AUTH_DB_NAME")
	userCollection, err := utils.GetMongoDbCollection(DbName, "user")
	if err != nil {
		return err
	}
	ctx := context.TODO()
	_, err = userCollection.InsertOne(ctx, user)
	return nil
}

func fetchUser(email string) (*User, error) {
	user := &User{}
	DbName := os.Getenv("AUTH_DB_NAME")
	userCollection, err := utils.GetMongoDbCollection(DbName, "user")
	if err != nil {
		return user, err
	}
	filter := bson.M{"email": email}
	ctx := context.TODO()
	result := userCollection.FindOne(ctx, filter)
	err = result.Decode(&user)
	if err != nil {
		return user, err
	}
	return user, err
}

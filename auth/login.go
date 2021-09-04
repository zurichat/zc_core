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
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/golang/gddo/httputil/header"
	"github.com/twinj/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"zuri.chat/zccore/utils"
)

func Login(w http.ResponseWriter, r *http.Request) {
	var u User
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

	err := dec.Decode(&u)
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

	// check if user password and email is in the database
	if ok, err := IsUserRegistered(u); ok == false {
		utils.GetError(err, http.StatusForbidden, w)
		return
	}
	tokenDetail, err := createToken(u.UserID)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}
	ctx := r.Context()
	saveErr := saveSessions(u.UserID, *tokenDetail, ctx)
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

// IsUserRegistered checks if user password and email are in the database. i.e user
// is already registerd.

func IsUserRegistered(user User) (bool, error) {
	userDB, err := fetchUser(user.Email)
	if err != nil {
		err = errors.New("No result for this credentials")
		return false, err
	}
	//find user in database with user info
	if user.Email == userDB.Email && user.Password == userDB.Password {
		return true, nil
	}
	err = errors.New("Invalid Login details")
	return false, err
}

// createToken creates access token for authenticated users
func createToken(ID primitive.ObjectID) (*TokenMetaData, error) {
	var err error
	tokenDetails := &TokenMetaData{
		AtExpires:   time.Now().Add(time.Minute * 15),
		RtExpires:   time.Now().Add(time.Hour * 24 * 7),
		AccessUuid:  uuid.NewV4().String(),
		RefreshUuid: uuid.NewV4().String(),
	}

	// Get access token
	accessClaims := jwt.MapClaims{}
	accessClaims["user_id"] = ID
	accessClaims["authorized"] = true
	accessClaims["access_uuid"] = tokenDetails.AccessUuid
	accessClaims["IAt"] = time.Now()
	accessClaims["exp"] = tokenDetails.AtExpires
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	tokenDetails.AccessToken, err = accessToken.SignedString([]byte(os.Getenv("HMAC_SECRET")))
	if err != nil {
		return nil, err
	}

	// Create Refresh token
	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = tokenDetails.RefreshUuid
	rtClaims["user_id"] = ID
	rtClaims["exp"] = tokenDetails.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	tokenDetails.RefreshToken, err = rt.SignedString([]byte(os.Getenv("HMAC_REFRESH_SECRET")))
	if err != nil {
		return nil, err
	}
	return tokenDetails, nil
}

func saveSessions(ID primitive.ObjectID, tokenDetails TokenMetaData, ctx context.Context) error {
	// create access token sessions
	accessTokenSession := &AcessSession{
		TokenUuid:   tokenDetails.AccessUuid,
		UserID:      ID,
		AccessToken: tokenDetails.AccessToken,
		CreatedAt:   time.Now().UTC(),
		ExpireOn:    tokenDetails.AtExpires,
	}
	refreshTokenSession := &RefreshSession{
		TokenUuid:    tokenDetails.RefreshUuid,
		UserID:       ID,
		RefreshToken: tokenDetails.RefreshToken,
		CreatedAt:    time.Now().UTC(),
		ExpireOn:     tokenDetails.RtExpires,
	}

	DbName := os.Getenv("AUTH_DB_NAME")
	sessionCollection, err := utils.GetMongoDbCollection(DbName, "Session")
	if err != nil {
		return err
	}
	now := time.Now()
	accessTime := tokenDetails.AtExpires.Sub(now)
	refreshTime := tokenDetails.RtExpires.Sub(now)
	routineCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	// TTL index
	indexAT := mongo.IndexModel{
		Keys:    bsonx.Doc{{Key: "access_token", Value: bsonx.Int32(1)}},
		Options: options.Index().SetExpireAfterSeconds(int32(accessTime.Seconds())),
	}
	indexRT := mongo.IndexModel{
		Keys:    bsonx.Doc{{Key: "refresh_token", Value: bsonx.Int32(1)}},
		Options: options.Index().SetExpireAfterSeconds(int32(refreshTime.Seconds())),
	}
	multiIndex := []mongo.IndexModel{indexAT, indexRT}

	_, err = sessionCollection.Indexes().CreateMany(routineCtx, multiIndex)

	_, err = sessionCollection.InsertOne(routineCtx, accessTokenSession)
	if err != nil {
		return err
	}

	_, err = sessionCollection.InsertOne(ctx, refreshTokenSession)
	if err != nil {
		return err
	}

	return nil
}

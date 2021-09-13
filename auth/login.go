package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/bson"
	"zuri.chat/zccore/utils"
)

const (
	secretKey       	= "5d5c7f94e29ba11f6822a2be310d3af4"
	sessionKey 			= "f6822af94e29ba112be310d3af45d5c7"
	session_collection 	= "session_store"	
	user_collection 	= "users"
)

var (
	validate           = validator.New()
	UserNotFound       = errors.New("User not found!")
	InvalidCredentials = errors.New("Invalid login credentials, confirm and try again")
)

func LoginIn(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	var creds Credentials
	if err := utils.ParseJsonFromRequest(request, &creds); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, response)
		return
	}
	if err := validate.Struct(creds); err != nil {
		utils.GetError(err, http.StatusBadRequest, response)
		return
	}

	user, err := FetchUserByEmail(bson.M{"email":  strings.ToLower(creds.Email)})
	if err != nil {
		utils.GetError(UserNotFound, http.StatusBadRequest, response)
		return
	}
	// check password
	check := CheckPassword(creds.Password, user.Password)
	if !check {
		utils.GetError(InvalidCredentials, http.StatusBadRequest, response)
		return
	}
	store := NewMongoStore(utils.GetCollection(session_collection), SESSION_MAX_AGE, true, []byte(secretKey))
	var session, e = store.Get(request, sessionKey)
	if e != nil {
		msg := fmt.Errorf("%s", e.Error())
		utils.GetError(msg, http.StatusBadRequest, response)
		return
	}

	// store session
	session.Values["id"] = user.ID.Hex()
	session.Values["email"] = user.Email

	if err = sessions.Save(request, response); err != nil {
		fmt.Printf("Error saving session: %s", err)
		return
	}

	resp := &Token{
		SessionID: session.ID,
		User: UserResponse{
			ID: user.ID,
			FirstName: user.FirstName,
			LastName: user.LastName,
			DisplayName: user.DisplayName,
			Email: user.Email,
			Phone: user.Phone,
			Status: int(user.Status),
			Timezone: user.Timezone,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,		
		},		
	}
	utils.GetSuccess("login successful", resp, response)
}


func LogOutUser(w http.ResponseWriter, r *http.Request) {
	store := NewMongoStore(
		utils.GetCollection(session_collection), 
		SESSION_MAX_AGE, 
		true, 
		[]byte(secretKey),
	)

	var session, err = store.Get(r, sessionKey)
	if err != nil {
		utils.GetError(NotAuthorized, http.StatusUnauthorized, w)
		return
	}

	if session.IsNew == true {
		utils.GetError(NoAuthToken, http.StatusUnauthorized, w)
		return
	}

	session.Options.MaxAge = -1

	if err = sessions.Save(r, w); err != nil {
		fmt.Printf("Error saving session: %s", err)
		return
	}

	utils.GetSuccess("logout successful", map[string]interface{}{}, w)
}


func VerifyTokenHandler(response http.ResponseWriter, request *http.Request) {
	// extract user id and email from context
	loggedIn := request.Context().Value("user").(*AuthUser)
	user, _ := FetchUserByEmail(bson.M{"email": strings.ToLower(loggedIn.Email)})

	resp := &VerifiedTokenResponse{
		true,
		UserResponse{
			ID: user.ID,
			FirstName: user.FirstName,
			LastName: user.LastName,
			DisplayName: user.DisplayName,
			Email: user.Email,
			Phone: user.Phone,
			Status: int(user.Status),
			Timezone: user.Timezone,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}

	utils.GetSuccess("verified", resp, response)
}
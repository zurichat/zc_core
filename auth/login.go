package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/bson"
	"zuri.chat/zccore/utils"
)

const (
	secretKey          = "5d5c7f94e29ba11f6822a2be310d3af4"
	sessionKey         = "f6822af94e29ba112be310d3af45d5c7"
	enckry             = "piqenpdan9n-94n49-e-9ad-aononoon"
	session_collection = "session_store"
	user_collection    = "users"
)

var (
	validate           = validator.New()
	UserNotFound       = errors.New("User not found!")
	InvalidCredentials = errors.New("Invalid login credentials, confirm and try again")
	hmacSampleSecret   = []byte("u7b8be9bd9b9ebd9b9dbdbee")
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

	user, err := FetchUserByEmail(bson.M{"email": strings.ToLower(creds.Email)})
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
	retoken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"session_name": Resptoken.SessionName,
		"cookie":       Resptoken.Cookie,
		"options":      Resptoken.Options,
		"id":           Resptoken.Id,
		"email":        user.Email,
	})

	tokenString, eert := retoken.SignedString(hmacSampleSecret)
	if eert != nil {
		utils.GetError(eert, http.StatusInternalServerError, response)
	}

	resp := &Token{
		SessionID: session.ID,
		User: UserResponse{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			Phone:     user.Phone,
			Timezone:  user.Timezone,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Token:     tokenString,
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
	var session *sessions.Session
	var err error
	session, err = store.Get(r, sessionKey)
	status, _, sessData := GetSessionDataFromToken(r, hmacSampleSecret)

	if err != nil && status == false {
		utils.GetError(NotAuthorized, http.StatusUnauthorized, w)
		return
	}
	var erro error
	if status == true {
		session, erro = NewS(store, sessData.Cookie, sessData.Id, sessData.Email, r, sessData.SessionName)
		fmt.Println(session)
		if err != nil && erro != nil {
			utils.GetError(NotAuthorized, http.StatusUnauthorized, w)
			return
		}
	}

	if session.IsNew == true {
		utils.GetError(NoAuthToken, http.StatusUnauthorized, w)
		return
	}

	fmt.Println(session)
	session.Options.MaxAge = -1

	if err = ClearSession(store, w, session); err != nil {
		fmt.Printf("Error saving session: %s", err)
		utils.GetError(fmt.Errorf("Logout Failed"), http.StatusUnauthorized, w)
		return
	}
	// if err = sessions.Save(r, w); err != nil {
	// 	fmt.Printf("Error saving session: %s", err)
	// 	utils.GetError(fmt.Errorf("Logout Failed"), http.StatusUnauthorized, w)
	// 	return
	// }

	utils.GetSuccess("logout successful", map[string]interface{}{}, w)
}

func VerifyTokenHandler(response http.ResponseWriter, request *http.Request) {
	// extract user id and email from context
	loggedIn := request.Context().Value("user").(*AuthUser)
	user, _ := FetchUserByEmail(bson.M{"email": strings.ToLower(loggedIn.Email)})

	resp := &VerifiedTokenResponse{
		true,
		UserResponse{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			Phone:     user.Phone,
			Timezone:  user.Timezone,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}

	utils.GetSuccess("verified", resp, response)
}

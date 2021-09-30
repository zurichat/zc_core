package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/user"
	"zuri.chat/zccore/utils"
)

// middleware to check if user is authorized
func (au *AuthHandler) IsAuthenticated(nextHandler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "application/json")

		store := NewMongoStore(utils.GetCollection(session_collection), au.configs.SessionMaxAge, true, []byte(au.configs.SecretKey))
		var session *sessions.Session
		var SessionEmail string
		var err error

		session, _ = store.Get(r, au.configs.SessionKey)
		status, _, sessData := GetSessionDataFromToken(r, []byte(au.configs.HmacSampleSecret))

		// if err == nil && !status {
		// 	utils.GetError(NotAuthorized, http.StatusUnauthorized, w)
		// 	return
		// }

		var erro error

		if status {
			session, erro = NewS(store, sessData.Cookie, sessData.Id, sessData.Email, r, sessData.SessionName, sessData.Gothic)
			if err != nil && erro != nil {
				utils.GetError(NotAuthorized, http.StatusUnauthorized, w)
				return
			}
		}

		if sessData.Gothic != nil {
			SessionEmail = sessData.GothicEmail
		} else if session.Values["email"] != nil {
			SessionEmail = session.Values["email"].(string)
		}

		// use is coming in newly, no cookies
		if session.IsNew {
			utils.GetError(NoAuthToken, http.StatusUnauthorized, w)
			return
		}

		objID, err := primitive.ObjectIDFromHex(session.ID)
		if err != nil {
			utils.GetError(ErrorInvalid, http.StatusUnauthorized, w)
			return
		}

		user := &AuthUser{
			ID:    objID,
			Email: SessionEmail,
		}

		ctx := context.WithValue(r.Context(), "user", user)
		nextHandler.ServeHTTP(w, r.WithContext(ctx))
	}
}

// OptionalAuthenticated calls the next's handler's ServeHTTP() with the request context unchanged
// if a user is not authenticated, else it modifies the request context with a copy of the user's
// details and passes the changed copy of the request to the next handler's ServeHTTP()
func (au *AuthHandler) OptionalAuthentication(nextHandler http.HandlerFunc, auth *AuthHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "application/json")

		store := NewMongoStore(utils.GetCollection(session_collection), au.configs.SessionMaxAge, true, []byte(au.configs.SecretKey))
		_, err := store.Get(r, au.configs.SessionKey)
		status, err, sessData := GetSessionDataFromToken(r, []byte(au.configs.HmacSampleSecret))

		if err != nil {
			utils.GetError(NotAuthorized, http.StatusUnauthorized, w)
			return
		}

		if !status && sessData.Email == "" {
			nextHandler.ServeHTTP(w, r)
			return
		} else {
			ctx := context.WithValue(r.Context(), UserDetails, &sessData)
			r = r.WithContext(ctx)
			nextHandler.ServeHTTP(w, r)
		}

	}
}

func (au *AuthHandler) IsAuthorized(nextHandler http.HandlerFunc, role string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var orgId string

		if mux.Vars(r)["id"] != "" {
			orgId = mux.Vars(r)["id"]
		}
		loggedInUser := r.Context().Value("user").(*AuthUser)
		lguser, ee := FetchUserByEmail(bson.M{"email": strings.ToLower(loggedInUser.Email)})
		if ee != nil {
			utils.GetError(errors.New("Error Fetching Logged in User"), http.StatusBadRequest, w)
		}
		user_id := lguser.ID

		// collections
		_, user_collection, member_collection := "organizations", "users", "members"
		// org_collection

		// fmt.Println(user_id)

		// Getting user's document from db
		var luHexid, _ = primitive.ObjectIDFromHex(user_id)
		userDoc, _ := utils.GetMongoDbDoc(user_collection, bson.M{"_id": luHexid})
		if userDoc == nil {
			utils.GetError(errors.New("User not found"), http.StatusBadRequest, w)
		}

		var authuser user.User
		mapstructure.Decode(userDoc, &authuser)

		if role == "zuri_admin" {
			if authuser.Role != "admin" {
				utils.GetError(errors.New("Access Denied"), http.StatusUnauthorized, w)
				return
			}

		} else {
			// Getting member's document from db
			orgMember, _ := utils.GetMongoDbDoc(member_collection, bson.M{"org_id": orgId, "email": authuser.Email})
			if orgMember == nil {
				fmt.Println("no org")
				utils.GetError(errors.New("Access Denied"), http.StatusUnauthorized, w)
				return
			}

			var memb RoleMember
			mapstructure.Decode(orgMember, &memb)

			// check role's access
			nA := map[string]int{"owner": 4, "admin": 3, "member": 2, "guest": 1}

			if nA[role] > nA[memb.Role] {
				utils.GetError(errors.New("Access Denied"), http.StatusUnauthorized, w)
				return
			}
		}

		user := &AuthUser{
			ID:    luHexid,
			Email: loggedInUser.Email,
		}
		ctx := context.WithValue(r.Context(), "user", user)
		nextHandler.ServeHTTP(w, r.WithContext(ctx))
	}
}

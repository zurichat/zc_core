package auth

import (
	"context"
	"errors"
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

// middleware to check if user is authorized.
func (au *AuthHandler) IsAuthenticated(nextHandler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "application/json")

		var (
			session      *sessions.Session
			SessionEmail string
			err          error
			erro         error
		)

		store := NewMongoStore(utils.GetCollection(sessionCollection), au.configs.SessionMaxAge, true, []byte(au.configs.SecretKey))
		session, _ = store.Get(r, au.configs.SessionKey)
		status, sessData, _ := GetSessionDataFromToken(r, []byte(au.configs.HmacSampleSecret))

		if status {
			session, erro = NewS(store, sessData.Cookie, sessData.ID, sessData.Email, r, sessData.SessionName, sessData.Gothic)
			if err != nil && erro != nil {
				utils.GetError(ErrNotAuthorized, http.StatusUnauthorized, w)
				return
			}
		}

		if sessData.Gothic != nil {
			SessionEmail = sessData.GothicEmail
		} else if session.Values["email"] != nil {
			SessionEmail, _ = session.Values["email"].(string)
		}

		// use is coming in newly, no cookies
		if session.IsNew {
			utils.GetError(ErrNoAuthToken, http.StatusUnauthorized, w)
			return
		}

		objID, err := primitive.ObjectIDFromHex(session.ID)
		if err != nil {
			utils.GetError(ErrorInvalid, http.StatusUnauthorized, w)
			return
		}

		u := &AuthUser{
			ID:    objID,
			Email: SessionEmail,
		}
		ctx := context.WithValue(r.Context(), UserContext, u)
		nextHandler.ServeHTTP(w, r.WithContext(ctx))
	}
}

// OptionalAuthenticated calls the next's handler's ServeHTTP() with the request context unchanged
// if a user is not authenticated, else it modifies the request context with a copy of the user's
// details and passes the changed copy of the request to the next handler's ServeHTTP().
func (au *AuthHandler) OptionalAuthentication(nextHandler http.HandlerFunc, auth *AuthHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "application/json")

		store := NewMongoStore(utils.GetCollection(sessionCollection), au.configs.SessionMaxAge, true, []byte(au.configs.SecretKey))
		_, er := store.Get(r, au.configs.SessionKey)
		status, sessData, err := GetSessionDataFromToken(r, []byte(au.configs.HmacSampleSecret))

		if er != nil || err != nil {
			utils.GetError(ErrNotAuthorized, http.StatusUnauthorized, w)
			return
		}

		if !status && sessData.Email == "" {
			nextHandler.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), UserDetails, &sessData)
		r = r.WithContext(ctx)
		nextHandler.ServeHTTP(w, r)
	}
}

func (au *AuthHandler) IsAuthorized(nextHandler http.HandlerFunc, role string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var (
			orgID    string
			authuser user.User
			memb     RoleMember
		)

		if mux.Vars(r)["id"] != "" {
			orgID = mux.Vars(r)["id"]
		}

		loggedInUser, _ := r.Context().Value("user").(*AuthUser)
		lguser, ee := FetchUserByEmail(bson.M{"email": strings.ToLower(loggedInUser.Email)})

		if ee != nil {
			utils.GetError(errors.New("error Fetching Logged in User"), http.StatusBadRequest, w)
		}

		userID := lguser.ID
		luHexid, _ := primitive.ObjectIDFromHex(userID)
		_, userCollection, memberCollection := "organizations", "users", "members"
		userDoc, _ := utils.GetMongoDbDoc(userCollection, bson.M{"_id": luHexid})

		if userDoc == nil {
			utils.GetError(errors.New("user not found"), http.StatusBadRequest, w)
			return
		}

		//nolint:errcheck //CODEI8:
		mapstructure.Decode(userDoc, &authuser)

		if role == "zuri_admin" {
			if authuser.Role != "admin" {
				utils.GetError(errors.New("access Denied"), http.StatusUnauthorized, w)
				return
			}
		} else {
			// Getting member's document from db
			orgMember, _ := utils.GetMongoDbDoc(memberCollection, bson.M{"org_id": orgID, "email": authuser.Email})
			if orgMember == nil {
				utils.GetError(errors.New("access Denied"), http.StatusUnauthorized, w)
				return
			}

			//nolint:errcheck //CODEI8:
			mapstructure.Decode(orgMember, &memb)

			// check role's access
			nA := map[string]int{"owner": 4, "admin": 3, "member": 2, "guest": 1}

			if nA[role] > nA[memb.Role] {
				utils.GetError(errors.New("access Denied"), http.StatusUnauthorized, w)
				return
			}
		}

		u := &AuthUser{
			ID:    luHexid,
			Email: loggedInUser.Email,
		}
		ctx := context.WithValue(r.Context(), UserContext, u)
		nextHandler.ServeHTTP(w, r.WithContext(ctx))
	}
}

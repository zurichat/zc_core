package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/utils"
)

// middleware to check if user is authorized
func (au *AuthHandler) IsAuthenticated(nextHandler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "application/json")

		store := NewMongoStore(utils.GetCollection(session_collection), au.configs.SessionMaxAge, true, []byte(secretKey))
		var session *sessions.Session
		var SessionEmail string
		var err error
		session, err = store.Get(r, sessionKey)
		status, _, sessData := GetSessionDataFromToken(r, hmacSampleSecret)

		if err != nil && status == false {
			utils.GetError(NotAuthorized, http.StatusUnauthorized, w)
			return
		}
		var erro error
		if status == true {
			session, erro = NewS(store, sessData.Cookie, sessData.Id, sessData.Email, r, sessData.SessionName, sessData.Gothic)
			fmt.Println(session)
			if err != nil && erro != nil {
				utils.GetError(NotAuthorized, http.StatusUnauthorized, w)
				return
			}

		}

		if status == false && sessData.Email == "" && sessData.Id == "" {
			utils.GetError(NotAuthorized, http.StatusUnauthorized, w)
			return
		}

		if sessData.Gothic != nil {
			SessionEmail = sessData.GothicEmail
		} else {
			SessionEmail = session.Values["email"].(string)
		}

		// use is coming in newly, no cookies
		if session.IsNew == true {
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

		store := NewMongoStore(utils.GetCollection(session_collection), au.configs.SessionMaxAge, true, []byte(secretKey))
		session, err := store.Get(r, sessionKey)
		status, _, sessData := GetSessionDataFromToken(r, hmacSampleSecret)

		if err != nil {
			fmt.Println(session)
			utils.GetError(NotAuthorized, http.StatusUnauthorized, w)
			return
		}

		if status == false && sessData.Email == "" {
			nextHandler.ServeHTTP(w, r)
			return
		} else {
			ctx := context.WithValue(r.Context(), UserDetails, &sessData)
			r = r.WithContext(ctx)
			nextHandler.ServeHTTP(w, r)
		}

	}
}

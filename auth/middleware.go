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

		store := NewMongoStore(utils.GetCollection(session_collection), SESSION_MAX_AGE, true, []byte(secretKey))
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
			Email: session.Values["email"].(string),
		}

		ctx := context.WithValue(r.Context(), "user", user)
		nextHandler.ServeHTTP(w, r.WithContext(ctx))
	}
}
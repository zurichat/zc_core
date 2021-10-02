package realtime

import (
	"encoding/json"
	"net/http"

	"github.com/mitchellh/mapstructure"
	"zuri.chat/zccore/auth"
	"zuri.chat/zccore/utils"
)

// Creates a 'not authorized' response for given user connection request
func CentrifugoNotAuthenticatedResponse(w http.ResponseWriter) {
	data := CentrifugoConnectResponse{}
	data.Result.User = ""
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(data)
}

// CentifugoConnectAuth returns the user ID of an authenticated user from the
// bearer token in the original connect request or returns an error if the user
// is not found
func CentifugoConnectAuth(r *http.Request) (userID string, err error) {
	// 1. Validate the token
	configuration := utils.NewConfigurations()
	signingKey := configuration.HmacSampleSecret
	status, err, sessionData := auth.GetSessionDataFromToken(r, []byte(signingKey))
	if err != nil {
		return "", err
	}

	// 2. Check for a user record that's assigned this token
	if status {
		userID, err = UserIDFromSession(sessionData, *configuration)
		if err != nil {
			return "", nil
		}
		// 3. Return user ID and nil error if user is found
		return userID, nil
	}
	return "", err
}

// Extract the token from the request header
func ExtractHeaderToken(r *http.Request) string {
	headerToken := r.Header.Get("Authorization")
	return headerToken
}

// Extract user id from token
func UserIDFromSession(sessionData auth.ResToken, conf utils.Configurations) (string, error) {
	var data map[string]interface{}
	mapstructure.Decode(sessionData, data)
	session, err := utils.GetMongoDbDoc(conf.SessionDbCollection, data)
	if err != nil {
		return "", err
	}
	return session["user_id"].(string), nil
}

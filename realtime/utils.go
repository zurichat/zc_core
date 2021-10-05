package realtime

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
	"zuri.chat/zccore/auth"
	"zuri.chat/zccore/utils"
)

var (
	ExceptOrigins      = []string{"https://zuri.chat/", "https://zuri.chat", "http://zuri.chat", "https://www.zuri.chat"}
	CDcollection       = "rtcconnections"
	MaxConnectionCount = 40
)

type ConnectionDocument struct {
	Origin string `bson:"origin"  json:"origin"`
	Expiry int    `bson:"expiry"  json:"expiry"`
}
type Disc struct {
	Code      int    `bson:"code"  json:"code"`
	Reconnect bool   `bson:"reconnect"  json:"reconnect"`
	Reason    string `bson:"reason"  json:"reason"`
}
type CustomAthResp struct {
	Disconnect Disc `bson:"disconnect"  json:"disconnect"`
}

func contains(v string, a []string) bool {
	for _, i := range a {
		if i == v {
			return true
		}
	}
	return false
}

func ConnectLimitError(count int) error {
	err := "Max Connection limit of " + fmt.Sprintf("%v", MaxConnectionCount) + " Exceeded"
	return errors.New(err)
}

func CheckOrigin(r *http.Request) (string, bool) {
	origin := r.Header["Origin"][0]
	if !contains(origin, ExceptOrigins) {
		return origin, false
	}
	return origin, true
}

func GetandSetDb(collection string, expiry int) {
	grt, cc_filter := make(map[string]interface{}), make(map[string]interface{})
	grt["$lt"] = int(time.Now().Unix())
	cc_filter["expiry"] = grt
	utils.DeleteManyMongoDoc(collection, cc_filter)
}

func CheckOriginConnections(origin string) (int, bool) {
	res, _ := utils.GetMongoDbDocs(CDcollection, bson.M{"origin": origin})
	c_count := len(res)
	if c_count >= MaxConnectionCount {
		return c_count, false
	}

	return c_count, true
}

func CustomAthResponse(w http.ResponseWriter, code int, reconnect bool, reason string) {
	inn := Disc{Code: code, Reconnect: reconnect, Reason: reason}
	fRes := CustomAthResp{Disconnect: inn}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(fRes); err != nil {
		log.Printf("Error sending response: %v", err)
	}
	fmt.Println(fRes)
}
func AuthorizeOrigin(r *http.Request) error {
	GetandSetDb(CDcollection, expiry)
	origin, status := CheckOrigin(r)
	if !status {
		nc_count, state := CheckOriginConnections(origin)
		if !state {
			return ConnectLimitError(nc_count)
		}
		dt := ConnectionDocument{Origin: origin, Expiry: int(time.Now().Unix()) + expiry}
		detail, _ := utils.StructToMap(dt)

		_, err := utils.CreateMongoDbDoc(CDcollection, detail)

		if err != nil {
			return err
		}
	}
	return nil
}

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
	mapstructure.Decode(sessionData, &data)
	session, err := utils.GetMongoDbDoc(conf.SessionDbCollection, data)
	if err != nil {
		return "", err
	}
	return session["user_id"].(string), nil
}

// Get session data from token string
func TokenStringClaims(bearerToken string, hmacSampleSecret []byte) (claimsInfo map[string]interface{}, err error) {

	if bearerToken == "" {
		return nil, errors.New("authorization access failed")
	}

	tokenKey, err := jwt.Parse(bearerToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return hmacSampleSecret, nil
	})

	if claims, ok := tokenKey.Claims.(jwt.MapClaims); ok && tokenKey.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}

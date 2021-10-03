package realtime

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"zuri.chat/zccore/utils"
)

var (
	ConectionCount string
	validate       = validator.New()
	expiry         = 60 * 30
)

type Channels struct {
	ChannelList []string `json:"channel" bson:"channel"`
}

type CentrifugoConnectResult struct {
	User     string `json:"user" bson:"user"`
	ExpireAt int    `json:"expire_at" bson:"expire_at"`
	// Channels Channels `json:"channels" bson:"channels"`
}

type CentrifugoConnectResponse struct {
	Result CentrifugoConnectResult `json:"result" bson:"result"`
}

type CentrifugoRefreshResponse struct {
	Result CentrifugoRefreshResult `json:"result" bson:"result"`
}
type CentrifugoRefreshResult struct {
	ExpireAt string `json:"expire_at" bson:"expire_at"`
}

type CentrifugoClientData map[string]string

type CentrifugoConnectRequest struct {
	Client    string               `json:"client" bson:"client"`
	Transport string               `json:"transport" bson:"transport"`
	Protocol  string               `json:"protocol" bson:"protocol"`
	Encoding  string               `json:"encoding" bson:"encoding"`
	Data      CentrifugoClientData `json:"data" bson:"data"`
}

func Auth(w http.ResponseWriter, r *http.Request) {
	// 1. Decode the request from centrifugo
	var creq CentrifugoConnectRequest
	err := json.NewDecoder(r.Body).Decode(&creq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 2. Authenticate client connect request
	token := creq.Data["bearer"]
	// 2.1: Validate token
	conf := utils.NewConfigurations()
	claims, err := TokenStringClaims(token, []byte(conf.HmacSampleSecret))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// 2.2: Get user ID from validated token
	userEmail := claims["email"]
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := utils.GetMongoDbDoc(conf.UserDbCollection, bson.M{"email": userEmail})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	primitiveID := user["_id"]
	userID := primitiveID.(primitive.ObjectID).Hex()
	fmt.Println(token, userID)

	result := &CentrifugoConnectResult{
		User:     userID,
		ExpireAt: int(time.Now().Unix()) + expiry,
	}

	data := &CentrifugoConnectResponse{
		Result: *result,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func Refresh(w http.ResponseWriter, r *http.Request) {

	data := CentrifugoRefreshResponse{}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(data)

}

func Test(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./realtime/test_rtc.html")
}

func PublishEvent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var event utils.Event
	err := utils.ParseJsonFromRequest(r, &event)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}
	if err = validate.Struct(event); err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}
	res := utils.Emitter(event)
	utils.GetSuccess("publish event status", res, w)

}

package realtime

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid"

	"zuri.chat/zccore/utils"
)

var (
	ConectionCount     string
	validate           = validator.New()
	MaxConnectionCount = 40
	expiry             = 60 * 30
)

type Channels struct {
	ChannelList []string `json:"channel" bson:"channel"`
}

type CentrifugoConnectResult struct {
	User     string   `json:"user" bson:"user"`
	ExpireAt int      `json:"expire_at" bson:"expire_at"`
	Channels Channels `json:"channels" bson:"channels"`
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

type CentrifugoConnectRequest struct {
	Client    string `json:"client" bson:"client"`
	Transport string `json:"transport" bson:"transport"`
	Protocol  string `json:"protocol" bson:"protocol"`
	Encoding  string `json:"encoding" bson:"encoding"`
}

func Auth(w http.ResponseWriter, r *http.Request) {
	// 1. Decode the request from centrifugo
	erro := AuthorizeOrigin(r)
	if erro != nil {
		CustomAthResponse(w, 4001, false, fmt.Sprintf("%v", erro))
		return
	}

	var creq CentrifugoConnectRequest
	err := json.NewDecoder(r.Body).Decode(&creq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 2. Authenticate user
	headerToken := ExtractHeaderToken(r)
	if headerToken == "" {
		CentrifugoNotAuthenticatedResponse(w)
	} else {
		// 3. Generate a response object. In final version you have to
		// check that this person is authenticated
		u, _ := uuid.NewV4()
		userID, err := CentifugoConnectAuth(r)
		if err != nil {
			CentrifugoNotAuthenticatedResponse(w)
		} else {
			data := CentrifugoConnectResponse{}
			data.Result.User = u.String()
			data.Result.User = userID
			data.Result.ExpireAt = time.Now().Second() + expiry

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(data)
		}
	}
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

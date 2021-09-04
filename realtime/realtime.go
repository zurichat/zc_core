package realtime

import (
	"encoding/json"
	"net/http"
)

type CentrifugoConnectResult struct {
	User string `json:"user" bson:"user"`
}

type CentrifugoConnectResponse struct {
	Result CentrifugoConnectResult `json:"result" bson:"oresult"`
}

func Auth(w http.ResponseWriter, r *http.Request) {

	data := CentrifugoConnectResponse{}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(data)
}

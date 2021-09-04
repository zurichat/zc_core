package realtime

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"

	uuid "github.com/gofrs/uuid"
)

type CentrifugoConnectResult struct {
	User string `json:"user" bson:"user"`
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

	fmt.Println("Entered Auth")

	// Save a copy of this request for debugging.
	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(requestDump))

	var creq CentrifugoConnectRequest

	err = json.NewDecoder(r.Body).Decode(&creq)
	if err != nil {
		// http.Error(w, err.Error(), http.StatusBadRequest)
		// return
	}

	// Do something with the Person struct...
	// fmt.Fprintf(w, "Person: %+v", creq)

	u, _ := uuid.NewV4()

	data := CentrifugoConnectResponse{}
	data.Result.User = u.String()

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

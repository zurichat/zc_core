package messaging

import (
	"fmt"
	"net/http"
)

func Message(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "This Is the Messaging Endpoint\n")
}

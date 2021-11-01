package organizations

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func getRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	return router
}

// Helper function to process a request and test its response
func getHTTPResponse(t *testing.T, r *mux.Router, req *http.Request) *httptest.ResponseRecorder {

	// Create a response recorder
	w := httptest.NewRecorder()

	// Create the service and process the above request.
	r.ServeHTTP(w, req)

	return w
}

func assertStatusCode(t *testing.T, got, expected int) {
	if got != expected {
		t.Errorf("got status %d expected status %d", got, expected)
	}
}

func assertResponseMessage(t *testing.T, got, expected string) {
	if got != expected {
		t.Errorf("got message: %q expected: %q", got, expected)
	}
}

func parseResponse(w *httptest.ResponseRecorder) map[string]interface{} {
	res := make(map[string]interface{})
	json.NewDecoder(w.Body).Decode(&res)
	return res
}
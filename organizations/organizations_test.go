package organizations

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"zuri.chat/zccore/utils"
)

var configs = utils.NewConfigurations()
var orgs = NewOrganizationHandler(configs, nil)
const defaultUser string = "testUser@gmail.com"

func TestCreateOrganization(t *testing.T) {
	t.Run("test for invalid json request body", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/organizations", nil)
		if err != nil {
			t.Fatal(err)
		}
		response := httptest.NewRecorder()
		orgs.Create(response, req)
		assertStatusCode(t, response.Code, http.StatusBadRequest)
	})

	t.Run("test for wrong request body key", func(t *testing.T) {
		var requestBody = []byte(`{"creat_email": "badmailformat.xyz"}`)

		req, err := http.NewRequest("POST", "/organizations", bytes.NewBuffer(requestBody))
		if err != nil {
			t.Fatal(err)
		}
		response := httptest.NewRecorder()
		orgs.Create(response, req)
		assertStatusCode(t, response.Code, http.StatusBadRequest)
		assertResponseMessage(t, parseResponse(response)["message"].(string), "invalid email format : ")
	})

	t.Run("test for bad email format", func(t *testing.T) {
		var requestBody = []byte(`{"creator_email": "badmailformat.xyz"}`)

		req, err := http.NewRequest("POST", "/organizations", bytes.NewBuffer(requestBody))
		if err != nil {
			t.Fatal(err)
		}
		response := httptest.NewRecorder()
		orgs.Create(response, req)
		assertStatusCode(t, response.Code, http.StatusBadRequest)
		assertResponseMessage(t, parseResponse(response)["message"].(string), "invalid email format : badmailformat.xyz")
	})

	t.Run("test for non existent user", func(t *testing.T) {
		var requestBody = []byte(`{"creator_email": "notuser@gmail.com"}`)

		req, err := http.NewRequest("POST", "/organizations", bytes.NewBuffer(requestBody))
		if err != nil {
			t.Fatal(err)
		}
		response := httptest.NewRecorder()
		orgs.Create(response, req)
		assertStatusCode(t, response.Code, http.StatusBadRequest)
		assertResponseMessage(t, parseResponse(response)["message"].(string), "user with this email does not exist")
	})

	/*
		The below test requires that the user has to exist first!
	*/

	t.Run("test for successful organization creation", func(t *testing.T) {
		var requestBody = []byte(`{"creator_email": "testUser@gmail.com"}`)

		req, err := http.NewRequest("POST", "/organizations", bytes.NewBuffer(requestBody))
		if err != nil {
			t.Fatal(err)
		}
		response := httptest.NewRecorder()
		orgs.Create(response, req)
		assertStatusCode(t, response.Code, http.StatusOK)

		// assert that the created org owner is a member of the org
		// assert that the created org id is added to the user organizations list
	})
}

func TestGetOrganization(t *testing.T) {
	t.Run("test for invalid id fails", func(t *testing.T) {
		r := getRouter()
		r.HandleFunc("/organizations/{id}", orgs.GetOrganization).Methods("GET")
		req, _ := http.NewRequest("GET", "/organizations/12345", nil)

		response := getHTTPResponse(t, r, req)

		assertStatusCode(t, response.Code, http.StatusBadRequest)
	})
	
	t.Run("test for unknown org id fails", func(t *testing.T) {
		r := getRouter()
		r.HandleFunc("/organizations/{id}", orgs.GetOrganization).Methods("GET")
		req, _ := http.NewRequest("GET", "/organizations/61695d8bb2cc8a9af4833d46", nil)

		response := getHTTPResponse(t, r, req)
		assertStatusCode(t, response.Code, http.StatusNotFound)
	})
}
package organizations

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"zuri.chat/zccore/auth"
	"zuri.chat/zccore/utils"
)

var configs = utils.NewConfigurations()
var orgs = NewOrganizationHandler(configs, nil)

const defaultUser string = "testuser@gmail.com"
var defaultOrgID string
var defaultOrgUrl string = "zurichat-bwz1418.zurichat.com"

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
		var requestBody = []byte(fmt.Sprintf(`{"creator_email": "%s"}`, defaultUser))

		req, err := http.NewRequest("POST", "/organizations", bytes.NewBuffer(requestBody))
		if err != nil {
			t.Fatal(err)
		}
		response := httptest.NewRecorder()
		orgs.Create(response, req)
		assertStatusCode(t, response.Code, http.StatusOK)

		// assert that the created org owner is a member of the org
		res := parseResponse(response)
		data := res["data"].(map[string]interface{})
		orgID := data["organization_id"].(string)

		memDoc, _ := utils.GetMongoDBDoc(MemberCollectionName, bson.M{"org_id": orgID, "email": defaultUser})
		if memDoc == nil{
			t.Errorf("user %s not found in org %s", defaultUser, orgID)
		}

		user, _ := auth.FetchUserByEmail(bson.M{"email": defaultUser})
		if user.Organizations[0] != orgID{
			t.Errorf("org %s, not in user %s workspaces",orgID, defaultUser)
		}
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

	t.Run("test can get org with valid id", func(t *testing.T) {
		r := getRouter()
		r.HandleFunc("/organizations/{id}", orgs.GetOrganization).Methods("GET")
		req, _ := http.NewRequest("GET", fmt.Sprintf("/organizations/%s", defaultOrgID), nil)

		response := getHTTPResponse(t, r, req)
		assertStatusCode(t, response.Code, http.StatusOK)
	})
}

func TestGetOrganizationByURL(t *testing.T) {
	t.Run("test for invalid url fails", func(t *testing.T) {
		r := getRouter()
		r.HandleFunc("/organizations/url/{url}", orgs.GetOrganizationByURL).Methods("GET")
		req, _ := http.NewRequest("GET", "/organizations/url/www.google.com", nil)

		response := getHTTPResponse(t, r, req)

		assertStatusCode(t, response.Code, http.StatusNotFound)
	})
	
	t.Run("test for invalid url fails", func(t *testing.T) {
		r := getRouter()
		r.HandleFunc("/organizations/url/{url}", orgs.GetOrganizationByURL).Methods("GET")
		req, _ := http.NewRequest("GET", fmt.Sprintf("/organizations/url/%s", defaultOrgUrl), nil)
		
		response := getHTTPResponse(t, r, req)

		assertStatusCode(t, response.Code, http.StatusOK)
	})
}

func TestDeleteOrganization(t *testing.T) {
	t.Run("test for invalid id fails", func(t *testing.T) {
		r := getRouter()
		r.HandleFunc("/organizations/{id}", orgs.DeleteOrganization).Methods("DELETE")
		req, _ := http.NewRequest("DELETE", "/organizations/12345", nil)

		response := getHTTPResponse(t, r, req)

		assertStatusCode(t, response.Code, http.StatusInternalServerError)
	})
	
	t.Run("test can delete organization", func(t *testing.T) {
		id, err := setUpOrganization()
		if err != nil{
			t.Fail()
		}

		r := getRouter()
		r.HandleFunc("/organizations/{id}", orgs.DeleteOrganization).Methods("DELETE")
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/organizations/%s", id), nil)
		
		response := getHTTPResponse(t, r, req)

		assertStatusCode(t, response.Code, http.StatusOK)
	})
}

func TestUpdateURL(t *testing.T) {
	t.Run("test for invalid id fails", func(t *testing.T) {
		r := getRouter()
		r.HandleFunc("/organizations/{id}", orgs.UpdateURL).Methods("PATCH")
		var requestBody = []byte(`{"workspace_url": "https://www.zuri.chat/zuri"}`)
	
		req, err := http.NewRequest("PATCH", "/organizations/12345", bytes.NewBuffer(requestBody))
		if err != nil {
			t.Fatal(err)
		}

		response := getHTTPResponse(t, r, req)

		assertStatusCode(t, response.Code, http.StatusBadRequest)
	})

	t.Run("test for invalid requestbody", func(t *testing.T) {
		r := getRouter()
		r.HandleFunc("/organizations/{id}", orgs.UpdateURL).Methods("PATCH")
	
		req, err := http.NewRequest("PATCH", "/organizations/614701b3845b436ea04d1122", nil)
		if err != nil {
			t.Fatal(err)
		}

		response := getHTTPResponse(t, r, req)

		assertStatusCode(t, response.Code, http.StatusUnprocessableEntity)
	})

	// test that update is successful
}
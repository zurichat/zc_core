package organizations

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"zuri.chat/zccore/service"
	"zuri.chat/zccore/utils"
)

var configs = utils.NewConfigurations()
var mailService = service.NewZcMailService(configs)

var orgs = NewOrganizationHandler(configs, mailService)

// test user email and password
var testUserEmail = "utukphd@gmail.com"
var testUserPassword = "ootook349"

func TestMain(m *testing.M) {
	// load .env file if it exists
	err := godotenv.Load("../.env")
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	fmt.Println("Environment variables successfully loaded. Starting application...")

	if err = utils.ConnectToDB(os.Getenv("CLUSTER_URL")); err != nil {
		fmt.Println("Could not connect to MongoDB")
	}
	m.Run()
}

func TestCreateOrganization(t *testing.T) {
	var requestBody = []byte(`{"creator_email": "utukphd@gmail.com"}`)

	req, err := http.NewRequest("POST", "/organizations", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	response := httptest.NewRecorder()
	orgs.Create(response, req)
	assertStatusCode(t, response.Code, http.StatusOK)
}

func TestGetOrganization(t *testing.T) {

}

func assertStatusCode(t *testing.T, got, expected int) {
	if got != expected {
		t.Errorf("got status %d expected stats %d", got, expected)
	}
}

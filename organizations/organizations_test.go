package organizations

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/suite"
	"zuri.chat/zccore/utils"
)

// We'll be able to store suite-wide
// variables and add methods to this
// test suite struct.
type OrganizationTestSuite struct {
	suite.Suite
	orghandler *OrganizationHandler
}
  
// This is an example test that will always succeed.
func (suite *OrganizationTestSuite) TestGetNonExistingIDFails() {
	mu := mux.NewRouter()
	mu.HandleFunc("/organizations/{id}", suite.orghandler.GetOrganization).Methods("GET")

	req, _ := http.NewRequest("GET", "/organizations/6145c915285e4a1840207403", nil)

	testHTTPResponse(suite.T(), mu, req, func(w *httptest.ResponseRecorder) bool {
		statusOK := w.Code == http.StatusNotFound
		return statusOK
	})
}

func (suite *OrganizationTestSuite) TestCreateOrgWithBadBodyRequestFail() {
	mu := mux.NewRouter()
	mu.HandleFunc("/organizations/{id}", suite.orghandler.Create).Methods("GET")
	
	badpayload := []byte(`{"creato_email":"test@test.com"}`)
	req, _ := http.NewRequest("GET", "/organizations/6145c915285e4a1840207403", bytes.NewBuffer(badpayload))

	testHTTPResponse(suite.T(), mu, req, func(w *httptest.ResponseRecorder) bool {
		statusOK := w.Code == http.StatusBadRequest
		return statusOK
	})
}

// This will run before before the tests in the suite are run.
func (suite *OrganizationTestSuite) SetupSuite() {
	configs := utils.NewConfigurations()
	suite.orghandler = NewOrganizationHandler(configs, nil)

	os.Setenv("DB_NAME", "zurichat_test")
	os.Setenv("CLUSTER_URL", "mongodb://127.0.0.1:27017")

	if err := utils.ConnectToDB(os.Getenv("CLUSTER_URL")); err != nil {
		log.Fatal("Could not connect to MongoDB")
	}
}
  
  // We need this function to kick off the test suite, otherwise
  // "go test" won't know about our tests.
func TestOrganizationTestSuite(t *testing.T) {
	suite.Run(t, new(OrganizationTestSuite))
}

func testHTTPResponse(t *testing.T, r *mux.Router, req *http.Request, f func(w *httptest.ResponseRecorder) bool) {
	// Create a response recorder
	t.Helper()

	w := httptest.NewRecorder()

	// Create the service and process the above request.
	r.ServeHTTP(w, req)

	if !f(w) {
		t.Fail()
	}
}
package organizations

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"zuri.chat/zccore/user"
	"zuri.chat/zccore/utils"
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

func TestMain(m *testing.M) {
	// load .env file if it exists
	err := godotenv.Load("../.testenv")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	fmt.Println("Environment variables successfully loaded. Starting application...")

	if err = utils.ConnectToDB(os.Getenv("CLUSTER_URL")); err != nil {
		log.Fatal("Could not connect to MongoDB")
	}

	// drop database if it is available 
	ctx := context.TODO()
	utils.GetDefaultMongoClient().Database(os.Getenv("DB_NAME")).Drop(ctx)
	fmt.Printf("\n\n")

	err = setUpUserAccount()
	if err != nil {
		log.Fatal(err.Error())
	}

	exitVal := m.Run()

	// drop database after running all tests
	ctx = context.TODO()
	utils.GetDefaultMongoClient().Database(os.Getenv("DB_NAME")).Drop(ctx)

    os.Exit(exitVal)
}

func setUpUserAccount() error{
	user := user.User{
		Email: defaultUser,
		Deactivated: false,
		IsVerified: true,
	}

	result, _ := utils.GetMongoDBDoc(UserCollectionName, bson.M{"email": user.Email})
	if result != nil {
		return fmt.Errorf("user %s exists", user.Email)
	}

	detail, _ := utils.StructToMap(user)
	_, err := utils.CreateMongoDBDoc(UserCollectionName, detail)

	if err != nil {
		return err
	}
	
	return err
}
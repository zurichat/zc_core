package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"

	"github.com/joho/godotenv"
	"zuri.chat/zccore/service"
	"zuri.chat/zccore/utils"
)

var (
	configs = utils.NewConfigurations()
	mailService = service.NewZcMailService(configs)
	au = NewAuthHandler(configs, mailService)
	// us = user.NewUserHandler(configs, mailService)
)

/*
	Tests
	1. TestLogin
	2. TestIsAuthenticated
	3. TestOptionalAuthentication
	4. TestIsAuthorized
	3. TestValidateCode
	4. TestVerifyAccount
	5. TestVerifyPasswordResetCode
	6. TestUpdatePassword
	7. TestRequestResetPasswordCode
*/

func TestMain(m *testing.M) {
	// load .env file if it exists
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	if err = utils.ConnectToDB(os.Getenv("CLUSTER_URL")); err != nil {
		log.Fatal("Could not connect to MongoDB")
	}

	fmt.Printf("\n\n")
	exitVal := m.Run()

	os.Exit(exitVal)
}

func TestLogin(t *testing.T) {
	requestURI := url.URL{ Path: "/auth/login"}
	
	// might need to seed data if database is new

	tests := []struct {
		Name         string
		RequestBody  Credentials
		ExpectedCode int		
	}{
		{
			Name:   "OK",
			RequestBody: Credentials{
				Email: "bbnaija2025@gmail.com",
				Password: "password3245",
			},
			ExpectedCode: http.StatusCreated,			
		},
		{
			Name: "should throw account not verified error",
			RequestBody: Credentials{
				Email: "bbnaija2025@gmail.com",
				Password: "password3245",				
			},
			ExpectedCode: http.StatusBadRequest,
		},
		{
			Name:   "should fail for incorrect password",
			RequestBody: Credentials{
				Email: "john.doe@workable.com",
				Password: "password223666",
			},
			ExpectedCode: http.StatusBadRequest,			
		},
		{
			Name:   "should throw error for invalid email",
			RequestBody: Credentials{
				Email: "enigbe.enike.com",
				Password: "password34223",
			},
			ExpectedCode: http.StatusBadRequest,			
		},				
	}

	// run test
	for _, test := range tests {
		fn := func(t *testing.T) {
			var b bytes.Buffer
			json.NewEncoder(&b).Encode(test.RequestBody)

			req, err := http.NewRequest(http.MethodPost, requestURI.String(), &b)
			if err != nil {
				t.Fatal(err)
			}
			
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Content-length", strconv.FormatInt(req.ContentLength, 10))

			defer func() {
				if err := req.Body.Close(); err != nil {
					t.Errorf("error encountered closing request body: %v", err)
				}
			}()			
			
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(au.LoginIn)
			handler.ServeHTTP(rr, req)	
			
			// fmt.Print(rr.Body)
			if status := rr.Code; status != test.ExpectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v", status, test.ExpectedCode)
				return
			}			
		}

		t.Run(test.Name, fn)
	}
}
package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"zuri.chat/zccore/auth"
	"zuri.chat/zccore/service"
	"zuri.chat/zccore/user"
	"zuri.chat/zccore/utils"
)

var (
	configs = utils.NewConfigurations()
	mailService = service.NewZcMailService(configs)
	au = auth.NewAuthHandler(configs, mailService)
	us = user.NewUserHandler(configs, mailService)
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

func TestLogin(t *testing.T) {
	requestURI := url.URL{ Path: "/auth/login"}
	
	// might need to seed data if database is new

	tests := []struct {
		Name         string
		RequestBody  auth.Credentials
		ExpectedCode int		
	}{
		{
			Name:   "OK",
			RequestBody: auth.Credentials{
				Email: "john.doe@workable.com",
				Password: "password",
			},
			ExpectedCode: http.StatusCreated,			
		},
		{
			Name:   "Incorrect Credentials",
			RequestBody: auth.Credentials{
				Email: "john.doe@workable.com",
				Password: "password223666",
			},
			ExpectedCode: http.StatusBadRequest,			
		},
		{
			Name:   "Invalid Email Address",
			RequestBody: auth.Credentials{
				Email: "enigbe.enike.com",
				Password: "password34223",
			},
			ExpectedCode: http.StatusUnprocessableEntity,			
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

			defer func() {
				if err := req.Body.Close(); err != nil {
					t.Errorf("error encountered closing request body: %v", err)
				}
			}()			
			
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(au.LoginIn)
			handler.ServeHTTP(rr, req)	
			
			status := rr.Code
			// fmt.Print(rr.Body)
			if status != test.ExpectedCode {
				t.Errorf("handler returned wrong status code: got %v want %v", status, test.ExpectedCode)
				return
			}			
		}

		t.Run(test.Name, fn)
	}
	
}
package auth

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"zuri.chat/zccore/plugin"
	"zuri.chat/zccore/user"
	"zuri.chat/zccore/utils"
)
const (
	secretKey = "5d5c7f94e29ba11f6822a2be310d3af4"
	user_collection = "users"
)

var validate = validator.New()

func printStruct(v interface{}) {
	fmt.Printf("%+v\n", v)
}

func LoginIn(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	var authDetails Authentication
	if err := utils.ParseJsonFromRequest(request, &authDetails); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, response)
		return
	}
	
	if err := validate.Struct(authDetails); err != nil {
		utils.GetError(err, http.StatusBadRequest, response)
		return
	}

	// var user user.User
	user := &user.User{}
	// check if user exists
	result, _ := utils.GetMongoDbDoc(user_collection, bson.M{"email": authDetails.Email})
	if result != nil {
		utils.GetError(errors.New("Something went wrong!"), http.StatusBadRequest, response)
		return
	}
	if err := plugin.MapToStruct(result, user); err != nil {}
	
	printStruct(result)
	// check password
	check := CheckPassword(authDetails.Password, user.Password)
	if !check {
		utils.GetError(
			errors.New("Invalid login credentials, confirm and try again"), 
			http.StatusBadRequest, 
			response,
		)		
	}

	vtoken, err := GenerateJWT(authDetails.Email, "")
	if err != nil {
		utils.GetError(errors.New("Failed to generate token"), http.StatusBadRequest, response)
		return
	}

	token := &Token{
		Email: user.Email,
		OrganizationID: "",
		TokenString: vtoken,		
	}
	utils.GetSuccess("login successful", token, response)
}
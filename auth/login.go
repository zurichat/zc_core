package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"zuri.chat/zccore/utils"
)

const (
	secretKey       = "5d5c7f94e29ba11f6822a2be310d3af4"
	user_collection = "users"
)

var (
	validate           = validator.New()
	UserNotFound       = errors.New("User not found!")
	InvalidCredentials = errors.New("Invalid login credentials, confirm and try again")
)

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

	user, err := fetchUserByEmail(bson.M{"email":  strings.ToLower(authDetails.Email)})
	if err != nil {
		utils.GetError(UserNotFound, http.StatusBadRequest, response)
		return
	}
	// check password
	check := CheckPassword(authDetails.Password, user.Password)
	if !check {
		utils.GetError(InvalidCredentials, http.StatusBadRequest, response)
		return
	}

	vtoken, err := GenerateJWT(user.ID.Hex(), authDetails.Email)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, response)
		return
	}

	token := &Token{
		TokenString: vtoken,
		User: *user,
	}
	utils.GetSuccess("login successful", token, response)
}

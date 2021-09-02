package user

import (
	"errors"
	"net/http"
	"time"

	"zuri.chat/zccore/utils"
)

// An end point to create new users
func UserRegistration(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	var user User
	
	err := utils.ParseJsonFromRequest(request, &user)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, response)
		return
	}

	isEmail := utils.IsValidEmail(user.Email)

	if !isEmail {
		utils.GetError(errors.New("email address is not valid"), http.StatusBadRequest, response)
		return
	} 
		
	user.CreatedAt = time.Now()

	detail, _ := utils.StructToMap(user, "bson")

	result, err := utils.CreateMongoDbDoc("users", detail)
	
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
		return
	}

	utils.GetSuccess("user created", result, response)

}
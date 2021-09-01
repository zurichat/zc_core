package auth

import (
	"errors"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"zuri.chat/zccore/models"
	"zuri.chat/zccore/utils"
)

func RegisterNewUser(w http.ResponseWriter, r *http.Request){
	var user models.User
	if err := utils.ParseJsonFromRequest(r, &user); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}
	_email := user.Email
	email := bson.M{"email": _email}

	res := utils.DocExists("Users", email)
	if (!res){
		utils.GetError(errors.New("user already exists with this email"), http.StatusBadRequest, w)
	}
	doc, _ := utils.StructToMap(user, "bson")
	result, err := utils.CreateMongoDbDoc("Users", doc)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}
	utils.GetSuccess("User successfully created", result, w)

}
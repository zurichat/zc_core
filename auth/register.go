package auth

import (
	"errors"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
	"zuri.chat/zccore/utils"
)

func RegisterNewUser(w http.ResponseWriter, r *http.Request){
	var user User
	if err := utils.ParseJsonFromRequest(r, &user); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}
	email := bson.M{"email": user.email}

	res := utils.DocExists("Users", email)
	if (!res){
		utils.GetError(errors.New("user already exists with this email"), http.StatusBadRequest, w)
	}
	password, _ := bcrypt.GenerateFromPassword([]byte(user.password), 14)
	user.password = password
	user.created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
    user.updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	doc, _ := utils.StructToMap(user, "bson")
	result, err := utils.CreateMongoDbDoc("Users", doc)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}
	utils.GetSuccess("User successfully created", result, w)

}
package auth

import (
	"errors"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
	"zuri.chat/zccore/utils"
)

type User struct {
	FirstName         string             `bson:"first_name" validate:"required,min=2,max=100"`
	LastName          string             `bson:"last_name" validate:"required,min=2,max=100"`
	Email             string             `bson:"email" validate:"email,required"`
	Password          []byte             `bson:"password" validate:"required,min=6"`
	Phone             string             `bson:"phone" validate:"required"`
	Company           string `bson:"company"`
	Timezone          string
	CreatedAt         time.Time `bson:"created_at"`
	UpdatedAt         time.Time `bson:"updated_at"`
	DeletedAt         time.Time `bson:"deleted_at"`
	
}

func RegisterNewUser(w http.ResponseWriter, r *http.Request){
	user := new(User)
	if err := utils.ParseJsonFromRequest(r, &user); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}
	email := bson.M{"email": user.Email}

	res := utils.DocExists("Users", email)
	if (!res){
		utils.GetError(errors.New("user already exists with this email"), http.StatusBadRequest, w)
	}
	password, _ := bcrypt.GenerateFromPassword([]byte(user.Password), 14)
	user.Password = password
	user.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
    user.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	doc, _ := utils.StructToMap(user, "bson")
	result, err := utils.CreateMongoDbDoc("Users", doc)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}
	utils.GetSuccess("User successfully created", result, w)

}
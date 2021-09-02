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
	ID                primitive.ObjectID `bson:"_id"`
	FirstName         string             `bson:"first_name" validate:"required,min=2,max=100"`
	LastName          string             `bson:"last_name" validate:"required,min=2,max=100"`
	Email             string             `bson:"email" validate:"email,required"`
	Password          string             `bson:"password" validate:"required,min=6"`
	Phone             string             `bson:"phone" validate:"required"`
	Status            Status
	Company           string `bson:"company"`
	Settings          *UserSettings
	Timezone          string
	CreatedAt         time.Time `bson:"created_at"`
	UpdatedAt         time.Time `bson:"updated_at"`
	DeletedAt         time.Time `bson:"deleted_at"`
	Workspaces        []*UserWorkspace
	EmailVerification UserEmailVerification
	PasswordResets    []*UserPasswordReset
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
	password, _ := bcrypt.GenerateFromPassword([]byte(user.password), 14)
	user.Password = password
	user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
    user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	doc, _ := utils.StructToMap(user, "bson")
	result, err := utils.CreateMongoDbDoc("Users", doc)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}
	utils.GetSuccess("User successfully created", result, w)

}
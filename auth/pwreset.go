package auth

import (
	"fmt"
	"net/http"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"zuri.chat/zccore/service"
	"zuri.chat/zccore/utils"
)

func (au *AuthHandler) VerifyMail(w http.ResponseWriter, r *http.Request) {}

func (au *AuthHandler) VerifyPasswordReset(w http.ResponseWriter, r *http.Request){}

// Send password reset code to user, auth not required
func (au *AuthHandler) RequestResetPasswordCode(w http.ResponseWriter, r *http.Request){
	email := struct {Email	string	`json:"email" validate:"email,required"`}{}

	if err := utils.ParseJsonFromRequest(r, &email); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	if err := validate.Struct(email); err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	user, err := FetchUserByEmail(bson.M{"email": strings.ToLower(email.Email)})
	if err != nil {
		utils.GetError(UserNotFound, http.StatusBadRequest, w)
		return
	}
	// Update user collection with UserPasswordReset - WIP

	data := &service.MailData{ 
		Username: user.FirstName, 
		Code: "xxxxx",
	}
	msger := au.mailService.NewMail([]string{user.Email}, "Reset Password Code", service.PasswordReset, data)
	
	if err := au.mailService.SendMail(msger); err != nil {
		fmt.Print(err.Error())
	}
}
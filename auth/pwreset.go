package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/service"
	"zuri.chat/zccore/utils"
)

func (au *AuthHandler) VerifyMail(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")
	c := struct {Code	string	`json:"code" validate:"required"`}{}

	if err := utils.ParseJsonFromRequest(r, &c); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}
	fmt.Print(c.Code)
	if err := validate.Struct(c); err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}
	
	filter := bson.M{"email_verification.token": c.Code}
	user, err := FetchUserByEmail(filter)
	if err != nil {
		utils.GetError(errors.New("Password reset code not found!"), http.StatusInternalServerError, w)
		return
	}
	
	// set email_verification null
	// update isverified to true
	id, _ := primitive.ObjectIDFromHex(user.ID)
	update := bson.M{"$set": bson.M{"email_verification": nil, "isverified": true}}

	if _, err := utils.GetCollection(user_collection).UpdateByID(
		context.Background(),
		id, update); err != nil {
			utils.GetError(err, http.StatusInternalServerError, w)
			return
		}

	utils.GetSuccess("Email verified, you can now login", nil, w)
}

func (au *AuthHandler) VerifyPasswordResetCode(w http.ResponseWriter, r *http.Request){}

// Send password reset code to user, auth not required
func (au *AuthHandler) RequestResetPasswordCode(w http.ResponseWriter, r *http.Request){
	w.Header().Add("content-type", "application/json")
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
	_, token := utils.RandomGen(6, "d")

	userPasswordReset := map[string]interface{}{
		"ip_address": strings.Split(r.RemoteAddr, ":")[0],
		"token": token,
		"expired_at": time.Now(),
		"updated_at": time.Now(),
		"created_at": time.Now(),
	}

	id, _ := primitive.ObjectIDFromHex(user.ID)
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"password_resets": userPasswordReset }}

	// update db;
	if _, err := utils.GetCollection(user_collection).UpdateOne(context.Background(), filter, update); err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return		
	}

	data := &service.MailData{ 
		Username: user.Email, 
		Code: userPasswordReset["token"].(string),
	}

	msger := au.mailService.NewMail([]string{user.Email}, "Reset Password Code", service.PasswordReset, data)

	if err := au.mailService.SendMail(msger); err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("Password reset code sent", nil, w)
}
package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"zuri.chat/zccore/service"
	"zuri.chat/zccore/user"
	"zuri.chat/zccore/utils"
)

var (
	ERROR_CONFIMATION_CODE = errors.New("Account confirmation code used or expired, confirm and try again")
	ERROR_RESET_CODE = errors.New("Invalid password reset code, already used or expired, confirm and try again")
)

func (au *AuthHandler) validateCode(r *http.Request, column string, errMsg error, vcode string)(*user.User, error) {
	c := struct {Code	string	`json:"code" validate:"required"`}{}

	var filter primitive.M
	if vcode == "" {
		if err := utils.ParseJsonFromRequest(r, &c); err != nil {
			return nil, err
		}

		if err := validate.Struct(c); err != nil {
			return nil, err
		}
		
		filter = bson.M{column: c.Code}
	} else {
		filter = bson.M{column: vcode}
	}

	user, err := FetchUserByEmail(filter)
	if err != nil {
		return nil, errMsg
	}	

	return user, nil
}


func (au *AuthHandler) VerifyAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")

	user, err := au.validateCode(r, "email_verification.token", ERROR_CONFIMATION_CODE, "")
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
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

func (au *AuthHandler) VerifyPasswordResetCode(w http.ResponseWriter, r *http.Request){
	w.Header().Add("content-type", "application/json")

	user, err := au.validateCode(r, "password_resets.token", ERROR_RESET_CODE, "")
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	resp := map[string]interface{}{
		"isverified": true,
		"verification_code": user.PasswordResets.Token,
	}

	utils.GetSuccess("Password reset code valid", resp, w)
}

func (au *AuthHandler) UpdatePassword(w http.ResponseWriter, r *http.Request){
	w.Header().Add("content-type", "application/json")

	rBody := struct{
		Password 			string	`json:"password" validate:"required"`
		ConfirmPassword 	string	`json:"confirm_password" validate:"required"`
	}{}

	params := mux.Vars(r)
	verificationToken := params["verification_code"]

	user, err := au.validateCode(r, "password_resets.token", ERROR_RESET_CODE, verificationToken)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}	

	if err := utils.ParseJsonFromRequest(r, &rBody); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	if err := validate.Struct(rBody); err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	if rBody.Password != rBody.ConfirmPassword {
		utils.GetError(ConfirmPasswordError, http.StatusBadRequest, w)
		return		
	}

	// update password & delete passwordreset object
	bytes, err := bcrypt.GenerateFromPassword([]byte(rBody.Password), 14)
	if err != nil { 
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	id, _ := primitive.ObjectIDFromHex(user.ID)
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"password_resets": nil, "password": string(bytes) }}

	if _, err := utils.GetCollection(user_collection).UpdateOne(context.Background(), filter, update); err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return		
	}	

	utils.GetSuccess("Password update successful", nil, w)
}

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

	// check if user already requested for password
	if user.PasswordResets != nil {
		x2x := map[string]interface{}{ "password_reset_code" : user.PasswordResets.Token }
		utils.GetSuccess("Password reset code already sent, check your email", x2x, w)
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

	msger := au.mailService.NewMail(
		[]string{user.Email}, 
		"Reset Password Code", service.PasswordReset, map[string]interface{}{
			"Username": user.Email, 
			"Code": userPasswordReset["token"].(string),
		})

	if err := au.mailService.SendMail(msger); err != nil {
		fmt.Printf("Error occured while sending mail: %s", err.Error())
	}

	utils.GetSuccess("Password reset code sent", nil, w)
}
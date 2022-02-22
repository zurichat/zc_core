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
	DefaultHashCode     = 14
	GenNumberLength     = 6
	ErrConfirmationCode = errors.New("Account confirmation code used or expired, confirm and try again")
	ErrResetCode        = errors.New("Invalid password reset code, already used or expired, confirm and try again")
)

func (au *AuthHandler) validateCode(r *http.Request, column string, errMsg error, vcode string) (*user.User, error) {
	c := struct {
		Code string `json:"code" validate:"required"`
	}{}

	var filter primitive.M

	if vcode == "" {
		if err := utils.ParseJSONFromRequest(r, &c); err != nil {
			return nil, err
		}

		if err := validate.Struct(c); err != nil {
			return nil, err
		}

		filter = bson.M{column: c.Code}
	} else {
		filter = bson.M{column: vcode}
	}

	u, err := FetchUserByEmail(filter)
	if err != nil {
		return nil, errMsg
	}

	return u, nil
}

func (au *AuthHandler) VerifyAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")

	u, err := au.validateCode(r, "email_verification.token", ErrConfirmationCode, "")
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// set email_verification null
	// update isverified to true
	id, _ := primitive.ObjectIDFromHex(u.ID)
	update := bson.M{"$set": bson.M{"email_verification": nil, "isverified": true}}

	if _, err := utils.GetCollection(userCollection).UpdateByID(
		context.Background(),
		id, update); err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("Email verified, you can now login", nil, w)
}

func (au *AuthHandler) VerifyPasswordResetCode(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")

	u, err := au.validateCode(r, "password_resets.token", ErrResetCode, "")
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	resp := map[string]interface{}{
		"isverified":        true,
		"verification_code": u.PasswordResets.Token,
	}

	utils.GetSuccess("Password reset code valid", resp, w)
}

func (au *AuthHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")

	rBody := struct {
		Password        string `json:"password" validate:"required,min=6"`
		ConfirmPassword string `json:"confirm_password" validate:"required"`
	}{}

	params := mux.Vars(r)
	verificationToken := params["verification_code"]

	u, err := au.validateCode(r, "password_resets.token", ErrResetCode, verificationToken)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	if e := utils.ParseJSONFromRequest(r, &rBody); e != nil {
		utils.GetError(e, http.StatusUnprocessableEntity, w)
		return
	}

	if er := validate.Struct(rBody); er != nil {
		utils.GetError(er, http.StatusBadRequest, w)
		return
	}

	if rBody.Password != rBody.ConfirmPassword {
		utils.GetError(ErrConfirmPassword, http.StatusBadRequest, w)
		return
	}

	// update password & delete passwordreset object
	bytes, err := bcrypt.GenerateFromPassword([]byte(rBody.Password), DefaultHashCode)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	id, _ := primitive.ObjectIDFromHex(u.ID)
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"password_resets": nil, "password": string(bytes)}}

	if _, err := utils.GetCollection(userCollection).UpdateOne(context.Background(), filter, update); err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("Password update successful", nil, w)
}

// Send password reset code to user, auth not required.
func (au *AuthHandler) RequestResetPasswordCode(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")

	email := struct {
		Email string `json:"email" validate:"email,required"`
	}{}

	if err := utils.ParseJSONFromRequest(r, &email); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	if err := validate.Struct(email); err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	u, err := FetchUserByEmail(bson.M{"email": strings.ToLower(email.Email)})
	if err != nil {
		utils.GetError(ErrUserNotFound, http.StatusBadRequest, w)
		return
	}

	// check if user already requested for password
	if u.PasswordResets != nil {
		x2x := map[string]interface{}{"password_reset_code": u.PasswordResets.Token}
		utils.GetSuccess("Password reset code already sent, check your email", x2x, w)

		return
	}

	// Update user collection with UserPasswordReset - WIP
	_, token := utils.RandomGen(GenNumberLength, "d")

	userPasswordReset := map[string]interface{}{
		"ip_address": strings.Split(r.RemoteAddr, ":")[0],
		"token":      token,
		"expired_at": time.Now(),
		"updated_at": time.Now(),
		"created_at": time.Now(),
	}

	id, _ := primitive.ObjectIDFromHex(u.ID)
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"password_resets": userPasswordReset}}

	// update db;
	if _, err := utils.GetCollection(userCollection).UpdateOne(context.Background(), filter, update); err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	msger := au.mailService.NewMail(
		[]string{u.Email},
		"Reset Password Code", service.PasswordReset, map[string]interface{}{
			"Username": u.Email,
			"Code":     userPasswordReset["token"].(string),
		})

	if err := au.mailService.SendMail(msger); err != nil {
		fmt.Printf("Error occurred while sending mail: %s", err.Error())
	}

	utils.GetSuccess("Password reset code sent", nil, w)
}

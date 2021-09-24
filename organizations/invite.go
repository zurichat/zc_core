package organizations

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"zuri.chat/zccore/service"
	"zuri.chat/zccore/utils"
)

// Team Invite to workspace
type UserHandler struct {
	configs     *utils.Configurations
	mailService service.MailService
}

var us UserHandler
var validate = validator.New()

func TeamInvitation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	orgURL := mux.Vars(r)["url"]
	org_collection := "organizations"

	orgDoc, _ := utils.GetMongoDbDoc(org_collection, bson.M{"workspace_url": orgURL})
	if orgDoc == nil {
		// fmt.Printf("workspace with url %s doesn't exist!", orgURL)
		utils.GetError(errors.New("organization does not exist"), http.StatusNotFound, w)
		return
	}

	email := struct {
		Email string `json:"email" validate:"email,required"`
	}{}

	if err := utils.ParseJsonFromRequest(r, &email); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	if err := validate.Struct(email); err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	if err := utils.ParseJsonFromRequest(r, &email); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	// userDoc, _ := utils.GetMongoDbDoc(user_collection, bson.M{"email": email.Email})
	// if userDoc == nil {
	// 	utils.GetError(errors.New("user with this email does not exist"), http.StatusBadRequest, w)
	// 	return
	// }

	// send invite email null
	uniquestring := randomString(36)
	link := "https://zuri.chat/confirmed_invitation/" + uniquestring
	invitationLink := map[string]interface{}{
		"ip_address": strings.Split(r.RemoteAddr, ":")[0],
		"link":       link,
		"expired_at": time.Now(),
		"updated_at": time.Now(),
		"created_at": time.Now(),
	}
	msger := us.mailService.NewMail(
		[]string{email.Email},
		"Workspace Invitation Link", service.TeamInvitation,
		&service.MailData{
			Username: email.Email,
			Url:      invitationLink["link"].(string),
		})

	if err := us.mailService.SendMail(msger); err != nil {
		fmt.Printf("Error occured while sending mail: %s", err.Error())
	}

	utils.GetSuccess("Invitation Email Sent", nil, w)
}

func randomString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwzyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_@")
	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

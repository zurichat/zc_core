package organizations

import (
	"errors"
	"fmt"
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
func TeamInvitation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	orgURL := mux.Vars(r)["url"]

	data, _ := utils.GetMongoDbDoc("organizations", bson.M{"workspace_url": orgURL})
	if data == nil {
		fmt.Printf("workspace with url %s doesn't exist!", orgURL)
		utils.GetError(errors.New("organization does not exist"), http.StatusNotFound, w)
		return
	}

	validate := validator.New()
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
	// send invite email null
	_, token := utils.RandomGen(6, "d")

	invitationCode := map[string]interface{}{
		"ip_address": strings.Split(r.RemoteAddr, ":")[0],
		"token":      token,
		"expired_at": time.Now(),
		"updated_at": time.Now(),
		"created_at": time.Now(),
	}
	type UserHandler struct {
		configs     *utils.Configurations
		mailService service.MailService
	}
	var us UserHandler
	msger := us.mailService.NewMail(
		[]string{email.Email},
		"Workspace Invitation Link", service.TeamInvitation,
		&service.MailData{
			Username: email.Email,
			Code:     invitationCode["token"].(string),
		})

	if err := us.mailService.SendMail(msger); err != nil {
		fmt.Printf("Error occured while sending mail: %s", err.Error())
	}

	utils.GetSuccess("Invitation Email Sent", nil, w)
}

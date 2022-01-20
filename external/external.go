package external

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"zuri.chat/zccore/logger"
	"zuri.chat/zccore/service"
	"zuri.chat/zccore/utils"
)

var (
	validate          = validator.New()
	ErrClientNotValid = errors.New("client type is not valid is not valid. Choose from windows, linux, mac, ios, android")
	ErrEmailNotValid  = errors.New("email address is not valid")
)

func (eh *Handler) EmailSubscription(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")


	var NewSubscription Subscription

	type subRes struct {
		status bool
	}

	err := json.NewDecoder(r.Body).Decode(&NewSubscription)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	SubDoc, _ := utils.GetMongoDBDoc(NewsletterCollection, bson.M{"email": NewSubscription.Email})
	if SubDoc != nil {
		logger.Info("%s already subscribed for newsletter", NewSubscription.Email)
		utils.GetSuccess("User already subscribed for newsletter", subRes{status: true}, w)

		return
	}

	msger := eh.mailService.NewMail(
	[]string{NewSubscription.Email}, "Zuri Chat Newsletter Subscription", service.EmailSubscription,
	map[string]interface{}{
		"Username": NewSubscription.Email,
	})

	if err = eh.mailService.SendMail(msger); err != nil {
		logger.Error("Error occurred while sending mail: %s", err.Error())
		utils.GetError(errors.New("Subscription failed"), http.StatusInternalServerError, w)

		return
	}

	coll := utils.GetCollection(NewsletterCollection)
	_, err = coll.InsertOne(r.Context(), NewSubscription)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("User successfully subscribed for newsletter", subRes{status: true}, w)
}

func (eh *Handler) DownloadClient(w http.ResponseWriter, r *http.Request) {
	var url string

	clientType := r.URL.Query().Get("client_type")
	email := r.URL.Query().Get("email")

	if !utils.IsValidEmail(email) {
		utils.GetError(ErrEmailNotValid, http.StatusBadRequest, w)
		return
	}

	windowsURL := "https://api.zuri.chat/files/applications/20210922182446_0.7z"
	defaultURL := "url not available at the moment"

	switch clientType {
	case "windows":
		url = windowsURL
	case "linux":
		url = defaultURL
	case "mac":
		url = defaultURL
	case "ios":
		url = defaultURL
	case "android":
		url = defaultURL
	default:
		utils.GetError(ErrClientNotValid, http.StatusBadRequest, w)
		return
	}

	msger := eh.mailService.NewMail(
		[]string{email},
		"Zuri Chat Desktop",
		service.DownloadClient,
		map[string]interface{}{
			"Username": email,
			"Code":     url,
		},
	)

	if err := eh.mailService.SendMail(msger); err != nil {
		fmt.Printf("Error occurred while sending mail: %s", err.Error())
	}

	utils.GetSuccess("Download Link Successfully sent", url, w)
}

func (eh *Handler) SendMail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var msgr *service.Mail

	bitSize := 8

	switch isCustomMail, _ := strconv.ParseInt(r.FormValue("custom_mail"), 0, bitSize); isCustomMail {
	case 1:
		// mailbody and content-type * content-type -> text/html and text/plain
		mail := struct {
			Email       string `json:"email" validate:"email,required"`
			Subject     string `json:"subject" validate:"required"`
			ContentType string `json:"content_type" validate:"required"`
			MailBody    string `json:"mail_body" validate:"required"`
		}{}

		if err := utils.ParseJSONFromRequest(r, &mail); err != nil {
			utils.GetError(err, http.StatusUnprocessableEntity, w)
			return
		}

		if err := validate.Struct(mail); err != nil {
			utils.GetError(err, http.StatusBadRequest, w)
			return
		}

		match, _ := regexp.MatchString("<(.|\n)*?>", mail.MailBody)
		if !match && mail.ContentType == "text/html" {
			utils.GetError(errors.New("valid html string is required for mailbody"), http.StatusBadRequest, w)
			return
		}

		msgr = eh.mailService.NewCustomMail(
			[]string{mail.Email},
			mail.Subject,
			mail.MailBody,
		)
	default:
		mail := struct {
			Email    string                 `json:"email" validate:"email,required"`
			Subject  string                 `json:"subject" validate:"required"`
			MailType int                    `json:"mail_type" validate:"required"`
			Data     map[string]interface{} `json:"data" validate:"required"`
		}{}

		if err := utils.ParseJSONFromRequest(r, &mail); err != nil {
			utils.GetError(err, http.StatusUnprocessableEntity, w)
			return
		}

		if err := validate.Struct(mail); err != nil {
			utils.GetError(err, http.StatusBadRequest, w)
			return
		}
		// ensure email is valid
		if !utils.IsValidEmail(strings.ToLower(mail.Email)) {
			utils.GetError(ErrEmailNotValid, http.StatusBadRequest, w)
			return
		}

		if _, ok := service.MailTypes[service.MailType(mail.MailType)]; !ok {
			utils.GetError(errors.New("invalid email type, email template does not exists"), http.StatusBadRequest, w)
			return
		}

		msgr = eh.mailService.NewMail(
			[]string{mail.Email},
			mail.Subject,
			service.MailType(mail.MailType),
			mail.Data,
		)
	}

	if err := eh.mailService.SendMail(msgr); err != nil {
		fmt.Printf("Error occurred while sending mail: %s", err.Error())
	}

	utils.GetSuccess("Mail sent successfully", nil, w)
}

func (eh *Handler) UnsubscribeEmail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	email := mux.Vars(r)["email"]
	field := make(map[string]interface{})

	field["email"] = email
	
	coll := utils.GetCollection(NewsletterCollection)

	response, err := coll.DeleteOne(r.Context(), field)

	if err != nil {
		logger.Error(err.Error())
	}

	if response.DeletedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("user unsubsribed successfully", nil, w)
}

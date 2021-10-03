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
	"go.mongodb.org/mongo-driver/bson"
	"zuri.chat/zccore/service"
	"zuri.chat/zccore/utils"
)

var (
	validate         = validator.New()
	CLIENT_NOT_VALID = errors.New("client type is not valid is not valid. Choose from windows, linux, mac, ios, android")
	EMAIL_NOT_VALID  = errors.New("Email address is not valid")
)

func (eh *ExternalHandler) EmailSubscription(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	newsletter_collection := "subscription"
	var NewSubscription Subscription
	type sub_res struct {
		status bool
	}
	err := json.NewDecoder(r.Body).Decode(&NewSubscription)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	SubDoc, _ := utils.GetMongoDbDoc(newsletter_collection, bson.M{"email": NewSubscription.Email})
	if SubDoc != nil {
		// fmt.Printf("user with email %s already subscribed!", NewSubscription.Email)
		msger := eh.mailService.NewMail(
			[]string{NewSubscription.Email}, "Zuri Chat Newsletter Subscription", service.EmailSubscription,
			map[string]interface{}{
				"Username": NewSubscription.Email,
			})

		if err := eh.mailService.SendMail(msger); err != nil {
			fmt.Printf("Error occured while sending mail: %s", err.Error())
		}
		utils.GetSuccess("Thanks for subscribing to for or Newsletter", sub_res{status: true}, w)
		return
	}

	coll := utils.GetCollection(newsletter_collection)
	res, err := coll.InsertOne(r.Context(), NewSubscription)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}
	fmt.Println(res.InsertedID)

	msger := eh.mailService.NewMail(
		[]string{NewSubscription.Email}, "Zuri Chat Newsletter Subscription", service.EmailSubscription,
		map[string]interface{}{
			"Username": NewSubscription.Email,
		})

	if err := eh.mailService.SendMail(msger); err != nil {
		fmt.Printf("Error occured while sending mail: %s", err.Error())
	}
	utils.GetSuccess("Thanks for subscribing for our Newsletter", sub_res{status: true}, w)

}

func (eh *ExternalHandler) DownloadClient(w http.ResponseWriter, r *http.Request) {
	var url string
	clientType := r.URL.Query().Get("client_type")
	email := r.URL.Query().Get("email")

	if !utils.IsValidEmail(email) {
		utils.GetError(EMAIL_NOT_VALID, http.StatusBadRequest, w)
		return
	}

	windows_url := "https://api.zuri.chat/files/applications/20210922182446_0.7z"
	linux_url := "url not avaliable at the moment"
	mac_url := "url not avaliable at the moment"
	ios_url := "url not avaliable"
	android_url := "url not avaliable at the moment"

	switch clientType {
	case "windows":
		url = windows_url
	case "linux":
		url = linux_url
	case "mac":
		url = mac_url
	case "ios":
		url = ios_url
	case "android":
		url = android_url
	default:
		utils.GetError(CLIENT_NOT_VALID, http.StatusBadRequest, w)
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
		fmt.Printf("Error occured while sending mail: %s", err.Error())
	}
	utils.GetSuccess("Download Link Successfully sent", url, w)

}

func (eh *ExternalHandler) SendMail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var msgr *service.Mail
	switch isCustomMail, _ := strconv.ParseInt(r.FormValue("custom_mail"), 0, 8); isCustomMail {
		case 1:
			//mailbody and content-type * content-type -> text/html and text/plain
			mail := struct {
				Email	 	string	`json:"email" validate:"email,required"`
				Subject	 	string	`json:"subject" validate:"required"`
				ContentType	string	`json:"content_type" validate:"required"`
				MailBody	string	`json:"mail_body" validate:"required"`
			} {}

			if err := utils.ParseJsonFromRequest(r, &mail); err != nil {
				utils.GetError(err, http.StatusUnprocessableEntity, w)
				return
			}
		
			if err := validate.Struct(mail); err != nil {
				utils.GetError(err, http.StatusBadRequest, w)
				return
			}

			match, _ := regexp.MatchString("<(.|\n)*?>", mail.MailBody)
			if !match && mail.ContentType == "text/html"  {
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
				Email	 string	`json:"email" validate:"email,required"`
				Subject	 string	`json:"subject" validate:"required"`
				MailType int	`json:"mail_type" validate:"required"`
				Data     map[string]interface{} `json:"data" validate:"required"`
			} {}
	
			if err := utils.ParseJsonFromRequest(r, &mail); err != nil {
				utils.GetError(err, http.StatusUnprocessableEntity, w)
				return
			}
		
			if err := validate.Struct(mail); err != nil {
				utils.GetError(err, http.StatusBadRequest, w)
				return
			}
			// ensure email is valid
			if !utils.IsValidEmail(strings.ToLower(mail.Email)) {
				utils.GetError(EMAIL_NOT_VALID, http.StatusBadRequest, w)
				return
			}	
		
			if _, ok := service.MailTypes[service.MailType(mail.MailType)]; !ok {
				utils.GetError(errors.New("invalid email type, email template does not exists!"), http.StatusBadRequest, w)
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
		fmt.Printf("Error occured while sending mail: %s", err.Error())
	}
	
	utils.GetSuccess("Mail sent successfully", nil, w)
}
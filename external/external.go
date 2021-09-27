package external

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"zuri.chat/zccore/service"
	"zuri.chat/zccore/utils"
)

var (
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
				"ZuriLogo": utils.ConvertImageTo64("./templates/email_sub/images/zuri_logo.png"),
				"Image2":   utils.ConvertImageTo64("./templates/email_sub/images/people_chatting.png"),
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
			"ZuriLogo": utils.ConvertImageTo64("./templates/email_sub/images/zuri_logo.png"),
			"Image2":   utils.ConvertImageTo64("./templates/email_sub/images/people_chatting.png"),
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

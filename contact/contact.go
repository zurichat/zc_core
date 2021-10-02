package contact

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"time"

	"zuri.chat/zccore/auth"
	"zuri.chat/zccore/service"
	"zuri.chat/zccore/utils"
)

func ContactUs(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")

	fmt.Println("Parsing Form Data")

	err := r.ParseMultipartForm(MAX_FILE_SIZE * 6)
	if err != nil {
		utils.GetError(errors.New("error parsing form data"), http.StatusBadRequest, w)
		return
	}

	// extract user email from session data or form field in request context
	userDetails, ok := r.Context().Value(auth.UserDetails).(*auth.ResToken)
	var email string
	if ok && userDetails != nil {
		email = userDetails.Email
	} else {
		email = r.Form.Get("email")
	}

	// extract subject and content data
	subject := r.Form.Get("subject")
	content := r.Form.Get("content")
	// parse attached files and save to file system
	attachments := r.MultipartForm.File["file"]

	validator := NewValidator()
	ValidateEmail(*validator, email)
	ValidateSubject(*validator, subject)
	ValidateContent(*validator, content)
	if len(attachments) > 0 {
		ValidateAttachedFiles(*validator, attachments)
	}

	if !validator.Valid() {
		utils.GetDetailedError("invalid form data", http.StatusBadRequest, validator.Errors, w)
		return
	}

	if len(attachments) > 0 {
		uploadFileInfo, err := SaveFileToFS(folderName, r)
		if err != nil {
			utils.GetError(err, http.StatusInternalServerError, w)
			return
		}
		contactFormData := GenerateContactData(email, subject, content, uploadFileInfo)

		data, err := utils.StructToMap(contactFormData)
		if err != nil {
			utils.GetError(err, http.StatusInternalServerError, w)
			return
		}

		mongoRes, err := utils.CreateMongoDbDoc("contact", data)
		if err != nil {
			utils.GetError(err, http.StatusInternalServerError, w)
			return
		}
		utils.GetSuccess("contact information sent successfully", mongoRes, w)
		return
	}
	var emptyFileInfo []service.MultipleTempResponse
	contactFormData := GenerateContactData(email, subject, content, emptyFileInfo)
	data, err := utils.StructToMap(contactFormData)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}
	mongoRes, err := utils.CreateMongoDbDoc("contact", data)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}
	utils.GetSuccess("contact information sent successfully", utils.M{"id": mongoRes.InsertedID}, w)

}

// GeneratePaths takes in file atachments and generates/returns a path slice
func GeneratePaths(attachments []*multipart.FileHeader) []string {
	var pathSlice []string
	if len(attachments) > 0 {
		for _, attch := range attachments {
			path := folderName + "/" + attch.Filename
			pathSlice = append(pathSlice, path)
		}
		return pathSlice
	}
	return pathSlice
}

// GenerateContactData returns a compact struct of contact form data given each piece
// of data sent by user
func GenerateContactData(email, subject, content string, filesInfo []service.MultipleTempResponse) ContactFormData {
	contactFormData := &ContactFormData{
		Subject:   subject,
		Content:   content,
		Email:     email,
		Files:     filesInfo,
		CreatedAt: time.Now(),
	}

	return *contactFormData
}

// SaveFileToFS saves each form file uploaded to the filesystem
func SaveFileToFS(folderName string, r *http.Request) ([]service.MultipleTempResponse, error) {
	multiTempRes, err := service.MultipleFileUpload(folderName, r)
	if err != nil {
		return nil, err
	}
	return multiTempRes, nil
}

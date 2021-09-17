package contact

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"zuri.chat/zccore/auth"
	"zuri.chat/zccore/utils"
)

func ContactUs(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")
	fmt.Println("Parsing Form Data")

	err := r.ParseMultipartForm(MAX_FILE_SIZE * 6)
	if err != nil {
		utils.GetError(errors.New("error parsing form data"), http.StatusBadRequest, w)
	}

	userDetails, ok := r.Context().Value(auth.UserDetails).(*auth.ResToken)
	var email string
	if ok == true && userDetails != nil {
		email = userDetails.Email
	} else {
		email = r.Form.Get("email")
	}

	subject := r.Form.Get("subject")
	content := r.Form.Get("content")
	attachments := r.MultipartForm.File["attachments"]

	validator := NewValidator()
	ValidateEmail(*validator, email)
	ValidateSubject(*validator, subject)
	ValidateContent(*validator, content)
	ValidateAttachedFiles(*validator, attachments)

	if !validator.Valid() {
		utils.GetDetailedError("invalid form data", http.StatusUnprocessableEntity, validator.Errors, w)
		return
	}

	for _, fileHeader := range attachments {
		err := SaveFileToFS(folderName, fileHeader)
		if err != nil {
			utils.GetError(err, http.StatusInternalServerError, w)
			return
		}
	}

	pathSlice := GeneratePaths(attachments)
	contactFormData := GenerateContactData(email, subject, content, pathSlice)

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
func GenerateContactData(email, subject, content string, paths []string) ContactFormData {
	contactFormData := &ContactFormData{
		Subject:     subject,
		Content:     content,
		Email:       email,
		Attachments: paths,
		CreatedAt:   time.Now(),
	}

	return *contactFormData
}

// SaveFileToFS saves each form file uploaded to the filesystem
func SaveFileToFS(folderName string, fileHeader *multipart.FileHeader) error {
	file, err := fileHeader.Open()
	defer file.Close()
	if err != nil {
		return err
	}

	_, err = os.Stat(folderName)
	if err != nil {
		err = os.Mkdir(folderName, 0755)
		if err != nil {
			return err
		}
	}

	destinationFile, err := os.Create(folderName + "/" + fileHeader.Filename)
	defer destinationFile.Close()
	if err != nil {
		return err
	}

	_, err = io.Copy(destinationFile, file)
	if err != nil {
		return err
	}

	return nil
}

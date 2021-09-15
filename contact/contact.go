package contact

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"zuri.chat/zccore/utils"
)

func ContactUs(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")
	fmt.Println("Parsing Form Data")

	// 1. Parse the form data
	err := r.ParseMultipartForm(2 << 10)
	if err != nil {
		utils.GetError(errors.New("error parsing form data"), http.StatusBadRequest, w)
	}
	// 2. Collect form data
	formData := r.MultipartForm
	// 3. Get attachment file header for each field
	subject := formData.Value["subject"]
	content := formData.Value["content"]
	supportEmail := formData.Value["support"]
	attachments := formData.File["attachments"]
	// 4. Validate each input value

	fmt.Println(subject, content, supportEmail, attachments)
	fmt.Println(attachments[0].Filename)

}

func Contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")
	fmt.Println("Parsing Form Data")

	// 1. Parse multipart form data
	err := r.ParseMultipartForm(MAX_FILE_SIZE * 6)
	if err != nil {
		utils.GetError(errors.New("error parsing form data"), http.StatusBadRequest, w)
	}

	// 2. Collect form values and files
	subject := r.Form.Get("subject")
	content := r.Form.Get("content")
	email := r.Form.Get("email")
	attachments := r.MultipartForm.File["attachments"]

	// 3. Validate form values and files
	// 3.1. Create new validator
	validator := NewValidator()
	// 3.2. Check form values
	ValidateEmail(*validator, email)
	ValidateSubject(*validator, subject)
	ValidateContent(*validator, content)
	ValidateAttachedFiles(*validator, attachments)

	if !validator.Valid() {
		utils.GetDetailedError("invalid form data", http.StatusUnprocessableEntity, validator.Errors, w)
		return
	}

	// 4. Save files to file system
	for _, fileHeader := range attachments {
		// Open form file
		file, err := fileHeader.Open()
		defer file.Close()
		if err != nil {
			utils.GetError(err, http.StatusInternalServerError, w)
			return
		}

		// Create destination folder and file
		folderName := "zc_contact"
		_, err = os.Stat(folderName)
		if err != nil {
			err = os.Mkdir("zc_contact", 0755)
			if err != nil {
				utils.GetError(err, http.StatusInternalServerError, w)
				return
			}
		}
		destinationFile, err := os.Create(folderName + "/" + fileHeader.Filename)
		defer destinationFile.Close()
		if err != nil {
			utils.GetError(err, http.StatusInternalServerError, w)
			return
		}

		// Copy form file to destination
		_, err = io.Copy(destinationFile, file)
		if err != nil {
			utils.GetError(err, http.StatusInternalServerError, w)
			return
		}
	}
}

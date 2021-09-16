package organizations

import (
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

var allowedMimeTypes = []string{"application/pdf",
	"image/png",
	"image/jpg",
	"text/plain",
	"image/jpeg",
	"application/msword",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document"}

// UploadFile uploads a file to the server
func contains(v string, a []string) bool {
	for _, i := range a {
		if i == v {
			return true
		}
	}
	return false
}

func UploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	file, handle, err := r.FormFile("file")
	if err != nil {
		fmt.Fprintf(w, "%v", err.Error())
		return
	}
	defer file.Close()

	mimeType := handle.Header.Get("Content-Type")
	fmt.Println(mimeType)
	switch {
	case contains(mimeType, allowedMimeTypes):
		saveFile(w, file, handle)
	default:
		jsonResponse(w, http.StatusBadRequest, "File type "+mimeType+" not allowed")
	}
}

func saveFile(w http.ResponseWriter, file multipart.File, handle *multipart.FileHeader) {
	data, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Fprintf(w, "%v", err.Error())
		return
	}

	err = ioutil.WriteFile("./files/"+handle.Filename, data, 0666)
	if err != nil {
		fmt.Fprintf(w, "%v", err.Error())
		return
	}
	jsonResponse(w, http.StatusCreated, "Archivo guardado exitosamente")
}

func jsonResponse(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, message)
}

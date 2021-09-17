package service

import (
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"zuri.chat/zccore/plugin"
	"zuri.chat/zccore/utils"
)

var allowedMimeTypes = []string{"application/pdf",
	"image/png", "image/jpg", "text/plain", "image/jpeg",
	"video/mp4", "video/mpeg", "video/ogg", "video/quicktime",
	"application/msword", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"}

type OneTempResponse struct {
	FileUrl string `json:"file_url"`
	Status  bool   `json:"status"`
}
type MultTempResponse struct {
	FilesInfo []MultipleTempResponse `json:"files_info"`
	Status    bool                   `json:"status"`
}

type MultipleTempResponse struct {
	OriginalName string `json:"original_name"`
	FileUrl      string `json:"file_url"`
}

// // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // //
// // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // //
// service file for uploading a single file to the files folder, please always specify a folder to saved your files in

func SingleFileUpload(folderName string, r *http.Request) (string, error) {
	var fileUrl string
	file, handle, err := r.FormFile("file")
	if err != nil {
		return "", err
	}
	defer file.Close()

	mimeType := handle.Header.Get("Content-Type")
	switch {
	case contains(mimeType, allowedMimeTypes):
		path, err := saveFile(folderName, file, handle, r)
		if err != nil {
			return "", err
		}
		fileUrl = path
	default:
		return "", fmt.Errorf("File type not Allow")
	}

	return fileUrl, nil
}

// // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // //
// // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // //

// // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // //
// // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // //
// Service function for uploading multiple files to the files folder, please always provide folder to save your files to

func MultipleFileUpload(folderName string, r *http.Request) ([]MultipleTempResponse, error) {
	if r.Method != "POST" {
		return nil, fmt.Errorf("method not allowed")
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return nil, err
	}

	files := r.MultipartForm.File["file"]
	var res []MultipleTempResponse
	for _, fileHeader := range files {

		file, err := fileHeader.Open()
		if err != nil {
			return nil, err
		}

		defer file.Close()

		buff := make([]byte, 512)
		_, err = file.Read(buff)
		if err != nil {
			return nil, err
		}

		filetype := http.DetectContentType(buff)
		if !contains(filetype, allowedMimeTypes) {
			return nil, fmt.Errorf("File type not allowed")
		}

		_, errr := file.Seek(0, io.SeekStart)
		if errr != nil {
			return nil, errr
		}

		if spl := strings.ReplaceAll(folderName, " ", ""); spl != "" {
			folderName = folderName + "/"
		} else {
			folderName = "mesc/"
			// folderName = ""
		}
		fileExtension := filepath.Ext(fileHeader.Filename)

		exeDir, newF := "./files/"+folderName, ""
		filenamePrefix := filepath.Join(exeDir, newF, buildFileName())
		filename, errr := pickFileName(filenamePrefix, fileExtension)
		if errr != nil {
			return nil, errr
		}

		err0 := os.MkdirAll(exeDir, os.ModePerm)
		if err != nil {
			return nil, err0
		}

		f, erri := os.Create(filename)
		if erri != nil {
			return nil, erri
		}

		defer f.Close()

		_, err = io.Copy(f, file)
		if err != nil {
			return nil, err
		}
		filename_e := strings.Join(strings.Split(filename, "\\"), "/")
		fileUrl := r.Host + "/" + filename_e
		lores := MultipleTempResponse{
			OriginalName: fileHeader.Filename,
			FileUrl:      fileUrl,
		}
		res = append(res, lores)
	}

	return res, nil
}

// // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // //
// // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // //
// Plugin file upload route functions

func UploadOneFile(w http.ResponseWriter, r *http.Request) {
	plugin_id := mux.Vars(r)["plugin_id"]
	_, err := plugin.FindPluginByID(r.Context(), plugin_id)
	if err != nil {
		utils.GetError(fmt.Errorf("Plugin does not exist"), http.StatusNotFound, w)
		return
	}
	url, err := SingleFileUpload(plugin_id, r)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}
	res := OneTempResponse{
		FileUrl: url,
		Status:  true,
	}
	utils.GetSuccess("File Upload Successful", res, w)

}

func UploadMultipleFiles(w http.ResponseWriter, r *http.Request) {
	plugin_id := mux.Vars(r)["plugin_id"]
	_, err := plugin.FindPluginByID(r.Context(), plugin_id)
	if err != nil {
		utils.GetError(fmt.Errorf("Plugin does not exist"), http.StatusNotFound, w)
		return
	}
	list, err := MultipleFileUpload(plugin_id, r)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}
	res := MultTempResponse{
		FilesInfo: list,
		Status:    true,
	}
	utils.GetSuccess("Files Uploaded Successfully", res, w)

}

// // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // //
// // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // //

// Functions below here are some inpackage functions used in the functions above

func saveFile(folderName string, file multipart.File, handle *multipart.FileHeader, r *http.Request) (string, error) {
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}
	if spl := strings.ReplaceAll(folderName, " ", ""); spl != "" {
		folderName = folderName + "/"
	} else {
		folderName = "mesc/"
		// folderName = ""
	}
	fileExtension := filepath.Ext(handle.Filename)

	exeDir, newF := "./files/"+folderName, ""
	filenamePrefix := filepath.Join(exeDir, newF, buildFileName())
	filename, errr := pickFileName(filenamePrefix, fileExtension)
	if errr != nil {
		return "", errr
	}

	// Create the uploads folder if it doesn't
	// already exist
	err0 := os.MkdirAll(exeDir, os.ModePerm)
	if err != nil {
		return "", err0
	}

	err = ioutil.WriteFile(filename, data, 0666)
	if err != nil {
		return "", err
	}
	filename_e := strings.Join(strings.Split(filename, "\\"), "/")
	fileUrl := r.Host + "/" + filename_e
	return fileUrl, nil
}

func contains(v string, a []string) bool {
	for _, i := range a {
		if i == v {
			return true
		}
	}
	return false
}

func buildFileName() string {
	return time.Now().Format("20060102150405")
}

func pickFileName(prefix string, suffix string) (string, error) {
	for i := 0; i < 100; i++ {
		fname := fmt.Sprintf("%s_%d%s", prefix, i, suffix)
		if _, err := os.Stat(fname); os.IsNotExist(err) {
			return fname, nil
		}

	}
	return "", fmt.Errorf("Unable to create a unique file with the prefix %v in 100 tries", prefix)
}

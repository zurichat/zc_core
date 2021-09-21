package service

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"

	// uuser "os/user"
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
	"application/msword", "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	"application/vnd.android.package-archive", "application/octet-stream",
	"application/x-rar-compressed"," application/octet-stream"," application/zip", "application/octet-stream", "application/x-zip-compressed", "multipart/x-zip",
	}

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
type DeleteFileRequest struct {
	FileUrl string `json:"file_url" validate:"required"`
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
		// cwd, _ := os.Getwd()
		exeDir, newF := "files/"+folderName, ""
		// mexeDir := filepath.Join(cwd,exeDir)
		filenamePrefix := filepath.Join(exeDir, newF, buildFileName())
		filename, errr := pickFileName(filenamePrefix, fileExtension)
		// wfilename := filepath.Join(cwd,filename)

		if errr != nil {
			return nil, errr
		}

		_, err2 := os.Stat(exeDir)
		if err2 != nil {
			// err1 := os.Mkdir(exeDir, 0777)
			// if err1 != nil {
			// 	return nil, err1
			// }
			err0 := os.MkdirAll(exeDir, 0777)
			if err0 != nil {
				return nil, err0
			}
		}

		destinationFile, erri := os.Create(filename)
		defer destinationFile.Close()
		if err != nil {
			return nil, erri
		}

		_, err = io.Copy(destinationFile, file)
		if err != nil {
			return nil, err
		}
		filename_e := strings.Join(strings.Split(filename, "\\"), "/")

		var urlPrefix string = "https://api.zuri.chat/"
		if r.Host == "127.0.0.1:8080" {
			urlPrefix = "127.0.0.1:8080/"
		}
		fileUrl := urlPrefix + filename_e
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
// Delete file service function

func DeleteFileFromServer(filePath string) error {
	e := os.Remove(filePath)
	if e != nil {
		return e
	}
	return nil
}

// // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // //
// // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // //

// // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // //
// // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // //
// profile image upload
func ProfileImageUpload(folderName string, r *http.Request) (string, error) {
	var allowedMimeImageTypes = []string{
		"image/bmp", "image/cis-cod", "image/gif", "image/ief", "image/jpeg", "image/jpeg", "image/jpeg", "image/pipeg	jfif", "image/svg+xml",
		"image/tiff", "image/tiff", "image/x-cmu-raster", "	image/x-cmx", "image/x-icon", "image/x-portable-anymap", "image/x-portable-bitmap",
		"image/x-portable-graymap", "image/x-portable-pixmap", "image/x-rgb", "image/x-xbitmap", "image/x-xpixmap", "image/x-xwindowdump",
		"image/png"}
	var fileUrl string
	file, handle, err := r.FormFile("image")
	if err != nil {
		return "", err
	}
	defer file.Close()

	mimeType := handle.Header.Get("Content-Type")
	switch {
	case contains(mimeType, allowedMimeImageTypes):
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
// Plugin file route functions

type Terror struct {
	Err error
	Msg string
}

func UploadOneFile(w http.ResponseWriter, r *http.Request) {
	plugin_id := mux.Vars(r)["plugin_id"]
	_, err := plugin.FindPluginByID(r.Context(), plugin_id)
	if err != nil {
		utils.GetError(fmt.Errorf("Acess Denied, Plugin does not exist"), http.StatusForbidden, w)
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
		utils.GetError(fmt.Errorf("Acess Denied, Plugin does not exist"), http.StatusForbidden, w)
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

func DeleteFile(w http.ResponseWriter, r *http.Request) {
	var delFile DeleteFileRequest
	plugin_id := mux.Vars(r)["plugin_id"]
	_, ee := plugin.FindPluginByID(r.Context(), plugin_id)
	if ee != nil {
		utils.GetError(fmt.Errorf("Access Denied, Plugin does not exist"), http.StatusForbidden, w)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&delFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(strings.Split(delFile.FileUrl, plugin_id)) == 1 {
		utils.GetError(fmt.Errorf("Delete Not Allowed for plugin of Id: "+plugin_id), http.StatusForbidden, w)
		return
	}
	var urldom string = "api.zuri.chat"
	if r.Host == "127.0.0.1:8080" {
		urldom = "127.0.0.1:8080"
	}
	filePath := "." + strings.Split(delFile.FileUrl, urldom)[1]
	cwd, _ := os.Getwd()
	filePath = filepath.Join(cwd, filePath)
	er := DeleteFileFromServer(filePath)
	if er != nil {
		utils.GetError(er, http.StatusBadRequest, w)
		return
	}
	utils.GetSuccess("Deleted Successfully", "", w)
}

// // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // //
// // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // //

// Functions below here are some inpackage functions used in the functions above

func saveFile(folderName string, file multipart.File, handle *multipart.FileHeader, r *http.Request) (string, error) {
	// cwd, _ := os.Getwd()
	// usdr,_ := uuser.Current()
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

	exeDir, newF := "files/"+folderName, ""
	// mexeDir := filepath.Join(cwd,exeDir)
	filenamePrefix := filepath.Join(exeDir, newF, buildFileName())
	filename, errr := pickFileName(filenamePrefix, fileExtension)
	// wfilename := filepath.Join(cwd,filename)
	if errr != nil {
		return "", errr
	}

	_, err2 := os.Stat(exeDir)
	if err2 != nil {
		// err1 := os.Mkdir(exeDir, os.ModePerm)
		// if err1 != nil {
		// 	return "", err1
		// }
		err0 := os.MkdirAll(exeDir, os.ModePerm)
		if err0 != nil {
			return "", err0
		}
	}

	destinationFile, erri := os.Create(filename)
	defer destinationFile.Close()
	if err != nil {
		return "", erri
	}

	err = ioutil.WriteFile(filename, data, 0777)
	if err != nil {
		return "", err
	}

	filename_e := strings.Join(strings.Split(filename, "\\"), "/")
	var urlPrefix string = "https://api.zuri.chat/"
	if r.Host == "127.0.0.1:8080" {
		urlPrefix = "127.0.0.1:8080/"
	}
	fileUrl := urlPrefix + filename_e
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

func MescFiles(w http.ResponseWriter, r *http.Request) {
	masc, mesc, apk_sec, exe_sec := utils.Env("APK_SEC"), utils.Env("EXE_SEC"), mux.Vars(r)["apk_sec"], mux.Vars(r)["exe_sec"]
	if !(masc == apk_sec && mesc == exe_sec){
		utils.GetError(fmt.Errorf("Acess Denied"), http.StatusForbidden, w)
		return
	}
	uploadPath := "applications"
	url, err := mescf(uploadPath, r)
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
func mescf(folderName string, r *http.Request) (string, error) {
	var fileUrl string
	file, handle, err := r.FormFile("app")
	if err != nil {
		return "", err
	}
	defer file.Close()

	path, err := saveFile(folderName, file, handle, r)
	if err != nil {
		return "", err
	}
	fileUrl = path

	return fileUrl, nil
}

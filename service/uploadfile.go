package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"

	"path/filepath"
	"strings"
	"time"

	"github.com/anthonynsimon/bild/imgio"
	"github.com/anthonynsimon/bild/transform"
	"github.com/gorilla/mux"
	"zuri.chat/zccore/plugin"
	"zuri.chat/zccore/utils"
)

var allowedMimeTypes = []string{"application/pdf",
	"image/png", "image/gif", "image/jpg", "text/plain", "image/jpeg",
	"video/mp4", "video/mpeg", "video/ogg", "video/quicktime",
	"application/msword", "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	"application/vnd.android.package-archive", "application/octet-stream",
	"application/x-rar-compressed", " application/octet-stream", " application/zip", "application/octet-stream", "application/x-zip-compressed", "multipart/x-zip",
}

const (
	localDH  = "127.0.0.1:8080"
	localDHL = "localhost:8080"
	mescPath = "mesc/"
	hostPath = "127.0.0.1:8080/"
)

var (
	permissionNumber fs.FileMode = 0777
	mg32             int64       = 32
	mg20             int64       = 20
	mg512                        = 512
)

type OneTempResponse struct {
	FileURL string `json:"file_url"`
	Status  bool   `json:"status"`
}
type MultTempResponse struct {
	FilesInfo []MultipleTempResponse `json:"files_info"`
	Status    bool                   `json:"status"`
}

type MultipleTempResponse struct {
	OriginalName string `json:"original_name"`
	FileURL      string `json:"file_url"`
}
type DeleteFileRequest struct {
	FileURL string `json:"file_url" validate:"required"`
}

// // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // //
// // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // //
// service file for uploading a single file to the files folder, please always specify a folder to saved your files in

func SingleFileUpload(folderName string, r *http.Request) (string, error) {
	var fileURL string

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

		fileURL = path
	default:
		return "", fmt.Errorf("file type not allow")
	}

	return fileURL, nil
}

// // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // //
// Service function for uploading multiple files to the files folder, please always provide folder to save your files to

func MultipleFileUpload(folderName string, r *http.Request) ([]MultipleTempResponse, error) {
	if r.Method != "POST" {
		return nil, fmt.Errorf("method not allowed")
	}

	if err := r.ParseMultipartForm(mg32 << mg20); err != nil {
		return nil, err
	}

	res := []MultipleTempResponse{}
	files := r.MultipartForm.File["file"]
	if files == nil {
		return nil, errors.New("empty file upload")
	}

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			return nil, err
		}

		defer file.Close()

		buff := make([]byte, mg512)

		_, err = file.Read(buff)
		if err != nil {
			return nil, err
		}

		filetype := http.DetectContentType(buff)
		if !contains(filetype, allowedMimeTypes) {
			return nil, fmt.Errorf("File type not allowed")
		}

		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			return nil, err
		}

		if spl := strings.ReplaceAll(folderName, " ", ""); spl != "" {
			folderName += "/"
		} else {
			folderName = mescPath
		}

		fileExtension := filepath.Ext(fileHeader.Filename)
		exeDir, newF := "files/"+folderName, ""
		filenamePrefix := filepath.Join(exeDir, newF, buildFileName())
		filename, err := pickFileName(filenamePrefix, fileExtension)

		if err != nil {
			return nil, err
		}

		_, err = os.Stat(exeDir)
		if err != nil {
			err0 := os.MkdirAll(exeDir, permissionNumber)
			if err0 != nil {
				return nil, err0
			}
		}

		destinationFile, err := os.Create(filename)

		if err != nil {
			return nil, err
		}

		defer destinationFile.Close()

		_, err = io.Copy(destinationFile, file)
		if err != nil {
			return nil, err
		}

		filenameE := strings.Join(strings.Split(filename, "\\"), "/")

		var urlPrefix = "https://api.zuri.chat/"
		if r.Host == localDH || r.Host == localDHL {
			urlPrefix = hostPath
		}

		fileURL := urlPrefix + filenameE

		lores := MultipleTempResponse{
			OriginalName: fileHeader.Filename,
			FileURL:      fileURL,
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

func Resize(f multipart.File, width, height int, r *http.Request, handle *multipart.FileHeader, folderName string) (string, error) {
	img, err := decodeFile(f, handle)
	if err != nil {
		return "", err
	}

	resized := transform.Resize(img, width, height, transform.Linear)

	imagePath, err := saveImageFile(folderName, resized, handle, r, imgio.PNGEncoder())
	if err != nil {
		return "", err
	}

	return imagePath, nil
}

func decodeFile(f multipart.File, handle *multipart.FileHeader) (image.Image, error) {
	fileExtension := strings.ToLower(filepath.Ext(handle.Filename))

	if fileExtension == ".gif" {
		_, err := gif.DecodeAll(f)
		if err != nil {
			fmt.Println(err)
			return nil, fmt.Errorf("error decoding")
		}

		return nil, nil
	} else if fileExtension == ".jpg" || fileExtension == ".jpeg" {
		img, err := jpeg.Decode(f)
		if err != nil {
			fmt.Println(err)
			return nil, fmt.Errorf("error decoding")
		}

		return img, nil
	}

	img, err := png.Decode(f)
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("error decoding")
	}

	return img, nil
}

func saveImageFile(folderName string, file *image.RGBA, handle *multipart.FileHeader, r *http.Request, encoder imgio.Encoder) (string, error) {
	if spl := strings.ReplaceAll(folderName, " ", ""); spl != "" {
		folderName += "/"
	} else {
		folderName = mescPath
	}

	fileExtension := filepath.Ext(handle.Filename)

	exeDir, newF := "files/"+folderName, ""
	filenamePrefix := filepath.Join(exeDir, newF, buildFileName())

	filename, errr := pickFileName(filenamePrefix, fileExtension)
	if errr != nil {
		return "", errr
	}

	_, err2 := os.Stat(exeDir)
	if err2 != nil {
		err0 := os.MkdirAll(exeDir, os.ModePerm)
		if err0 != nil {
			return "", err0
		}
	}

	filenameE := strings.Join(strings.Split(filename, "\\"), "/")
	if err := imgio.Save(filenameE, file, encoder); err != nil {
		return "", err
	}

	var urlPrefix = "https://api.zuri.chat/"

	if r.Host == localDH || r.Host == localDHL {
		urlPrefix = hostPath
	}

	fileURL := urlPrefix + filenameE

	return fileURL, nil
}

// // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // //
// // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // //
// profile image upload.
func ProfileImageUpload(folderName string, width, height int, r *http.Request) (string, error) {
	var allowedMimeImageTypes = []string{
		"image/bmp", "image/cis-cod", "image/gif", "image/ief", "image/jpeg", "image/jpeg", "image/jpeg", "image/pipeg	jfif", "image/svg+xml",
		"image/tiff", "image/tiff", "image/x-cmu-raster", "	image/x-cmx", "image/x-icon", "image/x-portable-anymap", "image/x-portable-bitmap",
		"image/x-portable-graymap", "image/x-portable-pixmap", "image/x-rgb", "image/x-xbitmap", "image/x-xpixmap", "image/x-xwindowdump",
		"image/png"}

	var fileURL string

	file, handle, err := r.FormFile("image")
	if err != nil {
		return "", err
	}
	defer file.Close()

	mimeType := handle.Header.Get("Content-Type")

	switch {
	case contains(mimeType, allowedMimeImageTypes):
		path, err := Resize(file, width, height, r, handle, folderName)
		if err != nil {
			return "", err
		}

		fileURL = path
	default:
		return "", fmt.Errorf("File type not Allow")
	}

	return fileURL, nil
}

// // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // //
// // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // //
// Plugin file route functions

type Terror struct {
	Err error
	Msg string
}

func UploadOneFile(w http.ResponseWriter, r *http.Request) {
	pluginID := mux.Vars(r)["plugin_id"]

	_, err := plugin.FindPluginByID(r.Context(), pluginID)
	if err != nil {
		utils.GetError(fmt.Errorf("acess Denied, Plugin does not exist"), http.StatusForbidden, w)
		return
	}

	url, err := SingleFileUpload(pluginID, r)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	res := OneTempResponse{
		FileURL: url,
		Status:  true,
	}
	utils.GetSuccess("File Upload Successful", res, w)
}

func UploadMultipleFiles(w http.ResponseWriter, r *http.Request) {
	pluginID := mux.Vars(r)["plugin_id"]

	_, err := plugin.FindPluginByID(r.Context(), pluginID)
	if err != nil {
		utils.GetError(fmt.Errorf("acess Denied, Plugin does not exist"), http.StatusForbidden, w)
		return
	}

	list, err := MultipleFileUpload(pluginID, r)
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

	pluginID := mux.Vars(r)["plugin_id"]

	_, ee := plugin.FindPluginByID(r.Context(), pluginID)
	if ee != nil {
		utils.GetError(fmt.Errorf("access Denied, Plugin does not exist"), http.StatusForbidden, w)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&delFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(strings.Split(delFile.FileURL, pluginID)) == 1 {
		utils.GetError(fmt.Errorf("Delete Not Allowed for plugin of Id: "+pluginID), http.StatusForbidden, w)
		return
	}

	var urldom = "api.zuri.chat"

	if r.Host == localDH || r.Host == localDHL {
		urldom = localDH
	}

	filePath := "." + strings.Split(delFile.FileURL, urldom)[1]
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
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	if spl := strings.ReplaceAll(folderName, " ", ""); spl != "" {
		folderName += "/"
	} else {
		folderName = mescPath
	}

	fileExtension := filepath.Ext(handle.Filename)

	exeDir, newF := "files/"+folderName, ""
	filenamePrefix := filepath.Join(exeDir, newF, buildFileName())

	filename, errr := pickFileName(filenamePrefix, fileExtension)
	if errr != nil {
		return "", errr
	}

	_, err2 := os.Stat(exeDir)
	if err2 != nil {
		err0 := os.MkdirAll(exeDir, os.ModePerm)
		if err0 != nil {
			return "", err0
		}
	}

	destinationFile, erri := os.Create(filename)

	if err != nil {
		return "", erri
	}

	defer destinationFile.Close()

	err = ioutil.WriteFile(filename, data, permissionNumber)
	if err != nil {
		return "", err
	}

	filenameE := strings.Join(strings.Split(filename, "\\"), "/")

	var urlPrefix = "https://api.zuri.chat/"

	if r.Host == localDH || r.Host == localDHL {
		urlPrefix = hostPath
	}

	fileURL := urlPrefix + filenameE

	return fileURL, nil
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

func pickFileName(prefix, suffix string) (string, error) {
	for i := 0; i < 100; i++ {
		fname := fmt.Sprintf("%s_%d%s", prefix, i, suffix)
		if _, err := os.Stat(fname); os.IsNotExist(err) {
			return fname, nil
		}
	}

	return "", fmt.Errorf("unable to create a unique file with the prefix %v in 100 tries", prefix)
}

func MescFiles(w http.ResponseWriter, r *http.Request) {
	masc, mesc, apkSec, exeSec := utils.Env("APK_SEC"), utils.Env("EXE_SEC"), mux.Vars(r)["apk_sec"], mux.Vars(r)["exe_sec"]
	if !(masc == apkSec && mesc == exeSec) {
		utils.GetError(fmt.Errorf("acess Denied"), http.StatusForbidden, w)
		return
	}

	uploadPath := "applications"

	url, err := mescf(uploadPath, r)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	res := OneTempResponse{
		FileURL: url,
		Status:  true,
	}

	utils.GetSuccess("File Upload Successful", res, w)
}
func mescf(folderName string, r *http.Request) (string, error) {
	var fileURL string

	file, handle, err := r.FormFile("app")
	if err != nil {
		return "", err
	}
	defer file.Close()

	path, err := saveFile(folderName, file, handle, r)
	if err != nil {
		return "", err
	}

	fileURL = path

	return fileURL, nil
}

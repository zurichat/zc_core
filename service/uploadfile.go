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
type DeleteFileRequest struct {
	FileUrl string `json:"file_url" validate:"required"`
}

// // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // //
// // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // //
// service file for uploading a single file to the files folder, please always specify a folder to saved your files in

func SingleFileUpload(folderName string, r *http.Request) (string, error, string) {
	var fileUrl string
	file, handle, err := r.FormFile("file")
	if err != nil {
		return "", err, "Error getting file from form data"
	}
	defer file.Close()

	mimeType := handle.Header.Get("Content-Type")
	switch {
	case contains(mimeType, allowedMimeTypes):
		path, err, msg := saveFile(folderName, file, handle, r)
		if err != nil {
			return "", err, msg
		}
		fileUrl = path
	default:
		return "", fmt.Errorf("File type not Allow"), "file type not allowed"
	}

	return fileUrl, nil, "success"
}

// // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // //
// // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // //

// // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // //
// // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // // //
// Service function for uploading multiple files to the files folder, please always provide folder to save your files to

func MultipleFileUpload(folderName string, r *http.Request) ([]MultipleTempResponse, error, string) {
	if r.Method != "POST" {
		return nil, fmt.Errorf("method not allowed"), "Request Method is not allowed"
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return nil, err, "Error Parsing Form"
	}

	files := r.MultipartForm.File["file"]
	var res []MultipleTempResponse
	for _, fileHeader := range files {

		file, err := fileHeader.Open()
		if err != nil {
			return nil, err, "Error Opening File Header"
		}

		defer file.Close()

		buff := make([]byte, 512)
		_, err = file.Read(buff)
		if err != nil {
			return nil, err, "Error Reading buffer"
		}

		filetype := http.DetectContentType(buff)
		if !contains(filetype, allowedMimeTypes) {
			return nil, fmt.Errorf("File type not allowed"), "File type is not allowed"
		}

		_, errr := file.Seek(0, io.SeekStart)
		if errr != nil {
			return nil, errr, "File Seeking Failed"
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
			return nil, errr, "Eror creating unique file name"
		}

		// err0 := os.MkdirAll(mexeDir, os.ModePerm)
		// if err != nil {
		// 	return nil, err0
		// }
		// _, err2 := os.Stat(exeDir)
 
		// if os.IsNotExist(err2) {
		// 	err1 := os.Mkdir(exeDir, 0777)
		// 	if err != nil {
		// 		return nil, err1, "Creating Dir with Mkdir Failed"
		// 	}
		// 	err0 := os.MkdirAll(exeDir, 0777)
		// 	if err != nil {
		// 		return nil, err0, "Creating Dir with MkdirAll Failed"
		// 	}
	
		// }
		_, err2 := os.Stat(exeDir)
		if err2 != nil {
			err1 := os.Mkdir(exeDir, 0777)
			if err != nil {
				return nil, err1, "Creating Dir with Mkdir Failed"
			}
			err0 := os.MkdirAll(exeDir, 0777)
			if err != nil {
				return nil, err0, "Creating Dir with MkdirAll Failed"
			}
		}

		// f, erri := os.Create(wfilename)
		// if erri != nil {
		// 	return nil, erri
		// }
		f, erri := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0777)
		if erri != nil {
			return nil, erri, "Opening File Failed"
		}

		defer f.Close()

		_, err = io.Copy(f, file)
		if err != nil {
			return nil, err, "copying data to file failed"
		}
		filename_e := strings.Join(strings.Split(filename, "\\"), "/")
		fileUrl := r.Host + "/" + filename_e
		lores := MultipleTempResponse{
			OriginalName: fileHeader.Filename,
			FileUrl:      fileUrl,
		}
		res = append(res, lores)
	}

	return res, nil, "successfully Upload files"
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
	url, err, msg := SingleFileUpload(plugin_id, r)
	if err != nil {
		err3 := Terror{
			Err: err, 
			Msg: msg,
		}
		utils.GetError(fmt.Errorf(fmt.Sprintf("%v", err3)), http.StatusBadRequest, w)
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
	list, err, msg := MultipleFileUpload(plugin_id, r)
	if err != nil {
		err3 := Terror{
			Err: err, 
			Msg: msg,
		}
		utils.GetError(fmt.Errorf(fmt.Sprintf("%v", err3)), http.StatusBadRequest, w)
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
	filePath := "." + strings.Split(delFile.FileUrl, r.Host)[1]
	cwd, _ := os.Getwd()
	filePath = filepath.Join(cwd,filePath)
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

func saveFile(folderName string, file multipart.File, handle *multipart.FileHeader, r *http.Request) (string, error, string) {
	// cwd, _ := os.Getwd()
	// usdr,_ := uuser.Current()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err, "error reading files"
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
		return "", errr, "error creating unique file name"
	}

	// _, err2 := os.Stat(exeDir)
 
	// if os.IsNotExist(err2) {
	// 	err1 := os.Mkdir(exeDir, 0777)
	// 	if err != nil {
	// 		return "", err1, "Creating Dir with Mkdir Failed"
	// 	}
	// 	err0 := os.MkdirAll(exeDir, 0777)
	// 	if err != nil {
	// 		return "", err0, "Creating Dir with MkdirAll Failed"
	// 	}
 
	// }

	_, err2 := os.Stat(exeDir)
	if err2 != nil {
		err1 := os.Mkdir(exeDir, 0777)
		if err != nil {
			return "", err1, "Creating Dir with Mkdir Failed"
		}
		err0 := os.MkdirAll(exeDir, 0777)
		if err != nil {
			return "", err0, "Creating Dir with MkdirAll Failed"
		}
	}

	_, erri := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0777)
		if erri != nil {
			return "", erri, "Creating file with openfile Failed"
	}

	err = ioutil.WriteFile(filename, data, 0777)
	if err != nil {
		return "", err, "Writing to file Failed"
	}
	filename_e := strings.Join(strings.Split(filename, "\\"), "/")
	fileUrl := r.Host + "/" + filename_e
	return fileUrl, nil, "Success"
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

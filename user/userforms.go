package user

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"zuri.chat/zccore/utils"
)

var tpl *template.Template

func init() {
	tpl = template.Must(template.ParseGlob("templates/*.gohtml"))
}

func UserForm(response http.ResponseWriter, request *http.Request) {
	tpl.ExecuteTemplate(response, "index.gohtml", nil)
}

func Processor(response http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		http.Redirect(response, request, "/usersform", http.StatusSeeOther)
		return
	}

	response.Header().Add("content-type", "application/x-www-form-urlencoded")

	first_name := request.FormValue("first_name")
	last_name := request.FormValue("last_name")
	password := request.FormValue("password")
	phone := request.FormValue("phone")
	display_name := request.FormValue("display_name")
	email := request.FormValue("email")

	user_collection := "users"

	var user User

	user.Email = email
	
	if !utils.IsValidEmail(user.Email) {
		utils.GetError(errors.New("email address is not valid"), http.StatusBadRequest, response)
		return
	}

	// confirm if user_email exists
	result, _ := utils.GetMongoDbDoc(user_collection, bson.M{"email": user.Email})
	if result != nil {
		fmt.Printf("users with email %s exists!", user.Email)
		utils.GetError(errors.New("operation failed"), http.StatusBadRequest, response)
		return
	}

	user.CreatedAt = time.Now()
	user.Password = password
	user.Phone = phone
	user.DisplayName = display_name
	user.Email = email
	user.FirstName = first_name
	user.LastName = last_name
	
	detail, _ := utils.StructToMap(user)

	res, err := utils.CreateMongoDbDoc(user_collection, detail)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, response)
		return
	}

	utils.GetSuccess("user created", res, response)
}

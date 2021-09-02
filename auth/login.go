package auth

import (
	"errors"
	"strings"

	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"

	"zuri.chat/zccore/utils"

)

var SECRET_KEY = []byte("gosecretkey")

type loginResponse struct {
	Name string `json:"name"`
	Email string `json:"email"`
    Token string `json:"token"`
}


func GenerateJWT()(string,error){
	token := jwt.New(jwt.SigningMethodHS256)
	tokenString, err :=  token.SignedString(SECRET_KEY)
	if err !=nil {
		fmt.Println("Error in JWT token generation")
		return "",err
	}
	return tokenString, nil
}

func UserLogin(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type","application/json")

	loginData := make(map[string]interface{})
 

	if err := utils.ParseJsonFromRequest(r, &loginData); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

    
    user_filter := make(map[string]interface{})
	user_filter["email"] = loginData["email"]

	// Check if email exists
	userModel, _ := utils.GetMongoDbDoc("users", user_filter)
	if userModel == nil {
		utils.GetError(errors.New("invalid email"), http.StatusBadRequest, w)
		return
	}

	// validate required fields
	// add required params into required array, make an empty array to hold error strings
   required, empty := []string{ "email", "password"}, make([]string, 0)


	
	// loop through and check for empty required params
	for _, value := range required {
		if str, ok := loginData[value].(string); ok {
			if strings.TrimSpace(str) == "" {
				empty = append(empty, strings.Join(strings.Split(value, "_"), " "))
			}
		} else {
			empty = append(empty, strings.Join(strings.Split(value, "_"), " "))
		}
	}
	if len(empty) > 0 {
		utils.GetError(errors.New(strings.Join(empty, ", ")+" required"), http.StatusBadRequest, w)
		return
	}

	 passErr := bcrypt.CompareHashAndPassword([]byte(userModel["password"].(string)),[]byte(loginData["password"].(string)))
	 if passErr != nil {
		fmt.Println(passErr)
		w.Write([]byte(`{"response":"Wrong Password!"}`))
		return
	 }else{
		 fmt.Println("Passwords match")
	 }
	 	


	jwtToken, err := GenerateJWT()
	if err != nil{
	  w.WriteHeader(http.StatusInternalServerError)
	  w.Write([]byte(`{"message":"`+err.Error()+`"}`))
	  return
	}
	w.Write([]byte(`{"token":"`+jwtToken+`"}`))


   // Instantiate login response struct 
	utils.GetSuccess("User Logged in successfully!", loginResponse{
		Name: userModel["first_name"].(string) + " " + userModel["last_name"].(string),
		Email: userModel["email"].(string),
		Token: jwtToken,
	}, w)

  }
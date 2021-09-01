package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
	"zuri.chat/zccore/user"
	"zuri.chat/zccore/utils"
)

var SECRET_KEY = []byte("gosecretkey")

type loginResponse struct {
	UserID string `json:"user_id"`
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
	// loginData := make(map[string]string)
	var user user.User
	db_name, _ :=  os.LookupEnv("DB_NAME")
	err := utils.ParseJsonFromRequest(r, &user);
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}


	collection, err := utils.GetMongoDbCollection(db_name, "users")
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = collection.FindOne(ctx, bson.M{"email":user.Email}).Decode(&user)
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"`+ err.Error() +`"}`))
		return
	}
	userPass := []byte(user.Password)
  
	bsp, err := bcrypt.GenerateFromPassword([]byte(userPass),bcrypt.DefaultCost)
    userPass = bsp
	if err != nil {
		log.Println(err)
	 }
	 passErr := bcrypt.CompareHashAndPassword([]byte(user.Password),[]byte(userPass))
	 if passErr != nil {
		fmt.Println(passErr)
		w.Write([]byte(`{"response":"Wrong Password!"}`))
		return
	 } else {
		fmt.Println("Passwords match")
	 }	


	jwtToken, err := GenerateJWT()
	if err != nil{
	  w.WriteHeader(http.StatusInternalServerError)
	  w.Write([]byte(`{"message":"`+err.Error()+`"}`))
	  return
	}
	w.Write([]byte(`{"token":"`+jwtToken+`"}`))

	// save organization
	// new_db_coll, err := utils.CreateMongoDbDoc("users", bson.M{
	// 	"email":       user.Email,
	// 	"password":    bsp,
	// })
	// if err != nil {
	// 	utils.GetError(err, http.StatusInternalServerError, w)
	// 	return
	// }
	// if err == nil {
	// 	fmt.Println("Database collection created.")
	// }
	utils.GetSuccess("User Logged in successfully!", loginResponse{
		UserID: user.UserID,
		Email: user.Email,
		Token: jwtToken,
	}, w)

  }

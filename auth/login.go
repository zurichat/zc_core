package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/user"
	"zuri.chat/zccore/utils"
)

const (
	enckry             = "piqenpdan9n-94n49-e-9ad-aononoon"
	session_collection = "session_store"
	user_collection    = "users"
)

var (
	validate            = validator.New()
	UserNotFound        = errors.New("User not found, confirm and try again!")
	InvalidCredentials  = errors.New("Invalid login credentials, confirm and try again")
	AccountConfirmError = errors.New("Your account is not verified, kindly check your email for verification code.")
	AccessExpired		= errors.New("error fetching user info, access token expired, kindly login again")
)

func (au *AuthHandler) GetAuthToken(user *user.User, sess *sessions.Session) (*Token, error) {
	retoken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"session_name": Resptoken.SessionName,
		"cookie":       Resptoken.Cookie,
		"options":      Resptoken.Options,
		"id":           Resptoken.Id,
		"email":        user.Email,
	})

	tokenString, err := retoken.SignedString([]byte(au.configs.HmacSampleSecret))
	if err != nil { return nil, err }

	resp := &Token{
		SessionID: sess.ID,
		User: UserResponse{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			Phone:     user.Phone,
			Timezone:  user.Timezone,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Token:     tokenString,
		},
	}
	
	return resp, nil
}

func (au *AuthHandler) LoginIn(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	var creds Credentials
	if err := utils.ParseJsonFromRequest(request, &creds); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, response)
		return
	}
	if err := validate.Struct(creds); err != nil {
		utils.GetError(err, http.StatusBadRequest, response)
		return
	}

	user, err := FetchUserByEmail(bson.M{"email": strings.ToLower(creds.Email)})
	if err != nil {
		utils.GetError(UserNotFound, http.StatusBadRequest, response)
		return
	}
	// check if user is verified
	if user.IsVerified != true {
		utils.GetError(AccountConfirmError, http.StatusBadRequest, response)
		return
	}

	// check password
	check := CheckPassword(creds.Password, user.Password)
	if !check {
		utils.GetError(InvalidCredentials, http.StatusBadRequest, response)
		return
	}
	store := NewMongoStore(utils.GetCollection(session_collection), au.configs.SessionMaxAge, true, []byte(au.configs.SecretKey))
	var session, e = store.Get(request, au.configs.SessionKey)
	if e != nil {
		msg := fmt.Errorf("%s", e.Error())
		utils.GetError(msg, http.StatusBadRequest, response)
		return
	}

	// store session
	session.Values["id"] = user.ID
	session.Values["email"] = user.Email

	if err = sessions.Save(request, response); err != nil {
		fmt.Printf("Error saving session: %s", err)
		return
	}
	resp, err := au.GetAuthToken(user, session)

	if err != nil {
		utils.GetError(err, http.StatusBadRequest, response)
		return
	}

	utils.GetSuccess("login successful", resp, response)
}

func (au *AuthHandler) LogOutUser(w http.ResponseWriter, r *http.Request) {
	store := NewMongoStore(
		utils.GetCollection(session_collection),
		au.configs.SessionMaxAge,
		true,
		[]byte(au.configs.SecretKey),
	)
	var session *sessions.Session
	var err error
	session, err = store.Get(r, au.configs.SessionKey)
	status, _, sessData := GetSessionDataFromToken(r, []byte(au.configs.HmacSampleSecret))

	if err != nil && status == false {
		utils.GetError(NotAuthorized, http.StatusUnauthorized, w)
		return
	}
	var erro error
	if status == true {
		session, erro = NewS(store, sessData.Cookie, sessData.Id, sessData.Email, r, sessData.SessionName, sessData.Gothic)
		// fmt.Println(session)
		if err != nil && erro != nil {
			utils.GetError(NotAuthorized, http.StatusUnauthorized, w)
			return
		}
	}

	if session.IsNew == true {
		utils.GetError(NoAuthToken, http.StatusUnauthorized, w)
		return
	}

	// fmt.Println(session)
	session.Options.MaxAge = -1

	if err = ClearSession(store, w, session); err != nil {
		fmt.Printf("Error saving session: %s", err)
		utils.GetError(fmt.Errorf("logout Failed"), http.StatusUnauthorized, w)
		return
	}
	// if err = sessions.Save(r, w); err != nil {
	// 	fmt.Printf("Error saving session: %s", err)
	// 	utils.GetError(fmt.Errorf("Logout Failed"), http.StatusUnauthorized, w)
	// 	return
	// }

	utils.GetSuccess("logout successful", map[string]interface{}{}, w)
}

func (au *AuthHandler) VerifyTokenHandler(response http.ResponseWriter, request *http.Request) {
	// extract user id and email from context
	loggedIn := request.Context().Value("user").(*AuthUser)
	user, _ := FetchUserByEmail(bson.M{"email": strings.ToLower(loggedIn.Email)})

	resp := &VerifiedTokenResponse{
		true,
		UserResponse{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			Phone:     user.Phone,
			Timezone:  user.Timezone,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}

	utils.GetSuccess("verified", resp, response)
}

func (au *AuthHandler) LogOutOtherSessions(w http.ResponseWriter, r *http.Request) {

	store := NewMongoStore(
		utils.GetCollection(session_collection),
		au.configs.SessionMaxAge,
		true,
		[]byte(au.configs.SecretKey),
	)
	var session *sessions.Session
	var err error

	// Get  current session
	session, err = store.Get(r, au.configs.SessionKey)
	status, _, sessData := GetSessionDataFromToken(r, []byte(au.configs.HmacSampleSecret))

	if err != nil && status == false {
		utils.GetError(NotAuthorized, http.StatusUnauthorized, w)
		return
	}

	// Handles token session
	if status == true {
		email := sessData.Email
		user, err := FetchUserByEmail(bson.M{"email": email})
		if err != nil {
			utils.GetError(err, http.StatusInternalServerError, w)
		}

		// Check that password is correct
		check := CheckPassword(r.FormValue("password"), user.Password)
		if !check {
			utils.GetError(errors.New("invalid password, confirm and try again"), http.StatusBadRequest, w)
			return
		}

		// delete other sessions apart from current one
		DeleteOtherSessions(user.ID, session.ID)

		// Handles cookie sessions
	} else {
		// Get current user ID
		id := session.Values["id"].(string)

		// Get user who has such ID
		user, er := FetchUserByID(id)
		if er != nil {
			utils.GetError(er, http.StatusInternalServerError, w)
		}

		// Check that password is correct
		check := CheckPassword(r.FormValue("password"), user.Password)
		if !check {
			utils.GetError(errors.New("invalid password, confirm and try again"), http.StatusBadRequest, w)
			return
		}

		// delete other sessions apart from current one
		DeleteOtherSessions(id, session.ID)
	}

	utils.GetSuccess("successfully logged out of other sessions", nil, w)
}

func (au *AuthHandler) SocialAuth(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type,access-control-allow-origin, access-control-allow-headers")

	// for response content type
	w.Header().Add("content-type", "application/json")

	// default providers
	providers := map[string]string{
		"google": au.configs.GoogleOAuthV3Url,
		"facebook": au.configs.FacebookOAuthUrl,
	}

	params := mux.Vars(r)
	social := struct {
		Provider    string `json:"provider" validate:"required"`
		AccessToken string `json:"access_token" validate:"required"`
	}{
		Provider:    params["provider"],
		AccessToken: params["access_token"],
	}
	
	if err := validate.Struct(social); err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	providerUrl, ok := providers[strings.ToLower(social.Provider)]
	if !ok {
		errMsg := fmt.Sprintf("Implementation error: %s does not exists!", social.Provider)
		utils.GetError(errors.New(errMsg), http.StatusBadRequest, w)
		return
	}

	url := strings.Replace(providerUrl, ":access_token", social.AccessToken, 1)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		utils.GetError(AccessExpired, http.StatusBadRequest, w)
		return
	}
	defer resp.Body.Close()

	store := NewMongoStore(utils.GetCollection(session_collection), au.configs.SessionMaxAge, true, []byte(au.configs.SecretKey))
	var session, e = store.Get(r, au.configs.SessionKey)
	if e != nil {
		utils.GetError(e, http.StatusBadRequest, w)
		return
	}

	switch p := strings.ToLower(social.Provider); p {
	case "google":
		socialUser := struct {
			ID				string		`json:"sub"`
			Email			string		`json:"email"`
			EmailVerified	string		`json:"email_verified"`
			FirstName		string		`json:"given_name"`
			LastName		string		`json:"family_name"`
			Picture			string		`json:"picture"`
		}{}

		//check if user exists
		json.NewDecoder(resp.Body).Decode(&socialUser)
		filter := bson.M{"$or": []bson.M{
			{"email": socialUser.Email},
			{"social.provider": p, "social.provider_id": socialUser.ID},
		} }

		vser, err := FetchUserByEmail(filter)
		if err != nil {
			// user not found, create one
			social := &user.Social{ID: socialUser.ID, Provider: p}
			b := &user.User{
				FirstName: socialUser.FirstName,
				LastName: socialUser.LastName,
				Email: socialUser.Email,
				Password: "",
				Deactivated: false,
				IsVerified: true,
				Social: social,
				Timezone : "Africa/Lagos", // set default timezone
				Organizations: []string{"614679ee1a5607b13c00bcb7"}, // set default org
				CreatedAt: time.Now(),
			}
			detail, _ := utils.StructToMap(b)
			res, err := utils.CreateMongoDbDoc(user_collection, detail)

			if err != nil {
				utils.GetError(err, http.StatusInternalServerError, w)
				return
			}

			session.Values["id"] = res.InsertedID
			session.Values["email"] = b.Email						
		} else {
			// update record
			social := map[string]interface{}{
				"provider_id": socialUser.ID,
				"provider": p,
			}

			id, _ := primitive.ObjectIDFromHex(vser.ID)
			filter := bson.M{"_id": id}
			update := bson.M{"$set": bson.M{"social": social, "email": strings.ToLower(socialUser.Email) }}
			
			if _, err := utils.GetCollection(user_collection).UpdateOne(context.Background(), filter, update); err != nil {
				utils.GetError(err, http.StatusInternalServerError, w)
				return		
			}

			session.Values["id"] = vser.ID
			session.Values["email"] = vser.Email
		}

		if err = sessions.Save(r, w); err != nil {
			fmt.Printf("Error saving session: %s", err)
			return
		}

		resp, err := au.GetAuthToken(vser, session)
		if err != nil {
			utils.GetError(err, http.StatusBadRequest, w)
			return
		}
	
		utils.GetSuccess("login successful", resp, w)	
		return 	
	case "facebook":
		utils.GetError(errors.New("Facebook: Pending implementation"), http.StatusBadRequest, w)
		return
	default:
		msg := fmt.Sprintf("Implementation error: %s does not exists!", p)
		utils.GetError(errors.New(msg), http.StatusBadRequest, w)
		return
	}
}
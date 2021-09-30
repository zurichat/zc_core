package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"zuri.chat/zccore/service"
	"zuri.chat/zccore/user"
	"zuri.chat/zccore/utils"

	"github.com/mitchellh/mapstructure"
)

type RoleMember struct {
	ID          string             `json:"id" bson:"_id"`
	OrgId       primitive.ObjectID `json:"org_id" bson:"org_id"`
	Files       []string           `json:"files" bson:"files"`
	ImageURL    string             `json:"image_url" bson:"image_url"`
	Name        string             `json:"name" bson:"name"`
	Email       string             `json:"email" bson:"email"`
	DisplayName string             `json:"display_name" bson:"display_name"`
	Bio         string             `json:"bio" bson:"bio"`
	Pronouns    string             `json:"pronouns" bson:"pronouns"`
	Phone       string             `json:"phone" bson:"phone"`
	TimeZone    string             `json:"time_zone" bson:"time_zone"`
	Role        string             `json:"role" bson:"role"`
	JoinedAt    time.Time          `json:"joined_at" bson:"joined_at"`
}

var SESSION_MAX_AGE int = int(time.Now().Unix() + (31536000 * 200))

var (
	NoAuthToken          = errors.New("No Authorization or session expired.")
	TokenExp             = errors.New("Session expired.")
	NotAuthorized        = errors.New("Not Authorized.")
	ConfirmPasswordError = errors.New("The password confirmation does not match")
	UserDetails          = UserKey("userDetails")
	AccessDenied         = errors.New("Access Denied")
)

type Credentials struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type Token struct {
	SessionID string       `json:"session_id"`
	User      UserResponse `json:"user"`
}

type AuthUser struct {
	ID    primitive.ObjectID `json:"id"`
	Email string             `json:"email"`
}

type MyCustomClaims struct {
	Authorized bool `json:"authorized"`
	User       AuthUser
	jwt.StandardClaims
}

type UserResponse struct {
	ID          string    `json:"id,omitempty"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	DisplayName string    `json:"display_name"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	Status      int       `json:"status"`
	Timezone    string    `json:"time_zone"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Token       string    `json:"token"`
}

type VerifiedTokenResponse struct {
	Verified bool         `json:"is_verified"`
	User     UserResponse `json:"user"`
}

type AuthHandler struct {
	configs     *utils.Configurations
	mailService service.MailService
}

type UserKey string

// Method to compare password
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func FetchUserByEmail(filter map[string]interface{}) (*user.User, error) {
	user := &user.User{}
	userCollection, err := utils.GetMongoDbCollection(os.Getenv("DB_NAME"), user_collection)
	if err != nil {
		return user, err
	}
	result := userCollection.FindOne(context.TODO(), filter)
	err = result.Decode(&user)
	return user, err
}

// Checks if a user is authorized to access a particular function, and either returns a 403 error or continues the process
// First is the Organisation's Id
// Second Option is the role necessary for accessing your endpoint, options are "owner" or "admin" or "member" or "guest"
// Third is response writer
// Fourth request reader
func IsAuthorized(orgId string, role string, w http.ResponseWriter, r *http.Request) bool {
	loggedInUser := r.Context().Value("user").(*AuthUser)
	lguser, ee := FetchUserByEmail(bson.M{"email": strings.ToLower(loggedInUser.Email)})
	if ee != nil {
		utils.GetError(errors.New("error fetching logged in User"), http.StatusBadRequest, w)
		return false
	}
	user_id := lguser.ID
	// collections
	_, user_collection, member_collection := "organizations", "users", "members"
	// org_collection

	// fmt.Println(user_id)

	// Getting user's document from db
	var luHexid, _ = primitive.ObjectIDFromHex(user_id)
	userDoc, _ := utils.GetMongoDbDoc(user_collection, bson.M{"_id": luHexid})
	if userDoc == nil {
		utils.GetError(errors.New("user not found"), http.StatusBadRequest, w)
		return false
	}

	var user user.User
	mapstructure.Decode(userDoc, &user)

	if role == "zuri_admin" {
		if user.Role == role {
			return true
		} else {
			utils.GetError(errors.New("access Denied"), http.StatusUnauthorized, w)
			return false
		}

	}

	// Getting member's document from db
	orgMember, _ := utils.GetMongoDbDoc(member_collection, bson.M{"org_id": orgId, "email": user.Email})
	if orgMember == nil {
		utils.GetError(errors.New("access Denied"), http.StatusUnauthorized, w)
		return false
	}

	var memb RoleMember
	mapstructure.Decode(orgMember, &memb)

	// check role's access
	nA := map[string]int{"owner": 4, "admin": 3, "member": 2, "guest": 1}

	if nA[role] > nA[memb.Role] {
		utils.GetError(errors.New("user Not Authorized"), http.StatusUnauthorized, w)
		return false
	}
	return true
}

// NOTE: Example of how to use isAuthorized function
// loggedInUser := r.Context().Value("user").(*auth.AuthUser)
// user, _ := auth.FetchUserByEmail(bson.M{"email": strings.ToLower(loggedInUser.Email)})
// sOrgId := mux.Vars(r)["id"]

// if !auth.IsAuthorized(sOrgId, "admin", w, r) {
// 	return
// }

func (au *AuthHandler) AuthTest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	store := NewMongoStore(utils.GetCollection(session_collection), au.configs.SessionMaxAge, true, []byte(au.configs.SecretKey))
	var session *sessions.Session
	var err error
	session, err = store.Get(r, au.configs.SessionKey)
	status, _, sessData := GetSessionDataFromToken(r, []byte(au.configs.HmacSampleSecret))

	if err != nil && !status {
		utils.GetError(NotAuthorized, http.StatusUnauthorized, w)
		return
	}
	var erro error
	if status {
		session, erro = NewS(store, sessData.Cookie, sessData.Id, sessData.Email, r, sessData.SessionName, sessData.Gothic)
		fmt.Println(session)
		if err != nil && erro != nil {
			utils.GetError(NotAuthorized, http.StatusUnauthorized, w)
			return
		}
	}

	// use is coming in newly, no cookies
	if session.IsNew {
		utils.GetError(NoAuthToken, http.StatusUnauthorized, w)
		return
	}

	objID, err := primitive.ObjectIDFromHex(session.ID)
	if err != nil {
		utils.GetError(ErrorInvalid, http.StatusUnauthorized, w)
		return
	}

	user := &AuthUser{
		ID:    objID,
		Email: session.Values["email"].(string),
	}
	utils.GetSuccess("Retrived", user, w)

}

// This confirm user password before deactivation
func (au *AuthHandler) ConfirmUserPassword(w http.ResponseWriter, r *http.Request) {
	loggedInUser := r.Context().Value("user").(*AuthUser)

	creds := struct {
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
	}{}

	if err := utils.ParseJsonFromRequest(r, &creds); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	if creds.Password != creds.ConfirmPassword {
		utils.GetError(ConfirmPasswordError, http.StatusBadRequest, w)
		return
	}

	user, err := FetchUserByEmail(bson.M{"email": strings.ToLower(loggedInUser.Email)})
	if err != nil {
		utils.GetError(UserNotFound, http.StatusBadRequest, w)
		return
	}
	// check password
	check := CheckPassword(creds.Password, user.Password)
	if !check {
		utils.GetError(errors.New("invalid credentials, confirm and try again"), http.StatusBadRequest, w)
		return
	}

	utils.GetSuccess("Password confirm successful", nil, w)
}

func GetSessionDataFromToken(r *http.Request, hmacSampleSecret []byte) (status bool, err error, data ResToken) {
	reqTokenh := r.Header.Get("Authorization")
	if reqTokenh == "" {
		return false, fmt.Errorf("authorization access failed"), ResToken{}
	}

	splitToken := strings.Split(reqTokenh, "Bearer ")
	if len(splitToken) < 2 {
		return false, fmt.Errorf("authorization access failed"), ResToken{}
	}

	reqToken := splitToken[1]

	token, err := jwt.Parse(reqToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return hmacSampleSecret, nil
	})

	if err != nil {
		return false, fmt.Errorf("failed"), ResToken{}
	}

	var retTokenD ResToken

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

		mapstructure.Decode(claims, &retTokenD)
		retTokenD.SessionName = fmt.Sprintf("%v", claims["session_name"])
		// fmt.Println(retTokenD)
		return true, nil, retTokenD

	} else {
		return false, fmt.Errorf("failed"), ResToken{}
	}
}

// Initiate
func NewAuthHandler(c *utils.Configurations, mail service.MailService) *AuthHandler {
	return &AuthHandler{configs: c, mailService: mail}
}

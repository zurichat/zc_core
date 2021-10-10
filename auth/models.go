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
	OrgID       primitive.ObjectID `json:"org_id" bson:"org_id"`
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

var (
	NumberTwo          = 2
	ErrNoAuthToken     = errors.New("no authorization or session expired")
	ErrTokenExp        = errors.New("session expired")
	ErrNotAuthorized   = errors.New("not authorized")
	ErrConfirmPassword = errors.New("the password confirmation does not match")
	ErrAccessDenied    = errors.New("access Denied")
	UserDetails        = UserKey("userDetails")
)

type Credentials struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type Token struct {
	SessionID string       `json:"session_id"`
	User      UserResponse `json:"user"`
}

//nolint:revive //CODEI8:
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

//nolint:revive //CODEI8:
type AuthHandler struct {
	configs     *utils.Configurations
	mailService service.MailService
}

type UserKey string

// Method to compare password.
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func FetchUserByEmail(filter map[string]interface{}) (*user.User, error) {
	u := &user.User{}
	userCollection, err := utils.GetMongoDBCollection(os.Getenv("DB_NAME"), userCollection)

	if err != nil {
		return u, err
	}

	result := userCollection.FindOne(context.TODO(), filter)
	err = result.Decode(&u)

	return u, err
}

// Checks if a user is authorized to access a particular function, and either returns a 403 error or continues the process
// First is the Organisation's Id
// Second Option is the role necessary for accessing your endpoint, options are "owner" or "admin" or "member" or "guest"
// Third is response writer
// Fourth request reader.
func IsAuthorized(orgID, role string, w http.ResponseWriter, r *http.Request) bool {
	loggedInUser, _ := r.Context().Value("user").(*AuthUser)
	lguser, ee := FetchUserByEmail(bson.M{"email": strings.ToLower(loggedInUser.Email)})

	if ee != nil {
		utils.GetError(errors.New("error fetching logged in User"), http.StatusBadRequest, w)
		return false
	}

	userID := lguser.ID
	luHexid, _ := primitive.ObjectIDFromHex(userID)
	_, userCollection, memberCollection := "organizations", "users", "members"
	userDoc, _ := utils.GetMongoDBDoc(userCollection, bson.M{"_id": luHexid})

	if userDoc == nil {
		utils.GetError(errors.New("user not found"), http.StatusBadRequest, w)
		return false
	}

	var (
		u    user.User
		memb RoleMember
	)

	//nolint:errcheck //CODEI8: please ignore
	mapstructure.Decode(userDoc, &u)

	if role == "zuri_admin" {
		if u.Role == role {
			return true
		}

		utils.GetError(errors.New("access Denied"), http.StatusUnauthorized, w)

		return false
	}

	// Getting member's document from db
	orgMember, _ := utils.GetMongoDBDoc(memberCollection, bson.M{"org_id": orgID, "email": u.Email})
	if orgMember == nil {
		utils.GetError(errors.New("access Denied"), http.StatusUnauthorized, w)
		return false
	}

	//nolint:errcheck //CODEI8:
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

	store := NewMongoStore(utils.GetCollection(sessionCollection), au.configs.SessionMaxAge, true, []byte(au.configs.SecretKey))

	var (
		session *sessions.Session
		err     error
		erro    error
	)

	session, err = store.Get(r, au.configs.SessionKey)
	status, sessData, _ := GetSessionDataFromToken(r, []byte(au.configs.HmacSampleSecret))

	if err != nil && !status {
		utils.GetError(ErrNotAuthorized, http.StatusUnauthorized, w)
		return
	}

	if status {
		session, erro = NewS(store, sessData.Cookie, sessData.ID, sessData.Email, r, sessData.SessionName, sessData.Gothic)
		if err != nil && erro != nil {
			utils.GetError(ErrNotAuthorized, http.StatusUnauthorized, w)
			return
		}
	}

	// use is coming in newly, no cookies
	if session.IsNew {
		utils.GetError(ErrNoAuthToken, http.StatusUnauthorized, w)
		return
	}

	objID, err := primitive.ObjectIDFromHex(session.ID)

	if err != nil {
		utils.GetError(ErrorInvalid, http.StatusUnauthorized, w)
		return
	}

	u := &AuthUser{
		ID:    objID,
		Email: session.Values["email"].(string),
	}
	utils.GetSuccess("Retrieved", u, w)
}

// This confirm user password before deactivation.
func (au *AuthHandler) ConfirmUserPassword(w http.ResponseWriter, r *http.Request) {
	loggedInUser, _ := r.Context().Value("user").(*AuthUser)

	creds := struct {
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
	}{}

	if err := utils.ParseJSONFromRequest(r, &creds); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	if creds.Password != creds.ConfirmPassword {
		utils.GetError(ErrConfirmPassword, http.StatusBadRequest, w)
		return
	}

	u, err := FetchUserByEmail(bson.M{"email": strings.ToLower(loggedInUser.Email)})

	if err != nil {
		utils.GetError(ErrUserNotFound, http.StatusBadRequest, w)
		return
	}
	// check password
	check := CheckPassword(creds.Password, u.Password)
	if !check {
		utils.GetError(errors.New("invalid credentials, confirm and try again"), http.StatusBadRequest, w)
		return
	}

	utils.GetSuccess("Password confirm successful", nil, w)
}

func GetSessionDataFromToken(r *http.Request, hmacSampleSecret []byte) (status bool, data ResToken, err error) {
	reqTokenh := r.Header.Get("Authorization")
	if reqTokenh == "" {
		return false, ResToken{}, fmt.Errorf("authorization access failed")
	}

	splitToken := strings.Split(reqTokenh, "Bearer ")
	if len(splitToken) < NumberTwo {
		return false, ResToken{}, fmt.Errorf("authorization access failed")
	}

	reqToken := splitToken[1]

	token, err := jwt.Parse(reqToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return hmacSampleSecret, nil
	})

	if err != nil {
		return false, ResToken{}, fmt.Errorf("failed")
	}

	var retTokenD ResToken

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		//nolint:errcheck //CODEI8:
		mapstructure.Decode(claims, &retTokenD)
		retTokenD.SessionName = fmt.Sprintf("%v", claims["session_name"])

		return true, retTokenD, nil
	}

	return false, ResToken{}, fmt.Errorf("failed")
}

func NewAuthHandler(c *utils.Configurations, mail service.MailService) *AuthHandler {
	return &AuthHandler{configs: c, mailService: mail}
}

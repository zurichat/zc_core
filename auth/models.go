package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
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

type contextKey int

const authUserKey contextKey = 0

var (
	NoAuthToken   = errors.New("No Authorization or session expired.")
	TokenExp      = errors.New("Session expired.")
	NotAuthorized = errors.New("Not Authorized.")
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
	ID          primitive.ObjectID `json:"id,omitempty"`
	FirstName   string             `json:"first_name"`
	LastName    string             `json:"last_name"`
	DisplayName string             `json:"display_name"`
	Email       string             `json:"email"`
	Phone       string             `json:"phone"`
	Status      int                `json:"status"`
	Timezone    string             `json:"time_zone"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

type VerifiedTokenResponse struct {
	Verified bool         `json:"is_verified"`
	User     UserResponse `json:"user"`
}

// Method to compare password
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func fetchUserByEmail(filter map[string]interface{}) (*user.User, error) {
	user := &user.User{}
	userCollection, err := utils.GetMongoDbCollection(os.Getenv("DB_NAME"), user_collection)
	if err != nil {
		return user, err
	}
	result := userCollection.FindOne(context.TODO(), filter)
	err = result.Decode(&user)
	return user, err
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

// middleware to check if user is authorized
func IsAuthenticated(nextHandler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "application/json")

		store := NewMongoStore(utils.GetCollection(session_collection), 3600, true, []byte(secretKey))
		var session, err = store.Get(r, sessionKey)
		if err != nil {
			utils.GetError(NotAuthorized, http.StatusUnauthorized, w)
			return
		}

		// use is coming in newly, no cookies
		if session.IsNew == true {
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

		ctx := context.WithValue(r.Context(), "user", user)
		nextHandler.ServeHTTP(w, r.WithContext(ctx))
	}
}

// Checks if a user is authorized to access a particular function, and either returns a 403 error or continues the process
// First Option is user id
// second is the Organisation's Id
// third Option is the role necessary for accessing your endpoint, options are "owner" or "admin" or "member" or "guest"
// fourth is response writer
func IsAuthorized(user_id string, orgId string, role string, w http.ResponseWriter) bool {

	// collections
	_, user_collection, member_collection := "organizations", "users", "members"
	// org_collection

	fmt.Println(user_id)

	// Getting user's document from db
	var luHexid, _ = primitive.ObjectIDFromHex(user_id)
	userDoc, _ := utils.GetMongoDbDoc(user_collection, bson.M{"_id": luHexid})
	if userDoc == nil {
		utils.GetError(errors.New("User not found"), http.StatusBadRequest, w)
		return false
	}

	var user user.User
	mapstructure.Decode(userDoc, &user)

	// Getting member's document from db
	orgMember, _ := utils.GetMongoDbDoc(member_collection, bson.M{"org_id": orgId, "email": user.Email})
	if orgMember == nil {
		utils.GetError(errors.New("Access Denied"), http.StatusUnauthorized, w)
		return false
	}

	var memb RoleMember
	mapstructure.Decode(orgMember, &memb)

	// check role's access
	nA := map[string]int{"owner": 4, "admin": 3, "member": 2, "guest": 1}

	if nA[role] > nA[memb.Role] {
		utils.GetError(errors.New("User Not Authorized"), http.StatusUnauthorized, w)
		return false
	}
	return true
}

// NOTE: Example of how to use isAuthorized function
// loggedInUser := r.Context().Value("user").(auth.AuthUser)
// orgId := mux.Vars(r)["id"]

// if !auth.IsAuthorized(loggedInUser.ID.Hex(), orgId, "member", w) {
// 	return
// }

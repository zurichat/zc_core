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
	NoAuthToken   = errors.New("No Authorization header provided.")
	TokenExp      = errors.New("Token expired.")
	NotAuthorized = errors.New("Not Authorized.")
)

type Authentication struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type Token struct {
	TokenString string       `json:"token"`
	User        UserResponse `json:"user"`
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

// Generate token
func GenerateJWT(userID, email string) (string, error) {
	SECRET_KEY, _ := os.LookupEnv("AUTH_SECRET_KEY")
	if SECRET_KEY == "" {
		SECRET_KEY = secretKey
	}

	var signKey = []byte(SECRET_KEY)

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return "", errors.New("Invalid user ObjectID")
	}

	claims := MyCustomClaims{
		true,
		AuthUser{
			ID:    objID,
			Email: email,
		},
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 12).Unix(), // 12 hours
			Issuer:    "api.zuri.chat",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(signKey)

	if err != nil {
		fmt.Errorf("Something went wrong: %s", err.Error())
		return "", err
	}

	return tokenString, nil
}

// middleware to check if user is authorized
func IsAuthenticated(nextHandler http.HandlerFunc) http.HandlerFunc {
	// token format "Authorization": "Bearer token"
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "application/json")

		if r.Header["Authorization"] == nil {
			utils.GetError(NoAuthToken, http.StatusUnauthorized, w)
			return
		}

		SECRET_KEY, _ := os.LookupEnv("AUTH_SECRET_KEY")
		if SECRET_KEY == "" {
			SECRET_KEY = secretKey
		}

		authToken := strings.Split(r.Header["Authorization"][0], " ")[1]

		token, err := jwt.ParseWithClaims(authToken, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		})

		if claims, ok := token.Claims.(*MyCustomClaims); ok && token.Valid {
			// Extract user information and add it to request context.
			ctx := context.WithValue(r.Context(), "user", claims.User)
			nextHandler.ServeHTTP(w, r.WithContext(ctx))
		} else {
			fmt.Print(err)
			utils.GetError(NotAuthorized, http.StatusUnauthorized, w)
			return
		}
	}
}

// Checks if a user is authorized to access a particular function, and either returns a 403 error or continues the process
// First Option is either a token or user id
// In the second option specify if you entered a token or id in the firstoption, options are "token" or "id"
// Third is the Organisation's Id
// Fourth Option is the role necessary for accessing your endpoint, options are "owner" or "admin" or "member" or "guest"
// Fifth is response writer
func IsAuthorized(tokenOrId string, idenType string, orgId string, role string, w http.ResponseWriter) bool {
	var user_id string
	// Get user Id from token if specified
	if idenType == "token" {
		status, uid, err := utils.TokenIsValid(tokenOrId)
		if status == false {
			utils.GetError(err, http.StatusUnauthorized, w)
		}
		user_id = uid

	} else if idenType == "id" {
		user_id = tokenOrId
	} else {
		fmt.Println("Specified incorrect identype in isAthorized function")
		utils.GetError(errors.New("Specified incorrect identype in isAthorized function"), http.StatusInternalServerError, w)
	}

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

// if !auth.IsAuthorized(loggedInUser.ID.Hex(), "id", orgId, "member", w) {
// 	return
// }

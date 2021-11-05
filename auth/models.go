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

var (
	NumberTwo          = 2
	ErrNoAuthToken     = errors.New("no authorization or session expired")
	ErrTokenExp        = errors.New("session expired")
	ErrNotAuthorized   = errors.New("not authorized")
	ErrConfirmPassword = errors.New("the password confirmation does not match")
	ErrAccessDenied    = errors.New("access Denied")
	UserDetails        = UserKey("userDetails")
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
func ComparePassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Deletes other sessions apart from current one.
func DeleteOtherSessions(userID, sessionID string) {
	uid, _ := primitive.ObjectIDFromHex(userID)
	sid, _ := primitive.ObjectIDFromHex(sessionID)
	filter := bson.M{
		"user_id": bson.M{"$eq": uid},
		"_id":     bson.M{"$ne": sid},
	}
	_, err := utils.DeleteManyMongoDBDoc(sessionCollection, filter)

	if err != nil {
		fmt.Printf("%v", err)
	}
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

// Finds User by ID.
func FetchUserByID(id string) (*user.User, error) {
	uid, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": uid}
	u := &user.User{}

	userCollection := utils.GetCollection(userCollection)
	result := userCollection.FindOne(context.TODO(), filter)
	err := result.Decode(&u)

	return u, err
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

func ClearSession(m *MongoStore, w http.ResponseWriter, session *sessions.Session) error {
	if err := m.delete(session); err != nil {
		return err
	}

	m.Token.SetToken(w, session.Name(), "", session.Options)

	Resptoken = ResToken{
		SessionName: session.Name(),
		Cookie:      "",
		Options:     session.Options,
	}

	return nil
}

func NewAuthHandler(c *utils.Configurations, mail service.MailService) *AuthHandler {
	return &AuthHandler{configs: c, mailService: mail}
}

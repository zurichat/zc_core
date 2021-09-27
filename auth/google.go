package auth

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
	"go.mongodb.org/mongo-driver/bson"
	"zuri.chat/zccore/user"
	"zuri.chat/zccore/utils"
)

const (
	CALLBACK_URL = "http://localhost:8080/auth/google/callback"
	ClientId = "166440807442-tsshvt0pub07pnp09q7e2frac0f24ljf.apps.googleusercontent.com"
	ClientSecret = "AW2eHnMh9O5uXHb9kLX6LiHQ"
)

var Store *MongoStore
var defaultStore *MongoStore

var keySet = false

func init() {
	key := []byte(secretKey)
	keySet = len(key) != 0
	Store = Store
	defaultStore = Store
}

func BeginAuthHandler(w http.ResponseWriter, r *http.Request) {
	url, err := GetAuthURL(w, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, err)
		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

var SetState = func(req *http.Request) string {
	state := req.URL.Query().Get("state")
	if len(state) > 0 {
		return state
	}

	return "state"

}

var GetState = func(req *http.Request) string {
	return req.URL.Query().Get("state")
}

func GetAuthURL(res http.ResponseWriter, req *http.Request) (string, error) {

	if !keySet && defaultStore == Store {
		fmt.Println("goth/gothic: no SESSION_SECRET environment variable is set. The default cookie store is not available and any calls will fail. Ignore this warning if you are using a different store.")
	}

	providerName, err := GetProviderName(req)
	if err != nil {
		return "", err
	}

	provider, err := goth.GetProvider(providerName)
	if err != nil {
		return "", err
	}
	sess, err := provider.BeginAuth(SetState(req))
	if err != nil {
		return "", err
	}

	url, err := sess.GetAuthURL()
	if err != nil {
		return "", err
	}
	Store := NewMongoStore(utils.GetCollection(session_collection), SESSION_MAX_AGE, true, []byte(secretKey))
	var session, e = Store.Get(req, sessionKey)
	if e != nil {
		return "", e
	}
	session.Values["gothic"] = sess.Marshal()
	// session.Values["id"] = user.ID
	// session.Values["email"] = user.Email
	// session.Values["id"] = "6145913b6c283f0bb74d3dc5"
	// session.Values["email"] = "gregoflash00@gmail.com"
	if err = SaveSocialSession(req, res, session, Store); err != nil {
		return "", err
	}

	return url, err
}

var CompleteUserAuth = func(res http.ResponseWriter, req *http.Request) (string, goth.User, error) {

	if !keySet && defaultStore == Store {
		fmt.Println("goth/gothic: no SESSION_SECRET environment variable is set. The default cookie store is not available and any calls will fail. Ignore this warning if you are using a different store.")
	}

	providerName, err := GetProviderName(req)
	if err != nil {
		return "", goth.User{}, err
	}

	provider, err := goth.GetProvider(providerName)
	if err != nil {
		return "", goth.User{}, err
	}
	Store := NewMongoStore(utils.GetCollection(session_collection), SESSION_MAX_AGE, true, []byte(secretKey))
	session, _ := Store.Get(req, sessionKey)

	if session.Values["gothic"] == nil {
		return "", goth.User{}, errors.New("could not find a matching session for this request")
	}

	sess, err := provider.UnmarshalSession(session.Values["gothic"].(string))
	if err != nil {
		return "", goth.User{}, err
	}

	_, err = sess.Authorize(provider, req.URL.Query())

	if err != nil {
		return "", goth.User{}, err
	}

	GothUser, _ := provider.FetchUser(sess)
	eer := CreateSocialUser(GothUser.Email)
	if eer != nil {
		return "", goth.User{}, eer
	}
	token, er := GenerateSocialToken(Store, session, GothUser, session.Values["gothic"])
	if er != nil {
		return "", goth.User{}, err
	}
	return token, GothUser, nil
}

func CreateSocialUser(useremail string) error {
	var user user.User
	result, _ := utils.GetMongoDbDoc(user_collection, bson.M{"email": useremail})
	if result != nil {
		return nil
	}

	user.Email = useremail
	user.CreatedAt = time.Now()
	user.Password = ""
	user.Deactivated = false
	user.IsVerified = true
	user.Social = true

	detail, _ := utils.StructToMap(user)

	_, err := utils.CreateMongoDbDoc(user_collection, detail)

	if err != nil {
		return err
	}
	return nil
}

func GenerateSocialToken(m *MongoStore, session *sessions.Session, Gothicuser goth.User, gothic interface{}) (string, error) {
	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, m.Codecs...)
	if err != nil {
		return "", err
	}
	Resptoken := ResToken{
		Id:          session.ID,
		SessionName: session.Name(),
		Cookie:      encoded,
		Options:     session.Options,
	}
	retoken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"session_name": Resptoken.SessionName,
		"cookie":       Resptoken.Cookie,
		"options":      Resptoken.Options,
		"id":           Resptoken.Id,
		"gothic":       gothic,
		"gothicemail":  Gothicuser.Email,
		// "social_session": session,
	})

	tokenString, eert := retoken.SignedString(hmacSampleSecret)
	if eert != nil {
		return "", err
	}
	return tokenString, nil
}

var GetProviderName = getProviderName

func getProviderName(req *http.Request) (string, error) {
	provider := req.URL.Query().Get("provider")
	if provider == "" {
		if p, ok := mux.Vars(req)["provider"]; ok {
			return p, nil
		}
	}
	if provider == "" {
		provider = req.URL.Query().Get(":provider")
	}
	if provider == "" {
		return provider, errors.New("you must select a provider")
	}
	return provider, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func (au *AuthHandler) BeginGoogleAuth(w http.ResponseWriter, r *http.Request) {
	goth.UseProviders(
		google.New(ClientId, ClientSecret, CALLBACK_URL, "email", "profile"),
	)
	BeginAuthHandler(w, r)
}

func (au *AuthHandler) CompleteGoogleAuth(w http.ResponseWriter, r *http.Request) {
	token, user, err := CompleteUserAuth(w, r)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}
	r_url := "https://zuri.chat/login/?token=" + token
	http.Redirect(w, r, r_url, http.StatusTemporaryRedirect)
	fmt.Println(user.Email)
}
func (au *AuthHandler) HtmlTemplate(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("template/index.html")
	t.Execute(w, false)
}

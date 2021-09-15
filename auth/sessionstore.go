package auth

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrorInvalid = errors.New("Zuri Core session: invalid session id")
	Resptoken    ResToken
)

type Session struct {
	Id       primitive.ObjectID `bson:"_id,omitempty`
	Data     string
	Modified time.Time
}

type ResToken struct {
	Id          string            `json:"id,omitempty"`
	Email       string            `json:"email,omitempty"`
	SessionName string            `json:"session_name"`
	Cookie      string            `json:"cookie,omitempty"`
	Options     *sessions.Options `json:"options,omitempty"`
}

type MongoStore struct {
	Codecs  []securecookie.Codec
	Options *sessions.Options
	Token   TokenGetSeter
	coll    *mongo.Collection
}

func NewMongoStore(c *mongo.Collection, maxAge int, ensureTTL bool, keyPairs ...[]byte) *MongoStore {
	store := &MongoStore{
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: maxAge,
		},
		Token: &CookieToken{},
		coll:  c,
	}

	store.MaxAge(maxAge)

	if ensureTTL {
		x := (time.Duration(maxAge) * time.Second)

		indexModel := mongo.IndexModel{
			Keys:    bson.M{"modified": 1},
			Options: options.Index().SetExpireAfterSeconds(int32(x.Seconds())),
		}
		c.Indexes().CreateOne(context.Background(), indexModel)
	}

	return store
}

func (m *MongoStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(m, name)
}

func (m *MongoStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(m, name)
	session.Options = &sessions.Options{
		Path:     m.Options.Path,
		Domain:   m.Options.Domain,
		MaxAge:   m.Options.MaxAge,
		Secure:   m.Options.Secure,
		HttpOnly: m.Options.HttpOnly,
	}

	session.IsNew = true

	var err error
	if cook, errToken := m.Token.GetToken(r, name); errToken == nil {
		err = securecookie.DecodeMulti(name, cook, &session.ID, m.Codecs...)
		if err == nil {
			err = m.load(session)
			if err == nil {
				session.IsNew = false
			} else {
				err = nil
			}
		}
	}

	return session, err
}
func NewS(m *MongoStore, cookie string, id string, email string, r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(m, name)
	session.Options = &sessions.Options{
		Path:     m.Options.Path,
		Domain:   m.Options.Domain,
		MaxAge:   m.Options.MaxAge,
		Secure:   m.Options.Secure,
		HttpOnly: m.Options.HttpOnly,
	}

	session.IsNew = true
	session.Values["id"] = id
	session.Values["email"] = email
	// session.ID = id
	var err error
	err = securecookie.DecodeMulti(name, cookie, &session.ID, m.Codecs...)
	if err == nil {
		errb := m.load(session)
		if errb == nil {
			session.IsNew = false
		} else {
			err = nil
		}
	}

	return session, err
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

func (m *MongoStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {

	if session.Options.MaxAge < 0 {
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

	if session.ID == "" {
		session.ID = primitive.NewObjectID().Hex()
	}

	if err := m.upsert(session); err != nil {
		return err
	}

	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, m.Codecs...)
	if err != nil {
		return err
	}
	Resptoken = ResToken{
		Id:          session.ID,
		SessionName: session.Name(),
		Cookie:      encoded,
		Options:     session.Options,
	}
	m.Token.SetToken(w, session.Name(), encoded, session.Options)
	return nil
}

func (m *MongoStore) MaxAge(age int) {
	m.Options.MaxAge = age

	for _, codec := range m.Codecs {
		if sc, ok := codec.(*securecookie.SecureCookie); ok {
			sc.MaxAge(age)
		}
	}
}

func (m *MongoStore) load(session *sessions.Session) error {
	ctx := context.Background()

	objID, err := primitive.ObjectIDFromHex(session.ID)
	if err != nil {
		return ErrorInvalid
	}

	s := Session{}
	if err := m.coll.FindOne(ctx, bson.M{"_id": objID}).Decode(&s); err != nil {
		return err
	}

	if err := securecookie.DecodeMulti(session.Name(), s.Data, &session.Values, m.Codecs...); err != nil {
		return err
	}
	return nil
}

func (m *MongoStore) upsert(session *sessions.Session) error {
	ctx := context.Background()

	objID, err := primitive.ObjectIDFromHex(session.ID)
	if err != nil {
		return errors.New("Zuri Core session: invalid session id")
	}

	var modified time.Time
	if val, ok := session.Values["modified"]; ok {
		modified, ok = val.(time.Time)
		if !ok {
			return ErrorInvalid
		}

	} else {
		modified = time.Now()
	}
	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values, m.Codecs...)
	s := Session{
		Id:       objID,
		Data:     encoded,
		Modified: modified,
	}

	opts := options.Update().SetUpsert(true)
	filter := bson.M{"_id": s.Id}
	update_data := bson.M{"$set": s}

	if _, err = m.coll.UpdateOne(ctx, filter, update_data, opts); err != nil {
		return err
	}

	return nil
}

func (m *MongoStore) delete(session *sessions.Session) error {
	ctx := context.Background()
	fmt.Print("I was called")

	objID, err := primitive.ObjectIDFromHex(session.ID)
	if err != nil {
		return ErrorInvalid
	}

	_, err = m.coll.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return err
	}

	return nil
}

var iv = []byte{35, 46, 57, 24, 85, 35, 24, 74, 87, 35, 88, 98, 66, 32, 14, 05}

func encodeBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func decodeBase64(s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}

func Encrypt(key, text string) string {
	fmt.Println(text)
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		panic(err)
	}
	plaintext := []byte(text)
	cfb := cipher.NewCFBEncrypter(block, iv)
	ciphertext := make([]byte, len(plaintext))
	cfb.XORKeyStream(ciphertext, plaintext)
	return encodeBase64(ciphertext)
}

func Decrypt(key, text string) string {
	fmt.Println(text)
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		panic(err)
	}
	ciphertext := decodeBase64(text)
	cfb := cipher.NewCFBEncrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	cfb.XORKeyStream(plaintext, ciphertext)
	return string(plaintext)
}

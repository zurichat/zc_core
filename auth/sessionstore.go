package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrorInvalid = errors.New("zuri core session: invalid session id")
	Resptoken    ResToken
)

type Session struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	UserID   primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"`
	Data     string
	Modified time.Time
}

type ResToken struct {
	ID            string            `json:"id,omitempty"`
	Email         string            `json:"email,omitempty"`
	SessionName   string            `json:"session_name"`
	Cookie        string            `json:"cookie,omitempty"`
	Options       *sessions.Options `json:"options,omitempty"`
	GothicUser    goth.User         `json:"gothuser"`
	GothicEmail   string            `json:"gothicemail"`
	Gothic        interface{}       `json:"gothic"`
	SocialSession *sessions.Session `json:"social_session"`
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

	// if ensureTTL {
	// 	x := (time.Duration(maxAge) * time.Second)

	// 	indexModel := mongo.IndexModel{
	// 		Keys: bson.M{"modified": 1},
	// 		Options: options.Index().SetExpireAfterSeconds(int32(x.Seconds())),
	// 	}
	// 	c.Indexes().CreateOne(context.Background(), indexModel)
	// }

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

func NewS(m *MongoStore, cookie, id, email string, r *http.Request, name string, gothic interface{}) (*sessions.Session, error) {
	session := sessions.NewSession(m, name)
	session.Options = &sessions.Options{
		Path:     m.Options.Path,
		Domain:   m.Options.Domain,
		MaxAge:   m.Options.MaxAge,
		Secure:   m.Options.Secure,
		HttpOnly: m.Options.HttpOnly,
	}

	session.IsNew = true
	err := securecookie.DecodeMulti(name, cookie, &session.ID, m.Codecs...)

	if err == nil {
		if gothic != nil {
			session.Values["gothic"] = gothic
		} else {
			session.Values["id"] = id
			session.Values["email"] = email
		}

		errb := m.load(session)
		if errb == nil {
			session.IsNew = false
		}
	}

	return session, err
}

func (m *MongoStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
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
		ID:          session.ID,
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

	var (
		userID primitive.ObjectID
		eeer   error
	)

	if session.Values["id"] != nil && session.Values["id"] != "" {
		userID, eeer = primitive.ObjectIDFromHex(session.Values["id"].(string))
	}

	if eeer != nil {
		userID, _ = primitive.ObjectIDFromHex("")
	}

	if err != nil {
		return errors.New("zuri core session: invalid session id")
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

	encoded, _ := securecookie.EncodeMulti(session.Name(), session.Values, m.Codecs...)
	s := Session{
		ID:       objID,
		UserID:   userID,
		Data:     encoded,
		Modified: modified,
	}
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"_id": s.ID}
	updateData := bson.M{"$set": s}

	if _, err = m.coll.UpdateOne(ctx, filter, updateData, opts); err != nil {
		return err
	}

	return nil
}

func (m *MongoStore) delete(session *sessions.Session) error {
	ctx := context.Background()

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
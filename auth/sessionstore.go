package auth

import (
	"context"
	"errors"
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
)

type Session struct{
	Id          primitive.ObjectID `bson:"_id,omitempty`
	Data 		string
	Modified	time.Time
}

type MongoStore struct {
	Codecs 	[]securecookie.Codec
	Options	*sessions.Options
	Token	TokenGetSeter
	coll	*mongo.Collection
}

func NewMongoStore(c *mongo.Collection, maxAge int, ensureTTL bool, keyPairs ...[]byte) *MongoStore {
	store := &MongoStore{
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
		Options: &sessions.Options{
			Path: "/",
			MaxAge: maxAge,
		},
		Token: &CookieToken{},
		coll: c,
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

func (m *MongoStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {

	if session.Options.MaxAge < 0 {
		if err := m.delete(session); err != nil {
			return err
		}
		m.Token.SetToken(w, session.Name(), "", session.Options)
		return nil
	}

	if session.ID == "" {
		session.ID = primitive.NewObjectID().Hex()
	}

	if err := m.upsert(session); err != nil {
		return err
	}

	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, m.Codecs...)
	if err != nil { return err }

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
	if _, err := primitive.ObjectIDFromHex(session.ID); err != nil {
		return ErrorInvalid
	}
	
	s := Session{}
	if err := m.coll.FindOne(ctx, bson.M{"_id": session.ID}).Decode(&s); err != nil {
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
		if !ok { return ErrorInvalid }

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

	objID, err := primitive.ObjectIDFromHex(session.ID)
	if err != nil {
		return ErrorInvalid
	}

	_, err = m.coll.DeleteOne(ctx, objID)
	if err != nil { return err }

	return nil
}

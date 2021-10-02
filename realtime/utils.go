package realtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"zuri.chat/zccore/utils"
)

var (
	ExceptOrigins = []string{"https://zuri.chat/", "https://zuri.chat", "http://zuri.chat", "https://www.zuri.chat"}
	CDcollection  = "rtcconnections"
	expiry        = 10
)

type ConnectionDocument struct {
	Origin string `bson:"origin"  json:"origin"`
	Expiry int    `bson:"expiry"  json:"expiry"`
}
type Disc struct {
	Code      int    `bson:"code"  json:"code"`
	Reconnect bool   `bson:"reconnect"  json:"reconnect"`
	Reason    string `bson:"reason"  json:"reason"`
}
type CustomAthResp struct {
	Disconnect Disc `bson:"disconnect"  json:"disconnect"`
}

func contains(v string, a []string) bool {
	for _, i := range a {
		if i == v {
			return true
		}
	}
	return false
}

func ConnectLimitError(count int) error {
	err := "Max Connection limit of " + fmt.Sprintf("%v", MaxConnectionCount) + " Exceeded"
	return errors.New(err)
}

func CheckOrigin(r *http.Request) (string, bool) {
	origin := r.Header["Origin"][0]
	if !contains(origin, ExceptOrigins) {
		return origin, false
	}
	return origin, true
}

func GetandSetDb(collection string, expiry int) {
	c := utils.GetCollection(collection)
	grt, cc_filter := make(map[string]interface{}), make(map[string]interface{})
	grt["$lt"] = int(time.Now().Unix())
	cc_filter["expiry"] = grt
	utils.DeleteManyMongoDoc(CDcollection, cc_filter)
	indexModel := mongo.IndexModel{
		Keys: bson.M{"modified": 1},
		// Options: options.Index().SetExpireAfterSeconds(int32(x.Seconds())),
		Options: options.Index().SetExpireAfterSeconds(int32(expiry)),
	}
	c.Indexes().CreateOne(context.Background(), indexModel)
}

func CheckOriginConnections(origin string) (int, bool) {
	res, _ := utils.GetMongoDbDocs(CDcollection, bson.M{"origin": origin})
	c_count := len(res)
	if c_count >= MaxConnectionCount {
		return c_count, false
	}

	return c_count, true
}

func CustomAthResponse(w http.ResponseWriter, code int, reconnect bool, reason string) {
	inn := Disc{Code: code, Reconnect: reconnect, Reason: reason}
	fRes := CustomAthResp{Disconnect: inn}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(fRes); err != nil {
		log.Printf("Error sending response: %v", err)
	}
	fmt.Println(fRes)
}
func AuthorizeOrigin(r *http.Request) error {
	GetandSetDb(CDcollection, expiry)
	origin, status := CheckOrigin(r)
	if !status {
		nc_count, state := CheckOriginConnections(origin)
		if !state {
			return ConnectLimitError(nc_count)
		}
		dt := ConnectionDocument{Origin: origin, Expiry: int(time.Now().Unix()) + expiry}
		detail, _ := utils.StructToMap(dt)

		_, err := utils.CreateMongoDbDoc(CDcollection, detail)

		if err != nil {
			return err
		}
	}
	return nil
}

package external

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"zuri.chat/zccore/utils"
)

func EmailSubscription(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	newsletter_collection := "subscription"
	var NewSubscription Subscription
	type sub_res struct {
		status bool
	}
	err := json.NewDecoder(r.Body).Decode(&NewSubscription)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	SubDoc, _ := utils.GetMongoDbDoc(newsletter_collection, bson.M{"email": NewSubscription.Email})
	if SubDoc != nil {
		// fmt.Printf("user with email %s already subscribed!", NewSubscription.Email)
		utils.GetSuccess("Thanks for subscribing to for or Newsletter", sub_res{status: true}, w)
		return
	}

	coll := utils.GetCollection(newsletter_collection)
	res, err := coll.InsertOne(r.Context(), NewSubscription)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}
	fmt.Println(res.InsertedID)
	utils.GetSuccess("Thanks for subscribing to for or Newsletter", sub_res{status: true}, w)

}

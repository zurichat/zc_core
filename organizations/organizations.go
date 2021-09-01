package organizations

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/utils"
)

func GetOrganization(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	collection := "organizations"

	orgId := mux.Vars(r)["id"]
	objId, _ := primitive.ObjectIDFromHex(orgId)

	save, err := utils.GetMongoDbDocs(collection, bson.M{"_id": objId})

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}
	utils.GetSuccess("organization retrieved successfully", save, w)
}

func Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// parse form data
	r.ParseForm()
	collection, user_collection := "organizations", "users"

	// validate required fields
	// add required params into required array, make an empty array to hold error strings, make map to hold valid form params for creating organization
	required, empty, form_params := []string{"user_id", "name", "email"}, make([]string, 0), make(map[string]interface{})

	// get the form params
	form_params["user_id"] = r.FormValue("user_id")
	form_params["name"] = r.FormValue("name")
	form_params["email"] = r.FormValue("email")

	// loop through and check for empty required params
	for _, value := range required {
		if str, ok := form_params[value].(string); ok {
			if strings.TrimSpace(str) == "" {
				empty = append(empty, strings.Join(strings.Split(value, "_"), " "))
			}
		} else {
			empty = append(empty, strings.Join(strings.Split(value, "_"), " "))
		}
	}
	if len(empty) > 0 {
		utils.GetError(errors.New(strings.Join(empty, ", ")+" required"), http.StatusBadRequest, w)
		return
	}

	// check if organization name is already taken
	org_filter := make(map[string]interface{})
	org_filter["name"] = form_params["name"]
	org, _ := utils.GetMongoDbDoc(collection, org_filter)
	if org != nil {
		utils.GetError(errors.New("organization name is already taken"), http.StatusBadRequest, w)
		return
	}

	// confirm if user_id exists
	user_filter := make(map[string]interface{})
	user_filter["user_id"] = form_params["user_id"]
	user, _ := utils.GetMongoDbDoc(user_collection, user_filter)
	if user == nil {
		utils.GetError(errors.New("invalid user id"), http.StatusBadRequest, w)
		return
	}

	// save organization
	save, err := utils.CreateMongoDbDoc(collection, form_params)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}
	utils.GetSuccess("organization created", save, w)
}

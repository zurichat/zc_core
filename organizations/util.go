package organizations

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"zuri.chat/zccore/auth"
	"zuri.chat/zccore/service"
	"zuri.chat/zccore/utils"
)

func NewOrganizationHandler(c *utils.Configurations, mail service.MailService) *OrganizationHandler {
	return &OrganizationHandler{configs: c, mailService: mail}
}

// gets the details of a member in a workspace using parameters such as email, username etc
// returns parameters based on the member struct.
func FetchMember(filter map[string]interface{}) (*Member, error) {
	member := &Member{}
	memberCollection, err := utils.GetMongoDBCollection(os.Getenv("DB_NAME"), MemberCollectionName)

	if err != nil {
		return nil, err
	}

	result := memberCollection.FindOne(context.TODO(), filter)

	err = mapstructure.Decode(result, &member)

	if err != nil {
		return nil, err
	}

	return member, nil
}

// check that an organization exist.
func ValidateOrg(orgID string) error {
	// check that org_id is valid
	pOrgID, err := primitive.ObjectIDFromHex(orgID)
	if err != nil {
		return errors.New("invalid organization id")
	}

	// check that org exists
	orgDoc, _ := utils.GetMongoDBDoc(OrganizationCollectionName, bson.M{"_id": pOrgID})
	if orgDoc == nil {
		fmt.Printf("org with id %s doesn't exist!", orgID)
		return errors.New("organization does not exist")
	}

	return nil
}

// check that a member belongs in the an organization.
func ValidateMember(orgID, memberID string) error {
	// check that org_id is valid
	pMemID, err := primitive.ObjectIDFromHex(memberID)
	if err != nil {
		return errors.New("invalid Member id")
	}

	// check that member exists
	memberDoc, _ := utils.GetMongoDBDoc(MemberCollectionName, bson.M{"_id": pMemID, "org_id": orgID})
	if memberDoc == nil {
		fmt.Printf("member with id %s doesn't exist!", memberID)
		return errors.New("member does not exist")
	}

	return nil
}

// create member instance.
func NewMember(email, userName, orgID, role string) Member {
	return Member{
		Email:    email,
		UserName: userName,
		OrgID:    orgID,
		Role:     role,
		Presence: "true",
		JoinedAt: time.Now(),
		Deleted:  false,
		Settings: new(Settings),
	}
}

// clear a member's status after a duration.
func ClearStatusRoutine(orgID, memberID string, ch chan int64, clearOld chan bool) {
	// get period from channel
	period := <-ch

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(period)*time.Minute)
	defer cancel()

	d := time.Duration(period) * time.Minute
	t := time.NewTimer(d)

	go func() {
		for {
			select {
			case <-clearOld:
				// force timer to stop when new time is available
				// this occures everytime a new time is set so that old times
				// running can be interrupted
				if !t.Stop() {
					<-t.C
				}

				// restart timer because a condition occurred
				newD := time.Duration(period) * time.Minute
				t.Reset(newD)

			case <-t.C:
				// clear status when the timer completes!
				ClearStatus(memberID, period)

			case <-ctx.Done():
				return
			}
		}
	}()
	<-ctx.Done()

	// publish update to subscriber
	eventChannel := fmt.Sprintf("organizations_%s", orgID)
	event := utils.Event{Identifier: memberID, Type: "User", Event: UpdateOrganizationMemberStatusCleared, Channel: eventChannel, Payload: make(map[string]interface{})}

	go utils.Emitter(event)
}

func ClearStatus(memberID string, duration int64) {
	// duration 1 represents dont_clear time
	if duration == 1 {
		return
	}

	pmemberID, _ := primitive.ObjectIDFromHex(memberID)

	memberRec, err := utils.GetMongoDBDoc(MemberCollectionName, bson.M{"_id": pmemberID})
	if err != nil {
		log.Println("error while trying to get member")
		return
	}

	var prevStatus Status

	// convert bson to struct
	bsonBytes, _ := bson.Marshal(memberRec["status"])

	if err = bson.Unmarshal(bsonBytes, &prevStatus); err != nil {
		log.Println(err)
		return
	}

	update, _ := utils.StructToMap(Status{StatusHistory: prevStatus.StatusHistory})

	memberStatus := make(map[string]interface{})
	memberStatus["status"] = update

	result, err := utils.UpdateOneMongoDBDoc(MemberCollectionName, memberID, memberStatus)
	if err != nil {
		log.Println(err)
		return
	}

	if result.ModifiedCount == 0 {
		log.Println(err)
		return
	}

	log.Printf("%s status cleared successfully. Duration: %d", memberID, duration)
}

func FetchOrganization(filter map[string]interface{}) (*Organization, error) {
	organization := &Organization{}
	orgCollection, err := utils.GetMongoDBCollection(os.Getenv("DB_NAME"), OrganizationCollectionName)

	if err != nil {
		return organization, err
	}

	result := orgCollection.FindOne(context.TODO(), filter)
	err = result.Decode(&organization)

	return organization, err
}

// func (o *Organization) OrgPlugins() []map[string]interface{} {
// 	return o.Plugins
// }

func (o *Organization) OrgPlugins() map[string]interface{} {
	return o.Plugins
}

// used to update any field in an organization.
func OrganizationUpdate(w http.ResponseWriter, r *http.Request, updateParam updateParam) {
	w.Header().Set("Content-Type", "application/json")

	orgID := mux.Vars(r)["id"]
	_, err := primitive.ObjectIDFromHex(orgID)

	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	if err = utils.ParseJSONFromRequest(r, &RequestData); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	orgFilter := make(map[string]interface{})
	orgFilter[updateParam.orgFilterKey] = RequestData[updateParam.requestDataKey]
	update, err := utils.UpdateOneMongoDBDoc(OrganizationCollectionName, orgID, orgFilter)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if update.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	eventChannel := fmt.Sprintf("organizations_%s", orgID)
	event := utils.Event{Identifier: orgID, Type: "Organization", Event: updateParam.eventKey, Channel: eventChannel, Payload: make(map[string]interface{})}

	go utils.Emitter(event)

	utils.GetSuccess(fmt.Sprintf("%s updated successfully", updateParam.successMessage), nil, w)
}

func HandleMemberSearch(orgID, memberID string, ch chan HandleMemberSearchResponse, wg *sync.WaitGroup) {
	defer wg.Done()

	memberIDhex, err := primitive.ObjectIDFromHex(memberID)
	if err != nil {
		resp := HandleMemberSearchResponse{Memberinfo: Member{}, Err: err}
		ch <- resp

		return
	}

	orgMember, err := utils.GetMongoDBDoc(MemberCollectionName, bson.M{
		"org_id":  orgID,
		"_id":     memberIDhex,
		"deleted": bson.M{"$ne": true},
	})

	if err != nil {
		resp := HandleMemberSearchResponse{Memberinfo: Member{}, Err: err}
		ch <- resp

		return
	}

	var member Member

	bsonBytes, _ := bson.Marshal(orgMember)
	if err := bson.Unmarshal(bsonBytes, &member); err != nil {
		return
	}

	resp := HandleMemberSearchResponse{Memberinfo: member, Err: nil}
	ch <- resp
}

func RemoveHistoryAtIndex(s []StatusHistory, index int) []StatusHistory {
	return append(s[:index], s[index+1:]...)
}

func InsertHistoryAtIndex(s []StatusHistory, history StatusHistory, index int) []StatusHistory {
	return append(s[:index], append([]StatusHistory{history}, s[index:]...)...)
}

type checkSettingsPayload func() (ok bool)

type settingsPayload struct {
	settings             interface{}
	checkSettingsPayload checkSettingsPayload
	field                string
}

// utility function to manage updates to a member's setting.
func updateMemberSettings(w http.ResponseWriter, r *http.Request, settingsPayload settingsPayload) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	orgID, memberID := vars["id"], vars["mem_id"]

	// check that org_id is valid
	err := ValidateOrg(orgID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// check that member_id is valid
	err = ValidateMember(orgID, memberID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	// Parse request from incoming payload
	err = utils.ParseJSONFromRequest(r, &settingsPayload.settings)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	if ok := settingsPayload.checkSettingsPayload(); !ok {
		return
	}

	// convert setting struct to map
	settingsMap, err := utils.StructToMap(settingsPayload.settings)
	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	memberSettings := make(map[string]interface{})
	memberSettings[settingsPayload.field] = settingsMap

	// fetch and update the document
	update, err := utils.UpdateOneMongoDBDoc(MemberCollectionName, memberID, memberSettings)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if update.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	// publish update to subscriber
	eventChannel := fmt.Sprintf("organizations_%s", orgID)
	event := utils.Event{Identifier: memberID, Type: "User", Event: UpdateOrganizationMemberSettings, Channel: eventChannel, Payload: make(map[string]interface{})}

	go utils.Emitter(event)

	utils.GetSuccess("Member settings updated successfully", nil, w)
}

// utility function to manage update to organization billing.
func updateBilling(w http.ResponseWriter, r *http.Request, settingsPayload settingsPayload) {
	w.Header().Set("Content-Type", "application/json")

	orgID := mux.Vars(r)["id"]

	if err := utils.ParseJSONFromRequest(r, &settingsPayload.settings); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	validate := validator.New()

	if err := validate.Struct(settingsPayload.settings); err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	loggedInUser, ok := r.Context().Value("user").(*auth.AuthUser)
	if !ok {
		utils.GetError(errors.New("invalid user"), http.StatusBadRequest, w)
		return
	}

	member, err := FetchMember(bson.M{"org_id": orgID, "email": loggedInUser.Email})
	if err != nil {
		utils.GetError(errors.New("access denied"), http.StatusNotFound, w)
		return
	}

	orgFilter := make(map[string]interface{})
	orgFilter[settingsPayload.field] = settingsPayload.settings

	update, err := utils.UpdateOneMongoDBDoc(OrganizationCollectionName, orgID, orgFilter)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if update.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusUnprocessableEntity, w)
		return
	}

	// publish update to subscriber
	eventChannel := fmt.Sprintf("organizations_%s", orgID)
	event := utils.Event{Identifier: member.ID, Type: "User", Event: UpdateOrganizationBillingSettings, Channel: eventChannel, Payload: make(map[string]interface{})}

	go utils.Emitter(event)

	utils.GetSuccess("organization billing updated successfully", nil, w)
}

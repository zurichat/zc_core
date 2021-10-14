package organizations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

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
		return member, err
	}

	result := memberCollection.FindOne(context.TODO(), filter)

	err = result.Decode(&member)

	return member, err
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
		ID:       primitive.NewObjectID(),
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

	pmemberID, err := primitive.ObjectIDFromHex(memberID)
	if err != nil {
		log.Println("Invalid id")
		return
	}

	memberRec, err := utils.GetMongoDBDoc(MemberCollectionName, bson.M{"_id": pmemberID})
	if err != nil {
		log.Println("error while trying to get member")
		return
	}

	var prevStatus Status

	// convert bson to struct
	bsonBytes, _ := bson.Marshal(memberRec["status"])

	if err = bson.Unmarshal(bsonBytes, &prevStatus); err != nil {
		log.Println("error while trying to unmarshal")
		return
	}

	update, _ := utils.StructToMap(Status{StatusHistory: prevStatus.StatusHistory})

	memberStatus := make(map[string]interface{})
	memberStatus["status"] = update

	_, err = utils.UpdateOneMongoDBDoc(MemberCollectionName, memberID, memberStatus)
	if err != nil {
		log.Println("could not clear status")
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

func GetOrgPluginCollectionName(orgName string) string {
	return strings.ToLower(orgName) + "_" + InstalledPluginsCollectionName
}

func (o *Organization) OrgPlugins() []map[string]interface{} {
	orgCollectionName := GetOrgPluginCollectionName(o.ID)

	orgPlugins, _ := utils.GetMongoDBDocs(orgCollectionName, nil)

	var pluginsMap []map[string]interface{}

	pluginJSON, _ := json.Marshal(orgPlugins)
	err := json.Unmarshal(pluginJSON, &pluginsMap)

	if err != nil {
		return nil
	}

	return pluginsMap
}

// used to update any field in an organization.
func OrganizationUpdate(w http.ResponseWriter, r *http.Request, updateParam updateParam) {
	w.Header().Set("Content-Type", "application/json")

	orgID := mux.Vars(r)["id"]
	requestData := make(map[string]string)

	if err := utils.ParseJSONFromRequest(r, &requestData); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	orgFilter := make(map[string]interface{})
	orgFilter[updateParam.orgFilterKey] = requestData[updateParam.requestDataKey]
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

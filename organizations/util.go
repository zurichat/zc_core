package organizations

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"zuri.chat/zccore/service"
	"zuri.chat/zccore/utils"
)

func NewOrganizationHandler(c *utils.Configurations, mail service.MailService) *OrganizationHandler {
	return &OrganizationHandler{configs: c, mailService: mail}
}

// gets the details of a member in a workspace using parameters such as email, username etc
// returns parameters based on the member struct
func FetchMember(filter map[string]interface{}) (*Member, error) {
	member_collection := MemberCollectionName
	member := &Member{}
	memberCollection, err := utils.GetMongoDbCollection(os.Getenv("DB_NAME"), member_collection)
	if err != nil {
		return member, err
	}
	result := memberCollection.FindOne(context.TODO(), filter)
	err = result.Decode(&member)
	return member, err
}

// check that an organization exist
func ValidateOrg(orgId string) error{
	
	// check that org_id is valid
	pOrgId, err := primitive.ObjectIDFromHex(orgId)
	if err != nil {
		return errors.New("invalid organization id")
	}

	// check that org exists
	orgDoc, _ := utils.GetMongoDbDoc(OrganizationCollectionName, bson.M{"_id": pOrgId})
	if orgDoc == nil {
		fmt.Printf("org with id %s doesn't exist!", orgId)
		return errors.New("organization does not exist")
	}

	return nil
}

// check that a member belongs in the an organization
func ValidateMember(orgId, member_Id string) error{
	
	// check that org_id is valid
	pMemId, err := primitive.ObjectIDFromHex(member_Id)
	if err != nil {
		return errors.New("invalid Member id")	
	}

	// check that member exists
	memberDoc, _ := utils.GetMongoDbDoc(MemberCollectionName, bson.M{"_id": pMemId, "org_id": orgId})
	if memberDoc == nil {
		fmt.Printf("member with id %s doesn't exist!", member_Id)
		return errors.New("member does not exist")
	}

	return nil
}

// create member instance 
func NewMember(email string, userName string, orgId string, role string, setting *Settings) Member {
	return Member{
		ID:       primitive.NewObjectID(),
		Email:    email,
		UserName: userName,
		OrgId:    orgId,
		Role:     role,
		Presence: "true", 
		JoinedAt: time.Now(),
		Deleted:  false,
		Settings: setting,
	}
}

// clear a member's status after a duration
func ClearStatus(orgId, member_id string, period int) {
	time.Sleep(time.Duration(period) * time.Minute)
	update, _ := utils.StructToMap(Status{})

	memberStatus := make(map[string]interface{})
	memberStatus["status"] = update

	_, err := utils.UpdateOneMongoDbDoc(MemberCollectionName, member_id, memberStatus)
	if err != nil {
		log.Println("could not clear status")
		return
	}

	log.Printf("%s status cleared successfully", member_id)

	// publish update to subscriber
	eventChannel := fmt.Sprintf("organizations_%s", orgId)
	event := utils.Event{Identifier: member_id, Type: "User", Event: UpdateOrganizationMemberStatusCleared, Channel: eventChannel, Payload: make(map[string]interface{})}
	go utils.Emitter(event)
}

func FetchOrganization(filter map[string]interface{}) (*Organization, error) {
	org_collection := OrganizationCollectionName
	organization := &Organization{}
	orgCollection, err := utils.GetMongoDbCollection(os.Getenv("DB_NAME"), org_collection)
	if err != nil {
		return organization, err
	}
	result := orgCollection.FindOne(context.TODO(), filter)
	err = result.Decode(&organization)
	return organization, err
}
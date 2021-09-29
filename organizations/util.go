package organizations

import (
	"context"
	"errors"
	"fmt"
	"os"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"zuri.chat/zccore/utils"
	"zuri.chat/zccore/service"
)


type OrganizationHandler struct {
	configs     *utils.Configurations
	mailService service.MailService
}

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
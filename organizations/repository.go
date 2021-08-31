package organizations

import (
	"context"
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"zuri.chat/zccore/utils"
)

type OrgRepository interface {
	Create(ctx context.Context, org Organization) (*mongo.InsertOneResult, error)
}

type organizationRepository struct {
	Collection *mongo.Collection
}

func NewOrgRepository(Collection *mongo.Collection) OrgRepository {
	return &organizationRepository{Collection}
}

func (c *organizationRepository) Create(ctx context.Context, org Organization) (*mongo.InsertOneResult, error) {
	
	
	if org.Name == ""{
		org.Name = "ZuriWorkspace"
	}
	
	org.Url= org.Name+".zuri.chat"
	org.DateCreated = time.Now()

	// convert to map object
	var inInterface map[string]interface{}
    inrec, _ := json.Marshal(org)
    json.Unmarshal(inrec, &inInterface)

	// store into db
	response, err := utils.CreateMongoDbDoc(c.Collection.Name(), inInterface)

	if err != nil {
		return nil, err
	}
	return response, nil
}
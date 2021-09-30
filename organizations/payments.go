package organizations

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/utils"
)

const (
	USD = "usd" // US Dollar ($)
	EUR = "eur" // Euro (€)
	GBP = "gbp" // British Pound Sterling (UK£)
	JPY = "jpy" // Japanese Yen (¥)
	CAD = "cad" // Canadian Dollar (CA$)
	HKD = "hkd" // Hong Kong Dollar (HK$)
	CNY = "cny" // Chinese Yuan (CN¥)
	AUD = "aud" // Australian Dollar (A$)
)

// converts amount in naira to equivalent token value
func GetTokenAmount(Amount float64, Currency string) (float64, error) {
	var ExchangeMap = map[string]float64{
		USD: 1,
		EUR: 0.86,
	}
	ConversionRate, ok := ExchangeMap[Currency]
	if !ok {
		return float64(0), errors.New("currency not yet supported")
	}
	return Amount * ConversionRate, nil
}

// takes as input org id and token amount and increments token by that amount
func IncrementToken(OrgId string, TokenAmount float64) error {
	OrgIdFromHex, err := primitive.ObjectIDFromHex(OrgId)
	if err != nil {
		return err
	}

	organization, err := FetchOrganization(bson.M{"_id": OrgIdFromHex})
	if err != nil {
		return err
	}

	organization.Tokens += TokenAmount
	update_data := make(map[string]interface{})
	update_data["tokens"] = organization.Tokens
	if _, err := utils.UpdateOneMongoDbDoc(OrganizationCollectionName, OrgId, update_data); err != nil {
		return err
	}
	return nil
}

// takes as input org id and token amount and decreases token by that amount if available, else returns error
func DeductToken(OrgId string, TokenAmount float64) error {

	OrgIdFromHex, err := primitive.ObjectIDFromHex(OrgId)
	if err != nil {
		return err
	}

	organization, err := FetchOrganization(bson.M{"_id": OrgIdFromHex})
	if err != nil {
		return err
	}

	if organization.Tokens < TokenAmount {
		return errors.New("insufficient token balance")
	}

	organization.Tokens -= TokenAmount
	update_data := make(map[string]interface{})
	update_data["tokens"] = organization.Tokens
	if _, err := utils.UpdateOneMongoDbDoc(OrganizationCollectionName, OrgId, update_data); err != nil {
		return err
	}
	return nil
}

func SubscriptionBilling(OrgId string, ProVersionRate float64) error {

	orgMembers, err := utils.GetMongoDbDocs(MemberCollectionName, bson.M{"org_id": OrgId})
	if err != nil {
		return err
	}

	amount := float64(len(orgMembers)) * ProVersionRate

	if err := DeductToken(OrgId, amount); err != nil {
		return err
	}
	return nil
}

func SendTokenBillingEmail() {

}

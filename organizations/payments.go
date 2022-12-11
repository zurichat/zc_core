package organizations

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/checkout/session"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/service"
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

// Credit Card Types accepted by the Stripe API.
const (
	AmericanExpress = "American Express"
	DinersClub      = "Diners Club"
	Discover        = "Discover"
	JCB             = "JCB"
	MasterCard      = "MasterCard"
	Visa            = "Visa"
	UnknownCard     = "Unknown"
)

// Converts amount in real currency to equivalent token value.
func GetTokenAmount(amount float64, currency string) (float64, error) {
	var ExchangeMap = map[string]float64{
		USD: 1,
		EUR: 0.86,
	}

	ConversionRate, ok := ExchangeMap[currency]

	if !ok {
		return float64(0), errors.New("currency not yet supported")
	}

	return amount * ConversionRate, nil
}

// Takes as input org id and token amount and increments token by that amount.
func IncrementToken(orgID, description string, tokenAmount float64) error {
	OrgIDFromHex, err := primitive.ObjectIDFromHex(orgID)
	if err != nil {
		return err
	}

	organization, err := FetchOrganization(bson.M{"_id": OrgIDFromHex})
	if err != nil {
		return err
	}

	organization.Tokens += tokenAmount

	updateData := make(map[string]interface{})
	updateData["tokens"] = organization.Tokens

	if _, err := utils.UpdateOneMongoDBDoc(OrganizationCollectionName, orgID, updateData); err != nil {
		return err
	}

	return nil
}

// Takes as input org id and token amount and decreases token by that amount if available, else returns error.
func DeductToken(orgID, description string, tokenAmount float64) error {

	if strings.Contains(orgID, "-org") {
		organization, err := FetchOrganization(bson.M{"_id": orgID})
		if err != nil {
			return err
		}
		if organization.Tokens < tokenAmount {
			return errors.New("insufficient token balance")
		}

		organization.Tokens -= tokenAmount

		updateData := make(map[string]interface{})
		updateData["tokens"] = organization.Tokens

		return err

	} else {
		OrgIDFromHex, err := primitive.ObjectIDFromHex(orgID)
		if err != nil {
			return err
		}
		organization, err := FetchOrganization(bson.M{"_id": OrgIDFromHex})
		if err != nil {
			return err
		}

		if organization.Tokens < tokenAmount {
			return errors.New("insufficient token balance")
		}

		organization.Tokens -= tokenAmount

		updateData := make(map[string]interface{})
		updateData["tokens"] = organization.Tokens

		return err
	}

}

func SubscriptionBilling(orgID string, proVersionRate float64) error {
	orgMembers, err := utils.GetMongoDBDocs(MemberCollectionName, bson.M{"org_id": orgID})
	if err != nil {
		return err
	}

	var description string

	amount := float64(len(orgMembers)) * proVersionRate
	numMembers := len(orgMembers)

	description = "Billing for Pro version subscription for " + strconv.Itoa(numMembers) + " member(s) at " + strconv.Itoa(int(proVersionRate)) + " token(s) per member per month"

	if err := DeductToken(orgID, description, amount); err != nil {
		return err
	}

	return nil
}

func SendTokenBillingEmail(orgID, description string, amount float64) error {
	OrgIDFromHex, err := primitive.ObjectIDFromHex(orgID)
	if err != nil {
		return err
	}

	org, _ := FetchOrganization(bson.M{"_id": OrgIDFromHex})
	orgMail := org.CreatorEmail
	balance := org.Tokens
	name := org.Name

	ms := service.NewZcMailService(utils.NewConfigurations())
	billingMail := ms.NewMail(
		[]string{orgMail},
		"Token Billing Notice",
		service.TokenBillingNotice,
		map[string]interface{}{
			"Description": description,
			"Cost":        amount,
			"Balance":     balance,
			"Name":        name,
		},
	)

	if err := ms.SendMail(billingMail); err != nil {
		return err
	}

	return nil
}

// Allows user to be able to load tokens into organization wallet.
func (oh *OrganizationHandler) AddToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var org bson.M

	orgID := mux.Vars(r)["id"]
	if strings.Contains(orgID, "-org") {
		org, _ = utils.GetMongoDBDoc(OrganizationCollectionName, bson.M{"_id": orgID})

		if org == nil {
			utils.GetError(fmt.Errorf("organization %s not found", orgID), http.StatusNotFound, w)
			return
		}
	} else {
		objID, err := primitive.ObjectIDFromHex(orgID)

		if err != nil {
			utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
			return
		}

		org, _ = utils.GetMongoDBDoc(OrganizationCollectionName, bson.M{"_id": objID})

		if org == nil {
			utils.GetError(fmt.Errorf("organization %s not found", orgID), http.StatusNotFound, w)
			return
		}
	}

	requestData := make(map[string]float64)
	if err := utils.ParseJSONFromRequest(r, &requestData); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	orgFilter := make(map[string]interface{})

	tokens, ok := requestData["amount"]
	if !ok {
		utils.GetError(errors.New("amount not supplied"), http.StatusUnprocessableEntity, w)
		return
	}

	orgFilter["tokens"] = org["tokens"].(float64) + (tokens * 0.2)

	update, err := utils.UpdateOneMongoDBDoc(OrganizationCollectionName, orgID, orgFilter)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	var transaction TokenTransaction

	transaction.Amount = tokens
	transaction.Currency = "usd"
	transaction.Description = "Purchase Token"
	transaction.OrgID = orgID
	transaction.TransactionID = utils.GenUUID()
	transaction.Type = "Purchase"
	transaction.Time = time.Now()
	transaction.Token = tokens * 0.2
	detail, _ := utils.StructToMap(transaction)

	res, err := utils.CreateMongoDBDoc(TokenTransactionCollectionName, detail)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if update.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("Successfully loaded token", res, w)
}

// Get an organization transaction record.
func (oh *OrganizationHandler) GetTokenTransaction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	orgID := mux.Vars(r)["id"]

	save, _ := utils.GetMongoDBDocs(TokenTransactionCollectionName, bson.M{"org_id": orgID})

	if save == nil {
		utils.GetError(fmt.Errorf("organization transaction %s not found", orgID), http.StatusNotFound, w)
		return
	}

	utils.GetSuccess("transactions retrieved successfully", save, w)
}

// Charge token.
func (oh *OrganizationHandler) ChargeTokens(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	orgID := mux.Vars(r)["id"]

	if err := utils.ParseJSONFromRequest(r, &RequestData); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	bitSize := 64

	amount, err := strconv.ParseFloat(RequestData["amount"], bitSize)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	description := RequestData["description"]

	if err := DeductToken(orgID, description, amount); err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	utils.GetSuccess("Billing successful for: "+description, nil, w)
}

// Create checkout session.
func (oh *OrganizationHandler) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	orgID := mux.Vars(r)["id"]

	err := ValidateOrg(orgID)
	if err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	requestData := make(map[string]int64)
	if err = utils.ParseJSONFromRequest(r, &requestData); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	amount, ok := requestData["amount"]
	if !ok {
		utils.GetError(errors.New("amount not supplied"), http.StatusUnprocessableEntity, w)
		return
	}

	stripeAmount := amount * 100
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("usd"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Purchase Token"),
					},
					UnitAmount: stripe.Int64(stripeAmount),
				},
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(os.Getenv("FRONT_END_URL") + "/admin/settings/billings?status=success&orgId=" + orgID),
		CancelURL:  stripe.String(os.Getenv("FRONT_END_URL") + "/admin/settings/billings?status=failed&orgId=" + orgID),
	}

	s, err := session.New(params)

	if err != nil {
		utils.GetError(errors.New("error trying to initiate payment"), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("successfully initiated payment", s.URL, w)
}

// Save card details of a user for future payment.
func (oh *OrganizationHandler) AddCard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var orgDoc bson.M
	orgID := mux.Vars(r)["id"]

	if strings.Contains(orgID, "-org") {
		orgDoc, _ = utils.GetMongoDBDoc(OrganizationCollectionName, bson.M{"_id": orgID})
		if orgDoc == nil {
			utils.GetError(errors.New("organization does not exist"), http.StatusBadRequest, w)
			return
		}
	} else {

		// Checks if organization id is valid
		orgIDHex, err := primitive.ObjectIDFromHex(orgID)
		if err != nil {
			utils.GetError(errors.New("invalid organization id"), http.StatusBadRequest, w)
			return
		}

		// Checks if organization exists in the database
		orgDoc, _ = utils.GetMongoDBDoc(OrganizationCollectionName, bson.M{"_id": orgIDHex})
		if orgDoc == nil {
			utils.GetError(errors.New("organization does not exist"), http.StatusBadRequest, w)
			return
		}
	}

	MemberID := mux.Vars(r)["mem_id"]
	objID, err := primitive.ObjectIDFromHex(MemberID)

	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	member, _ := utils.GetMongoDBDoc(MemberCollectionName, bson.M{"_id": objID})

	if member == nil {
		utils.GetError(fmt.Errorf("member %s not found", MemberID), http.StatusNotFound, w)
		return
	}

	var newcard Card

	if err = utils.ParseJSONFromRequest(r, &newcard); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	newcard.OrgID = orgID
	newcard.MemberID = MemberID
	newcard.CVCCheck = string(utils.GCMEncrypt([]byte(newcard.CVCCheck), "zcore_key"))
	newcard.CardNumber = string(utils.GCMEncrypt([]byte(newcard.CardNumber), "zcore_key"))

	// convert card struct to map
	card, err := utils.StructToMap(newcard)

	if err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	res, err := utils.CreateMongoDBDoc(CardCollectionName, card)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("successfully added new card", res, w)
}

// Delete card detail of a user.
func (oh *OrganizationHandler) DeleteCard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	orgID := mux.Vars(r)["id"]
	var orgDoc bson.M

	// Checks if organization id is valid

	if strings.Contains(orgID, "-org") {
		orgDoc, _ = utils.GetMongoDBDoc(OrganizationCollectionName, bson.M{"_id": orgID})
		if orgDoc == nil {
			utils.GetError(errors.New("organization does not exist"), http.StatusBadRequest, w)
			return
		}
	} else {
		orgIDHex, err := primitive.ObjectIDFromHex(orgID)
		if err != nil {
			utils.GetError(errors.New("invalid organization id"), http.StatusBadRequest, w)
			return
		}
		// Checks if organization exists in the database
		orgDoc, _ = utils.GetMongoDBDoc(OrganizationCollectionName, bson.M{"_id": orgIDHex})
		if orgDoc == nil {
			utils.GetError(errors.New("organization does not exist"), http.StatusBadRequest, w)
			return
		}
	}

	MemberID := mux.Vars(r)["mem_id"]
	objID, err := primitive.ObjectIDFromHex(MemberID)

	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	member, _ := utils.GetMongoDBDoc(MemberCollectionName, bson.M{"_id": objID})

	if member == nil {
		utils.GetError(fmt.Errorf("member %s not found", MemberID), http.StatusNotFound, w)
		return
	}

	MemberCard := mux.Vars(r)["card_id"]
	res, err := utils.DeleteOneMongoDBDoc(CardCollectionName, MemberCard)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if res.DeletedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("card successfully deleted", res, w)
}

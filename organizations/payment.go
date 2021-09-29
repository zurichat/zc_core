package organizations

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/checkout/session"
	"zuri.chat/zccore/utils"
)

// ISO  3- digit currency codes for major currencies(not the full list)
const (
	USD = "usd"
	AUD = "aud"
)

// type CreditLoad struct {
// 	OrgId         string    `json:"org_id" bson:"org_id"`
// 	Currency      string    `json:"currency" bson:"currency"`
// 	Stripetoken   string    `json:"stripe_token" bson:"stripe_token"`
// 	Email         string    `json:"email" bson:"email"`
// 	Amount        float64   `json:"amount" bson:"amount"`
// 	Card          *Card     `json:"card" bson:"card"`
// 	Time          time.Time `json:"time" bson:"time"`
// 	TransactionId string    `json:"transaction_id" bson:"transaction_id"`
// }

// type Card struct {
// 	Id       string `json:"id" bson:"id"`
// 	Name     string `json:"name, omitempty" bson:"name, omitempty"`
// 	CardType string `json:"card_type" bson:"card_type"`
// 	ExpMonth int    `json:"exp_month" bson:"exp_month"`
// 	ExpYear  int    `json:"exp_year" bson:"exp_year"`
// 	Last4    string `json:"last4" bson:"last4"`
// 	CVCCheck string `json:"cvc_check, omitempty" bson:"cvc_check, omitempty"`
// }

// type StripeToken struct{}

// func (oh *OrganizationHandler) Payment(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")
// 	err := stripe.SetKeyEnv()

// 	if err != nil {
// 			log.Fatal(err)
// 	}
// 	params:= stripe.ChargeParams
// 	_ , err = stripe.Charges.Create(&params)

// 	if err == nil {
// 			fmt.Fprint(W, "payment successful")
// 	} else{
// 			fmt.Fprint(w, "payment unsuccessful"+err.error())
// 	}
// }

func (oh *OrganizationHandler) ConvertToToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	orgId := mux.Vars(r)["id"]
	requestData := make(map[string]string)
	if err := utils.ParseJsonFromRequest(r, &requestData); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	org_filter := make(map[string]interface{})
	org_filter["tokens"] = requestData["tokens"]
	update, err := utils.UpdateOneMongoDbDoc(OrganizationCollectionName, orgId, org_filter)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if update.ModifiedCount == 0 {
		utils.GetError(errors.New("operation failed"), http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("tokens converted successfully", nil, w)
}

func (oh *OrganizationHandler) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			&stripe.CheckoutSessionLineItemParams{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("usd"),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("T-shirt"),
					},
					UnitAmount: stripe.Int64(2000),
				},
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String("https://example.com/success"),
		CancelURL:  stripe.String("https://example.com/cancel"),
	}

	s, _ := session.New(params)

	utils.GetSuccess("Payment successfully created", s, w)
	return
}

// Check organization id
// get the amount from body data
// convert the amount to toekn
// update token field on that organization

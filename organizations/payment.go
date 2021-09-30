package organizations

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/utils"
)

func (oh *OrganizationHandler) AddToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	orgId := mux.Vars(r)["id"]
	objId, err := primitive.ObjectIDFromHex(orgId)

	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	org, _ := utils.GetMongoDbDoc(OrganizationCollectionName, bson.M{"_id": objId})

	if org == nil {
		utils.GetError(fmt.Errorf("organization %s not found", orgId), http.StatusNotFound, w)
		return
	}

	requestData := make(map[string]float64)
	if err := utils.ParseJsonFromRequest(r, &requestData); err != nil {
		utils.GetError(err, http.StatusUnprocessableEntity, w)
		return
	}

	org_filter := make(map[string]interface{})
	tokens, ok := requestData["amount"]
	if !ok {
		utils.GetError(errors.New("amount not supplied"), http.StatusUnprocessableEntity, w)
		return
	}

	org_filter["tokens"] = org["tokens"].(float64) + (tokens * 0.2)

	update, err := utils.UpdateOneMongoDbDoc(OrganizationCollectionName, orgId, org_filter)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}
	var transaction TokenTransaction

	transaction.Amount = tokens
	transaction.Currency = "usd"
	transaction.Description = "Purchase Token"
	transaction.OrgId = orgId
	transaction.TransactionId = utils.GenUUID()
	transaction.Type = "Purchase"
	transaction.Time = time.Now()
	transaction.Token = tokens * 0.2
	detail, _ := utils.StructToMap(transaction)

	res, err := utils.CreateMongoDbDoc(TokenTransactionCollectionName, detail)

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

// Get an organization record
func (oh *OrganizationHandler) GetTokenTransaction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	orgId := mux.Vars(r)["id"]

	save, _ := utils.GetMongoDbDocs(TokenTransactionCollectionName, bson.M{"org_id": orgId})

	if save == nil {
		utils.GetError(fmt.Errorf("organization transaction %s not found", orgId), http.StatusNotFound, w)
		return
	}

	utils.GetSuccess("transactions retrieved successfully", save, w)
}

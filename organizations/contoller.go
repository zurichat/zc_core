package organizations

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"zuri.chat/zccore/utils"
)

type OrgController struct {
	Service OrgService
}

func NewOrgController(r *mux.Router, os OrgService) {

	handler := &OrgController{
		Service: os,
	}

	r.HandleFunc("/organizations", handler.CreateOrganization).Methods("POST")
	// r.HandleFunc("/organizations", handler.GetOrganizations).Methods("GET")
	// r.HandleFunc("/organizations/{org_id}", handler.GetOrganization).Methods("GET")	
	// r.HandleFunc("/organizations/{org_id}", handler.DeleteOrganization).Methods("DELETE")
}

func (a *OrgController) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var newOrg Organization
	err := decoder.Decode(&newOrg)

	if err != nil {
		panic(err)
	}

	// validate that email is not empty and it meets the format
	if !utils.IsValidEmail(newOrg.OwnerEmail){
		utils.GetError(fmt.Errorf("invalid email format : %s", newOrg.OwnerEmail), http.StatusInternalServerError, w)
		return
	}

	res, err := a.Service.Create(r.Context(), newOrg)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
	}

	utils.GetSuccess("Success", res, w)
}

func (a *OrgController) GetOrganizations(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Endpoint to fetch all organizations")
}

func (a *OrgController) GetOrganization(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Endpoint to fetch all organizations")
}
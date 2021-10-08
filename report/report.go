package report

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/utils"
)

// Add a report.
func AddReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var report Report

	err := json.NewDecoder(r.Body).Decode(&report)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	orgID := mux.Vars(r)["id"]
	objID, err := primitive.ObjectIDFromHex(orgID)

	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	orgDoc, _ := utils.GetMongoDbDoc(OrganizationCollectionName, bson.M{"_id": objID})

	if orgDoc == nil {
		utils.GetError(errors.New("organization with id "+orgID+" doesn't exist!"), http.StatusBadRequest, w)
		return
	}

	report.Organization = orgID
	report.Date = time.Now()

	if !utils.IsValidEmail(report.ReporterEmail) {
		utils.GetError(fmt.Errorf("invalid email format : %s", report.ReporterEmail), http.StatusBadRequest, w)
		return
	}

	// check that reporter is in the organization
	reporterDoc, _ := utils.GetMongoDbDoc(MemberCollectionName, bson.M{"org_id": orgID, "email": report.ReporterEmail})
	if reporterDoc == nil {
		utils.GetError(errors.New("reporter must be a member of this organization"), http.StatusBadRequest, w)
		return
	}

	if !utils.IsValidEmail(report.OffenderEmail) {
		utils.GetError(fmt.Errorf("invalid email format : %s", report.OffenderEmail), http.StatusBadRequest, w)
		return
	}

	// check that offender is in the organization
	offenderDoc, _ := utils.GetMongoDbDoc(MemberCollectionName, bson.M{"org_id": orgID, "email": report.OffenderEmail})
	if offenderDoc == nil {
		utils.GetError(errors.New("offender must be a member of this organization"), http.StatusBadRequest, w)
		return
	}

	if report.Organization == "" {
		utils.GetError(errors.New("organization id required"), http.StatusBadRequest, w)
		return
	}

	if report.Subject == "" {
		utils.GetError(errors.New("report's subject required"), http.StatusBadRequest, w)
		return
	}

	if report.Body == "" {
		utils.GetError(errors.New("report's body required"), http.StatusBadRequest, w)
		return
	}

	var reportMap map[string]interface{}
	err = utils.ConvertStructure(report, &reportMap)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	save, err := utils.CreateMongoDbDoc(ReportCollectionName, reportMap)
	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("report created", utils.M{"report_id": save.InsertedID}, w)
}

// Get a report.
func GetReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	orgID := mux.Vars(r)["id"]

	reportID := mux.Vars(r)["report_id"]
	reportObjID, err := primitive.ObjectIDFromHex(reportID)

	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	doc, _ := utils.GetMongoDbDoc(ReportCollectionName, bson.M{"organization_id": orgID, "_id": reportObjID})

	if doc == nil {
		utils.GetError(fmt.Errorf("report %s not found", orgID), http.StatusNotFound, w)
		return
	}

	var report Report
	err = utils.ConvertStructure(doc, &report)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("report  retrieved successfully", report, w)
}

// Get reports.
func GetReports(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	orgID := mux.Vars(r)["id"]

	doc, _ := utils.GetMongoDbDocs(ReportCollectionName, bson.M{"organization_id": orgID})

	report := []Report{}

	if doc == nil {
		utils.GetSuccess("no report has been added yet", report, w)
		return
	}

	err := utils.ConvertStructure(doc, &report)

	if err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("report  retrieved successfully", report, w)
}

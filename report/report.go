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

// Get a report
func GetReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	reportId := mux.Vars(r)["id"]
	objId, err := primitive.ObjectIDFromHex(reportId)
	
	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusBadRequest, w)
		return
	}

	doc, _ := utils.GetMongoDbDoc(ReportCollectionName, bson.M{"_id": objId})

	if doc == nil {
		utils.GetError(fmt.Errorf("report %s not found", reportId), http.StatusNotFound, w)
		return
	}

	var report Report
	utils.ConvertStructure(doc, &report)

	utils.GetSuccess("report  retrieved successfully", report, w)
}


func AddReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var report Report

	err := json.NewDecoder(r.Body).Decode(&report)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	report.Date = time.Now()

	var reportMap map[string]interface{}
	utils.ConvertStructure(report, &reportMap)	

	save, err := utils.CreateMongoDbDoc(ReportCollectionName, reportMap)
	if err != nil {
		fmt.Println(err)
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	if report.ReporterName == "" {
		utils.GetError(errors.New("reporter's name required"), http.StatusBadRequest, w)
		return
	}

	if report.ReporteeName == "" {
		utils.GetError(errors.New("reportee's name required"), http.StatusBadRequest, w)
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

	utils.GetSuccess("report created", save, w)
}
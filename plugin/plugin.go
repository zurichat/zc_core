package plugin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/utils"
)


func SyncUpdate(w http.ResponseWriter, r *http.Request) {
	pp := SyncUpdateRequest{}

	ppID, err := primitive.ObjectIDFromHex(mux.Vars(r)["id"])

	if err != nil {
		utils.GetError(errors.WithMessage(err, "incorrect id"), http.StatusUnprocessableEntity, w)
		return
	}

	//nolint:govet //dod-san: ignore error shadowing.
	if err := utils.ParseJSONFromRequest(r, &pp); err != nil {
		utils.GetError(errors.WithMessage(err, "error processing request"), http.StatusUnprocessableEntity, w)
		return
	}

	pluginDetails, _ := utils.GetMongoDBDoc(PluginCollectionName, bson.M{"_id": ppID})

	if pluginDetails == nil {
		utils.GetError(errors.WithMessage(fmt.Errorf("plugin not found"), "error processing request"), http.StatusUnprocessableEntity, w)
		return
	}

	var splugin Plugin

	if err = mapstructure.Decode(pluginDetails, &splugin); err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	sort.SliceStable(splugin.Queue, func(i, j int) bool {
		return splugin.Queue[i].ID < splugin.Queue[j].ID
	})

	for i := 0; i < len(splugin.Queue); i++ {
		onestruct := splugin.Queue[i]
		if onestruct.ID <= pp.ID {
			splugin.Queue = append(splugin.Queue[:i], splugin.Queue[i+1:]...)
			i-- // Important: decrease index
		}
	}

	updateFields := make(map[string]interface{})

	updateFields["queue"] = splugin.Queue
	_, ee := utils.UpdateOneMongoDBDoc(PluginCollectionName, mux.Vars(r)["id"], updateFields)

	if ee != nil {
		utils.GetError(ee, http.StatusInternalServerError, w)
		return
	}

	utils.GetSuccess("synchronization updated successful", nil, w)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	coll := utils.GetCollection("plugins")

	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		utils.GetError(errors.New("invalid id"), http.StatusUnprocessableEntity, w)
		return
	}

	_, err = coll.DeleteOne(r.Context(), bson.M{"_id": objectID})

	if err != nil {
		utils.GetError(errors.WithMessage(err, "error deleting plugin"), http.StatusBadRequest, w)
		return
	}

	w.WriteHeader(http.StatusNoContent)

	w.Header().Set("content-type", "application/json")

	utils.GetSuccess("plugin deleted", nil, w)
}

type Handler struct {
	Service
	validate *validator.Validate
}

func (h *Handler) readJSON(r *http.Request, out interface{}) error {
	return json.NewDecoder(r.Body).Decode(out)
}

func (h *Handler) writeJSON(w http.ResponseWriter, code int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	return json.NewEncoder(w).Encode(data)
}

type ResponseModel struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func (h *Handler) successResponse(w http.ResponseWriter, code int, msg string, data interface{}) {
	resp := &ResponseModel{
		Status:  "success",
		Message: msg,
		Data:    data,
	}
	h.writeJSON(w, code, resp)
}

func (h *Handler) errorResponse(w http.ResponseWriter, code int, message string) {
	resp := ResponseModel{"error", message, nil}
	h.writeJSON(w, code, resp)
}

func NewHandler(s Service) *Handler {
	return &Handler{s, validator.New()}
}

func (ph *Handler) Register(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Name           string   `json:"name" validate:"required"`
		Description    string   `json:"description" validate:"required"`
		DeveloperName  string   `json:"developer_name" validate:"required"`
		DeveloperEmail string   `json:"developer_email" validate:"required"`
		TemplateURL    string   `json:"template_url" validate:"required"`
		SidebarURL     string   `json:"sidebar_url" validate:"required"`
		InstallURL     string   `json:"install_url" validate:"required"`
		IconURL        string   `json:"icon_url"`
		Images         []string `json:"images,omitempty"`
		Version        string   `json:"version"`
		Category       string   `json:"category"`
		Tags           []string `json:"tags,omitempty"`
	}{}

	if err := ph.readJSON(r, &data); err != nil {
		ph.errorResponse(w, http.StatusUnprocessableEntity, ErrorMessage(err))
		LogError(err)

		return
	}

	if err := ph.validate.Struct(data); err != nil {
		err = Errorf(EINVALID, "validation error: %v", err)
		ph.errorResponse(w, http.StatusBadRequest, ErrorMessage(err))
		LogError(err)

		return
	}

	if p, err := ph.Service.FindOne(r.Context(), bson.M{
		"template_url": data.TemplateURL,
	}); err == nil && p != nil {
		err := Errorf(EDUPLICATE, "plugin exists")
		ph.errorResponse(w, http.StatusForbidden, ErrorMessage(err))
		LogError(err)

		return
	}

	newPlugin := &Plugin{}

	if err := mapstructure.Decode(data, newPlugin); err != nil {
		ph.errorResponse(w, http.StatusInternalServerError, ErrorMessage(err))
		LogError(err)

		return
	}

	newPlugin.Approved = true
	newPlugin.ApprovedAt = time.Now().String()

	if err := ph.Service.Create(r.Context(), newPlugin); err != nil {
		ph.errorResponse(w, http.StatusInternalServerError, ErrorMessage(err))
		LogError(err)

		return
	}

	ph.successResponse(w, http.StatusCreated, "plugin created", utils.M{"plugin": newPlugin})
}

func (ph *Handler) Update(w http.ResponseWriter, r *http.Request) {
	pp := Patch{}
	id := mux.Vars(r)["id"]

	if err := ph.readJSON(r, &pp); err != nil {
		ph.errorResponse(w, http.StatusUnprocessableEntity, ErrorMessage(err))
		LogError(err)

		return
	}

	objID, err := primitive.ObjectIDFromHex(id)
	
	if err != nil {
       ph.errorResponse(w, http.StatusBadRequest, ErrorMessage(Errorf(ENOENT, "plugin with id %s not found", id)))
       return
	}

	if err := ph.Service.Update(r.Context(), bson.M{"_id": objID}, pp); err != nil {
		ph.errorResponse(w, http.StatusInternalServerError, ErrorMessage(err))
		LogError(err)

		return
	}

	ph.successResponse(w, http.StatusOK, "plugin updated", nil)
}

func(ph *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	objID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		err = Errorf(EINVALID, "cannot proccess request: invalid object id")
		ph.errorResponse(w, http.StatusUnprocessableEntity, ErrorMessage(err))
		
		return
	}

	if err := ph.Service.Delete(r.Context(), bson.M{"_id": objID}); err != nil {
		ph.errorResponse(w, http.StatusInternalServerError, ErrorMessage(err))
		LogError(err)

		return
	}

	ph.successResponse(w, http.StatusOK, "plugin deleted", nil)
}

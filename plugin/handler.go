package plugin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Handler struct {
	Service
	validate *validator.Validate
}

// D is a generic map[string]interface{}.
type D map[string]interface{}

func NewHandler(s Service) *Handler {
	return &Handler{s, validator.New()}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
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

	if err := h.readJSON(r, &data); err != nil {
		h.errorResponse(w, http.StatusUnprocessableEntity, ErrorMessage(err))
		LogError(err)

		return
	}

	if err := h.validate.Struct(data); err != nil {
		err = Errorf(EINVALID, "validation error: %v", err)
		h.errorResponse(w, http.StatusBadRequest, ErrorMessage(err))
		LogError(err)

		return
	}

	if p, err := h.Service.FindOne(r.Context(), bson.M{
		"template_url": data.TemplateURL,
	}); err == nil && p != nil {
		err := Errorf(EDUPLICATE, "plugin exists")
		h.errorResponse(w, http.StatusForbidden, ErrorMessage(err))
		LogError(err)

		return
	}

	newPlugin := &Plugin{}

	if err := mapstructure.Decode(data, newPlugin); err != nil {
		h.errorResponse(w, http.StatusInternalServerError, ErrorMessage(err))
		LogError(err)

		return
	}

	newPlugin.Approved = true
	newPlugin.ApprovedAt = time.Now().String()

	if err := h.Service.Create(r.Context(), newPlugin); err != nil {
		h.errorResponse(w, http.StatusInternalServerError, ErrorMessage(err))
		LogError(err)

		return
	}

	h.successResponse(w, http.StatusCreated, "plugin created", D{"plugin": newPlugin})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	pp := Patch{}
	id := mux.Vars(r)["id"]

	if err := h.readJSON(r, &pp); err != nil {
		h.errorResponse(w, http.StatusUnprocessableEntity, ErrorMessage(err))
		LogError(err)

		return
	}

	objID, err := primitive.ObjectIDFromHex(id)
	
	if err != nil {
       h.errorResponse(w, http.StatusBadRequest, ErrorMessage(Errorf(ENOENT, "plugin with id %s not found", id)))
       return
	}

	if err := h.Service.Update(r.Context(), bson.M{"_id": objID}, pp); err != nil {
		h.errorResponse(w, http.StatusInternalServerError, ErrorMessage(err))
		LogError(err)

		return
	}

	h.successResponse(w, http.StatusOK, "plugin updated", nil)
}

func(h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	objID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		err = Errorf(EINVALID, "cannot process request: invalid object id")
		h.errorResponse(w, http.StatusUnprocessableEntity, ErrorMessage(err))
		
		return
	}

	if err := h.Service.Delete(r.Context(), bson.M{"_id": objID}); err != nil {
		h.errorResponse(w, http.StatusInternalServerError, ErrorMessage(err))
		LogError(err)

		return
	}

	h.successResponse(w, http.StatusOK, "plugin deleted", nil)
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
	
	//nolint:errcheck // why do you need me to perform err checking here 🙈?
	h.writeJSON(w, code, resp)
}

func (h *Handler) errorResponse(w http.ResponseWriter, code int, message string) {
	resp := ResponseModel{"error", message, nil}
	
	//nolint:errcheck // again, why do you need me to perform err checking here 🙈?
	h.writeJSON(w, code, resp)
}

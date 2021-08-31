package organizations

import (
	// "encoding/json"
	// "fmt"
	"encoding/json"
	"net/http"
	// "os"

	"zuri.chat/zccore/utils"
)


type Organisation struct {
	OrganisationId string
}

func GetPluginsOrganizations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	plugins_collection := "plugins"

	decoder := json.NewDecoder(r.Body)
    var org map[string]interface{}
    err := decoder.Decode(&org)
    if err != nil {
		utils.GetError(err, http.StatusNotFound, w)
    }


	/* filter := map[string]interface{}{
		"organisation_id" : "1",
	} */
	
	result, err := utils.GetMongoDbDocs(plugins_collection, org)

	if err != nil {
		utils.GetError(err, http.StatusNotFound, w)
	}

	utils.GetSuccess("Plugins returned successfully", result, w)
}

/* type malformedRequest struct {
    status int
    msg    string
} */

/* func (mr *malformedRequest) Error() string {
    return mr.msg
} */

/* func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
    if r.Header.Get("Content-Type") != "" {
        value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
        if value != "application/json" {
            msg := "Content-Type header is not application/json"
            return &malformedRequest{status: http.StatusUnsupportedMediaType, msg: msg}
        }
    }

    r.Body = http.MaxBytesReader(w, r.Body, 1048576)

    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields()

    err := dec.Decode(&dst)
    if err != nil {
        var syntaxError *json.SyntaxError
        var unmarshalTypeError *json.UnmarshalTypeError

        switch {
        case errors.As(err, &syntaxError):
            msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
            return &malformedRequest{status: http.StatusBadRequest, msg: msg}

        case errors.Is(err, io.ErrUnexpectedEOF):
            msg := fmt.Sprintf("Request body contains badly-formed JSON")
            return &malformedRequest{status: http.StatusBadRequest, msg: msg}

        case errors.As(err, &unmarshalTypeError):
            msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
            return &malformedRequest{status: http.StatusBadRequest, msg: msg}

        case strings.HasPrefix(err.Error(), "json: unknown field "):
            fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
            msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
            return &malformedRequest{status: http.StatusBadRequest, msg: msg}

        case errors.Is(err, io.EOF):
            msg := "Request body must not be empty"
            return &malformedRequest{status: http.StatusBadRequest, msg: msg}

        case err.Error() == "http: request body too large":
            msg := "Request body must not be larger than 1MB"
            return &malformedRequest{status: http.StatusRequestEntityTooLarge, msg: msg}

        default:
            return err
        }
    }

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
        msg := "Request body must only contain a single JSON object"
        return &malformedRequest{status: http.StatusBadRequest, msg: msg}
    }

    return nil
} */
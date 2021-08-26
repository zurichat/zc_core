package plugin

import (
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/utils"
)

const PLUGIN_COLLECTION_NAME = "plugins"

// Plugin (App) model
type Plugin struct {
	ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id"`
	Name         string             `json:"name" bson:"name"`
	Description  string             `json:"description" bson:"description"`
	InstallCount int                `json:"install_count" bson:"install_count"`
	Approved     bool               `json:"approved" bson:"approved"`
	ApprovedAt   time.Time          `json:"approved_at" bson:"approved_at"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`
}

func Create(w http.ResponseWriter, r *http.Request) {
	p := &Plugin{}

	if err := utils.ParseJSONFromRequest(r, p); err != nil {
		utils.GetError(err, http.StatusBadRequest, w)
		return
	}

	if err := createPlugin(p); err != nil {
		utils.GetError(err, http.StatusInternalServerError, w)
		return
	}

	w.WriteHeader(http.StatusCreated)
	utils.GetSuccess("Successfully created plugin", p, w)
}

func createPlugin(p *Plugin) error {
	m, _ := utils.StructToMap(p, "bson")
	m["created_at"] = time.Now()
	res, err := utils.CreateMongoDbDoc(PLUGIN_COLLECTION_NAME, m)
	if err != nil {
		return err
	}
	p.ID = res.InsertedID.(primitive.ObjectID)
	p.CreatedAt = m["created_at"].(time.Time)
	return nil
}

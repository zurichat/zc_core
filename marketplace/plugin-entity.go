import "go.mongodb.org/mongo-driver/bson/primitive"

type Category struct {
	ID   primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name string             `json:"name" bson:"name"`
	Slug string             `json:"slug" bson:"slug"`
}

type Categories struct {
	Categories []Category
}
type Features struct {
	Features []string
}

type Plugins struct {
	ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name         string             `json:"name" bson:"name"`
	Description  string             `json:"description" bson:"description"`
	InstallCount int                `json:"install_count,omitempty" bson:"install___count,omitempty"`
	Categories   *Categories        `bson:",omitempty" bson:",omitempty"`
	Features     *Features          `bson:",omitempty" bson:",omitempty"`
	ApprovedAt   bool               `json:"approved_at,omitempty" bson:"approved___at,omitempty"`
}

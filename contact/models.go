package contact

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"zuri.chat/zccore/service"
)

type ContactFormData struct {
	ID        primitive.ObjectID             `json:"id,omitempty"`
	Subject   string                         `json:"subject"`
	Content   string                         `json:"content"`
	Files     []service.MultipleTempResponse `json:"attachments"`
	Email     string                         `json:"email"`
	CreatedAt time.Time                      `json:"created_at"`
}

package contact

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ContactFormData struct {
	ID          primitive.ObjectID `json:"id,omitempty"`
	Subject     string             `json:"subject"`
	Content     string             `json:"content"`
	Attachments []string           `json:"attachments"`
	Email       string             `json:"email"`
	CreatedAt   time.Time          `json:"created_at"`
}

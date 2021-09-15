package contact

import (
	"mime/multipart"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ContactForm struct {
	ID          primitive.ObjectID      `json:"id,omitempty"`
	Subject     string                  `json:"subject"`
	Content     string                  `json:"content"`
	Attachments []*multipart.FileHeader `json:"attachments"`
	Email       string                  `json:"email"`
	CreatedAt   time.Time               `json:"created_at"`
}

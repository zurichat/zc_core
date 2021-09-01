package user


import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

/* 
  Had to add this struct here to implement user login
*/

type User struct {
	ID                  primitive.ObjectID     `bson:"_id"`
	Email               string                 `bson:"email" validate:"required"`
	UserID              string                 `bson:"user_id"`
	FirstName           string                 `bson:"first_name"`
	LastName            string                 `bson:"last_name"`
	Password            string                 `bson:"password" validate:"required"`
	Phone               string                 `bson:"phone"`
	Status              string                 `bson:"status"`
	OrganizationID      string                 `bson:"organization_id"`
	EmailVerificationID string                 `bson:"email_verification_id"`
	PasswordResetIDs    []string               `bson:"password_reset_ids"`
	Settings            map[string]interface{} `bson:"settings"`
	WorkSpaces          map[string]interface{} `bson:"workspaces"`	
	DisplayName         string                 `bson:"display_name" `
	DisplayImage        string                 `bson:"display_image"`
	About               string                 `bson:"about"`
	Timezone            string                 `bson:"timezone"`
	CreatedAt           time.Time              `bson:"created_at"`
	UpdatedAt           time.Time              `bson:"updated_at"`
	DeletedAt           time.Time              `bson:"deleted_at"`
}
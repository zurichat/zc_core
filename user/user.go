package user

import (
	"fmt"
	"net/http"
	"time"

	"zuri.chat/zccore/utils"
)

const USER_COLL_NAME = "users"

type Settings struct {
	Role string `json:"role"`
}

type User struct {
	FirstName         string    `json:"first_name"`
	LastName          string    `json:"last_name"`
	DisplayName       string    `json:"display_name"`
	Email             string    `json:"email"`
	Password          string    `json:"password"`
	Status            string    `json:"status"`
	Company           string    `json:"company"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	DeletedAt         time.Time `json:"deleted_at"`
	Phone             string    `json:"phone"`
	Timezone          string    `json:"timezone"`
	EmailVerification `json:"email_verification"`
	Settings          `json:"settings"`
	// Company
	// PasswordResets
	// Workspaces
}

func Create(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello world")
}

func createUser(user *User) error {

	_, err := utils.CreateMongoDbDoc(USER_COLL_NAME)
}

func GetUserByID() {}

func GetUsers() {}

func UpdateUser() {}

func DeleteUser() {}

// Validate performs basic validations and returns an error if the user contains invalid fields.
func (u *User) Validate() error {
	return nil
}

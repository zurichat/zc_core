package external

import (
	"time"

	"zuri.chat/zccore/service"
	"zuri.chat/zccore/utils"
)

type Subscription struct {
	Email        string    `json:"email" bson:"email"`
	SubscribedAt time.Time `json:"subscribed_at" bson:"subscribed_at"`
}

type Handler struct {
	configs     *utils.Configurations
	mailService service.MailService
}

func NewExternalHandler(c *utils.Configurations, mail service.MailService) *Handler {
	return &Handler{configs: c, mailService: mail}
}

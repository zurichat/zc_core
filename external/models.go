package external

import "time"

type Subscription struct {
	Email        string    `json:"email" bson:"email"`
	SubscribedAt time.Time `json:"subscribed_at" bson:"subscribed_at"`
}

package user

import "time"

type EmailVerification struct {
	Verified  bool      `json:"verified"`
	Token     string    `json:"token"`
	ExpiredAt time.Time `json:"expired_at"`
}

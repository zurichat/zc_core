package SuidService

import (
	"fmt"
	"math/rand"
	"time"
)

type SuID interface {
	GenerateId(name string, n int) string
}

type suid struct {
}

func (s suid) GenerateId(name string, n int) string {
	rand.Seed(time.Now().UnixNano())
	const charset = "0123456789abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	url := fmt.Sprintf("%s-%s", name, string(b))
	return url
}

func NewSuid() SuID {
	return &suid{}
}

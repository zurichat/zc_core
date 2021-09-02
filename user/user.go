package user

import (
	"fmt"
	"net/http"
)

func CreateUser(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello world")
}

func CreateUserProfile(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello world")
}

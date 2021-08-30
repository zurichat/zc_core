package app

import (
	"fmt"
	"net/http"
)

func AppHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Welcome to Zuri Core App")
}
package data

import (
	"fmt"
	"net/http"
)

func WriteData(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "This is you writing data\n")
}

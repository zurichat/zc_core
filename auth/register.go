package auth

import (
	"net/http"

	"zuri.chat/zccore/utils"
)

// func register(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")

// 	// parse form data
// 	r.ParseForm()
//     user_collection := "users"
// 	// validate required fields
// 	// add required params into required array, make an empty array to hold error strings, make map to hold valid form params for creating organization
// 	required, empty, form_params := []string{"user_id", "name", "email"}, make([]string, 0), make(map[string]interface{})


// 	utils.GetSuccess("User registered successfully", save, w)
// }
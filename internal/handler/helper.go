package handler

import (
	"encoding/json"
	"net/http"
)

// Response ...
type Response struct {
	Message string
	Error   string
}

// respondWithError ...
func respondWithError(w http.ResponseWriter, code int, message string, e string) {
	w.WriteHeader(code)
	if e := json.NewEncoder(w).Encode(Response{Message: message, Error: e}); e != nil {
		panic(e)
	}
}

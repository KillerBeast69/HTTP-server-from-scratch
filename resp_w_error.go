package main

import "net/http"

func respond_with_error(w http.ResponseWriter, code int, msg string) {
	type err_response struct {
		Error string `json:"error"`
	}
	respond_with_json(w, code, err_response{Error: msg})
}

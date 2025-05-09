package router

import (
	"net/http"
)

//SetJSON ...
func SetJSON(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// RemoveContextTypeJSON ...
func RemoveContextTypeJSON(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Del("Content-Type")
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

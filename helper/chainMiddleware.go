package helper

import (
	"net/http"
)

// ChainMiddleware applies multiple middleware functions to an http.HandlerFunc
func ChainMiddleware(h http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
}

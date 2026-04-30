package handler

import (
	"fmt"
	"net/http"
	"strings"

	"backapp/internal/middleware"
)

func buildPublicUploadURL(r *http.Request, path string) string {
	scheme := "http"
	if middleware.IsRequestSecure(r) {
		scheme = "https"
	}

	host := forwardedHost(r)
	return fmt.Sprintf("%s://%s%s", scheme, host, path)
}

func forwardedHost(r *http.Request) string {
	forwardedHost := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-Host"), ",")[0])
	if forwardedHost != "" {
		return forwardedHost
	}

	return r.Host
}

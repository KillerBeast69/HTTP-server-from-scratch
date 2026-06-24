package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("no authorization header included")
	}

	prefix := "ApiKey "
	if !strings.HasPrefix(authHeader, prefix) {
		return "", errors.New("malformed authorization header")
	}

	key := strings.TrimPrefix(authHeader, prefix)
	key = strings.TrimSpace(key)

	if key == "" {
		return "", errors.New("empty api key")
	}

	return key, nil
}

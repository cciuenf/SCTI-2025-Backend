package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"scti/internal/models"
	u "scti/internal/utilities"
	"strings"
)

// extractSlugAndValidate extracts slug from URL path and validates it's not empty
func extractSlugAndValidate(r *http.Request) (string, error) {
	slug := r.PathValue("slug")
	if slug == "" {
		return "", errors.New("the event slug can't be empty")
	}
	return strings.ToLower(slug), nil
}

// getUserFromContext extracts and returns the user from context based on JWT claims
func getUserFromContext(getUserByID func(string) (models.User, error), r *http.Request) (models.User, error) {
	claims := u.GetUserFromContext(r.Context())
	if claims == nil {
		return models.User{}, errors.New("error getting data from claims")
	}
	user, err := getUserByID(claims.ID)
	if err != nil {
		return models.User{}, errors.New("error getting user: " + err.Error())
	}
	return user, nil
}

// decodeRequestBody decodes the request body into the provided struct
func decodeRequestBody(r *http.Request, target interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		return errors.New("error parsing request body: " + err.Error())
	}
	return nil
}

// handleError sends a standardized error response
func handleError(w http.ResponseWriter, err error, statusCode int) {
	u.SendError(w, []string{err.Error()}, "event-stack", statusCode)
}

// handleSuccess sends a standardized success response
func handleSuccess(w http.ResponseWriter, data interface{}, message string, statusCode int) {
	u.SendSuccess(w, data, message, statusCode)
}

package utilities

import (
	"context"
	"log"

	"scti/internal/models"
)

func GetUserFromContext(ctx context.Context) *models.UserClaims {
	claims, ok := ctx.Value("user").(*models.UserClaims)
	if !ok {
		log.Println("No user claims found in context")
		return nil
	}
	return claims
}

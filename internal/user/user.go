package user

import (
	"github.com/google/uuid"
	"strings"
	"time"
)

type User struct {
	ID, OrganizationID uuid.UUID `json:"id"`
	Email, Name        string    `json:"email"`
	CreatedAt          time.Time `json:"created_at"`
}

func Valid(email, name string) bool {
	return strings.Contains(email, "@") && strings.TrimSpace(name) != "" && len(email) <= 320 && len(name) <= 255
}

package organization

import (
	"github.com/google/uuid"
	"strings"
	"time"
)

type Organization struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func ValidName(s string) bool { return strings.TrimSpace(s) != "" && len(s) <= 255 }

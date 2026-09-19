package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewComment(t *testing.T) {
	now := time.Date(2026, 9, 18, 10, 0, 0, 0, time.UTC)

	comment, err := NewComment("comment-1", "ticket-1", "user-1", "Looking into it.", true, now)
	if err != nil {
		t.Fatalf("NewComment() error = %v", err)
	}

	if !comment.Internal {
		t.Fatalf("Internal = false, want true")
	}
	if !comment.CreatedAt.Equal(now) || !comment.UpdatedAt.Equal(now) {
		t.Fatalf("timestamps = (%v, %v), want %v", comment.CreatedAt, comment.UpdatedAt, now)
	}
}

func TestCommentValidate(t *testing.T) {
	now := time.Date(2026, 9, 18, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		comment Comment
		want    error
	}{
		{
			name:    "valid public reply",
			comment: Comment{ID: "comment-1", TicketID: "ticket-1", AuthorID: "user-1", Body: "Thanks.", CreatedAt: now, UpdatedAt: now},
		},
		{
			name:    "missing body",
			comment: Comment{ID: "comment-1", TicketID: "ticket-1", AuthorID: "user-1", Body: " ", CreatedAt: now, UpdatedAt: now},
			want:    ErrMissingRequiredField,
		},
		{
			name:    "missing ticket id",
			comment: Comment{ID: "comment-1", AuthorID: "user-1", Body: "Thanks.", CreatedAt: now, UpdatedAt: now},
			want:    ErrMissingRequiredField,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.comment.Validate()
			if tt.want == nil {
				if err != nil {
					t.Fatalf("Validate() error = %v", err)
				}
				return
			}
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want errors.Is(..., %v)", err, tt.want)
			}
		})
	}
}

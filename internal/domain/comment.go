package domain

import "time"

type Comment struct {
	ID        CommentID
	TicketID  TicketID
	AuthorID  UserID
	Body      string
	Internal  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewComment(id CommentID, ticketID TicketID, authorID UserID, body string, internal bool, at time.Time) (*Comment, error) {
	comment := &Comment{
		ID:        id,
		TicketID:  ticketID,
		AuthorID:  authorID,
		Body:      body,
		Internal:  internal,
		CreatedAt: at,
		UpdatedAt: at,
	}

	if err := comment.Validate(); err != nil {
		return nil, err
	}

	return comment, nil
}

func (c Comment) Validate() error {
	if err := required("id", c.ID.IsZero()); err != nil {
		return err
	}
	if err := required("ticket_id", c.TicketID.IsZero()); err != nil {
		return err
	}
	if err := required("author_id", c.AuthorID.IsZero()); err != nil {
		return err
	}
	if err := required("body", blank(c.Body)); err != nil {
		return err
	}
	if err := required("created_at", c.CreatedAt.IsZero()); err != nil {
		return err
	}
	if err := required("updated_at", c.UpdatedAt.IsZero()); err != nil {
		return err
	}

	return nil
}

package models

import (
	"github.com/google/uuid"
)

type Meta struct {
	ID       uuid.UUID
	Relation uuid.UUID
	Name     string
	Data     string
}

type LoginPassword struct {
	ID       *uuid.UUID
	UserID   uuid.UUID
	Login    string
	Password string
}

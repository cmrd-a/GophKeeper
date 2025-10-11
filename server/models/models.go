package models

import (
	"time"

	"github.com/google/uuid"
)

// Common fields for all vault items.
type VaultItem struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Meta struct {
	ID        uuid.UUID
	Relation  uuid.UUID
	Name      string
	Data      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type LoginPassword struct {
	VaultItem

	Login    string
	Password string
}

type TextData struct {
	VaultItem

	Text string
}

type BinaryData struct {
	VaultItem

	Data []byte
}

type CardData struct {
	VaultItem

	Number  []byte // encrypted
	CVV     []byte // encrypted
	Holder  string
	Expires time.Time
}

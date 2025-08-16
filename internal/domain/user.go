package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID        uuid.UUID
	Name      string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewUser creates a new user
func NewUser(name, email string) *User {
	return &User{
		ID:        uuid.New(),
		Name:      name,
		Email:     email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// UpdateDetails updates user name and email
func (u *User) UpdateDetails(name, email string) {
	u.Name = name
	u.Email = email
	u.UpdatedAt = time.Now()
}


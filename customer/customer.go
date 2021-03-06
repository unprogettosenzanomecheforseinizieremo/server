package customer

import (
	"github.com/google/uuid"
	"time"
)

// Represent the customer
// At the moment it can't be update
type Customer struct {
	ID                 ID        `json:"id"`
	Email              Email     `json:"email"`
	Password           Password  `json:"-"`
	Status             Status    `json:"status"`
	ActivateHash       string    `json:"-"`
	ChangePasswordHash string    `json:"-"`
	ChangeEmailHash    string    `json:"-"`
	Created            time.Time `json:"created"`
	Activated          time.Time `json:"activated,omitempty"`
	Updated            time.Time `json:"updated,omitempty"`
}

// New returns a customer created for the first time
func New(id ID, email Email, password Password) *Customer {
	return &Customer{
		ID:           id,
		Email:        email,
		Status:       NotActivated,
		ActivateHash: uuid.New().String(),
		Password:     password,
		Created:      time.Now(),
	}
}

// IsActive checks if a customer is active
func (c *Customer) IsActive() bool {
	return c.Status == Activated
}

// Activate activates the customer comparing the hash
func (c *Customer) Activate(hash string) bool {
	if hash != c.ActivateHash || c.ActivateHash == "" {
		return false
	}

	c.ActivateHash = ""
	c.Status = Activated
	c.Activated = time.Now()
	return true
}

// GenerateChangePasswordHash changes the customer password
func (c *Customer) GenerateChangePasswordHash() {
	c.ChangePasswordHash = uuid.New().String()
	c.Updated = time.Now()
}

// ChangePassword changes the customer password
func (c *Customer) ChangePassword(p Password, hash string) bool {
	if hash != c.ChangePasswordHash || c.ChangePasswordHash == "" {
		return false
	}
	c.ChangePasswordHash = ""
	c.Password = p
	c.Updated = time.Now()
	return true
}

// GenerateChangeEmailHash changes the customer password
func (c *Customer) GenerateChangeEmailHash() {
	c.ChangeEmailHash = uuid.New().String()
	c.Updated = time.Now()
}

// ChangeEmail changes the customer email
func (c *Customer) ChangeEmail(e Email, hash string) bool {
	if hash != c.ChangeEmailHash || c.ChangeEmailHash == "" {
		return false
	}
	c.ChangeEmailHash = ""
	c.Email = e
	c.Status = PendingEmail
	c.ActivateHash = uuid.New().String()
	c.Updated = time.Now()
	return true
}

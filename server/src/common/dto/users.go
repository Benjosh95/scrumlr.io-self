package dto

import (
	"net/http"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"scrumlr.io/server/database"
	"scrumlr.io/server/database/types"
)

// User is the response for all user requests.
type User struct {
	// The id of the user
	ID uuid.UUID `json:"id"`

	// The user's display name
	Name string `json:"name"`

	// The user's avatar configuration
	Avatar *types.Avatar `json:"avatar,omitempty"`

	// The user's passkey credentials
	Credentials []webauthn.Credential `json:"credentials"` // TODO: OK? omitempty?
}

func (u *User) From(user database.User) *User {
	u.ID = user.ID
	u.Name = user.Name
	u.Avatar = user.Avatar
	u.Credentials = user.Credentials
	return u
}

func (*User) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}

func (u *User) WebAuthnIcon() string {
	return "" // empty string OK
}
func (u *User) WebAuthnID() []byte {
	return []byte(u.ID.String())
}
func (u *User) WebAuthnName() string {
	return u.Name
}
func (u *User) WebAuthnDisplayName() string {
	return u.Name
}
func (u *User) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}

type UserUpdateRequest struct {
	ID          uuid.UUID             `json:"-"`
	Name        string                `json:"name"`
	Avatar      *types.Avatar         `json:"avatar,omitempty"`
	Credentials []webauthn.Credential `json:"credentials"`
}

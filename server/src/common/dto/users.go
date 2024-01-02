package dto

import (
	"net/http"
	"time"

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
	Credentials []types.ExtendedCredential `json:"credentials"` // TODO: omitempty?
}

type UserUpdateRequest struct {
	ID          uuid.UUID                  `json:"-"`
	Name        string                     `json:"name"`
	Avatar      *types.Avatar              `json:"avatar,omitempty"`
	Credentials []types.ExtendedCredential `json:"credentials"` // TODO: omitempty?
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

// ConvertToWebAuthnCredential converts ExtendedCredential to webauthn.Credential
func ConvertToWebAuthnCredentials(extendedCredentials []types.ExtendedCredential) []webauthn.Credential {
	webAuthnCredentials := make([]webauthn.Credential, len(extendedCredentials))

	for i, extendedCredential := range extendedCredentials {
		webAuthnCredentials[i] = webauthn.Credential{
			ID:              extendedCredential.ID,
			PublicKey:       extendedCredential.PublicKey,
			AttestationType: extendedCredential.AttestationType,
			Transport:       extendedCredential.Transport,
			Flags:           extendedCredential.Flags,
			Authenticator:   extendedCredential.Authenticator,
		}
	}

	return webAuthnCredentials
}

// ConvertToExtendedCredential converts webauthn.Credential to ExtendedCredential
func ConvertToExtendedCredential(webAuthnCredential webauthn.Credential) types.ExtendedCredential {
	return types.ExtendedCredential{
		Credential: webAuthnCredential,
		CreatedAt:  time.Now(),
		LastUsedAt: time.Now(),
	}
}

func (u *User) WebAuthnIcon() string {
	return ""
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
	convertedWebAuthnCredentials := ConvertToWebAuthnCredentials(u.Credentials)
	return convertedWebAuthnCredentials
}

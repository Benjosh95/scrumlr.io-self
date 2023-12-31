package types

import (
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
)

type ExtendedCredential struct {
	webauthn.Credential
	DisplayName string    `json:"displayName,omitempty"`
	LastUsedAt  time.Time `json:"lastUsedAt"`
	CreatedAt   time.Time `json:"createdAt"`
}

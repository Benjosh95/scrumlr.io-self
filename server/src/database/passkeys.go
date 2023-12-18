package database

import (
	"context"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"scrumlr.io/server/common"
)

type PasskeySession struct {
	bun.BaseModel `bun:"table:passkey_sessions"`
	ID            uuid.UUID
	User          uuid.UUID
	Challenge     string
	// AllowedCredentialIDs [][]byte
	AllowedCredentialIDs [][]byte
	Expires              time.Time
	UserVerification     protocol.UserVerificationRequirement
	Extensions           protocol.AuthenticationExtensions
	CreatedAt            time.Time
}

type PasskeySessionInsert struct {
	bun.BaseModel `bun:"table:passkey_sessions"`
	User          uuid.UUID
	Challenge     string
	// AllowedCredentialIDs [][]byte
	AllowedCredentialIDs [][]byte
	Expires              time.Time
	UserVerification     protocol.UserVerificationRequirement
	Extensions           protocol.AuthenticationExtensions
}

func (d *Database) GetPasskeySession(userId uuid.UUID) (PasskeySession, error) {
	var passkeySession PasskeySession
	err := d.db.NewSelect().Model(&passkeySession).Where("\"user\" = ?", userId).Scan(context.Background())
	return passkeySession, err
}

func (d *Database) CreatePasskeySession(insert PasskeySessionInsert) (PasskeySession, error) {
	// Declare a variable to store the created or updated PasskeySession
	var createdSession PasskeySession

	// Check if a PasskeySession with the given user ID already exists
	_, err := d.GetPasskeySession(insert.User)
	if err == nil {
		// PasskeySession with the given user ID already exists, update it
		_, err := d.db.NewUpdate().
			Model(&insert).
			Where("\"user\" = ?", insert.User).
			Returning("*"). // Return all columns of the updated record
			Exec(common.ContextWithValues(context.Background(), "Database", d), &createdSession)

		// Return the updated PasskeySession and any error that occurred
		return createdSession, err
	}

	// PasskeySession with the given user ID does not exist, insert a new one
	_, err = d.db.NewInsert().
		Model(&insert).
		Returning("*"). // Return all columns of the inserted record
		Exec(common.ContextWithValues(context.Background(), "Database", d), &createdSession)

	// Return the created PasskeySession and any error that occurred
	return createdSession, err
}

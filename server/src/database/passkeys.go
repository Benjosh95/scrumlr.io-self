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
	bun.BaseModel        `bun:"table:passkey_sessions"`
	ID                   uuid.UUID
	User                 uuid.UUID
	Challenge            string
	AllowedCredentialIDs [][]byte
	Expires              time.Time
	UserVerification     protocol.UserVerificationRequirement
	Extensions           protocol.AuthenticationExtensions
	CreatedAt            time.Time
}

type PasskeySessionInsert struct {
	bun.BaseModel        `bun:"table:passkey_sessions"`
	User                 uuid.UUID
	Challenge            string
	AllowedCredentialIDs [][]byte
	Expires              time.Time
	UserVerification     protocol.UserVerificationRequirement
	Extensions           protocol.AuthenticationExtensions
}

func (d *Database) GetSessionByUserId(userId uuid.UUID) (PasskeySession, error) {
	var passkeySession PasskeySession
	err := d.db.NewSelect().Model(&passkeySession).Where("\"user\" = ?", userId).Scan(context.Background())
	return passkeySession, err
}

func (d *Database) GetSessionById(id uuid.UUID) (PasskeySession, error) {
	var passkeySession PasskeySession
	err := d.db.NewSelect().Model(&passkeySession).Where("id = ?", id).Scan(context.Background(), &passkeySession)
	return passkeySession, err
}

func (d *Database) CreateSession(insert PasskeySessionInsert) (PasskeySession, error) {
	var createdSession PasskeySession

	// Check if a PasskeySession with the given user ID already exists
	session, err := d.GetSessionByUserId(insert.User)
	isNullUUID := session.User == uuid.Nil

	if err == nil && !isNullUUID {
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

	return createdSession, err
}

package passkeys

import (
	"context"

	"scrumlr.io/server/common"
	"scrumlr.io/server/database"
	"scrumlr.io/server/logger"
	"scrumlr.io/server/realtime"
	"scrumlr.io/server/services"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

type PasskeyService struct {
	database DB
	realtime *realtime.Broker
}

func NewPasskeyService(db DB, rt *realtime.Broker) services.Passkeys {
	b := new(PasskeyService)
	b.database = db
	b.realtime = rt
	return b
}

type DB interface {
	CreatePasskeySession(database.PasskeySessionInsert) (database.PasskeySession, error)
	GetPasskeySession(uuid.UUID) (database.PasskeySession, error)
}

func (s *PasskeyService) GetSession(ctx context.Context, userId uuid.UUID) (*webauthn.SessionData, error) {
	log := logger.FromContext(ctx)

	session, err := s.database.GetPasskeySession(userId)
	if err != nil {
		log.Errorw("failed to get passkey session", "err", err)
		return nil, common.InternalServerError

	}

	return PasskeySessionToWebAuthnSession(session), nil
}

func (s *PasskeyService) CreateSession(ctx context.Context, session *webauthn.SessionData) error {
	log := logger.FromContext(ctx)

	userID, err := uuid.Parse(string(session.UserID))
	if err != nil {
		log.Errorw("failed to parse userId", "err", err)
		return common.InternalServerError
	}

	passkeySessionInsert := WebAuthnSessionToPasskeySession(session, userID)
	_, err = s.database.CreatePasskeySession(passkeySessionInsert)
	if err != nil {
		log.Errorw("failed to create passkey session", "err", err)
		return common.InternalServerError
	}

	return err
}

// two conversion functions to make typing work with go-webauthn library's "sessionData" type
func WebAuthnSessionToPasskeySession(session *webauthn.SessionData, userID uuid.UUID) database.PasskeySessionInsert {
	return database.PasskeySessionInsert{
		User:                 userID,
		Challenge:            session.Challenge,
		AllowedCredentialIDs: session.AllowedCredentialIDs,
		Expires:              session.Expires,
		UserVerification:     session.UserVerification,
		Extensions:           session.Extensions,
	}
}

func PasskeySessionToWebAuthnSession(session database.PasskeySession) *webauthn.SessionData {
	return &webauthn.SessionData{
		UserID:               []byte(session.User.String()),
		Challenge:            session.Challenge,
		AllowedCredentialIDs: session.AllowedCredentialIDs,
		Expires:              session.Expires,
		UserVerification:     session.UserVerification,
		Extensions:           session.Extensions,
	}
}

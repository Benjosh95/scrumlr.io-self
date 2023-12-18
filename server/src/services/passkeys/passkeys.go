package passkeys

import (
	"context"
	"fmt"

	"scrumlr.io/server/database"
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
	// TODO: b.database.AttachObserver((database.NotesObserver)(b))
	return b
}

type DB interface {
	CreatePasskeySession(database.PasskeySessionInsert) (database.PasskeySession, error)
	GetPasskeySession(uuid.UUID) (database.PasskeySession, error)
}

func (s *PasskeyService) GetSession(ctx context.Context, userId uuid.UUID) (*webauthn.SessionData, error) {
	session, err := s.database.GetPasskeySession(userId)
	if err != nil {
		return nil, err
	}
	fmt.Print("GetSession -------- session:", session)
	return PasskeySessionToWebAuthnSession(session), nil
}

func (s *PasskeyService) CreateSession(ctx context.Context, session *webauthn.SessionData) (string, error) {
	// TODO: log := logger.FromContext(ctx)

	//TODO:
	userID, err := uuid.Parse(string(session.UserID))
	if err != nil {
		return "", err
	}

	//use type from database??
	// passkeySession, err := s.database.CreatePasskeySession(database.PasskeySessionInsert{
	// 	User:                 userID,
	// 	Challenge:            session.Challenge,
	// 	AllowedCredentialIDs: session.AllowedCredentialIDs,
	// 	Expires:              session.Expires,
	// 	UserVerification:     session.UserVerification,
	// 	Extensions:           session.Extensions,
	// })

	passkeySessionInsert := WebAuthnSessionToPasskeySession(session, userID)
	createdSession, err := s.database.CreatePasskeySession(passkeySessionInsert)
	if err != nil {
		return "", err
	}

	fmt.Print("createdSession", createdSession)

	return "Created Session", err // TODO
}

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

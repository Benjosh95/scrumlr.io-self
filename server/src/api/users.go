package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/render"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"scrumlr.io/server/common"
	"scrumlr.io/server/common/dto"
	"scrumlr.io/server/logger"
)

// getUser get a user
func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("User").(uuid.UUID)

	user, err := s.users.Get(r.Context(), userId)
	if err != nil {
		common.Throw(w, r, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.Respond(w, r, user)
}

func (s *Server) updateUser(w http.ResponseWriter, r *http.Request) {
	log := logger.FromRequest(r)

	user := r.Context().Value("User").(uuid.UUID)

	var body dto.UserUpdateRequest
	if err := render.Decode(r, &body); err != nil {
		common.Throw(w, r, common.BadRequestError(err))
		return
	}

	body.ID = user

	updatedUser, err := s.users.Update(r.Context(), body)
	if err != nil {
		log.Errorw("failed to update user", "err", err)
		common.Throw(w, r, common.InternalServerError)
		return
	}

	render.Status(r, http.StatusOK)
	render.Respond(w, r, updatedUser)
}

//Passkeys functionality

var (
	globalSession_Reg *webauthn.SessionData
	globalSession_Log *webauthn.SessionData
)

func (s *Server) generateRegistrationOptions(w http.ResponseWriter, r *http.Request) {
	// TODO: Loging
	// log := logger.FromRequest(r)

	userId := r.Context().Value("User").(uuid.UUID)
	user, err := s.users.Get(r.Context(), userId)
	if err != nil {
		common.Throw(w, r, err)
		return
	}

	options, session, err := s.webAuthn.BeginRegistration(user)
	if err != nil {
		common.Throw(w, r, err)
		// http.Error(w, "Error generating registration options", http.StatusInternalServerError)
		return
	}

	// store the session data in your data store or session management system
	globalSession_Reg = session
	str, err := s.passkeys.CreateSession(r.Context(), session)
	if err != nil {
		common.Throw(w, r, err)
		// http.Error(w, "Error generating registration options", http.StatusInternalServerError)
		return
	}
	fmt.Print("str", str)

	//return the generated registration options
	render.JSON(w, r, options)
}

func (s *Server) verifyRegistration(w http.ResponseWriter, r *http.Request) {
	log := logger.FromRequest(r)

	userId := r.Context().Value("User").(uuid.UUID)
	user, err := s.users.Get(r.Context(), userId)
	if err != nil {
		common.Throw(w, r, err)
		return
	}

	// Get the session data stored from the function above
	session, err := s.passkeys.GetSession(r.Context(), userId) //is a user always only having one session at once???
	// session := *globalSession_Reg
	fmt.Printf("session-------\nChallenge: %s\nUserID: %v\nAllowedCredentialIDs: %v\nExpires: %s\nUserVerification: %s\nExtensions: %v\n",
		session.Challenge, session.UserID, session.AllowedCredentialIDs, session.Expires, session.UserVerification, session.Extensions)

	credential, err := s.webAuthn.FinishRegistration(user, *session, r)
	if err != nil {
		common.Throw(w, r, err)
		// http.Error(w, "Error generating registration options", http.StatusInternalServerError)
		return
	}

	// Add the new credential to the user's existing credentials
	user.Credentials = append(user.Credentials, *credential)

	// TODO: is it rly done with requests here, and not with the user type directly => could force type conversion
	// Prepare the update request
	updateRequest := dto.UserUpdateRequest{
		ID:          user.ID,
		Name:        user.Name,
		Avatar:      user.Avatar,
		Credentials: user.Credentials,
	}

	// perform the update
	updatedUser, err := s.users.Update(r.Context(), updateRequest)
	if err != nil {
		log.Errorw("failed to update user", "err", err)
		common.Throw(w, r, common.InternalServerError)
		return
	}

	// TODO: return id of "newly" registered passkey
	render.JSON(w, r, updatedUser.Credentials) // Handle next steps
}

func (s *Server) generateAuthenticationOptions(w http.ResponseWriter, r *http.Request) {
	//TODO: logging

	userId := r.Context().Value("User").(uuid.UUID)
	user, err := s.users.Get(r.Context(), userId)
	if err != nil {
		common.Throw(w, r, err)
		return
	}

	options, session, err := s.webAuthn.BeginLogin(user) //user muss credential(s) abfrage unterstützen
	if err != nil {
		// TODO: Handle Error and return.
		return
	}

	// TODO:
	// store the session values
	// datastore.SaveSession(session)
	globalSession_Log = session

	// return the options generated
	// options.publicKey contain our registration options
	render.Status(r, http.StatusOK)
	render.Respond(w, r, options)
}

func (s *Server) verifyAuthentication(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("User").(uuid.UUID)
	user, err := s.users.Get(r.Context(), userId)
	if err != nil {
		common.Throw(w, r, err)
		return
	}

	// TODO: Get the session data stored from the function above
	session := *globalSession_Log //should have two different sessions? for reg and log?

	credential, err := s.webAuthn.FinishLogin(user, session, r)
	if err != nil {
		// Handle Error and return.
		return
	}
	fmt.Print("updated credential------", credential)

	// TODO: Handle credential.Authenticator.CloneWarning ??

	// If login was successful, update the credential object
	// Pseudocode to update the user credential.
	// user.UpdateCredential(credential)
	// datastore.SaveUser(user)

	render.Status(r, http.StatusOK)
	render.Respond(w, r, "Login Success")
}

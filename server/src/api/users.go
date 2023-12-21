package api

import (
	"context"
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

	// userId := r.Context().Value("User").(uuid.UUID)
	// user, err := s.users.Get(r.Context(), userId)
	// if err != nil {
	// 	common.Throw(w, r, err)
	// 	return
	// }

	options, session, err := s.webAuthn.BeginDiscoverableLogin() //user muss credential(s) abfrage unterstützen
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
	// userId := r.Context().Value("User").(uuid.UUID) // Continue
	// user, err := s.users.Get(r.Context(), userId)
	// if err != nil {
	// 	common.Throw(w, r, err)
	// 	return
	// }

	// TODO: Put functionalities into service not in the handler here
	//TODO: extract the rawID and userhandle from webauthn assertionResponse in body
	// var assertionResponse struct {
	// 	RawID      []byte `json:"rawId"`
	// 	UserHandle []byte `json:"userHandle"`
	// }

	// if err := json.NewDecoder(r.Body).Decode(&assertionResponse); err != nil {
	// 	// Handle JSON decoding error
	// 	http.Error(w, "Error decoding JSON", http.StatusBadRequest)
	// 	return
	// }

	// TODO: Get the session data stored from the function above
	session := *globalSession_Log //should have two different sessions? for reg and log?

	// get session to get challenge,
	// get r to extract signature
	// use publickey from db to (decrypt/Unsign) signature and match with challenge from session
	credential, err := s.webAuthn.FinishDiscoverableLogin(s.myDiscoverableUserHandler,
		session,
		r,
	)
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

// CONTINUE: LOOK Blue Post-it
// // TODO: Implement the DiscoverableUserHandler function
// func (s *Server) myDiscoverableUserHandler(rawID, userHandle []byte) (*dto.User, error) {
// 	// Perform a lookup in your data store to find the user based on rawID or userHandle
// 	// Replace this with your actual logic to fetch the user from your data store.

// 	// CONTINUE
// 	// find the user from RawID and UserHandle / which means RawID matches the ID of a Credential which is linked to a user
// 	// and this user is also used to validate the response
// 	// Platzhalterrisch inc.
// 	uuidTempUserString := "1c607741-bec1-4d97-b909-f988a453abdd"
// 	// Parse the string to obtain a UUID
// 	parsedUUID, _ := uuid.Parse(uuidTempUserString)
// 	user, _ := s.users.Get(context.Background(), parsedUUID)

// return user, nil
// Example: Assume you have a function s.users.GetUserByRawID that retrieves a user by rawID.
// user, err := s.users.GetUserByRawID(rawID)
// if err != nil {
// 	return User{}, err
// }

// Alternatively, you can use userHandle to fetch the user.
// user, err := s.users.GetUserByUserHandle(userHandle)
// if err != nil {
//     return User{}, err
// }

// Implement the DiscoverableUserHandler interface
func (s *Server) myDiscoverableUserHandler(rawID, userHandle []byte) (webauthn.User, error) {
	// Find the user based on rawID or userHandle in your data store
	uuidTempUserString := "1c607741-bec1-4d97-b909-f988a453abdd"
	// Parse the string to obtain a UUID
	parsedUUID, _ := uuid.Parse(uuidTempUserString)
	user, err := s.users.Get(context.Background(), parsedUUID)
	if err != nil {
		// Handle the error (e.g., user not found)
		return nil, err
	}

	// Return the user, which should implement the webauthn.User interface
	return user, nil
}

package api

import (
	"context"
	"fmt"
	"math"
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
	render.Status(r, http.StatusOK)
	render.Respond(w, r, options)
}

func (s *Server) verifyAuthentication(w http.ResponseWriter, r *http.Request) {

	// get userhandle to identify user of the passed passkey
	//can this be made shorter/simpler? Maybe get userhandle base64 string directly and decode it.
	// parsedResponse, err := protocol.ParseCredentialRequestResponse(r)
	// if err != nil {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }

	// base64String := base64.StdEncoding.EncodeToString(parsedResponse.Response.UserHandle)
	// decodedBytes, err := base64.StdEncoding.DecodeString(base64String)
	// if err != nil {
	// 	fmt.Println("Error decoding base64:", err)
	// 	return
	// }
	// userIdString := string(decodedBytes)  //Refactor
	// userId, _ := uuid.Parse(userIdString) //Refactor

	//test Temp alternative
	userId, _ := uuid.Parse("1c607741-bec1-4d97-b909-f988a453abdd")

	//Get User of the assertionResponseRequest
	user, err := s.users.Get(r.Context(), userId)
	if err != nil {
		common.Throw(w, r, err)
		return
	}
	fmt.Print("Retrieved_user", user)

	// TODO: Get the session data stored from the function above
	session := *globalSession_Log //should have two different sessions? for reg and log? or just one which gets overwritten?
	fmt.Print("session = ", session)

	// pass handler to retrieve correct user from db with matching credentialID and RawID
	// pass session to get challenge,
	// pass r to extract signature
	// use publickey from db to (decrypt/Unsign) signature and match with challenge from passed session
	credential, err := s.webAuthn.FinishDiscoverableLogin(s.discoverableUserHandler,
		session,
		r,
	)
	if err != nil {
		fmt.Print("FinishDiscoverableLogin ERROR:  = ", err) // only modified at changes?
		return
	}

	//JWT
	tokenString, err := s.auth.Sign(map[string]interface{}{"id": user.ID})
	if err != nil {
		// log.Errorw("unable to generate token string", "err", err)
		fmt.Print("meeh")
		common.Throw(w, r, common.InternalServerError)
		return
	}

	fmt.Print("tokenString", tokenString)

	cookie := http.Cookie{Name: "jwt", Value: tokenString, Path: "/", HttpOnly: true, MaxAge: math.MaxInt32}
	common.SealCookie(r, &cookie)
	http.SetCookie(w, &cookie)

	// TODO: Handle credential.Authenticator.CloneWarning ??

	// TODO:
	// If login was successful, update the credential object
	// user.UpdateCredential(credential)
	// datastore.SaveUser(user)
	fmt.Print("updated credential = ", credential) // only modified at changes?

	render.Status(r, http.StatusOK)
	render.Respond(w, r, user)
	// render.Respond(w, r, "Login Success")
}

// Implementing the DiscoverableUserHandler interface
func (s *Server) discoverableUserHandler(rawID, userHandle []byte) (webauthn.User, error) {
	// Perform a lookup in your data store to find the user based on rawID or userHandle
	// Replace this with your actual logic to fetch the user from your data store.

	// find the user from RawID and UserHandle / which means RawID matches the ID of a Credential which is linked to a user
	// and this user is also used to validate the response

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

	// TEMPORARY SOLUTION
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

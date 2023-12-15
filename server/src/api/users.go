package api

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/go-chi/render"
	"github.com/go-webauthn/webauthn/protocol"
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

// ///////////// Wtf ”#####
// Your existing MockUser struct
// Your existing MockUser struct
type MockUser struct {
	ID          []byte
	Name        string
	DisplayName string
	// Add other fields as needed
	Credentials []webauthn.Credential
}

// MockUserImpl implements the webauthn.User interface
type MockUserImpl struct {
	MockUser
}

// WebAuthnIcon implements the WebAuthnIcon method of the webauthn.User interface
func (u *MockUserImpl) WebAuthnIcon() string {
	// You can return an empty string or a default value here
	return ""
}

// WebAuthnID implements the WebAuthnID method of the webauthn.User interface
func (u *MockUserImpl) WebAuthnID() []byte {
	return u.ID
}

// WebAuthnName implements the WebAuthnName method of the webauthn.User interface
func (u *MockUserImpl) WebAuthnName() string {
	return u.Name
}

// WebAuthnDisplayName implements the WebAuthnDisplayName method of the webauthn.User interface
func (u *MockUserImpl) WebAuthnDisplayName() string {
	return u.DisplayName
}

// WebAuthnCredentials implements the WebAuthnCredentials method of the webauthn.User interface
func (u *MockUserImpl) WebAuthnCredentials() []webauthn.Credential {
	// Implement the logic to retrieve or create credentials for the user
	// Return a slice of webauthn.Credential
	return u.Credentials
}

func (u *MockUserImpl) AddCredential(credential *webauthn.Credential) {
	u.Credentials = append(u.Credentials, *credential)
}

func (u *MockUserImpl) UpdateCredential(credential *webauthn.Credential) {
	for idx, elem := range u.Credentials {
		if reflect.DeepEqual(elem.ID, credential.ID) {
			u.Credentials[idx] = *credential
			return
		}
	}
}

// CreationResponse represents the response containing both options and session data
type CreationResponse struct {
	Options *protocol.CredentialCreation
	Session *webauthn.SessionData
}

var (
	globalUser        *MockUserImpl
	globalSession_Reg *webauthn.SessionData
	globalSession_Log *webauthn.SessionData
)

func (s *Server) getCreationOptions(w http.ResponseWriter, r *http.Request) {
	// Mock User here
	mockUser := &MockUser{
		ID:          []byte("mockUserID"),
		Name:        "Mock User",
		DisplayName: "Mock Display Name",
		// Add other fields as needed
		Credentials: make([]webauthn.Credential, 0),
	}

	user := &MockUserImpl{MockUser: *mockUser}

	options, session, err := s.webAuthnInstance.BeginRegistration(user)
	if err != nil {
		// handle error
		http.Error(w, "Error generating registration options", http.StatusInternalServerError)
		return
	}

	// Create a struct to hold both options and session
	response := CreationResponse{
		Options: options,
		Session: session,
	}

	// Store the sessionData values (and user)
	// You may want to store the session data in your data store or session management system
	globalSession_Reg = session
	globalUser = user
	// Return the response as JSON
	render.JSON(w, r, response)
}

func (s *Server) finishRegistration(w http.ResponseWriter, r *http.Request) {
	// user := datastore.GetUser() // Get the user
	user := globalUser
	// Get the session data stored from the function above
	// session := datastore.GetSession()
	session := *globalSession_Reg

	credential, err := s.webAuthnInstance.FinishRegistration(user, session, r) // wie muss request (body) aussehen
	if err != nil {
		// Handle Error and return.

		return
	}

	// If creation was successful, store the credential object
	// Pseudocode to add the user credential.
	user.AddCredential(credential)
	// datastore.SaveUser(user)
	fmt.Print(credential)

	// JSONResponse(w, "Registration Success", http.StatusOK) // Handle next steps
	render.JSON(w, r, http.StatusOK) // Handle next steps
}

func (s *Server) beginLogin(w http.ResponseWriter, r *http.Request) {
	user := globalUser // Find the user

	options, session, err := s.webAuthnInstance.BeginLogin(user) //user muss credential(s) abfrage unterstützen
	if err != nil {
		// Handle Error and return.
		fmt.Print(err)
		fmt.Print("uwe")
		return
	}

	// store the session values
	// datastore.SaveSession(session)
	globalSession_Log = session

	// return the options generated
	// options.publicKey contain our registration options
	render.Status(r, http.StatusOK)
	render.Respond(w, r, options)
}

func (s *Server) finishLogin(w http.ResponseWriter, r *http.Request) {
	// user := datastore.GetUser() // Get the user
	user := globalUser

	// Get the session data stored from the function above
	session := *globalSession_Log //should have two different sessions? for reg and log?

	credential, err := s.webAuthnInstance.FinishLogin(user, session, r)
	if err != nil {
		// Handle Error and return.

		return
	}

	// Handle credential.Authenticator.CloneWarning

	// If login was successful, update the credential object
	// Pseudocode to update the user credential.
	user.UpdateCredential(credential)
	// datastore.SaveUser(user)

	render.Status(r, http.StatusOK)
	render.Respond(w, r, "Login Success")
}

package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"scrumlr.io/server/common"
	"scrumlr.io/server/logger"

	"github.com/go-chi/render"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/markbates/goth/gothic"
	"scrumlr.io/server/common/dto"
	"scrumlr.io/server/database/types"
)

// AnonymousSignUpRequest represents the request to create a new anonymous user.
type AnonymousSignUpRequest struct {
	// The display name of the user.
	Name string
}

// AuthVerificationRequest helps to access the userhandle within the AssertionResponse
type AuthVerificationRequest struct {
	Response struct {
		UserHandle string `json:"userhandle"`
	} `json:"response"`
}

// signInAnonymously create a new anonymous user
func (s *Server) signInAnonymously(w http.ResponseWriter, r *http.Request) {
	log := logger.FromRequest(r)

	var body AnonymousSignUpRequest
	if err := render.Decode(r, &body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := s.users.LoginAnonymous(r.Context(), body.Name)
	if err != nil {
		log.Errorw("could not create user", "req", body, "err", err)
		common.Throw(w, r, common.InternalServerError)
		return
	}

	tokenString, err := s.auth.Sign(map[string]interface{}{"id": user.ID})
	if err != nil {
		log.Errorw("unable to generate token string", "err", err)
		common.Throw(w, r, common.InternalServerError)
		return
	}

	cookie := http.Cookie{Name: "jwt", Value: tokenString, Path: "/", HttpOnly: true, MaxAge: math.MaxInt32}
	common.SealCookie(r, &cookie)
	http.SetCookie(w, &cookie)

	render.Status(r, http.StatusCreated)
	render.Respond(w, r, user)
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{Name: "jwt", Value: "deleted", Path: "/", MaxAge: -1, Expires: time.UnixMilli(0)}
	common.SealCookie(r, &cookie)
	http.SetCookie(w, &cookie)

	if common.GetHostWithoutPort(r) != common.GetTopLevelHost(r) {
		cookieWithSubdomain := http.Cookie{Name: "jwt", Value: "deleted", Path: "/", MaxAge: -1, Expires: time.UnixMilli(0)}
		common.SealCookie(r, &cookieWithSubdomain)
		cookieWithSubdomain.Domain = common.GetHostWithoutPort(r)
		http.SetCookie(w, &cookieWithSubdomain)
	}

	render.Status(r, http.StatusNoContent)
	render.Respond(w, r, nil)
}

// beginAuthProviderVerification will redirect the user to the specified auth provider consent page
func (s *Server) beginAuthProviderVerification(w http.ResponseWriter, r *http.Request) {
	gothic.BeginAuthHandler(w, r)
}

// verifyAuthProviderCallback will verify the auth provider call, create or update a user and redirect to the page provider with the state
func (s *Server) verifyAuthProviderCallback(w http.ResponseWriter, r *http.Request) {
	externalUser, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	provider := strings.ToUpper(externalUser.Provider)
	var internalUser *dto.User
	switch provider {
	case (string)(types.AccountTypeGoogle):
		internalUser, err = s.users.CreateGoogleUser(r.Context(), externalUser.UserID, externalUser.NickName, externalUser.AvatarURL)
	case (string)(types.AccountTypeGitHub):
		internalUser, err = s.users.CreateGitHubUser(r.Context(), externalUser.UserID, externalUser.NickName, externalUser.AvatarURL)
	case (string)(types.AccountTypeMicrosoft):
		internalUser, err = s.users.CreateMicrosoftUser(r.Context(), externalUser.UserID, externalUser.NickName, externalUser.AvatarURL)
	case (string)(types.AccountTypeAzureAd):
		internalUser, err = s.users.CreateAzureAdUser(r.Context(), externalUser.UserID, externalUser.NickName, externalUser.AvatarURL)
	case (string)(types.AccountTypeApple):
		internalUser, err = s.users.CreateAppleUser(r.Context(), externalUser.UserID, externalUser.NickName, externalUser.AvatarURL)
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tokenString, _ := s.auth.Sign(map[string]interface{}{"id": internalUser.ID})
	cookie := http.Cookie{Name: "jwt", Value: tokenString, Path: "/", Expires: time.Now().AddDate(0, 0, 3*7)}
	common.SealCookie(r, &cookie)
	http.SetCookie(w, &cookie)

	state := gothic.GetState(r)
	stateSplit := strings.Split(state, "__")
	if len(stateSplit) > 1 {
		w.Header().Set("Location", stateSplit[1])
		w.WriteHeader(http.StatusSeeOther)
	}
	if s.basePath == "/" {
		w.Header().Set("Location", fmt.Sprintf("%s://%s/", common.GetProtocol(r), r.Host))
	} else {
		w.Header().Set("Location", fmt.Sprintf("%s://%s%s/", common.GetProtocol(r), r.Host, s.basePath))
	}
	w.WriteHeader(http.StatusSeeOther)
}

// hold login session in  memory
var globalSession_Log *webauthn.SessionData

func (s *Server) generateRegistrationOptions(w http.ResponseWriter, r *http.Request) {
	log := logger.FromRequest(r)

	userId := r.Context().Value("User").(uuid.UUID)
	user, err := s.users.Get(r.Context(), userId)
	if err != nil {
		log.Errorw("failed to get user", "err", err)
		common.Throw(w, r, common.InternalServerError)
		return
	}

	// excludeCredentials so a user only can have a single passkey on its authenticator for that Account, ExcludeCredentials = ExistingCredentials
	var excludeCredentials []protocol.CredentialDescriptor
	for _, credential := range user.Credentials {
		excludeCredentials = append(excludeCredentials, protocol.CredentialDescriptor{
			CredentialID: []byte(credential.ID),
			Type:         protocol.PublicKeyCredentialType,
		})
	}

	// requireResidentKey = true, residentKey = required to enable Conditional UI => Passkeys
	authSelect := protocol.AuthenticatorSelection{
		RequireResidentKey: protocol.ResidentKeyRequired(),
		ResidentKey:        protocol.ResidentKeyRequirementRequired,
		UserVerification:   protocol.VerificationPreferred,
	}

	// generates registrationOptions for Browser API navigator.credentials.create(X)
	options, session, err := s.webAuthn.BeginRegistration(
		user,
		webauthn.WithAuthenticatorSelection(authSelect),
		webauthn.WithExclusions(excludeCredentials),
	)
	if err != nil {
		log.Errorw("failed to generate registration options", "err", err)
		common.Throw(w, r, common.InternalServerError)
		return
	}

	// stores session data for verification
	err = s.passkeys.CreateSession(r.Context(), session)
	if err != nil {
		log.Errorw("failed to create passkey session", "err", err)
		common.Throw(w, r, common.InternalServerError)
		return
	}

	render.JSON(w, r, options)
}

func (s *Server) verifyRegistration(w http.ResponseWriter, r *http.Request) {
	log := logger.FromRequest(r)

	userId := r.Context().Value("User").(uuid.UUID)
	user, err := s.users.Get(r.Context(), userId)
	if err != nil {
		log.Errorw("failed to get user", "err", err)
		common.Throw(w, r, common.InternalServerError)
		return
	}

	// gets the session data for verification
	session, err := s.passkeys.GetSession(r.Context(), userId)
	if err != nil {
		log.Errorw("failed to get passkey session", "err", err)
		common.Throw(w, r, common.InternalServerError)
		return
	}

	// does the Verification
	credential, err := s.webAuthn.FinishRegistration(user, *session, r)
	if err != nil {
		log.Errorw("failed to finish registration", "err", err)
		common.Throw(w, r, common.InternalServerError)
		return
	}

	// add the new credential to the user's existing credentials
	user.Credentials = append(user.Credentials, dto.ConvertToExtendedCredential(*credential))
	updateRequest := dto.UserUpdateRequest{
		ID:          user.ID,
		Name:        user.Name,
		Avatar:      user.Avatar,
		Credentials: user.Credentials,
	}
	updatedUser, err := s.users.Update(r.Context(), updateRequest)
	if err != nil {
		log.Errorw("failed to update user", "err", err)
		common.Throw(w, r, common.InternalServerError)
		return
	}

	render.JSON(w, r, updatedUser.Credentials)
}

func (s *Server) generateAuthenticationOptions(w http.ResponseWriter, r *http.Request) {
	log := logger.FromRequest(r)

	// generates the authentication options and session data including the challenge
	options, session, err := s.webAuthn.BeginDiscoverableLogin()
	if err != nil {
		log.Errorw("failed to generate authentication options and session", "err", err)
		common.Throw(w, r, common.InternalServerError)
		return
	}

	// TODO: Auth Session wird im hauptspeicher gehalten. Gibt es alternativen oder Ok so?
	globalSession_Log = session

	render.JSON(w, r, options)
}

func (s *Server) verifyAuthentication(w http.ResponseWriter, r *http.Request) {
	log := logger.FromRequest(r)

	// https://blog.flexicondev.com/read-go-http-request-body-multiple-times
	// reading the body and assigning it back allows multiple reads of request.body.
	// (r.body of type io.ReadCloser can normaly be read only once)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	parsedResponse, err := protocol.ParseCredentialRequestResponseBody(r.Body)
	if err != nil {
		log.Errorw("failed to parse assertion response", "err", err)
		common.Throw(w, r, common.InternalServerError)
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	userHandle := parsedResponse.Response.UserHandle
	userId, _ := uuid.Parse(string(userHandle))

	user, err := s.users.Get(r.Context(), userId)
	if err != nil {
		log.Errorw("failed to get user", "err", err)
		common.Throw(w, r, common.InternalServerError)
		return
	}

	// Todo: Muss im Hauptspeicher gehalten werden? -> NOPE muss über sessionID gelöst werden
	session := *globalSession_Log
	// r.Cookie()

	// pass handler which is retrieving correct user from db by userHandle or rawId
	// pass session to extract challenge,
	// pass r to extract signature from r.body.response.signature
	// get and use publickey from db to (decrypt/Unsign) signature and match with challenge from passed session
	credential, err := s.webAuthn.FinishDiscoverableLogin(s.discoverableUserHandler,
		session,
		r,
	)

	// Todo: can be done smarter
	if err != nil {
		log.Errorw("failed to finish login", "err", err)

		switch err.Error() {
		case "Unable to find the credential for the returned credential ID":
			common.Throw(w, r, common.NotFoundError)
		default:
			common.Throw(w, r, common.InternalServerError)
		}
		return
	}

	// generates jwt
	tokenString, err := s.auth.Sign(map[string]interface{}{"id": user.ID})
	if err != nil {
		log.Errorw("unable to generate token string", "err", err)
		common.Throw(w, r, common.InternalServerError)
		return
	}
	// sets the jwt
	cookie := http.Cookie{Name: "jwt", Value: tokenString, Path: "/", HttpOnly: true, MaxAge: math.MaxInt32}
	common.SealCookie(r, &cookie)
	http.SetCookie(w, &cookie)

	// TODO: Handle credential.Authenticator.CloneWarning ??

	// updates "lastUsedAt" attribute of users Credential that was used to log in successfully
	for i, userCredential := range user.Credentials {
		if bytes.Equal(userCredential.ID, credential.ID) {
			user.Credentials[i].LastUsedAt = time.Now()
			break
		}
	}
	updateRequest := dto.UserUpdateRequest{
		ID:          user.ID,
		Name:        user.Name,
		Avatar:      user.Avatar,
		Credentials: user.Credentials,
	}
	updatedUser, err := s.users.Update(r.Context(), updateRequest)
	if err != nil {
		log.Errorw("failed to update user", "err", err)
		common.Throw(w, r, common.InternalServerError)
		return
	}

	render.JSON(w, r, updatedUser)
}

// Performs a lookup in the db to find the user based on userHandle (=userId) or alternatively by rawID (=CredentialId)
func (s *Server) discoverableUserHandler(rawID, userHandle []byte) (webauthn.User, error) {
	userId, err := uuid.Parse(string(userHandle))
	if err != nil {
		return nil, err
	}

	user, err := s.users.Get(context.Background(), userId)
	if err != nil {
		return nil, err
	}

	return user, nil
}

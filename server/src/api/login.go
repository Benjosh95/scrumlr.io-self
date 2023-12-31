package api

import (
	"bytes"
	"context"
	"fmt"
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

	// get userhandle to identify user of the passed passkey
	// can this be made shorter/simpler? Maybe get userhandle base64 string directly and decode it.
	// parsedResponse, err := protocol.ParseCredentialRequestResponse(r)
	// if err != nil {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }
	// fmt.Print("parsedResponse of r = ", parsedResponse.Response)
	// base64String := base64.StdEncoding.EncodeToString(parsedResponse.Response.UserHandle)
	// decodedBytes, err := base64.StdEncoding.DecodeString(base64String)
	// if err != nil {
	// 	fmt.Println("Error decoding base64:", err)
	// 	return
	// }
	// userIdString := string(decodedBytes)  //Refactor
	// userId, _ := uuid.Parse(userIdString) //Refactor

	//test Temp alternative
	//CONTINUE userid ist vom Userhandle zu entnehmen und nicht hardcoded wie hier
	//Beim extraheieren vom userhandle und convertieren zwischen []byte, base64 und uuid
	//dabei gab es fehler die sich aber erst im r http.Request bei der finishdiscoverableLogin function gezeigt haben.
	//vermutlich beim Code hierüber irg was anders machen, weil so hardcoded funktioniert es.

	// var requestBody AuthVerificationRequest
	// if err := render.Decode(r, &requestBody); err != nil {
	// 	common.Throw(w, r, common.BadRequestError(err))
	// 	return
	// }
	// decodedBytes, err := base64.StdEncoding.DecodeString(requestBody.Response.UserHandle)
	// if err != nil {
	// 	fmt.Println("Error decoding base64:", err)
	// 	return
	// }
	// userId, err := uuid.Parse(string(decodedBytes))
	// if err != nil {
	// 	fmt.Println("Error converting decoded bytes to uuid:", err)
	// 	return
	// }

	//Hard coded workaround for code above. I just dont know why "parse error on assetion" happens in finishdiscoverableLogin function...
	userId, _ := uuid.Parse("0215ad0e-e57f-4173-81af-ffb79719af65")

	//Get User of the assertionResponseRequest
	user, err := s.users.Get(r.Context(), userId)
	if err != nil {
		log.Errorw("failed to get user", "err", err)
		common.Throw(w, r, common.InternalServerError)
		return
	}
	fmt.Print("Retrieved_user", user)

	// Muss im Hauptspeicher gehalten werden.
	session := *globalSession_Log
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

	//generates JWT
	tokenString, err := s.auth.Sign(map[string]interface{}{"id": user.ID})
	if err != nil {
		log.Errorw("unable to generate token string", "err", err)
		common.Throw(w, r, common.InternalServerError)
		return
	}

	cookie := http.Cookie{Name: "jwt", Value: tokenString, Path: "/", HttpOnly: true, MaxAge: math.MaxInt32}
	common.SealCookie(r, &cookie)
	http.SetCookie(w, &cookie)

	// TODO: Handle credential.Authenticator.CloneWarning ??

	// modifies "lastUsedAt" attribute of Credential which was used to log in successfully
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

	// updates the user with the modified credentials
	updatedUser, err := s.users.Update(r.Context(), updateRequest)
	if err != nil {
		log.Errorw("failed to update user", "err", err)
		common.Throw(w, r, common.InternalServerError)
		return
	}

	render.JSON(w, r, updatedUser)
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
	uuidTempUserString := "0215ad0e-e57f-4173-81af-ffb79719af65"
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

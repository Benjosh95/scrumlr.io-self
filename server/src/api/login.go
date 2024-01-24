package api

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"scrumlr.io/server/common"
	"scrumlr.io/server/logger"

	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/markbates/goth/gothic"
	"scrumlr.io/server/common/dto"
	"scrumlr.io/server/database/types"
)

// AnonymousSignUpRequest represents the request to create a new anonymous user.
type AnonymousSignUpRequest struct {
	// The display name of the user.
	Name string
}

type PasskeySignInResponse struct {
	Token string `json:"token"`
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

// TODO: document it
func (s *Server) signInWithPasskey(w http.ResponseWriter, r *http.Request) {
	// take assertionResponse from request and finalize login with it and the pk_api and store the returned token
	tenantId := "64284d4b-750b-4c6b-a809-9601c6cd6ae4"
	url := fmt.Sprintf("https://passkeys.hanko.io/%s/login/finalize", tenantId)

	req, _ := http.NewRequest("POST", url, r.Body)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer res.Body.Close()

	var pkSignInResponse PasskeySignInResponse
	if err := json.NewDecoder(res.Body).Decode(&pkSignInResponse); err != nil {
		fmt.Println("Error decoding response:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	hankoToken := pkSignInResponse.Token

	// call hanko pk_api to get jwks, verify hankotoken with it and extract the sub in the userId
	jwksURL := fmt.Sprintf("https://passkeys.hanko.io/%s/.well-known/jwks.json", tenantId)

	JWKS, err := jwk.Fetch(r.Context(), jwksURL)
	if err != nil {
		fmt.Print("err:", err)
	}

	// under the hood: extract kId from header of jwt/hankoToken -> take pubKey from jwks with matching kId -> verify jwt with it.
	parsedToken, err := jwt.Parse([]byte(hankoToken), jwt.WithKeySet(JWKS))
	if err != nil {
		fmt.Println("Token parsing failed:", err)
		return
	}

	sub, success := parsedToken.Get("sub")
	if !success {
		fmt.Println("No sub field in token:", err)
		return
	}

	// get user by sub/userId of parsed hankoToken
	userId, _ := uuid.Parse(sub.(string))
	user, err := s.users.Get(r.Context(), userId)
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	//create/sign token with userId and set it for the client in its client
	tokenString, err := s.auth.Sign(map[string]interface{}{"id": user.ID})
	if err != nil {
		fmt.Println("unable to generate token string:", err)
		return
	}

	cookie := http.Cookie{Name: "jwt", Value: tokenString, Path: "/", HttpOnly: true, MaxAge: math.MaxInt32}
	common.SealCookie(r, &cookie)
	http.SetCookie(w, &cookie)

	//return user reference in hankoToken
	render.Status(r, http.StatusOK)
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

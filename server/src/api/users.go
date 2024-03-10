package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/render"
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

type createAssociationTokenResponse struct {
	HTTPStatusCode int    `json:"httpStatusCode"`
	Message        string `json:"message"`
	RequestData    struct {
		RequestID string `json:"requestID"`
		Link      string `json:"link"`
	} `json:"requestData"`
	Runtime float64 `json:"runtime"`
	Data    struct {
		Token           string `json:"token"`
		RejectionReason string `json:"rejectionReason"`
	} `json:"data"`
}

func (s *Server) getAssociationToken(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("User").(uuid.UUID)
	user, err := s.users.Get(r.Context(), userId)
	if err != nil {
		common.Throw(w, r, err)
		return
	}

	payload := map[string]interface{}{
		"loginIdentifier":     user.Name,
		"loginIdentifierType": "email",
		"requestID":           "req-557...663",
		"clientInfo": map[string]interface{}{
			"remoteAddress": "::ffff:172.18.0.1",
			"userAgent":     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36",
		},
	}

	jsonValue, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "https://backendapi.corbado.io/v1/associationTokens", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("pro-1697688654308671815", "corbado1_Tcr7ynGVzoRCU7YwPx6MsYfNXLS9YH")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}

	var response createAssociationTokenResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		fmt.Println("Error decoding response:", err)
		return
	}

	fmt.Printf("response TOKEN of the associationTokens endpoint %v", response.Data.Token)
	token := response.Data.Token
	defer resp.Body.Close()

	render.Status(r, http.StatusOK)
	render.Respond(w, r, token)
}

type WebhookRequest struct {
	ID        string // Unique ID per webhook request
	ProjectID string // Your project ID from the developer panel
	Action    string // The specific action of the webhook request
	Data      struct {
		Username string // Data object dependent on action
	}
}

func (s *Server) userExistsWebhook(w http.ResponseWriter, r *http.Request) {

	var request WebhookRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		fmt.Println("failed while decoding request: ", err)
		return
	}
	fmt.Printf("actionType: %s", request.Action)
	// switch request.Action {
	// case "authMethods":
	// 	user, _ := s.users.GetByUsername(r.Context(), request.Data.Username)
	// 	request.Data.Username
	// }
	// s.corbado.CorbadoSDK.Projects().AuthMethodsList()
}

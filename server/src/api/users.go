package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

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

const tenantID = "64284d4b-750b-4c6b-a809-9601c6cd6ae4"
const apiKey = "81EX6eV-rysIp2t7m8ZxYoJxNf2oJ9W2e5w_TW84qOJZ55YYWxRCuMj6Xl03BmuU8CFDbiP-yzOTmx_2IgmqWA=="

// Check if environment variables are set, if so, override the hardcoded values
// if envTenantID := os.Getenv("PASSKEY_TENANT_ID"); envTenantID != "" {
// 	tenantID = envTenantID
// }

//	if envAPIKey := os.Getenv("PASSKEY_SECRET_API_KEY"); envAPIKey != "" {
//		apiKey = envAPIKey
//	}
// var decodedSecret, err = base64.StdEncoding.DecodeString(apiKey)

var baseURL = fmt.Sprintf("https://passkeys.hanko.io/%s", tenantID)
var headers = map[string]string{"apiKey": apiKey, "Content-Type": "application/json"}

// TODO: move
func (s *Server) startRegistration(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Base URL: %s\n", baseURL)
	fmt.Printf("Headers: %+v\n", headers)

	userId := r.Context().Value("User").(uuid.UUID)
	user, _ := s.users.Get(r.Context(), userId)

	payload := map[string]string{
		"user_id":  user.ID.String(),
		"username": user.Name,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := http.Post(baseURL+"/registration/initialize", "application/json", strings.NewReader(string(body)))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	// Decode the response
	var creationOptions map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&creationOptions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// creationOptions is an object that can directly be passed to create()
	// (the function that opens the "create passkey" dialog)
	// in the frontend.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(creationOptions)
}

func (s *Server) finalizeRegistration(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}
	// Create a new request with POST method
	req, err := http.NewRequest("POST", baseURL+"/registration/finalize", r.Body)
	if err != nil {
		fmt.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apiKey", apiKey)
	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()
	// Decode the response
	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		fmt.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// The response from the Passkey API contains a JWT (`data.token`).
	// data = map[token:eyJhbGciOiJSUzI1NiIsImtpZCI6IjQzMDk1OTc1LThhN2QtNDJkZS04MTliLTYzNWQ4YWQwMDZkMSIsInR5cCI6IkpXVCJ9.eyJhdWQiOlsibG9jYWxob3N0Il0sImNyZWQiOiJ4Y3VIS25JVGV3QUFiT0RUWXB2azVCT2FWV1JEOWRMWmowaE5VSFYwMHIwIiwiZXhwIjoxNzA0NDU2NTE5LCJpYXQiOjE3MDQ0NTYyMTksInN1YiI6IjE1MWUwYzM3LWNkMTMtNDIxYS1iNjU2LWMwMTRmYzE3YmM1MyJ9.tF2lRA5xGZYHg8eHzkNLstcaygwmm4OVUutkwQowX1Cr0JJBVMUUBY_WGL5V07C5XNeFwEEPaT4D3YFh9t7glGRHsn5S2vifVshS_Q9-nESH5wSO-NkrMdjAz98bj-tepzpXEOaHrX4AXGMgDMF_9mkCXVxRNCjc1J0-URC4bMHG3spewLUD--yDV93OHdzO2xvuantwnzfmHGTrBUb0_g2QfkOrQz3MUgFSBvxvp-KWp8TImhY6Qj4XUe1hp-H7ROBeoX3iMYTWzs5j2fTBdADJBM3Sq5Xst1T1J9iwty-QvrRwOjX0rfEBrf1j5DnvGnjCutxNqKXJF51trWoxjlHJGYssQYP3hjrwjwR_k139SDq-rjg9ciXX_5xtDG3JZViOTE0E8Df0bMwElm2EjYwQM5UjOCyHDMuPQOXa2PFD6T8auqmGIAVtQNKWXa4YdjgHhPIuovGk4uP-Ph1SCk_m0Krf6YDEuNRPsuIWu-JwhaLQ2SwE-f97e-v3NiVSP0CMRVJP_PjUR9SHljnOAY_OylK9r28Pl7cxrt62UXtwQmEYoy4oBIHEAq8oTYUpvm7nvBRzWNyVJo11LsTJ_csClWOihSwcBJIwjpWHz_K21F5tH1tmcPBxXIt9DBsrAcBiGPK703CH9HAeiBluNz2rd6VikMzkaAcWlbtFsXI]
	// What you do with this JWT is up to you.

	// Here, we don't need to do anything with it, since the user already
	// is logged in. In the login endpoints, later in this guide, we'll
	// use the data contained in the JWT to create a session for our user.

	//TODO Set the created JWT in cookie for client and redirect to /new? or do this clientside

	// Redirect to /success (replace this with your actual success page)
	// http.Redirect(w, r, "/success", http.StatusFound)
	
	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]interface{}{
		"status":  "success",
		"message": "Data received successfully",
		"token":   data["token"],
	})
}

func printStructFields(s interface{}) {
	v := reflect.ValueOf(s)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fieldName := t.Field(i).Name
		fieldValue := v.Field(i).Interface()

		fmt.Printf("%s: %v\n", fieldName, fieldValue)
	}
}

func printRequestBody(r io.ReadCloser) {
	defer r.Close() // Make sure to close the body when you're done with it

	// Read the content of the body into a byte slice
	bodyBytes, err := io.ReadAll(r)
	if err != nil {
		fmt.Println("Error reading request body:", err)
		return
	}

	// Convert the byte slice to a string and print
	bodyString := string(bodyBytes)
	fmt.Println("Request Body:", bodyString)
}

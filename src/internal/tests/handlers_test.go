package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"scti/internal/models"
	"scti/internal/utilities"

	"github.com/stretchr/testify/assert"
)

func (s *APISuite) RegisterAndLogin(uid string) (string, string) {
	email := fmt.Sprintf("user_%s@example.com", uid)
	password := "testpassword123"
	name := fmt.Sprintf("TestName_%s", uid)
	lastName := "TestLast"

	registerReq := models.UserRegister{
		Email:    email,
		Password: password,
		Name:     name,
		LastName: lastName,
	}

	s.Run("Register User", func() {
		code, resp := s.request(http.MethodPost, "/register", registerReq)
		assert.Equal(s.T(), http.StatusCreated, code)
		assert.True(s.T(), resp.Success)
		assert.NotNil(s.T(), resp.Data)

		data := resp.Data.(map[string]interface{})
		assert.NotEmpty(s.T(), data["access_token"])
		assert.NotEmpty(s.T(), data["refresh_token"])
	})

	var userAccessToken, userRefreshToken string
	s.Run("Login with same user", func() {
		loginReq := models.UserLogin{
			Email:    email,
			Password: password,
		}

		code, resp := s.request(http.MethodPost, "/login", loginReq)
		s.assertSuccess(code, resp)

		data := resp.Data.(map[string]interface{})
		assert.NotEmpty(s.T(), data["access_token"])
		userAccessToken = data["access_token"].(string)
		assert.NotEmpty(s.T(), data["refresh_token"])
		userRefreshToken = data["refresh_token"].(string)
	})
	return userAccessToken, userRefreshToken
}

func (s *APISuite) VerifyTokens(userAccessToken, userRefreshToken string) {
	s.Run("Verify user's tokens", func() {
		req := httptest.NewRequest(http.MethodPost, "/verify-tokens", nil)
		req.Header.Set("Authorization", "Bearer "+userAccessToken)
		req.Header.Set("Refresh", "Bearer "+userRefreshToken)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		assert.Equal(s.T(), http.StatusOK, w.Code)
		var resp utilities.Response
		_ = json.NewDecoder(w.Body).Decode(&resp)
		assert.True(s.T(), resp.Success)
		assert.Empty(s.T(), resp.Errors)
	})
}

func (s *APISuite) Logout(userAccessToken, userRefreshToken string) {
	s.Run("Logout user", func() {
		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		req.Header.Set("Authorization", "Bearer "+userAccessToken)
		req.Header.Set("Refresh", "Bearer "+userRefreshToken)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		assert.Equal(s.T(), http.StatusOK, w.Code)
		var resp utilities.Response
		_ = json.NewDecoder(w.Body).Decode(&resp)
		s.assertSuccess(w.Code, resp)
	})
}

func (s *APISuite) Login(uid string) (string, string) {
	var userAccessToken, userRefreshToken string
	email := fmt.Sprintf("user_%s@example.com", uid)
	password := "testpassword123"

	s.Run("Login", func() {
		loginReq := models.UserLogin{
			Email:    email,
			Password: password,
		}

		code, resp := s.request(http.MethodPost, "/login", loginReq)
		s.assertSuccess(code, resp)

		data := resp.Data.(map[string]interface{})
		assert.NotEmpty(s.T(), data["access_token"])
		userAccessToken = data["access_token"].(string)
		assert.NotEmpty(s.T(), data["refresh_token"])
		userRefreshToken = data["refresh_token"].(string)
	})
	return userAccessToken, userRefreshToken
}

func (s *APISuite) LoginEmailPassword(email, password string) (string, string) {
	var userAccessToken, userRefreshToken string

	s.Run("Login", func() {
		loginReq := models.UserLogin{
			Email:    email,
			Password: password,
		}

		code, resp := s.request(http.MethodPost, "/login", loginReq)
		s.assertSuccess(code, resp)

		data := resp.Data.(map[string]interface{})
		assert.NotEmpty(s.T(), data["access_token"])
		userAccessToken = data["access_token"].(string)
		assert.NotEmpty(s.T(), data["refresh_token"])
		userRefreshToken = data["refresh_token"].(string)
	})
	return userAccessToken, userRefreshToken
}

func (s *APISuite) RevokeRefreshToken(userAccessToken, userRefreshToken string) {
	s.Run("Revoke refresh user's token", func() {
		var revokeTokenReq struct {
			RefreshToken string `json:"refresh_token"`
		}

		revokeTokenReq.RefreshToken = userRefreshToken

		body, err := json.Marshal(revokeTokenReq)
		if err != nil {
			assert.True(s.T(), false)
			return
		}

		req := httptest.NewRequest(http.MethodPost, "/revoke-refresh-token", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+userAccessToken)
		req.Header.Set("Refresh", "Bearer "+userRefreshToken)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		assert.Equal(s.T(), http.StatusOK, w.Code)
		var resp utilities.Response
		_ = json.NewDecoder(w.Body).Decode(&resp)
		assert.True(s.T(), resp.Success)
		assert.Empty(s.T(), resp.Errors)
	})
}

func (s *APISuite) GetEvents() {
	s.Run("Get Events", func() {
		req := httptest.NewRequest(http.MethodGet, "/events", nil)
		req.Header.Set("accept", "application/json")

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		assert.Equal(s.T(), http.StatusOK, w.Code)

		var resp utilities.Response
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(s.T(), err)

		assert.True(s.T(), resp.Success)
		assert.Empty(s.T(), resp.Errors)

		assert.NotNil(s.T(), resp.Data)
	})
}

func (s *APISuite) ChangeName(userAccessToken, userRefreshToken, uid string) {
	s.Run("Change name user's", func() {
		var changeNameReq struct {
			Name     string `json:"name"`
			LastName string `json:"last_name"`
		}

		changeNameReq.Name = fmt.Sprintf("TestNameChanged_%s", uid)
		changeNameReq.LastName = "TestLastChanged"

		body, err := json.Marshal(changeNameReq)
		if err != nil {
			assert.True(s.T(), false)
			return
		}

		req := httptest.NewRequest(http.MethodPost, "/change-name", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+userAccessToken)
		req.Header.Set("Refresh", "Bearer "+userRefreshToken)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		assert.Equal(s.T(), http.StatusOK, w.Code)
		var resp utilities.Response
		_ = json.NewDecoder(w.Body).Decode(&resp)
		assert.True(s.T(), resp.Success)
		assert.Empty(s.T(), resp.Errors)
	})
}

func (s *APISuite) SwitchEventCreatorStatus(userAccessToken, userRefreshToken, uid string) {
	s.Run("switch event creator status", func() {
		var SECSReq struct {
			Email     string `json:"email"`
		}

		SECSReq.Email = fmt.Sprintf("user_%s@example.com", uid)

		body, err := json.Marshal(SECSReq)
		if err != nil {
			assert.True(s.T(), false)
			return
		}

		req := httptest.NewRequest(http.MethodPost, "/switch-event-creator-status", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+userAccessToken)
		req.Header.Set("Refresh", "Bearer "+userRefreshToken)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		assert.Equal(s.T(), http.StatusOK, w.Code)
		var resp utilities.Response
		_ = json.NewDecoder(w.Body).Decode(&resp)
		assert.True(s.T(), resp.Success)
		assert.Empty(s.T(), resp.Errors)
	})
}

func (s *APISuite) CreateEvent(userAccessToken, userRefreshToken, slug string) {
	s.Run("Create event", func() {

		eventReq := map[string]interface{}{
			"name":               "Go Workshop",
			"slug":               slug,
			"description":        "Learn Go programming",
			"location":           "Room 101",
			"start_date":         "2025-05-01T14:00:00Z",
			"end_date":           "2025-05-01T17:00:00Z",
			"is_hidden":          false,
			"is_blocked":         false,
			"max_tokens_per_user": 1,
		}

		body, err := json.Marshal(eventReq)
		assert.NoError(s.T(), err)

		req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+userAccessToken)
		req.Header.Set("Refresh", "Bearer "+userRefreshToken)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		assert.Equal(s.T(), http.StatusOK, w.Code)
		var resp utilities.Response
		err = json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(s.T(), err)
		assert.True(s.T(), resp.Success)
		assert.Empty(s.T(), resp.Errors)
	})
}

func (s *APISuite) GetEventsUser(userAccessToken, userRefreshToken string) {
	s.Run("Get events created by a user", func() {
		req := httptest.NewRequest(http.MethodGet, "/events/created", nil)
		req.Header.Set("Authorization", "Bearer "+userAccessToken)
		req.Header.Set("Refresh", "Bearer "+userRefreshToken)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		assert.Equal(s.T(), http.StatusOK, w.Code)
		var resp utilities.Response
		_ = json.NewDecoder(w.Body).Decode(&resp)
		assert.True(s.T(), resp.Success)
		assert.Empty(s.T(), resp.Errors)
	})
}

func (s *APISuite) GetPublicEvents() {
	s.Run("Get Public Events", func() {
		req := httptest.NewRequest(http.MethodGet, "/events/public", nil)
		req.Header.Set("accept", "application/json")

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		assert.Equal(s.T(), http.StatusOK, w.Code)

		var resp utilities.Response
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(s.T(), err)

		assert.True(s.T(), resp.Success)
		assert.Empty(s.T(), resp.Errors)

		assert.NotNil(s.T(), resp.Data)
	})
}

func (s *APISuite) GetEventSlug(slug string) {
	s.Run("Get Event by Slug", func() {
		url := fmt.Sprintf("/events/%s", slug)
		req := httptest.NewRequest(http.MethodGet, url, nil)
		req.Header.Set("accept", "application/json")

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		assert.Equal(s.T(), http.StatusOK, w.Code)

		var resp utilities.Response
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(s.T(), err)

		assert.True(s.T(), resp.Success)
		assert.Empty(s.T(), resp.Errors)

		assert.NotNil(s.T(), resp.Data)
	})
}

func (s *APISuite) DeleteEvent(userAccessToken, userRefreshToken, slug string) {
	s.Run("Delete event", func() {


		url := fmt.Sprintf("/events/%s", slug)
		req := httptest.NewRequest(http.MethodDelete, url, nil)
		req.Header.Set("Authorization", "Bearer "+userAccessToken)
		req.Header.Set("Refresh", "Bearer "+userRefreshToken)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		assert.Equal(s.T(), http.StatusOK, w.Code)

		var resp utilities.Response
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(s.T(), err)
		assert.True(s.T(), resp.Success)
		assert.Empty(s.T(), resp.Errors)
	})
}

func (s *APISuite) UpdateEvent(userAccessToken, userRefreshToken, slug string) {
	s.Run("Create event", func() {

		eventReq := map[string]interface{}{
			"name":               "Go Workshop",
			"slug":               slug,
			"description":        "Learn Go programming UPDATED",
			"location":           "Room 101",
			"start_date":         "2025-05-01T14:00:00Z",
			"end_date":           "2025-05-01T17:00:00Z",
			"is_hidden":          false,
			"is_blocked":         false,
			"max_tokens_per_user": 1,
		}

		body, err := json.Marshal(eventReq)
		assert.NoError(s.T(), err)



		url := fmt.Sprintf("/events/%s", slug)
		req := httptest.NewRequest(http.MethodPatch, url, bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+userAccessToken)
		req.Header.Set("Refresh", "Bearer "+userRefreshToken)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		assert.Equal(s.T(), http.StatusOK, w.Code)
		var resp utilities.Response
		err = json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(s.T(), err)
		assert.True(s.T(), resp.Success)
		assert.Empty(s.T(), resp.Errors)
	})
}

func (s *APISuite) DemoteUserEvent(userAccessToken, userRefreshToken, uid, slug string) {
	s.Run("Demote user in event", func() {
		var DemEvReq struct {
			Email     string `json:"email"`
		}

		DemEvReq.Email = fmt.Sprintf("user_%s@example.com", uid)

		body, err := json.Marshal(DemEvReq)
		if err != nil {
			assert.True(s.T(), false)
			return
		}

		url := fmt.Sprintf("/events/%s/demote", slug)
		req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+userAccessToken)
		req.Header.Set("Refresh", "Bearer "+userRefreshToken)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		assert.Equal(s.T(), http.StatusOK, w.Code)
		var resp utilities.Response
		_ = json.NewDecoder(w.Body).Decode(&resp)
		assert.True(s.T(), resp.Success)
		assert.Empty(s.T(), resp.Errors)
	})
}

func (s *APISuite) PromoteUserEvent(userAccessToken, userRefreshToken, uid, slug string) {
	s.Run("Promote user in event", func() {
		var PromEvReq struct {
			Email     string `json:"email"`
		}

		PromEvReq.Email = fmt.Sprintf("user_%s@example.com", uid)

		body, err := json.Marshal(PromEvReq)
		if err != nil {
			assert.True(s.T(), false)
			return
		}

		url := fmt.Sprintf("/events/%s/promote", slug)
		req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+userAccessToken)
		req.Header.Set("Refresh", "Bearer "+userRefreshToken)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		assert.Equal(s.T(), http.StatusOK, w.Code)
		var resp utilities.Response
		_ = json.NewDecoder(w.Body).Decode(&resp)
		assert.True(s.T(), resp.Success)
		assert.Empty(s.T(), resp.Errors)
	})
}

func (s *APISuite) RegisterUserEvent(userAccessToken, userRefreshToken, slug string) {
	s.Run("Register user in event", func() {
		url := fmt.Sprintf("/events/%s/register", slug)
		req := httptest.NewRequest(http.MethodPost, url, nil)
		req.Header.Set("Authorization", "Bearer "+userAccessToken)
		req.Header.Set("Refresh", "Bearer "+userRefreshToken)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		assert.Equal(s.T(), http.StatusOK, w.Code)
		var resp utilities.Response
		_ = json.NewDecoder(w.Body).Decode(&resp)
		assert.True(s.T(), resp.Success)
		assert.Empty(s.T(), resp.Errors)
	})
}

func (s *APISuite) UnregisterUserEvent(userAccessToken, userRefreshToken, slug string) {
	s.Run("Unregister user in event", func() {
		url := fmt.Sprintf("/events/%s/unregister", slug)
		req := httptest.NewRequest(http.MethodPost, url, nil)
		req.Header.Set("Authorization", "Bearer "+userAccessToken)
		req.Header.Set("Refresh", "Bearer "+userRefreshToken)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		assert.Equal(s.T(), http.StatusOK, w.Code)
		var resp utilities.Response
		_ = json.NewDecoder(w.Body).Decode(&resp)
		assert.True(s.T(), resp.Success)
		assert.Empty(s.T(), resp.Errors)
	})
}

func (s *APISuite) UserEvents(userAccessToken, userRefreshToken string) {
	s.Run("events with this user", func() {
		req := httptest.NewRequest(http.MethodGet, "/user-events", nil)
		req.Header.Set("Authorization", "Bearer "+userAccessToken)
		req.Header.Set("Refresh", "Bearer "+userRefreshToken)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		assert.Equal(s.T(), http.StatusOK, w.Code)
		var resp utilities.Response
		_ = json.NewDecoder(w.Body).Decode(&resp)
		assert.True(s.T(), resp.Success)
		assert.Empty(s.T(), resp.Errors)
	})
}

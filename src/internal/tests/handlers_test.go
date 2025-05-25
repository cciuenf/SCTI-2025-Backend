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

package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"scti/internal/models"
	"scti/internal/utilities"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var (
	UserAccessToken   string
	UserRefreshToken  string
	SuperAccessToken  string
	SuperRefreshToken string
)

func (s *APISuite) testRegisterAndLogin() {
	// Unique ID for test traceability
	uid := uuid.NewString()[:8]
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

	s.Run("Login with same user", func() {
		loginReq := models.UserLogin{
			Email:    email,
			Password: password,
		}

		code, resp := s.request(http.MethodPost, "/login", loginReq)
		s.assertSuccess(code, resp)

		data := resp.Data.(map[string]interface{})
		assert.NotEmpty(s.T(), data["access_token"])
		UserAccessToken = data["access_token"].(string)
		assert.NotEmpty(s.T(), data["refresh_token"])
		UserRefreshToken = data["refresh_token"].(string)
	})
}

func (s *APISuite) testVerifyTokens() {
	s.Run("Verify user's tokens", func() {
		req := httptest.NewRequest(http.MethodPost, "/verify-tokens", nil)
		req.Header.Set("Authorization", "Bearer "+UserAccessToken)
		req.Header.Set("Refresh", "Bearer "+UserRefreshToken)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		assert.Equal(s.T(), http.StatusOK, w.Code)
		var resp utilities.Response
		_ = json.NewDecoder(w.Body).Decode(&resp)
		assert.True(s.T(), resp.Success)
		assert.Empty(s.T(), resp.Errors)
	})
}

func (s *APISuite) testLogout() {
	s.Run("Logout user", func() {
		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		req.Header.Set("Authorization", "Bearer "+UserAccessToken)
		req.Header.Set("Refresh", "Bearer "+UserRefreshToken)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, req)

		assert.Equal(s.T(), http.StatusOK, w.Code)
		var resp utilities.Response
		_ = json.NewDecoder(w.Body).Decode(&resp)
		s.assertSuccess(w.Code, resp)
	})
}

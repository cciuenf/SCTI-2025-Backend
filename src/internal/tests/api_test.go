package handlers_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"scti/config"
	"scti/internal/db"
	"scti/internal/router"
	"scti/internal/utilities"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/google/uuid"
)

type APISuite struct {
	suite.Suite
	router http.Handler
}

func (s *APISuite) SetupSuite() {
	os.Setenv("TEST_MODE", "true")
	cfg := config.LoadConfig("../../.env")
	database := db.Connect(*cfg)
	s.router = router.InitializeMux(database, cfg)
}

func TestAPISuite(t *testing.T) {
	suite.Run(t, new(APISuite))
}

func (s *APISuite) TestUserFlow() {
		// Unique ID for test traceability
	uid := uuid.NewString()[:8]

	var access_token, refresh_token string
	s.Run("1_RegisterAndLogin", func() {
		access_token, refresh_token = s.RegisterAndLogin(uid)
	})
	s.Run("2_VerifyTokens", func() {
		s.VerifyTokens(access_token, refresh_token)
	})
	s.Run("3_RevokeRefreshToken", func() {
		s.RevokeRefreshToken(access_token, refresh_token)
	})
	s.Run("4_Login", func() {
		access_token, refresh_token = s.Login(uid)
	})
	s.Run("4_Logout", func() {
		s.Logout(access_token, refresh_token)
	})
		s.Run("5_GetEvents", func() {
		s.GetEvents()
	})
}

func (s *APISuite) request(method, path string, body any) (int, utilities.Response) {
	var buf io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		buf = bytes.NewBuffer(b)
	}

	req := httptest.NewRequest(method, path, buf)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	var resp utilities.Response
	_ = json.NewDecoder(w.Body).Decode(&resp)

	return w.Code, resp
}

func (s *APISuite) assertSuccess(code int, resp utilities.Response) {
	assert.Equal(s.T(), http.StatusOK, code)
	assert.True(s.T(), resp.Success)
	assert.Empty(s.T(), resp.Errors)
}

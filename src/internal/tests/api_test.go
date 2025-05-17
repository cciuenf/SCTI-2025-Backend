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
)

type APISuite struct {
	suite.Suite
	router http.Handler
}

func (s *APISuite) SetupSuite() {
	os.Setenv("TEST_MODE", "true")
	cfg := config.LoadConfig("../../.env")
	database := db.Connect(*cfg)
	s.router = *router.InitializeMux(database, cfg)
}

func TestAPISuite(t *testing.T) {
	suite.Run(t, new(APISuite))
}

func (s *APISuite) TestUserFlow() {
	s.Run("1_RegisterAndLogin", s.testRegisterAndLogin)
	s.Run("2_VerifyTokens", s.testVerifyTokens)
	s.Run("3_Logout", s.testLogout)
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

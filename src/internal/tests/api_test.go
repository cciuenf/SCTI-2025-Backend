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

	"github.com/google/uuid"
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
	s.router = router.InitializeMux(database, cfg)
}

func TestAPISuite(t *testing.T) {
	suite.Run(t, new(APISuite))
}

func (s *APISuite) TestUserFlow() {
	// Unique ID for test traceability
	uid := uuid.NewString()[:8]
	slug := uuid.New().String()[:8]

	var access_token, refresh_token string

	//Cria o usuário, verifica os tokens e revoga
	s.Run("RegisterAndLogin", func() {
		access_token, refresh_token = s.RegisterAndLogin(uid)
	})
	s.Run("VerifyTokens", func() {
		s.VerifyTokens(access_token, refresh_token)
	})
	s.Run("RevokeRefreshToken", func() {
		s.RevokeRefreshToken(access_token, refresh_token)
	})

// loga o usuário, troca o nome
	s.Run("Login", func() {
		access_token, refresh_token = s.Login(uid)
	})
	s.Run("ChangeName", func() {
		s.ChangeName(access_token, refresh_token, uid)
	})

// loga o superUser, transforma o usuario em criador de eventos
	s.Run("8_LoginSuperUser", func() {
		access_token, refresh_token = s.LoginEmailPassword(os.Getenv("SCTI_EMAIL"), os.Getenv("MASTER_USER_PASS"))
	})
	s.Run("9_SwitchEventCreatorStauts", func() {
		s.SwitchEventCreatorStatus(access_token, refresh_token, uid)
	})

	// o usuario cria um evento, apaga, cria outro, atualiza, se registra nele, vê os eventos registrados
	s.Run("Login", func() {
		access_token, refresh_token = s.Login(uid)
	})
	s.Run("CreateEvent", func() {
		s.CreateEvent(access_token, refresh_token, slug)
	})
	s.Run("DeleteEvent", func() {
		s.DeleteEvent(access_token, refresh_token, slug)
	})
	slug = uuid.New().String()[:8]
	s.Run("CreateEvent", func() {
		s.CreateEvent(access_token, refresh_token, slug)
	})
	s.Run("UpdateEvent", func() {
		s.UpdateEvent(access_token, refresh_token, slug)
	})
	s.Run("RegisterUserEvent", func() {
		s.RegisterUserEvent(access_token, refresh_token, slug)
	})
	s.Run("GetEventsUser", func() {
		s.GetEventsUser(access_token, refresh_token)
	})

	// novo usuario, se registra em um evento
	uid2 := uuid.NewString()[:8]
	s.Run("RegisterAndLogin", func() {
		access_token, refresh_token = s.RegisterAndLogin(uid2)
	})
	s.Run("RegisterUserEvent", func() {
		s.RegisterUserEvent(access_token, refresh_token, slug)
	})


	// loga o superUser, transforma o usuario2 em gerenciador de evento
	s.Run("8_LoginSuperUser", func() {
		access_token, refresh_token = s.LoginEmailPassword(os.Getenv("SCTI_EMAIL"), os.Getenv("MASTER_USER_PASS"))
	})
	s.Run("PromoteUserEvent", func() {
		s.PromoteUserEvent(access_token, refresh_token, uid2, slug)
	})

	// login usuario2, desregistra e desloga
	s.Run("Login", func() {
		access_token, refresh_token = s.Login(uid2)
	})
	s.Run("UnregisterUserEvent", func() {
		s.UnregisterUserEvent(access_token, refresh_token, slug)
	})
	s.Run("UserEvents", func() {
		s.UserEvents(access_token, refresh_token)
	})
	s.Run("Logout", func() {
		s.Logout(access_token, refresh_token)
	})

// verifica os eventos, não precisa de login
	s.Run("GetEvents", func() {
		s.GetEvents()
	})
	s.Run("GetPublicEvents", func() {
		s.GetPublicEvents()
	})
	s.Run("GetEventSlug", func() {
		s.GetEventSlug(slug)
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

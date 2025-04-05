package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"scti/internal/services"
)

type AuthHandler struct {
	AuthService *services.AuthService
}

func NewAuthHandler(service *services.AuthService) *AuthHandler {
	return &AuthHandler{AuthService: service}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	// Simples validação (caso precise)
	if data.Email == "" || data.Password == "" || data.Name == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if err := h.AuthService.Register(data.Email, data.Password, data.Name); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	token, err := h.AuthService.Login(data.Email, data.Password)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Resposta com o token
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, token)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if data.Email == "" || data.Password == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	token, err := h.AuthService.Login(data.Email, data.Password)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, token)
}

func (h *AuthHandler) VerifyJWT(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "JWT is valid") // Simples resposta, caso você queira validar o JWT de fato, você pode adicionar lógica aqui.
}

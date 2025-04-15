package handlers

import (
	"encoding/json"
	"net/http"
	"scti/internal/models"
	"scti/internal/services"
	u "scti/internal/utilities"
)

type EventHandler struct {
	EventService *services.EventService
}

func NewEventHandler(service *services.EventService) *EventHandler {
	return &EventHandler{EventService: service}
}

func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var event models.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		u.Send(w, "Erro ao ler JSON", nil, http.StatusBadRequest)
		return
	}

	claims := u.GetUserFromContext(r.Context())
	user, err := h.EventService.GetUserByID(claims.ID)
	if err != nil {
		u.Send(w, "Erro ao verificar usuário: "+err.Error(), nil, http.StatusInternalServerError)
		return
	} else if !user.IsMasterUser {
		u.Send(w, "Apenas usuários mestres podem criar eventos", nil, http.StatusForbidden)
		return
	}

	err = h.EventService.CreateEvent(&event)
	if err != nil {
		u.Send(w, err.Error(), nil, http.StatusBadRequest)
		return
	}

	u.Send(w, "", event, http.StatusOK)
}

func (h *EventHandler) GetEventBySlug(w http.ResponseWriter, r *http.Request) {
	eventSlug := r.PathValue("slug")
	if eventSlug == "" {
		u.Send(w, "slug não pode ser vazio", nil, http.StatusBadRequest)
		return
	}

	event, err := h.EventService.GetEventBySlug(eventSlug)
	if err != nil {
		u.Send(w, err.Error(), nil, http.StatusBadRequest)
		return
	}

	u.Send(w, "", event, http.StatusOK)
}

func (h *EventHandler) GetAllEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.EventService.GetAllEvents()
	if err != nil {
		u.Send(w, err.Error(), nil, http.StatusBadRequest)
		return
	}

	u.Send(w, "", events, http.StatusOK)
}

func (h *EventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	var event models.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		u.Send(w, "Erro ao ler JSON"+err.Error(), nil, http.StatusBadRequest)
		return
	}

	claims := u.GetUserFromContext(r.Context())
	user, err := h.EventService.GetUserByID(claims.ID)
	if err != nil {
		u.Send(w, "Erro ao verificar usuário: "+err.Error(), nil, http.StatusInternalServerError)
		return
	} else if !user.IsMasterUser {
		u.Send(w, "Apenas usuários mestres podem modificar eventos", nil, http.StatusForbidden)
		return
	}

	UpdatedEvent, err := h.EventService.UpdateEvent(&event)
	if err != nil {
		u.Send(w, err.Error(), nil, http.StatusBadRequest)
		return
	}

	u.Send(w, "", UpdatedEvent, http.StatusOK)
}

func (h *EventHandler) UpdateEventBySlug(w http.ResponseWriter, r *http.Request) {
	var event models.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		u.Send(w, "Erro ao ler JSON"+err.Error(), nil, http.StatusBadRequest)
		return
	}

	claims := u.GetUserFromContext(r.Context())
	user, err := h.EventService.GetUserByID(claims.ID)
	if err != nil {
		u.Send(w, "Erro ao verificar usuário: "+err.Error(), nil, http.StatusInternalServerError)
		return
	} else if !user.IsMasterUser {
		u.Send(w, "Apenas usuários mestres podem modificar eventos", nil, http.StatusForbidden)
		return
	}

	slug := r.PathValue("slug")
	if slug == "" {
		u.Send(w, "slug não pode ser vazio", nil, http.StatusBadRequest)
		return
	}

	UpdatedEvent, err := h.EventService.UpdateEventBySlug(slug, &event)
	if err != nil {
		u.Send(w, err.Error(), nil, http.StatusBadRequest)
		return
	}

	u.Send(w, "", UpdatedEvent, http.StatusOK)
}

// TODO: Cannot delete event if there are paid users registered to it
func (h *EventHandler) DeleteEventBySlug(w http.ResponseWriter, r *http.Request) {
	claims := u.GetUserFromContext(r.Context())
	user, err := h.EventService.GetUserByID(claims.ID)
	if err != nil {
		u.Send(w, "Erro ao verificar usuário: "+err.Error(), nil, http.StatusInternalServerError)
		return
	} else if !user.IsMasterUser {
		u.Send(w, "Apenas usuários mestres podem modificar eventos", nil, http.StatusForbidden)
		return
	}

	slug := r.PathValue("slug")
	if slug == "" {
		u.Send(w, "slug não pode ser vazio", nil, http.StatusBadRequest)
		return
	}

	err = h.EventService.DeleteEventBySlug(slug)
	if err != nil {
		u.Send(w, err.Error(), nil, http.StatusBadRequest)
		return
	}

	u.Send(w, "Deleted Event", nil, http.StatusOK)
}

func (h *EventHandler) RegisterToEvent(w http.ResponseWriter, r *http.Request) {
	claims := u.GetUserFromContext(r.Context())
	slug := r.PathValue("slug")
	if slug == "" {
		u.Send(w, "slug não pode ser vazio", nil, http.StatusBadRequest)
		return
	}

	err := h.EventService.RegisterToEvent(claims.ID, slug)
	if err != nil {
		u.Send(w, err.Error(), nil, http.StatusBadRequest)
		return
	}

	u.Send(w, "Registered to Event", nil, http.StatusOK)
}

// TODO: Cannot unregister if the user has paid for the event
func (h *EventHandler) UnregisterToEvent(w http.ResponseWriter, r *http.Request) {
	claims := u.GetUserFromContext(r.Context())
	slug := r.PathValue("slug")
	if slug == "" {
		u.Send(w, "slug não pode ser vazio", nil, http.StatusBadRequest)
		return
	}

	err := h.EventService.UnregisterToEvent(claims.ID, slug)
	if err != nil {
		u.Send(w, err.Error(), nil, http.StatusBadRequest)
		return
	}

	u.Send(w, "Unregistered from Event", nil, http.StatusOK)
}

func (h *EventHandler) GetEventAtendeesBySlug(w http.ResponseWriter, r *http.Request) {
	eventSlug := r.PathValue("slug")
	if eventSlug == "" {
		u.Send(w, "slug não pode ser vazio", nil, http.StatusBadRequest)
		return
	}

	atendees, err := h.EventService.GetEventAtendeesBySlug(eventSlug)
	if err != nil {
		u.Send(w, err.Error(), nil, http.StatusBadRequest)
		return
	}

	u.Send(w, "", atendees, http.StatusOK)
}

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
		u.SendError(w, []string{"error parsing response body: " + err.Error()}, "event-stack", http.StatusBadRequest)
		return
	}

	claims := u.GetUserFromContext(r.Context())
	user, err := h.EventService.GetUserByID(claims.ID)
	if err != nil {
		u.SendError(w, []string{"error getting user: " + err.Error()}, "event-stack", http.StatusInternalServerError)
		return
	} else if !user.IsMasterUser {
		u.SendError(w, []string{"only master users can create events"}, "event-stack", http.StatusForbidden)
		return
	}

	err = h.EventService.CreateEvent(&event)
	if err != nil {
		u.SendError(w, []string{"error creating event: " + err.Error()}, "event-stack", http.StatusBadRequest)
		return
	}

	u.SendSuccess(w, event, "", http.StatusOK)
}

func (h *EventHandler) GetEventBySlug(w http.ResponseWriter, r *http.Request) {
	eventSlug := r.PathValue("slug")
	if eventSlug == "" {
		u.SendError(w, []string{"the event slug can't be empty"}, "event-stack", http.StatusBadRequest)
		return
	}

	event, err := h.EventService.GetEventBySlug(eventSlug)
	if err != nil {
		u.SendError(w, []string{"error getting event: " + err.Error()}, "event-stack", http.StatusBadRequest)
		return
	}

	u.SendSuccess(w, event, "", http.StatusOK)
}

func (h *EventHandler) GetAllEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.EventService.GetAllEvents()
	if err != nil {
		u.SendError(w, []string{"error getting all events: " + err.Error()}, "event-stack", http.StatusBadRequest)
		return
	}

	u.SendSuccess(w, events, "", http.StatusOK)
}

func (h *EventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	var event models.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		u.SendError(w, []string{"error reading json: " + err.Error()}, "event-stack", http.StatusBadRequest)
		return
	}

	claims := u.GetUserFromContext(r.Context())
	user, err := h.EventService.GetUserByID(claims.ID)
	if err != nil {
		u.SendError(w, []string{"error getting user: " + err.Error()}, "event-stack", http.StatusInternalServerError)
		return
	} else if !user.IsMasterUser {
		u.SendError(w, []string{"only master users can edit events"}, "event-stack", http.StatusForbidden)
		return
	}

	UpdatedEvent, err := h.EventService.UpdateEvent(&event)
	if err != nil {
		u.SendError(w, []string{"error updating event: " + err.Error()}, "event-stack", http.StatusBadRequest)
		return
	}

	u.SendSuccess(w, UpdatedEvent, "", http.StatusOK)
}

func (h *EventHandler) UpdateEventBySlug(w http.ResponseWriter, r *http.Request) {
	var event models.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		u.SendError(w, []string{"error reading json: " + err.Error()}, "event-stack", http.StatusBadRequest)
		return
	}

	slug := r.PathValue("slug")
	if slug == "" {
		u.SendError(w, []string{"the event slug can't be empty"}, "event-stack", http.StatusBadRequest)
		return
	}

	claims := u.GetUserFromContext(r.Context())
	user, err := h.EventService.GetUserByID(claims.ID)
	if err != nil {
		u.SendError(w, []string{"error getting user: " + err.Error()}, "event-stack", http.StatusInternalServerError)
		return
	} else if !user.IsMasterUser {
		u.SendError(w, []string{"only master users can edit events"}, "event-stack", http.StatusForbidden)
		return
	}

	UpdatedEvent, err := h.EventService.UpdateEventBySlug(slug, &event)
	if err != nil {
		u.SendError(w, []string{"error updating event: " + err.Error()}, "event-stack", http.StatusBadRequest)
		return
	}

	u.SendSuccess(w, UpdatedEvent, "", http.StatusOK)
}

// TODO: Cannot delete event if there are paid users registered to it
func (h *EventHandler) DeleteEventBySlug(w http.ResponseWriter, r *http.Request) {
	claims := u.GetUserFromContext(r.Context())

	slug := r.PathValue("slug")
	if slug == "" {
		u.SendError(w, []string{"the event slug can't be empty"}, "event-stack", http.StatusBadRequest)
		return
	}

	user, err := h.EventService.GetUserByID(claims.ID)
	if err != nil {
		u.SendError(w, []string{"error getting user: " + err.Error()}, "event-stack", http.StatusInternalServerError)
		return
	} else if !user.IsMasterUser {
		u.SendError(w, []string{"only master users can delete events"}, "event-stack", http.StatusForbidden)
		return
	}

	err = h.EventService.DeleteEventBySlug(slug)
	if err != nil {
		u.SendError(w, []string{"error deleting event: " + err.Error()}, "event-stack", http.StatusBadRequest)
		return
	}

	u.SendSuccess(w, nil, "deleted event", http.StatusOK)
}

func (h *EventHandler) RegisterToEvent(w http.ResponseWriter, r *http.Request) {
	claims := u.GetUserFromContext(r.Context())
	slug := r.PathValue("slug")
	if slug == "" {
		u.SendError(w, []string{"the event slug can't be empty"}, "event-stack", http.StatusBadRequest)
		return
	}

	err := h.EventService.RegisterToEvent(claims.ID, slug)
	if err != nil {
		u.SendError(w, []string{"error registering to event: " + err.Error()}, "event-stack", http.StatusBadRequest)
		return
	}

	u.SendSuccess(w, nil, "registered to event sucessfully", http.StatusOK)
}

// TODO: Cannot unregister if the user has paid for the event
func (h *EventHandler) UnregisterToEvent(w http.ResponseWriter, r *http.Request) {
	claims := u.GetUserFromContext(r.Context())
	slug := r.PathValue("slug")
	if slug == "" {
		u.SendError(w, []string{"the event slug can't be empty"}, "event-stack", http.StatusBadRequest)
		return
	}

	err := h.EventService.UnregisterToEvent(claims.ID, slug)
	if err != nil {
		u.SendError(w, []string{"error unregistering to event" + err.Error()}, "event-stack", http.StatusBadRequest)
		return
	}

	u.SendSuccess(w, "unregistered from event successfully", "event-stack", http.StatusOK)
}

func (h *EventHandler) GetEventAtendeesBySlug(w http.ResponseWriter, r *http.Request) {
	eventSlug := r.PathValue("slug")
	if eventSlug == "" {
		u.SendError(w, []string{"the event slug can't be empty"}, "event-stack", http.StatusBadRequest)
		return
	}

	atendees, err := h.EventService.GetEventAtendeesBySlug(eventSlug)
	if err != nil {
		u.SendError(w, []string{"error getting event atendees: " + err.Error()}, "event-stack", http.StatusBadRequest)
		return
	}

	u.SendSuccess(w, atendees, "", http.StatusOK)
}

func (h *EventHandler) PromoteUserOfEventBySlug(w http.ResponseWriter, r *http.Request) {
	claims := u.GetUserFromContext(r.Context())
	slug := r.PathValue("slug")
	if slug == "" {
		u.SendError(w, []string{"the event slug can't be empty"}, "event-stack", http.StatusBadRequest)
		return
	}

	type emailBody struct {
		Email string `json:"email"`
	}

	var body emailBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		u.SendError(w, []string{"error reading json: " + err.Error()}, "event-stack", http.StatusBadRequest)
		return
	}

	err := h.EventService.PromoteUserOfEventBySlug(body.Email, claims.ID, slug)
	if err != nil {
		u.SendError(w, []string{"error promoting user: " + err.Error()}, "event-stack", http.StatusBadRequest)
		return
	}

	u.SendSuccess(w, nil, "successfully promoted user of "+slug, http.StatusOK)
}

func (h *EventHandler) DemoteUserOfEventBySlug(w http.ResponseWriter, r *http.Request) {
	claims := u.GetUserFromContext(r.Context())
	slug := r.PathValue("slug")
	if slug == "" {
		u.SendError(w, []string{"the event slug can't be empty"}, "event-stack", http.StatusBadRequest)
		return
	}

	type emailBody struct {
		Email string `json:"email"`
	}

	var body emailBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		u.SendError(w, []string{"error reading json: " + err.Error()}, "event-stack", http.StatusBadRequest)
		return
	}

	err := h.EventService.DemoteUserOfEventBySlug(body.Email, claims.ID, slug)
	if err != nil {
		u.SendError(w, []string{"error demoting user: " + err.Error()}, "event-stack", http.StatusBadRequest)
		return
	}

	u.SendSuccess(w, nil, "successfully demoted user of "+slug, http.StatusOK)
}

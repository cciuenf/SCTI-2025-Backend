package handlers

import (
	"encoding/json"
	"net/http"
	"scti/internal/models"
	"scti/internal/services"
	u "scti/internal/utilities"
	"time"
)

type EventHandler struct {
	EventService *services.EventService
}

func NewEventHandler(service *services.EventService) *EventHandler {
	return &EventHandler{EventService: service}
}

type CreateEventRequest struct {
	Slug        string    `json:"slug" example:"gws"`
	Name        string    `json:"name" example:"Go Workshop"`
	Description string    `json:"description" example:"Learn Go programming"`
	StartDate   time.Time `json:"start_date" example:"2025-05-01T14:00:00Z"`
	EndDate     time.Time `json:"end_date" example:"2025-05-01T17:00:00Z"`
	Location    string    `json:"location" example:"Room 101"`
}

// CreateEvent godoc
// @Summary      Create a new event
// @Description  Creates a new event. Only master users can create events
// @Tags         events
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        request body CreateEventRequest true "Event creation info"
// @Success      200  {object}  NoMessageSuccessResponse{data=models.Event}
// @Failure      400  {object}  EventStandardErrorResponse
// @Failure      401  {object}  EventStandardErrorResponse
// @Failure      403  {object}  EventStandardErrorResponse
// @Router       /events [post]
func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var event models.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		u.SendError(w, []string{"error parsing response body: " + err.Error()}, "event-stack", http.StatusBadRequest)
		return
	}

	claims := u.GetUserFromContext(r.Context())
	user, err := h.EventService.GetUserByID(claims.ID)
	if err != nil {
		u.SendError(w, []string{"error getting user: " + err.Error()}, "event-stack", http.StatusBadRequest)
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

// GetEventBySlug godoc
// @Summary      Get event by slug
// @Description  Returns an event's details by its slug
// @Tags         events
// @Produce      json
// @Param        slug path string true "Event slug"
// @Success      200  {object}  NoMessageSuccessResponse{data=models.Event}
// @Failure      400  {object}  EventStandardErrorResponse
// @Router       /events/{slug} [get]
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

// GetEventBySlugWithActivities godoc
// @Summary      Get event by slug with all its activities
// @Description  Returns an event's details by its slug with all its activities filled with their data
// @Tags         events
// @Produce      json
// @Param        slug path string true "Event slug"
// @Success      200  {object}  NoMessageSuccessResponse{data=models.Event}
// @Failure      400  {object}  EventStandardErrorResponse
// @Router       /events/{slug}/activities [get]
func (h *EventHandler) GetEventBySlugWithActivities(w http.ResponseWriter, r *http.Request) {
	eventSlug := r.PathValue("slug")
	if eventSlug == "" {
		u.SendError(w, []string{"the event slug can't be empty"}, "event-stack", http.StatusBadRequest)
		return
	}

	event, err := h.EventService.GetEventBySlugWithActivities(eventSlug)
	if err != nil {
		u.SendError(w, []string{"error getting event: " + err.Error()}, "event-stack", http.StatusBadRequest)
		return
	}

	u.SendSuccess(w, event, "", http.StatusOK)
}

// GetAllEvents godoc
// @Summary      Get all events
// @Description  Returns a list of all events
// @Tags         events
// @Produce      json
// @Success      200  {object}  NoMessageSuccessResponse{data=[]models.Event}
// @Failure      400  {object}  EventStandardErrorResponse
// @Router       /events [get]
func (h *EventHandler) GetAllEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.EventService.GetAllEvents()
	if err != nil {
		u.SendError(w, []string{"error getting all events: " + err.Error()}, "event-stack", http.StatusBadRequest)
		return
	}

	u.SendSuccess(w, events, "", http.StatusOK)
}

type UpdateEventByIDRequest struct {
	ID          string    `json:"id" example:"be7b5a6d-ae48-4bda-b01e-58a12eeb65f5"`
	Slug        string    `json:"slug" example:"uw"`
	Name        string    `json:"name" example:"Updated Workshop"`
	Description string    `json:"description" example:"Updated workshop description"`
	Location    string    `json:"location" example:"Room 202"`
	StartDate   time.Time `json:"start_date" example:"2030-11-11T00:00:00Z"`
	EndDate     time.Time `json:"end_date" example:"2030-11-11T23:59:59Z"`
}

// UpdateEvent godoc
// @Summary      Update an event by ID
// @Description  Updates an existing event using its ID. Only master users can update events
// @Tags         events
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        request body UpdateEventByIDRequest true "Event update info with ID"
// @Success      200  {object}  NoMessageSuccessResponse{data=models.Event}
// @Failure      400  {object}  EventStandardErrorResponse
// @Failure      401  {object}  EventStandardErrorResponse
// @Failure      403  {object}  EventStandardErrorResponse
// @Router       /events [patch]
func (h *EventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	var event models.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		u.SendError(w, []string{"error reading json: " + err.Error()}, "event-stack", http.StatusBadRequest)
		return
	}

	claims := u.GetUserFromContext(r.Context())
	user, err := h.EventService.GetUserByID(claims.ID)
	if err != nil {
		u.SendError(w, []string{"error getting user: " + err.Error()}, "event-stack", http.StatusBadRequest)
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

type UpdateEventRequest struct {
	Slug        string    `json:"slug" example:"uw"`
	Name        string    `json:"name" example:"Updated Workshop"`
	Description string    `json:"description" example:"Updated workshop description"`
	Location    string    `json:"location" example:"Room 202"`
	StartDate   time.Time `json:"start_date" example:"2030-11-11T00:00:00Z"`
	EndDate     time.Time `json:"end_date" example:"2030-11-11T23:59:59Z"`
}

// UpdateEventBySlug godoc
// @Summary      Update an event by slug
// @Description  Updates an existing event using its slug. Only master users can update events
// @Tags         events
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug"
// @Param        request body UpdateEventRequest true "Event update info"
// @Success      200  {object}  NoMessageSuccessResponse{data=models.Event}
// @Failure      400  {object}  EventStandardErrorResponse
// @Failure      401  {object}  EventStandardErrorResponse
// @Failure      403  {object}  EventStandardErrorResponse
// @Router       /events/{slug} [patch]
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
		u.SendError(w, []string{"error getting user: " + err.Error()}, "event-stack", http.StatusBadRequest)
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
// TODO: Implement a block event, that hides and blocks thatt event from being registered to
// DeleteEventBySlug godoc
// @Summary      Delete an event by slug
// @Description  Deletes an existing event using its slug. Only master users can delete events
// @Tags         events
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug"
// @Success      200  {object}  NoDataSuccessResponse
// @Failure      400  {object}  EventStandardErrorResponse
// @Failure      401  {object}  EventStandardErrorResponse
// @Failure      403  {object}  EventStandardErrorResponse
// @Router       /events/{slug} [delete]
func (h *EventHandler) DeleteEventBySlug(w http.ResponseWriter, r *http.Request) {
	claims := u.GetUserFromContext(r.Context())

	slug := r.PathValue("slug")
	if slug == "" {
		u.SendError(w, []string{"the event slug can't be empty"}, "event-stack", http.StatusBadRequest)
		return
	}

	user, err := h.EventService.GetUserByID(claims.ID)
	if err != nil {
		u.SendError(w, []string{"error getting user: " + err.Error()}, "event-stack", http.StatusBadRequest)
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

// RegisterToEvent godoc
// @Summary      Register to an event
// @Description  Registers the authenticated user to an event by its slug
// @Tags         events
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug"
// @Success      200  {object}  NoDataSuccessResponse
// @Failure      400  {object}  EventStandardErrorResponse
// @Failure      401  {object}  EventStandardErrorResponse
// @Router       /events/{slug}/attend [post]
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
// UnregisterToEvent godoc
// @Summary      Unregister from an event
// @Description  Unregisters the authenticated user from an event by its slug
// @Tags         events
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug"
// @Success      200  {object}  NoDataSuccessResponse
// @Failure      400  {object}  EventStandardErrorResponse
// @Failure      401  {object}  EventStandardErrorResponse
// @Router       /events/{slug}/unattend [post]
func (h *EventHandler) UnregisterFromEvent(w http.ResponseWriter, r *http.Request) {
	claims := u.GetUserFromContext(r.Context())
	slug := r.PathValue("slug")
	if slug == "" {
		u.SendError(w, []string{"the event slug can't be empty"}, "event-stack", http.StatusBadRequest)
		return
	}

	err := h.EventService.UnregisterFromEvent(claims.ID, slug)
	if err != nil {
		u.SendError(w, []string{"error unregistering to event: " + err.Error()}, "event-stack", http.StatusBadRequest)
		return
	}

	u.SendSuccess(w, "unregistered from event successfully", "event-stack", http.StatusOK)
}

// GetEventAtendeesBySlug godoc
// @Summary      Get event attendees
// @Description  Returns a list of all user IDs registered to an event by its slug
// @Tags         events
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug"
// @Success      200  {object}  NoMessageSuccessResponse{data=[]models.EventUser}
// @Failure      400  {object}  EventStandardErrorResponse
// @Failure      401  {object}  EventStandardErrorResponse
// @Router       /events/{slug}/attendees [get]
func (h *EventHandler) GetEventAttendeesBySlug(w http.ResponseWriter, r *http.Request) {
	eventSlug := r.PathValue("slug")
	if eventSlug == "" {
		u.SendError(w, []string{"the event slug can't be empty"}, "event-stack", http.StatusBadRequest)
		return
	}

	attendees, err := h.EventService.GetEventAttendeesBySlug(eventSlug)
	if err != nil {
		u.SendError(w, []string{"error getting event attendees: " + err.Error()}, "event-stack", http.StatusBadRequest)
		return
	}

	u.SendSuccess(w, attendees, "", http.StatusOK)
}

type PromoteUserRequest struct {
	Email string `json:"email" example:"user@example.com"`
}

// PromoteUserOfEventBySlug godoc
// @Summary      Promote user in event
// @Description  Promotes a user to organizer role in an event. The following rules apply:
// @Description  - Only master users and master admins can promote others
// @Description  - Master users can promote normal users to admin or admins to master admin
// @Description  - Master admins can only promote normal users to admin
// @Description  - Users must be registered to the event to be promoted
// @Description  - Users cannot promote themselves
// @Description  - Master users cannot be promoted
// @Tags         events
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug"
// @Param        request body PromoteUserRequest true "User email to promote"
// @Success      200  {object}  NoDataSuccessResponse
// @Failure      400  {object}  EventStandardErrorResponse
// @Failure      401  {object}  EventStandardErrorResponse
// @Failure      403  {object}  EventStandardErrorResponse
// @Router       /events/{slug}/promote [post]
func (h *EventHandler) PromoteUserOfEventBySlug(w http.ResponseWriter, r *http.Request) {
	claims := u.GetUserFromContext(r.Context())
	slug := r.PathValue("slug")
	if slug == "" {
		u.SendError(w, []string{"the event slug can't be empty"}, "event-stack", http.StatusBadRequest)
		return
	}

	var body PromoteUserRequest
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

type DemoteUserRequest struct {
	Email string `json:"email" example:"user@example.com"`
}

// DemoteUserOfEventBySlug godoc
// @Summary      Demote user in event
// @Description  Demotes a user from their admin role in an event. The following rules apply:
// @Description  - Only master users and master admins can demote others
// @Description  - Master users can demote any admin (master or normal)
// @Description  - Master admins can only demote normal admins
// @Description  - Users cannot demote themselves
// @Description  - Master users cannot be demoted
// @Description  - Target must be an admin of the event
// @Description  - Targets can be demoted if they unregister from the event
// @Tags         events
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug"
// @Param        request body DemoteUserRequest true "User email to demote"
// @Success      200  {object}  NoDataSuccessResponse
// @Failure      400  {object}  EventStandardErrorResponse
// @Failure      401  {object}  EventStandardErrorResponse
// @Failure      403  {object}  EventStandardErrorResponse
// @Router       /events/{slug}/demote [post]
func (h *EventHandler) DemoteUserOfEventBySlug(w http.ResponseWriter, r *http.Request) {
	claims := u.GetUserFromContext(r.Context())
	slug := r.PathValue("slug")
	if slug == "" {
		u.SendError(w, []string{"the event slug can't be empty"}, "event-stack", http.StatusBadRequest)
		return
	}

	var body DemoteUserRequest
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

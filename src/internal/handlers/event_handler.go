package handlers

import (
	"errors"
	"net/http"
	"scti/internal/models"
	"scti/internal/services"
)

type EventHandler struct {
	EventService *services.EventService
}

func NewEventHandler(service *services.EventService) *EventHandler {
	return &EventHandler{EventService: service}
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
// @Param        request body models.CreateEventRequest true "Event creation info"
// @Success      200  {object}  NoMessageSuccessResponse{data=models.Event}
// @Failure      400  {object}  EventStandardErrorResponse
// @Failure      401  {object}  EventStandardErrorResponse
// @Failure      403  {object}  EventStandardErrorResponse
// @Router       /events [post]
func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var reqBody models.CreateEventRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	user, err := getUserFromContext(h.EventService.GetUserByID, r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	event, err := h.EventService.CreateEvent(user, reqBody)
	if err != nil {
		handleError(w, errors.New("error creating event: "+err.Error()), http.StatusBadRequest)
		return
	}

	handleSuccess(w, event, "", http.StatusOK)
}

// GetEvent godoc
// @Summary      Get event by slug
// @Description  Returns an event's details by its slug
// @Tags         events
// @Produce      json
// @Param        slug path string true "Event slug"
// @Success      200  {object}  NoMessageSuccessResponse{data=models.Event}
// @Failure      400  {object}  EventStandardErrorResponse
// @Router       /events/{slug} [get]
func (h *EventHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	slug, err := extractSlugAndValidate(r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	event, err := h.EventService.GetEvent(slug)
	if err != nil {
		handleError(w, errors.New("error getting event: "+err.Error()), http.StatusBadRequest)
		return
	}

	handleSuccess(w, event, "", http.StatusOK)
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
		handleError(w, errors.New("error getting all events: "+err.Error()), http.StatusBadRequest)
		return
	}

	handleSuccess(w, events, "", http.StatusOK)
}

// GetEventsCreatedByUser godoc
// @Summary      Get events created by a user
// @Description  Returns a list of all events created by a user
// @Tags         events
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Success      200  {object}  NoMessageSuccessResponse{data=[]models.Event}
// @Failure      400  {object}  EventStandardErrorResponse
// @Failure      401  {object}  EventStandardErrorResponse
// @Failure      403  {object}  EventStandardErrorResponse
// @Router       /events/created [get]
func (h *EventHandler) GetEventsCreatedByUser(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromContext(h.EventService.GetUserByID, r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	events, err := h.EventService.GetEventsCreatedByUser(user)
	if err != nil {
		handleError(w, errors.New("error getting events created by user: "+err.Error()), http.StatusBadRequest)
		return
	}

	handleSuccess(w, events, "", http.StatusOK)
}

// GetAllPublicEvents godoc
// @Summary      Get all public events
// @Description  Returns a list of all public events (where IsPublic=true)
// @Tags         events
// @Produce      json
// @Success      200  {object}  NoMessageSuccessResponse{data=[]models.Event}
// @Failure      400  {object}  EventStandardErrorResponse
// @Router       /events/public [get]
func (h *EventHandler) GetAllPublicEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.EventService.GetAllPublicEvents()
	if err != nil {
		handleError(w, errors.New("error getting all events: "+err.Error()), http.StatusBadRequest)
		return
	}

	handleSuccess(w, events, "", http.StatusOK)
}

// UpdateEvent godoc
// @Summary      Update an event by slug
// @Description  Updates an existing event using its slug. Only master users can update events
// @Tags         events
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug"
// @Param        request body models.UpdateEventRequest true "Event update info"
// @Success      200  {object}  NoMessageSuccessResponse{data=models.Event}
// @Failure      400  {object}  EventStandardErrorResponse
// @Failure      401  {object}  EventStandardErrorResponse
// @Failure      403  {object}  EventStandardErrorResponse
// @Router       /events/{slug} [patch]
func (h *EventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	slug, err := extractSlugAndValidate(r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	var reqBody models.UpdateEventRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	user, err := getUserFromContext(h.EventService.GetUserByID, r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	updatedEvent, err := h.EventService.UpdateEvent(user, slug, &reqBody)
	if err != nil {
		handleError(w, errors.New("error updating event: "+err.Error()), http.StatusBadRequest)
		return
	}

	handleSuccess(w, updatedEvent, "", http.StatusOK)
}

// DeleteEvent godoc
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
func (h *EventHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	slug, err := extractSlugAndValidate(r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	user, err := getUserFromContext(h.EventService.GetUserByID, r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	if err := h.EventService.DeleteEvent(user, slug); err != nil {
		handleError(w, errors.New("error deleting event: "+err.Error()), http.StatusBadRequest)
		return
	}

	handleSuccess(w, nil, "deleted event", http.StatusOK)
}

// Saving the qr code as a png file in the server
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
// @Router       /events/{slug}/register [post]
func (h *EventHandler) RegisterToEvent(w http.ResponseWriter, r *http.Request) {
	slug, err := extractSlugAndValidate(r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	user, err := getUserFromContext(h.EventService.GetUserByID, r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	if err := h.EventService.RegisterUserToEvent(user, slug); err != nil {
		handleError(w, errors.New("error registering to event: "+err.Error()), http.StatusBadRequest)
		return
	}

	handleSuccess(w, nil, "registered to event", http.StatusOK)
}

// UnregisterFromEvent godoc
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
// @Router       /events/{slug}/unregister [post]
func (h *EventHandler) UnregisterFromEvent(w http.ResponseWriter, r *http.Request) {
	slug, err := extractSlugAndValidate(r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	user, err := getUserFromContext(h.EventService.GetUserByID, r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	if err := h.EventService.UnregisterUserFromEvent(user, slug); err != nil {
		handleError(w, errors.New("error unregistering from event: "+err.Error()), http.StatusBadRequest)
		return
	}

	handleSuccess(w, nil, "unregistered from event", http.StatusOK)
}

type UserAdminActionRequest struct {
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
// @Param        request body UserAdminActionRequest true "User email to promote"
// @Success      200  {object}  NoDataSuccessResponse
// @Failure      400  {object}  EventStandardErrorResponse
// @Failure      401  {object}  EventStandardErrorResponse
// @Failure      403  {object}  EventStandardErrorResponse
// @Router       /events/{slug}/promote [post]
func (h *EventHandler) PromoteUserOfEventBySlug(w http.ResponseWriter, r *http.Request) {
	slug, err := extractSlugAndValidate(r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	var reqBody UserAdminActionRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	user, err := getUserFromContext(h.EventService.GetUserByID, r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	if err := h.EventService.PromoteUserOfEventBySlug(user, reqBody.Email, slug); err != nil {
		handleError(w, errors.New("error promoting user: "+err.Error()), http.StatusBadRequest)
		return
	}

	handleSuccess(w, nil, "promoted user", http.StatusOK)
}

// DemoteUserOfEventBySlug godoc
// @Summary      Demote user in event
// @Description  Demotes a user from their admin role in an event. The following rules apply:
// @Description  - Only super users, event creators and master admins can demote others
// @Description  - Super users and event creators can demote any admin (master or normal)
// @Description  - Master admins can only demote normal admins
// @Description  - Users cannot demote themselves
// @Description  - Super users and event creators cannot be demoted
// @Description  - Super users and event creators cannot be promoted
// @Description  - Target must be an admin of the event
// @Description  - Targets can be demoted if they unregister from the event
// @Tags         events
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug"
// @Param        request body UserAdminActionRequest true "User email to demote"
// @Success      200  {object}  NoDataSuccessResponse
// @Failure      400  {object}  EventStandardErrorResponse
// @Failure      401  {object}  EventStandardErrorResponse
// @Failure      403  {object}  EventStandardErrorResponse
// @Router       /events/{slug}/demote [post]
func (h *EventHandler) DemoteUserOfEventBySlug(w http.ResponseWriter, r *http.Request) {
	slug, err := extractSlugAndValidate(r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	var reqBody UserAdminActionRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	user, err := getUserFromContext(h.EventService.GetUserByID, r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	if err := h.EventService.DemoteUserOfEventBySlug(user, reqBody.Email, slug); err != nil {
		handleError(w, errors.New("error demoting user: "+err.Error()), http.StatusBadRequest)
		return
	}

	handleSuccess(w, nil, "demoted user", http.StatusOK)
}

// GetUserEvents godoc
// @Summary      Get user events
// @Description  Returns a list of all events for the authenticated user
// @Tags         events
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Success      200  {object}  NoMessageSuccessResponse{data=[]models.Event}
// @Failure      400  {object}  EventStandardErrorResponse
// @Failure      401  {object}  EventStandardErrorResponse
// @Router       /user-events [get]
func (h *EventHandler) GetUserEvents(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromContext(h.EventService.GetUserByID, r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	events, err := h.EventService.GetUserEvents(user)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	handleSuccess(w, events, "", http.StatusOK)
}

// IsUserPaid godoc
// @Summary      Tells if the user paid for the event or not
// @Description  Returns a boolean indicating if the user is a paid attendant
// @Tags         events
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        request body models.IsUserPaidRequest true "User to check status"
// @Success      200  {object}  NoMessageSuccessResponse{data=bool}
// @Failure      400  {object}  EventStandardErrorResponse
// @Failure      401  {object}  EventStandardErrorResponse
// @Router       /events/{slug}/is-paid [get]
func (h *EventHandler) IsUserPaid(w http.ResponseWriter, r *http.Request) {
	slug, err := extractSlugAndValidate(r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	user, err := getUserFromContext(h.EventService.GetUserByID, r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	var reqBody models.IsUserPaidRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	hasTicket, err := h.EventService.IsUserPaid(user, slug, reqBody.ID)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	handleSuccess(w, hasTicket, "", http.StatusOK)
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
// @Param        request body models.CreateEventRequest true "Event creation info"
// @Success      200  {object}  NoMessageSuccessResponse{data=models.Event}
// @Failure      400  {object}  EventStandardErrorResponse
// @Failure      401  {object}  EventStandardErrorResponse
// @Failure      403  {object}  EventStandardErrorResponse
// @Router       /events/{slug}/coffee [post]
func (h *EventHandler) CreateCoffee(w http.ResponseWriter, r *http.Request) {
	var reqBody models.CreateCoffeeRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	slug, err := extractSlugAndValidate(r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	user, err := getUserFromContext(h.EventService.GetUserByID, r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	coffee, err := h.EventService.CreateCoffee(user, slug, reqBody)
	if err != nil {
		handleError(w, errors.New("error creating coffee break: "+err.Error()), http.StatusBadRequest)
		return
	}

	handleSuccess(w, coffee, "", http.StatusOK)
}

// GetAllEvents godoc
// @Summary      Get all events
// @Description  Returns a list of all events
// @Tags         events
// @Produce      json
// @Success      200  {object}  NoMessageSuccessResponse{data=[]models.Event}
// @Failure      400  {object}  EventStandardErrorResponse
// @Router       /events/{slug}/coffee [get]
func (h *EventHandler) GetAllCoffees(w http.ResponseWriter, r *http.Request) {
	slug, err := extractSlugAndValidate(r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	coffees, err := h.EventService.GetAllCoffees(slug)
	if err != nil {
		handleError(w, errors.New("error getting all coffees: "+err.Error()), http.StatusBadRequest)
		return
	}

	handleSuccess(w, coffees, "", http.StatusOK)
}

// UpdateEvent godoc
// @Summary      Update an event by slug
// @Description  Updates an existing event using its slug. Only master users can update events
// @Tags         events
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug"
// @Param        request body models.UpdateEventRequest true "Event update info"
// @Success      200  {object}  NoMessageSuccessResponse{data=models.Event}
// @Failure      400  {object}  EventStandardErrorResponse
// @Failure      401  {object}  EventStandardErrorResponse
// @Failure      403  {object}  EventStandardErrorResponse
// @Router       /events/{slug}/coffee [patch]
func (h *EventHandler) UpdateCoffee(w http.ResponseWriter, r *http.Request) {
	slug, err := extractSlugAndValidate(r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	var reqBody models.UpdateCoffeeRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	user, err := getUserFromContext(h.EventService.GetUserByID, r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	updatedCoffee, err := h.EventService.UpdateCoffee(user, slug, &reqBody)
	if err != nil {
		handleError(w, errors.New("error updating coffee break: "+err.Error()), http.StatusBadRequest)
		return
	}

	handleSuccess(w, updatedCoffee, "", http.StatusOK)
}

// DeleteEvent godoc
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
// @Router       /events/{slug}/coffee [delete]
func (h *EventHandler) DeleteCoffee(w http.ResponseWriter, r *http.Request) {
	slug, err := extractSlugAndValidate(r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	var reqBody models.DeleteCoffeeRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	user, err := getUserFromContext(h.EventService.GetUserByID, r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	if err := h.EventService.DeleteCoffee(user, slug, reqBody); err != nil {
		handleError(w, errors.New("error deleting event: "+err.Error()), http.StatusBadRequest)
		return
	}

	handleSuccess(w, nil, "deleted event", http.StatusOK)
}

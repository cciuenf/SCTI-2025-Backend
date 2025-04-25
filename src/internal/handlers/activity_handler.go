package handlers

import (
	"encoding/json"
	"net/http"
	"scti/internal/models"
	u "scti/internal/utilities"
	"time"
)

// Exists within event handler, we just use another file for ease of reading

type CreateEventActivityRequest struct {
	Name        string `json:"name" example:"Replicando o github com React.js"`
	Description string `json:"description" example:"Nesse mini curso vamos criar uma réplica do Github com React"`
	Speaker     string `json:"speaker" example:"Marcos Junior"`
	Location    string `json:"location" example:"Laborátorio de Computação 1"`
	MaxCapacity int    `json:"max_capacity" example:"25"` // 0 means unlimited

	Type models.ActivityType `json:"type" example:"mini-curso"`

	StartTime time.Time `json:"start_time" example:"2025-05-01T14:00:00Z"`
	EndTime   time.Time `json:"end_time" example:"2025-05-01T16:00:00Z"`

	IsMandatory bool `json:"is_mandatory" example:"false"` // If the user needs to be registered automatically
	HasFee      bool `json:"has_fee" example:"true"`       // If the user needs a token or not to enter

	IsStandalone   bool   `json:"is_standalone" example:"false"` // If it can be registered to or exist without an event
	StandaloneSlug string `json:"standalone_slug" example:""`    // Used as event slug
}

// CreateEventActivity godoc
// @Summary      Creates a new activity for that event
// @Description  Creates a new activity for that event. Only master users or master admins can create event activities
// @Description  Activities created on this endpoint will not be created as standalone, but can be made standalone after creation
// @Tags         events
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        request body CreateEventActivityRequest true "Activity creation info"
// @Success      200  {object}  NoMessageSuccessResponse{data=models.Activity}
// @Failure      400  {object}  EventStandardErrorResponse
// @Failure      401  {object}  EventStandardErrorResponse
// @Router       /events/{slug}/activity [post]
func (h *EventHandler) CreateEventActivity(w http.ResponseWriter, r *http.Request) {
	eventSlug := r.PathValue("slug")
	if eventSlug == "" {
		u.SendError(w, []string{"the event slug can't be empty"}, "event-stack", http.StatusBadRequest)
		return
	}

	var activity models.Activity
	if err := json.NewDecoder(r.Body).Decode(&activity); err != nil {
		u.SendError(w, []string{"error parsing response body: " + err.Error()}, "event-stack", http.StatusBadRequest)
		return
	}

	claims := u.GetUserFromContext(r.Context())
	user, err := h.EventService.GetUserByID(claims.ID)
	if err != nil {
		u.SendError(w, []string{"error getting user: " + err.Error()}, "event-stack", http.StatusBadRequest)
		return
	}
	if status, err := h.EventService.IsAdminTypeOf(user, models.AdminTypeMaster, eventSlug); !status || err != nil {
		if err != nil {
			u.SendError(w, []string{"error getting user's admin status: " + err.Error()}, "event-stack", http.StatusUnauthorized)
		}
		u.SendError(w, []string{"user is not master admin of the event!"}, "event-stack", http.StatusUnauthorized)
	}

	err = h.EventService.CreateEventActivity(&activity, eventSlug)
	if err != nil {
		u.SendError(w, []string{"error creating activity: " + err.Error()}, "event-stack", http.StatusBadRequest)
		return
	}

	u.SendSuccess(w, activity, "", http.StatusOK)
}

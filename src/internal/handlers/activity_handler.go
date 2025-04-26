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
// @Description  Activities created with 0 MaxCapaxity will have unlimited capacity
// @Tags         activities
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug (if from event)"
// @Param        request body CreateEventActivityRequest true "Activity creation info"
// @Success      200  {object}  NoMessageSuccessResponse{data=models.Activity}
// @Failure      400  {object}  EventStandardErrorResponse
// @Failure      401  {object}  EventStandardErrorResponse
// @Router       /events/{slug}/activity [post]
// @Router       /activity [post]
func (h *EventHandler) CreateEventActivity(w http.ResponseWriter, r *http.Request) {
	eventSlug := r.PathValue("slug")

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

	if eventSlug != "" {
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
	} else {
		if user.IsMasterUser {
			err = h.EventService.CreateEventlessActivity(&activity)
			if err != nil {
				u.SendError(w, []string{"error creating fully standalone activity: " + err.Error()}, "event-stack", http.StatusBadRequest)
				return
			}
		}
	}

	u.SendSuccess(w, activity, "", http.StatusOK)
}

type RegisterToActivityRequest struct {
	ActivityID string `json:"activity_id" example:"uuid"`
}

// RegisterToActivity godoc
// @Summary      Registers an user to an activity
// @Description  Registers an user to an activity if the user is in the event of the activity
// @Description  or if the activity is a standalone activity
// @Tags         activities
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug (if from event)"
// @Param        request body RegisterToActivityRequest true "Activity in which to register to"
// @Success      200  {object}  StandardSuccessResponse
// @Failure      400  {object}  EventStandardErrorResponse
// @Failure      401  {object}  EventStandardErrorResponse
// @Router       /activity/register [post]
// @Router       /events/{slug}/activity/register [post]
func (h *EventHandler) RegisterUserToActivity(w http.ResponseWriter, r *http.Request) {
	var activity RegisterToActivityRequest
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

	eventSlug := r.PathValue("slug")
	if eventSlug == "" {
		err = h.EventService.RegisterUserToStandaloneActivity(user, activity.ActivityID)
		if err != nil {
			u.SendError(w, []string{"error registering to standalone activity: " + err.Error()}, "event-stack", http.StatusBadRequest)
			return
		}
	} else {
		err = h.EventService.RegisterUserToActivityFromEvent(user, activity.ActivityID)
		if err != nil {
			u.SendError(w, []string{"error registering to activity: " + err.Error()}, "event-stack", http.StatusBadRequest)
			return
		}
	}

	u.SendSuccess(w, nil, "registered user to activity", http.StatusOK)
}

type UnregisterFromActivityRequest struct {
	ActivityID string `json:"activity_id" example:"uuid"`
}

// UnregisterToActivity godoc
// @Summary      Unregisters an user from an activity
// @Description  Unregisters an user from an activity independently if the user is in the event or not
// @Description  Doesn't let the user unregister if the activity was attended
// @Tags         activities
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug (if from event)"
// @Param        request body UnregisterFromActivityRequest true "Activity in which to unregister from"
// @Success      200  {object}  StandardSuccessResponse
// @Failure      400  {object}  EventStandardErrorResponse
// @Failure      401  {object}  EventStandardErrorResponse
// @Router       /activity/unregister [post]
// @Router       /events/{slug}/activity/unregister [post]
func (h *EventHandler) UnregisterUserFromActivity(w http.ResponseWriter, r *http.Request) {
	var activity UnregisterFromActivityRequest
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

	err = h.EventService.UnregisterUserFromActivity(user, activity.ActivityID)
	if err != nil {
		u.SendError(w, []string{"error unregistering from activity: " + err.Error()}, "event-stack", http.StatusBadRequest)
		return
	}

	u.SendSuccess(w, nil, "unregistered user from activity", http.StatusOK)
}

// GetAllActivitiesFromEvent godoc
// @Summary      Returns a list of all of that event's activities
// @Description  Returns a list of all of that event's activities
// @Tags         activities
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug"
// @Success      200  {object}  NoMessageSuccessResponse{data=[]models.Activity}
// @Failure      400  {object}  EventStandardErrorResponse
// @Failure      401  {object}  EventStandardErrorResponse
// @Router       /events/{slug}/activities [get]
func (h *EventHandler) GetAllActivitiesFromEvent(w http.ResponseWriter, r *http.Request) {
	eventSlug := r.PathValue("slug")
	if eventSlug == "" {
		u.SendError(w, []string{"the event slug can't be empty"}, "event-stack", http.StatusBadRequest)
		return
	}

	activities, err := h.EventService.GetAllEventActivities(eventSlug)
	if err != nil {
		u.SendError(w, []string{"error getting event activities: " + err.Error()}, "event-stack", http.StatusBadRequest)
		return
	}

	u.SendSuccess(w, activities, "", http.StatusOK)
}

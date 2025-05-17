package handlers

import (
	"errors"
	"net/http"
	"scti/internal/models"
	"scti/internal/services"
	"strings"
)

type ActivityHandler struct {
	ActivityService *services.ActivityService
}

func NewActivityHandler(activityService *services.ActivityService) *ActivityHandler {
	return &ActivityHandler{
		ActivityService: activityService,
	}
}

// CreateEventActivity godoc
// @Summary      Create a new activity for an event
// @Description  Creates a new activity for the specified event
// @Tags         activities
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug"
// @Param        request body models.CreateActivityRequest true "Activity creation info"
// @Success      200  {object}  NoMessageSuccessResponse{data=models.Activity}
// @Failure      400  {object}  ActivityStandardErrorResponse
// @Failure      401  {object}  ActivityStandardErrorResponse
// @Failure      403  {object}  ActivityStandardErrorResponse
// @Router       /events/{slug}/activity [post]
func (h *ActivityHandler) CreateEventActivity(w http.ResponseWriter, r *http.Request) {
	slug, err := extractSlugAndValidate(r)
	if err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	var reqBody models.CreateActivityRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	user, err := getUserFromContext(h.ActivityService.ActivityRepo.GetUserByID, r)
	if err != nil {
		if strings.Contains(err.Error(), "claims") {
			UnauthorizedError(w, err, "activity")
		} else {
			BadRequestError(w, err, "activity")
		}
		return
	}

	activity, err := h.ActivityService.CreateEventActivity(user, slug, reqBody)
	if err != nil {
		HandleErrMsg("Error creating activity", err, w).Stack("activity").BadRequest()
		return
	}

	handleSuccess(w, activity, "", http.StatusOK)
}

// GetAllActivitiesFromEvent godoc
// @Summary      Get all activities for an event
// @Description  Returns all activities for the specified event
// @Tags         activities
// @Produce      json
// @Param        slug path string true "Event slug"
// @Success      200  {object}  NoMessageSuccessResponse{data=[]models.Activity}
// @Failure      400  {object}  ActivityStandardErrorResponse
// @Router       /events/{slug}/activities [get]
func (h *ActivityHandler) GetAllActivitiesFromEvent(w http.ResponseWriter, r *http.Request) {
	slug, err := extractSlugAndValidate(r)
	if err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	activities, err := h.ActivityService.GetAllActivitiesFromEvent(slug)
	if err != nil {
		HandleErrMsg("error getting activities", err, w).Stack("activity").BadRequest()
		return
	}

	handleSuccess(w, activities, "", http.StatusOK)
}

// UpdateEventActivity godoc
// @Summary      Update an activity
// @Description  Updates an existing activity for the specified event
// @Tags         activities
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug"
// @Param        request body models.ActivityUpdateRequest true "Activity update info with ID"
// @Success      200  {object}  NoMessageSuccessResponse{data=models.Activity}
// @Failure      400  {object}  ActivityStandardErrorResponse
// @Failure      401  {object}  ActivityStandardErrorResponse
// @Failure      403  {object}  ActivityStandardErrorResponse
// @Router       /events/{slug}/activity [patch]
func (h *ActivityHandler) UpdateEventActivity(w http.ResponseWriter, r *http.Request) {
	slug, err := extractSlugAndValidate(r)
	if err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	var reqBody models.ActivityUpdateRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	if reqBody.ActivityID == "" {
		BadRequestError(w, NewErr("activity ID is required"), "activity")
		return
	}

	user, err := getUserFromContext(h.ActivityService.ActivityRepo.GetUserByID, r)
	if err != nil {
		if strings.Contains(err.Error(), "claims") {
			UnauthorizedError(w, err, "activity")
		} else {
			BadRequestError(w, err, "activity")
		}
		return
	}

	activity, err := h.ActivityService.UpdateEventActivity(user, slug, reqBody.ActivityID, reqBody)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			NotFoundError(w, err, "Activity", "activity")
		} else if strings.Contains(err.Error(), "permission") {
			ForbiddenError(w, err, "activity")
		} else {
			HandleErrMsg("Error updating activity", err, w).Stack("activity").BadRequest()
		}
		return
	}

	handleSuccess(w, activity, "", http.StatusOK)
}

// TODO: Prohibit deleting activities that have attendees or paid participants
// DeleteEventActivity godoc
// @Summary      Delete an activity
// @Description  Deletes an existing activity from the specified event
// @Tags         activities
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug"
// @Param        request body models.ActivityDeleteRequest true "Activity deletion info"
// @Success      200  {object}  NoDataSuccessResponse
// @Failure      400  {object}  ActivityStandardErrorResponse
// @Failure      401  {object}  ActivityStandardErrorResponse
// @Failure      403  {object}  ActivityStandardErrorResponse
// @Router       /events/{slug}/activity [delete]
func (h *ActivityHandler) DeleteEventActivity(w http.ResponseWriter, r *http.Request) {
	slug, err := extractSlugAndValidate(r)
	if err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	var reqBody models.ActivityDeleteRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	if reqBody.ActivityID == "" {
		HandleErrMsg("activity ID is required", nil, w).Stack("activity").BadRequest()
		return
	}

	user, err := getUserFromContext(h.ActivityService.ActivityRepo.GetUserByID, r)
	if err != nil {
		if strings.Contains(err.Error(), "claims") {
			UnauthorizedError(w, err, "activity")
		} else {
			BadRequestError(w, err, "activity")
		}
		return
	}

	if err := h.ActivityService.DeleteEventActivity(user, slug, reqBody.ActivityID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			NotFoundError(w, err, "Activity", "activity")
		} else if strings.Contains(err.Error(), "permission") {
			ForbiddenError(w, err, "activity")
		} else if strings.Contains(err.Error(), "attendees") {
			ConflictError(w, err, "Activity with existing attendees", "activity")
		} else {
			HandleErr(err, w).Msg("Error deleting activity").Stack("activity").BadRequest()
		}
		return
	}

	handleSuccess(w, nil, "deleted activity", http.StatusOK)
}

// RegisterUserToActivity godoc
// @Summary      Register to an activity
// @Description  Registers the authenticated user to an activity within an event they are already registered for
// @Tags         activities
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug"
// @Param        request body models.ActivityRegistrationRequest true "Activity registration info"
// @Success      200  {object}  NoDataSuccessResponse
// @Failure      400  {object}  ActivityStandardErrorResponse
// @Failure      401  {object}  ActivityStandardErrorResponse
// @Router       /events/{slug}/activity/register [post]
func (h *ActivityHandler) RegisterUserToActivity(w http.ResponseWriter, r *http.Request) {
	slug, err := extractSlugAndValidate(r)
	if err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	var reqBody models.ActivityRegistrationRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	if reqBody.ActivityID == "" {
		HandleErrMsg("activity ID is required", nil, w).Stack("activity").BadRequest()
		return
	}

	user, err := getUserFromContext(h.ActivityService.ActivityRepo.GetUserByID, r)
	if err != nil {
		if strings.Contains(err.Error(), "claims") {
			UnauthorizedError(w, err, "activity")
		} else {
			BadRequestError(w, err, "activity")
		}
		return
	}

	if err := h.ActivityService.RegisterUserToActivity(user, slug, reqBody.ActivityID); err != nil {
		if strings.Contains(err.Error(), "capacity") {
			capacityErr := errors.New("maximum capacity reached")
			HandleErrMsg("activity is at full capacity", capacityErr, w).Stack("activity").Conflict()
		} else if strings.Contains(err.Error(), "already registered") {
			ConflictError(w, err, "Registration", "activity")
		} else if strings.Contains(err.Error(), "event not registered") {
			ForbiddenError(w, errors.New("must register for event first"), "activity")
		} else {
			HandleErr(err, w).Msg("Error registering to activity").Stack("activity").BadRequest()
		}
		return
	}

	handleSuccess(w, nil, "registered to activity successfully", http.StatusOK)
}

// UnregisterUserFromActivity godoc
// @Summary      Unregister from an activity
// @Description  Unregisters the authenticated user from an activity within an event
// @Tags         activities
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug"
// @Param        request body models.ActivityRegistrationRequest true "Activity registration info"
// @Success      200  {object}  NoDataSuccessResponse
// @Failure      400  {object}  ActivityStandardErrorResponse
// @Failure      401  {object}  ActivityStandardErrorResponse
// @Router       /events/{slug}/activity/unregister [post]
func (h *ActivityHandler) UnregisterUserFromActivity(w http.ResponseWriter, r *http.Request) {
	slug, err := extractSlugAndValidate(r)
	if err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	var reqBody models.ActivityRegistrationRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	if reqBody.ActivityID == "" {
		BadRequestError(w, NewErr("activity ID is required"), "activity")
		return
	}

	user, err := getUserFromContext(h.ActivityService.ActivityRepo.GetUserByID, r)
	if err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	if err := h.ActivityService.UnregisterUserFromActivity(user, slug, reqBody.ActivityID); err != nil {
		HandleErrMsg("error unregistering from activity", err, w).Stack("activity").BadRequest()
		return
	}

	handleSuccess(w, nil, "unregistered from activity successfully", http.StatusOK)
}

// RegisterUserToStandaloneActivity godoc
// @Summary      Register to a standalone activity
// @Description  Registers the authenticated user to a standalone activity without requiring event registration
// @Tags         activities
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug"
// @Param        request body models.ActivityRegistrationRequest true "Activity registration info"
// @Success      200  {object}  NoDataSuccessResponse
// @Failure      400  {object}  ActivityStandardErrorResponse
// @Failure      401  {object}  ActivityStandardErrorResponse
// @Router       /events/{slug}/activity/register-standalone [post]
func (h *ActivityHandler) RegisterUserToStandaloneActivity(w http.ResponseWriter, r *http.Request) {
	slug, err := extractSlugAndValidate(r)
	if err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	var reqBody models.ActivityRegistrationRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	if reqBody.ActivityID == "" {
		BadRequestError(w, NewErr("activity ID is required"), "activity")
		return
	}

	user, err := getUserFromContext(h.ActivityService.ActivityRepo.GetUserByID, r)
	if err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	if err := h.ActivityService.RegisterUserToStandaloneActivity(user, slug, reqBody.ActivityID); err != nil {
		HandleErrMsg("error registering to standalone activity", err, w).Stack("activity").BadRequest()
		return
	}

	handleSuccess(w, nil, "registered to standalone activity successfully", http.StatusOK)
}

// UnregisterUserFromStandaloneActivity godoc
// @Summary      Unregister from a standalone activity
// @Description  Unregisters the authenticated user from a standalone activity
// @Tags         activities
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug"
// @Param        request body models.ActivityRegistrationRequest true "Activity registration info"
// @Success      200  {object}  NoDataSuccessResponse
// @Failure      400  {object}  ActivityStandardErrorResponse
// @Failure      401  {object}  ActivityStandardErrorResponse
// @Router       /events/{slug}/activity/unregister-standalone [post]
func (h *ActivityHandler) UnregisterUserFromStandaloneActivity(w http.ResponseWriter, r *http.Request) {
	slug, err := extractSlugAndValidate(r)
	if err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	var reqBody models.ActivityRegistrationRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	if reqBody.ActivityID == "" {
		BadRequestError(w, NewErr("activity ID is required"), "activity")
		return
	}

	user, err := getUserFromContext(h.ActivityService.ActivityRepo.GetUserByID, r)
	if err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	if err := h.ActivityService.UnregisterUserFromStandaloneActivity(user, slug, reqBody.ActivityID); err != nil {
		HandleErrMsg("error unregistering from standalone activity", err, w).Stack("activity").BadRequest()
		return
	}

	handleSuccess(w, nil, "unregistered from standalone activity successfully", http.StatusOK)
}

// AttendActivity godoc
// @Summary      Mark attendance for an activity
// @Description  Marks a user as having attended an activity (admin only)
// @Tags         activities
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug"
// @Param        request body models.ActivityRegistrationRequest true "Attendance info"
// @Success      200  {object}  NoDataSuccessResponse
// @Failure      400  {object}  ActivityStandardErrorResponse
// @Failure      401  {object}  ActivityStandardErrorResponse
// @Failure      403  {object}  ActivityStandardErrorResponse
// @Router       /events/{slug}/activity/attend [post]
func (h *ActivityHandler) AttendActivity(w http.ResponseWriter, r *http.Request) {
	slug, err := extractSlugAndValidate(r)
	if err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	var reqBody models.ActivityRegistrationRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	if reqBody.ActivityID == "" || reqBody.UserID == "" {
		BadRequestError(w, NewErr("activity ID and user ID are required"), "activity")
		return
	}

	admin, err := getUserFromContext(h.ActivityService.ActivityRepo.GetUserByID, r)
	if err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	if err := h.ActivityService.AttendActivity(admin, slug, reqBody.ActivityID, reqBody.UserID); err != nil {
		HandleErrMsg("error marking attendance", err, w).Stack("activity").BadRequest()
		return
	}

	handleSuccess(w, nil, "attendance marked successfully", http.StatusOK)
}

// UnattendActivity godoc
// @Summary      Remove attendance for an activity
// @Description  Removes a user's attendance record for an activity (master admin only)
// @Tags         activities
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug"
// @Param        request body models.ActivityRegistrationRequest true "Attendance info"
// @Success      200  {object}  NoDataSuccessResponse
// @Failure      400  {object}  ActivityStandardErrorResponse
// @Failure      401  {object}  ActivityStandardErrorResponse
// @Failure      403  {object}  ActivityStandardErrorResponse
// @Router       /events/{slug}/activity/unattend [post]
func (h *ActivityHandler) UnattendActivity(w http.ResponseWriter, r *http.Request) {
	slug, err := extractSlugAndValidate(r)
	if err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	var reqBody models.ActivityRegistrationRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	if reqBody.ActivityID == "" || reqBody.UserID == "" {
		BadRequestError(w, NewErr("activity ID and user ID are required"), "activity")
		return
	}

	admin, err := getUserFromContext(h.ActivityService.ActivityRepo.GetUserByID, r)
	if err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	if err := h.ActivityService.UnattendActivity(admin, slug, reqBody.ActivityID, reqBody.UserID); err != nil {
		HandleErrMsg("error removing attendance", err, w).Stack("activity").BadRequest()
		return
	}

	handleSuccess(w, nil, "attendance removed successfully", http.StatusOK)
}

// GetActivityRegistrations godoc
// @Summary      Retrieves a list of registrations of an activity
// @Description  The end point returns a list of all registrations of a specified activity (all admins)
// @Tags         activities
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug"
// @Param        request body models.GetAttendeesRequest true "ActivityID"
// @Success      200  {object}  NoMessageSuccessResponse{data=models.ActivityRegistration}
// @Failure      400  {object}  ActivityStandardErrorResponse
// @Failure      401  {object}  ActivityStandardErrorResponse
// @Failure      403  {object}  ActivityStandardErrorResponse
// @Router       /events/{slug}/activity/attendees [get]
func (h *ActivityHandler) GetActivityRegistrations(w http.ResponseWriter, r *http.Request) {
	slug, err := extractSlugAndValidate(r)
	if err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	var reqBody models.GetAttendeesRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	if reqBody.ID == "" {
		BadRequestError(w, NewErr("activity ID is required"), "activity")
		return
	}

	admin, err := getUserFromContext(h.ActivityService.ActivityRepo.GetUserByID, r)
	if err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	var registrations []models.ActivityRegistration
	if registrations, err = h.ActivityService.GetActivityRegistrations(admin, slug, reqBody.ID); err != nil {
		HandleErrMsg("error getting registrations", err, w).Stack("activity").BadRequest()
		return
	}

	handleSuccess(w, registrations, "", http.StatusOK)
}

// GetUserAccesses godoc
// @Summary      Retrieves a list of accesses for a user
// @Description  The end point returns a list of all accesses for a specified user
// @Tags         activities
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Success      200  {object}  NoMessageSuccessResponse{data=[]models.AccessTarget}
// @Failure      400  {object}  ActivityStandardErrorResponse
// @Failure      401  {object}  ActivityStandardErrorResponse
// @Failure      403  {object}  ActivityStandardErrorResponse
// @Router       /user-accesses [get]
func (h *ActivityHandler) GetUserAccesses(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromContext(h.ActivityService.ActivityRepo.GetUserByID, r)
	if err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	var accesses []models.AccessTarget
	if accesses, err = h.ActivityService.GetUserAccesses(user.ID); err != nil {
		HandleErrMsg("error getting accesses", err, w).Stack("activity").BadRequest()
		return
	}

	handleSuccess(w, accesses, "", http.StatusOK)
}

// GetUserAccessesFromEvent godoc
// @Summary      Retrieves a list of accesses for a user from an event
// @Description  The end point returns a list of all accesses for a specified user from a specified event
// @Tags         activities
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug"
// @Success      200  {object}  NoMessageSuccessResponse{data=[]models.AccessTarget}
// @Failure      400  {object}  ActivityStandardErrorResponse
// @Failure      401  {object}  ActivityStandardErrorResponse
// @Failure      403  {object}  ActivityStandardErrorResponse
// @Router       /events/{slug}/accesses [get]
func (h *ActivityHandler) GetUserAccessesFromEvent(w http.ResponseWriter, r *http.Request) {
	slug, err := extractSlugAndValidate(r)
	if err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	user, err := getUserFromContext(h.ActivityService.ActivityRepo.GetUserByID, r)
	if err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	var accesses []models.AccessTarget
	if accesses, err = h.ActivityService.GetUserAccessesFromEvent(user.ID, slug); err != nil {
		HandleErrMsg("error getting accesses", err, w).Stack("activity").BadRequest()
		return
	}

	handleSuccess(w, accesses, "", http.StatusOK)
}

// GetUserActivities godoc
// @Summary      Retrieves a list of activities for a user
// @Description  The end point returns a list of all activities for a specified user
// @Tags         activities
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Success      200  {object}  NoMessageSuccessResponse{data=[]models.Activity}
// @Failure      400  {object}  ActivityStandardErrorResponse
// @Failure      401  {object}  ActivityStandardErrorResponse
// @Failure      403  {object}  ActivityStandardErrorResponse
// @Router       /user-activities [get]
func (h *ActivityHandler) GetUserActivities(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromContext(h.ActivityService.ActivityRepo.GetUserByID, r)
	if err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	var activities []models.Activity
	if activities, err = h.ActivityService.GetUserActivities(user); err != nil {
		HandleErrMsg("error getting activities", err, w).Stack("activity").BadRequest()
		return
	}

	handleSuccess(w, activities, "", http.StatusOK)
}

// GetUserActivitiesFromEvent godoc
// @Summary      Retrieves a list of activities for a user from an event
// @Description  The end point returns a list of all activities for a specified user from a specified event
// @Tags         activities
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        slug path string true "Event slug"
// @Success      200  {object}  NoMessageSuccessResponse{data=[]models.Activity}
// @Failure      400  {object}  ActivityStandardErrorResponse
// @Failure      401  {object}  ActivityStandardErrorResponse
// @Failure      403  {object}  ActivityStandardErrorResponse
// @Router       /events/{slug}/user-activities [get]
func (h *ActivityHandler) GetUserActivitiesFromEvent(w http.ResponseWriter, r *http.Request) {
	slug, err := extractSlugAndValidate(r)
	if err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	user, err := getUserFromContext(h.ActivityService.ActivityRepo.GetUserByID, r)
	if err != nil {
		BadRequestError(w, err, "activity")
		return
	}

	var activities []models.Activity
	if activities, err = h.ActivityService.GetUserActivitiesFromEvent(user, slug); err != nil {
		HandleErrMsg("error getting activities", err, w).Stack("activity").BadRequest()
		return
	}

	handleSuccess(w, activities, "", http.StatusOK)
}

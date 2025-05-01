package handlers

import (
	"errors"
	"net/http"
	"scti/internal/models"
	"scti/internal/services"
)

type ActivityHandler struct {
	ActivityService *services.ActivityService
}

func NewActivityHandler(activityService *services.ActivityService) *ActivityHandler {
	return &ActivityHandler{
		ActivityService: activityService,
	}
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
		handleError(w, err, http.StatusBadRequest)
		return
	}

	activities, err := h.ActivityService.GetAllActivitiesFromEvent(slug)
	if err != nil {
		handleError(w, errors.New("error getting activities: "+err.Error()), http.StatusBadRequest)
		return
	}

	handleSuccess(w, activities, "", http.StatusOK)
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
		handleError(w, err, http.StatusBadRequest)
		return
	}

	var reqBody models.CreateActivityRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	user, err := getUserFromContext(h.ActivityService.ActivityRepo.GetUserByID, r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	activity, err := h.ActivityService.CreateEventActivity(user, slug, reqBody)
	if err != nil {
		handleError(w, errors.New("error creating activity: "+err.Error()), http.StatusBadRequest)
		return
	}

	handleSuccess(w, activity, "", http.StatusOK)
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
		handleError(w, err, http.StatusBadRequest)
		return
	}

	var reqBody models.ActivityUpdateRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	if reqBody.ActivityID == "" {
		handleError(w, errors.New("activity ID is required"), http.StatusBadRequest)
		return
	}

	user, err := getUserFromContext(h.ActivityService.ActivityRepo.GetUserByID, r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	activity, err := h.ActivityService.UpdateEventActivity(user, slug, reqBody.ActivityID, reqBody.Activity)
	if err != nil {
		handleError(w, errors.New("error updating activity: "+err.Error()), http.StatusBadRequest)
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
		handleError(w, err, http.StatusBadRequest)
		return
	}

	var reqBody models.ActivityDeleteRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	if reqBody.ActivityID == "" {
		handleError(w, errors.New("activity ID is required"), http.StatusBadRequest)
		return
	}

	user, err := getUserFromContext(h.ActivityService.ActivityRepo.GetUserByID, r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	if err := h.ActivityService.DeleteEventActivity(user, slug, reqBody.ActivityID); err != nil {
		handleError(w, errors.New("error deleting activity: "+err.Error()), http.StatusBadRequest)
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
		handleError(w, err, http.StatusBadRequest)
		return
	}

	var reqBody models.ActivityRegistrationRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	if reqBody.ActivityID == "" {
		handleError(w, errors.New("activity ID is required"), http.StatusBadRequest)
		return
	}

	user, err := getUserFromContext(h.ActivityService.ActivityRepo.GetUserByID, r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	if err := h.ActivityService.RegisterUserToActivity(user, slug, reqBody.ActivityID); err != nil {
		handleError(w, errors.New("error registering to activity: "+err.Error()), http.StatusBadRequest)
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
		handleError(w, err, http.StatusBadRequest)
		return
	}

	var reqBody models.ActivityRegistrationRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	if reqBody.ActivityID == "" {
		handleError(w, errors.New("activity ID is required"), http.StatusBadRequest)
		return
	}

	user, err := getUserFromContext(h.ActivityService.ActivityRepo.GetUserByID, r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	if err := h.ActivityService.UnregisterUserFromActivity(user, slug, reqBody.ActivityID); err != nil {
		handleError(w, errors.New("error unregistering from activity: "+err.Error()), http.StatusBadRequest)
		return
	}

	handleSuccess(w, nil, "unregistered from activity successfully", http.StatusOK)
}

// TODO: Implement not permitting to register to acitivity if the user has another activity registered at the same time
// TODO: Implement not permitting to register to acitivity if the activity has already concluded
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
		handleError(w, err, http.StatusBadRequest)
		return
	}

	var reqBody models.ActivityRegistrationRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	if reqBody.ActivityID == "" {
		handleError(w, errors.New("activity ID is required"), http.StatusBadRequest)
		return
	}

	user, err := getUserFromContext(h.ActivityService.ActivityRepo.GetUserByID, r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	if err := h.ActivityService.RegisterUserToStandaloneActivity(user, slug, reqBody.ActivityID); err != nil {
		handleError(w, errors.New("error registering to standalone activity: "+err.Error()), http.StatusBadRequest)
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
		handleError(w, err, http.StatusBadRequest)
		return
	}

	var reqBody models.ActivityRegistrationRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	if reqBody.ActivityID == "" {
		handleError(w, errors.New("activity ID is required"), http.StatusBadRequest)
		return
	}

	user, err := getUserFromContext(h.ActivityService.ActivityRepo.GetUserByID, r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	if err := h.ActivityService.UnregisterUserFromStandaloneActivity(user, slug, reqBody.ActivityID); err != nil {
		handleError(w, errors.New("error unregistering from standalone activity: "+err.Error()), http.StatusBadRequest)
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
		handleError(w, err, http.StatusBadRequest)
		return
	}

	var reqBody models.ActivityRegistrationRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	if reqBody.ActivityID == "" || reqBody.UserID == "" {
		handleError(w, errors.New("activity ID and user ID are required"), http.StatusBadRequest)
		return
	}

	admin, err := getUserFromContext(h.ActivityService.ActivityRepo.GetUserByID, r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	if err := h.ActivityService.AttendActivity(admin, slug, reqBody.ActivityID, reqBody.UserID); err != nil {
		handleError(w, errors.New("error marking attendance: "+err.Error()), http.StatusBadRequest)
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
		handleError(w, err, http.StatusBadRequest)
		return
	}

	var reqBody models.ActivityRegistrationRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	if reqBody.ActivityID == "" || reqBody.UserID == "" {
		handleError(w, errors.New("activity ID and user ID are required"), http.StatusBadRequest)
		return
	}

	admin, err := getUserFromContext(h.ActivityService.ActivityRepo.GetUserByID, r)
	if err != nil {
		handleError(w, err, http.StatusBadRequest)
		return
	}

	if err := h.ActivityService.UnattendActivity(admin, slug, reqBody.ActivityID, reqBody.UserID); err != nil {
		handleError(w, errors.New("error removing attendance: "+err.Error()), http.StatusBadRequest)
		return
	}

	handleSuccess(w, nil, "attendance removed successfully", http.StatusOK)
}

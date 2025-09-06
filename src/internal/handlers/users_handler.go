package handlers

import (
	"errors"
	"net/http"

	"scti/internal/services"
)

type UsersHandler struct {
	UserService *services.UserService
}

func NewUsersHandler(userService *services.UserService) *UsersHandler {
	return &UsersHandler{UserService: userService}
}

type CreateEventCreatorRequest struct {
	Email string `json:"email"`
}

// @Summary      Create an event creator
// @Description  Create an event creator
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        Authorization header string true "Bearer {access_token}"
// @Param        Refresh header string true "Bearer {refresh_token}"
// @Param        request body CreateEventCreatorRequest true "Create event creator request"
// @Success      200  {object}  NoMessageSuccessResponse
// @Failure      400  {object}  AuthStandardErrorResponse
// @Router       /users/create-event-creator [post]
func (h *UsersHandler) CreateEventCreator(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromContext(h.UserService.UserRepo.GetUserByID, r)
	if err != nil {
		BadRequestError(w, err, "user")
		return
	}

	var reqBody CreateEventCreatorRequest
	if err := decodeRequestBody(r, &reqBody); err != nil {
		BadRequestError(w, err, "user")
		return
	}

	if reqBody.Email == "" {
		BadRequestError(w, errors.New("email is required"), "user")
		return
	}

	creator, err := h.UserService.CreateEventCreator(&user, reqBody.Email)
	if err != nil {
		HandleErrMsg("error creating event creator", err, w).Stack("users").BadRequest()
		return
	}

	handleSuccess(w, creator, "", http.StatusCreated)
}

// GetUserInfo godoc
// @Summary      Get user info from ID
// @Description  Get user info from ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id path string true "User ID"
// @Success      200  {object}  NoMessageSuccessResponse{data=models.UserInfo}
// @Failure      400  {object}  AuthStandardErrorResponse
// @Router       /users/{id} [get]
func (h *UsersHandler) GetUserInfoFromID(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")

	user, err := h.UserService.GetUserInfoFromID(userID)
	if err != nil {
		HandleErrMsg("error getting user info", err, w).Stack("users").BadRequest()
		return
	}

	handleSuccess(w, user, "", http.StatusOK)
}

type UserInfoBatch struct {
	Id_array []string `json:"id_array"`
}

// GetUserInfoBatched godoc
// @Summary      Get user info from ID array
// @Description  Get user info from ID array
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body UserInfoBatch true "Array list of all users IDs"
// @Success      200  {object}  NoMessageSuccessResponse{data=UserInfoBatch}
// @Failure      400  {object}  AuthStandardErrorResponse
// @Router       /users/batch [post]
func (h *UsersHandler) GetUserInfoBatched(w http.ResponseWriter, r *http.Request) {
	var reqBody *UserInfoBatch
	if err := decodeRequestBody(r, &reqBody); err != nil {
		BadRequestError(w, err, "user")
		return
	}
	if reqBody.Id_array == nil {
		BadRequestError(w, errors.New("array list is required"), "user")
		return
	}

	users_info, err := h.UserService.GetUserInfoFromIDBatch(reqBody.Id_array)
	if err != nil {
		HandleErrMsg("error getting users infos", err, w).Stack("users").BadRequest()
		return
	}
	handleSuccess(w, users_info, "", http.StatusCreated)
}

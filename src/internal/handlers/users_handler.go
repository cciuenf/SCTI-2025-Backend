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
// @Param        request body CreateEventCreatorRequest true "Create event creator request"
// @Success      200  {object}  NoMessageSuccessResponse
// @Failure      400  {object}  AuthStandardErrorResponse
// @Failure      401  {object}  AuthStandardErrorResponse
// @Failure      403  {object}  AuthStandardErrorResponse
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

package handlers

type StandardSuccessResponse struct {
	Success bool   `json:"success" example:"true"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty" example:"some success message"`
}

type NoMessageSuccessResponse struct {
	Success bool   `json:"success" example:"true"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

type NoDataSuccessResponse struct {
	Success bool   `json:"success" example:"true"`
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty" example:"some success message"`
}

type AuthStandardErrorResponse struct {
	Success bool     `json:"success" example:"false"`
	Stack   string   `json:"stack,omitempty" example:"auth-stack"`
	Errors  []string `json:"errors" example:"some error message"`
}

type EventStandardErrorResponse struct {
	Success bool     `json:"success" example:"false"`
	Stack   string   `json:"stack,omitempty" example:"event-stack"`
	Errors  []string `json:"errors" example:"some error message"`
}

type ActivityStandardErrorResponse struct {
	Success bool     `json:"success" example:"false"`
	Stack   string   `json:"stack,omitempty" example:"activity-stack"`
	Errors  []string `json:"errors" example:"some error message"`
}

type ProductStandardErrorResponse struct {
	Success bool     `json:"success" example:"false"`
	Stack   string   `json:"stack,omitempty" example:"product-stack"`
	Errors  []string `json:"errors" example:"some error message"`
}

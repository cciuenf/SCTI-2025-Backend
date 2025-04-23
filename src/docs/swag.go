package docs

// @title           SCTI 2025 API
// @version         1.0
// @description     API Server for SCTI 2025 Event Management
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  your-email@domain.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

type StandardResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type UserRegisterRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"password123"`
	Name     string `json:"name" example:"John"`
	LastName string `json:"lastName" example:"Doe"`
}

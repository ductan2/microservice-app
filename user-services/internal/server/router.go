package server

import (
	"user-services/internal/api/controllers"
	"user-services/internal/api/repositories"
	routers "user-services/internal/api/routes"
	"user-services/internal/api/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Deps struct {
	DB *gorm.DB
}

func NewRouter(deps Deps) *gin.Engine {
	r := gin.New()

	// Middlewares
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/health", controllers.Health)

	// Initialize repositories
	userRepo := repositories.NewUserRepository(deps.DB)
	userProfileRepo := repositories.NewUserProfileRepository(deps.DB)
	auditLogRepo := repositories.NewAuditLogRepository(deps.DB)
	outboxRepo := repositories.NewOutboxRepository(deps.DB)
	sessionRepo := repositories.NewSessionRepository(deps.DB)
	refreshTokenRepo := repositories.NewRefreshTokenRepository(deps.DB)
	mfaRepo := repositories.NewMFARepository(deps.DB)
	loginAttemptRepo := repositories.NewLoginAttemptRepository(deps.DB)
	passwordResetRepo := repositories.NewPasswordResetRepository(deps.DB)

	// Initialize services
	authService := services.NewAuthService(userRepo, userProfileRepo, auditLogRepo, outboxRepo, sessionRepo, refreshTokenRepo, mfaRepo, loginAttemptRepo)
	profileService := services.NewUserProfileService(userProfileRepo)
	passwordService := services.NewPasswordService(userRepo, passwordResetRepo, auditLogRepo, outboxRepo, userProfileRepo)

	// Initialize controllers
	authCtrl := controllers.NewAuthController(authService)
	profileCtrl := controllers.NewProfileController(profileService)
	passwordCtrl := controllers.NewPasswordController(passwordService)

	api := r.Group("/api/v1")
	{
		routers.RegisterAuthRoutes(api, *authCtrl)
		routers.RegisterProfileRoutes(api, profileCtrl)
		routers.RegisterPasswordRoutes(api, passwordCtrl)
	}

	return r
}

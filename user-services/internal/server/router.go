package server

import (
	"user-services/internal/api/controllers"
	"user-services/internal/api/repositories"
	routers "user-services/internal/api/routes"
	"user-services/internal/api/services"
	"user-services/internal/cache"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Deps struct {
	DB          *gorm.DB
	RedisClient *redis.Client
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

	// Initialize Redis session cache
	sessionCache := cache.NewSessionCache(deps.RedisClient)

	// Initialize services
	authService := services.NewAuthService(userRepo, userProfileRepo, auditLogRepo, outboxRepo, sessionRepo, refreshTokenRepo, mfaRepo, loginAttemptRepo, sessionCache)
	profileService := services.NewUserProfileService(userProfileRepo)
	currentUserService := services.NewCurrentUserService(userRepo)
	passwordService := services.NewPasswordService(userRepo, passwordResetRepo, auditLogRepo, outboxRepo, userProfileRepo)
	mfaService := services.NewMFAService(mfaRepo, userRepo)
	sessionService := services.NewSessionService(sessionRepo, sessionCache)
	userService := services.NewUserService(userRepo)

	// Initialize controllers
	userCtrl := controllers.NewUserController(authService, profileService, currentUserService, userService)
	passwordCtrl := controllers.NewPasswordController(passwordService)
	mfaCtrl := controllers.NewMFAController(mfaService)
	sessionCtrl := controllers.NewSessionController(sessionService)

	api := r.Group("/api/v1")
	{
		routers.RegisterUserRoutes(api, userCtrl)
		routers.RegisterPasswordRoutes(api, passwordCtrl, sessionCache)
		routers.RegisterMFARoutes(api, mfaCtrl, sessionCache)
		routers.RegisterSessionRoutes(api, sessionCtrl, sessionCache)
	}

	return r
}

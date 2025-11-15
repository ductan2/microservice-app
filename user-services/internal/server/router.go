package server

import (
	"user-services/internal/api/controllers"
	"user-services/internal/api/middleware"
	"user-services/internal/api/repositories"
	routers "user-services/internal/api/routes"
	"user-services/internal/api/services"
	"user-services/internal/cache"
	"user-services/internal/config"

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

	// Load configuration
	cfg := config.GetConfig()

	// Initialize rate limiter
	rateLimiter := middleware.NewRedisRateLimiter(deps.RedisClient, cfg)

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
	activitySessionRepo := repositories.NewActivitySessionRepository(deps.DB)
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
	activitySessionService := services.NewActivitySessionService(activitySessionRepo, deps.DB)
	// Initialize services
	tokenService := services.NewTokenService(refreshTokenRepo, sessionRepo)

	// Initialize controllers
	userCtrl := controllers.NewUserController(authService, profileService, currentUserService, userService, sessionService, rateLimiter, deps.RedisClient)
	passwordCtrl := controllers.NewPasswordController(passwordService)
	mfaCtrl := controllers.NewMFAController(mfaService)
	sessionCtrl := controllers.NewSessionController(sessionService)
	activitySessionCtrl := controllers.NewActivitySessionController(activitySessionService)

	api := r.Group("/api/v1")
	{
		routers.RegisterUserRoutes(api, userCtrl, rateLimiter, cfg)
		routers.RegisterPasswordRoutes(api, passwordCtrl, sessionCache, rateLimiter, cfg)
		routers.RegisterAuthRoutes(api, controllers.NewTokenController(tokenService, rateLimiter), rateLimiter, cfg)
		routers.RegisterMFARoutes(api, mfaCtrl, sessionCache)
		routers.RegisterSessionRoutes(api, sessionCtrl, sessionCache)
		routers.RegisterActivitySessionRoutes(api, activitySessionCtrl, sessionCache)
	}

	return r
}

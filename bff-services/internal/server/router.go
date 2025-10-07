package server

import (
	"bff-services/internal/api/controllers"
	"bff-services/internal/cache"
	"bff-services/internal/config"
	middleware "bff-services/internal/middlewares"
	"bff-services/internal/services"

	"github.com/gin-gonic/gin"
)

type Deps struct {
	UserService    services.UserService
	ContentService services.ContentService
	LessonService  services.LessonService
	SessionCache   *cache.SessionCache
}

func NewRouter(deps Deps) *gin.Engine {
	r := gin.New()

	// Middlewares
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(func(c *gin.Context) {
		origin := config.GetCORSOrigin()
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/health", controllers.Health)

	var (
		authCtrl     *controllers.AuthController
		profileCtrl  *controllers.ProfileController
		passwordCtrl *controllers.PasswordController
		mfaCtrl      *controllers.MFAController
		sessionCtrl  *controllers.SessionController
		contentCtrl  *controllers.ContentController
		usersCtrl    *controllers.UsersController
	)
	if deps.UserService != nil {
		authCtrl = controllers.NewAuthController(deps.UserService)
		profileCtrl = controllers.NewProfileController(deps.UserService)
		passwordCtrl = controllers.NewPasswordController(deps.UserService)
		mfaCtrl = controllers.NewMFAController(deps.UserService)
		sessionCtrl = controllers.NewSessionController(deps.UserService)
	}
	if deps.ContentService != nil {
		contentCtrl = controllers.NewContentController(deps.ContentService)
	}
	if deps.UserService != nil && deps.LessonService != nil {
		usersCtrl = controllers.NewUsersController(deps.UserService, deps.LessonService)
	}

	api := r.Group("/api/v1")
	{
		api.GET("/health", controllers.Health)
		if authCtrl != nil {
			api.POST("/register", authCtrl.Register)
			api.POST("/login", authCtrl.Login)
			api.POST("/logout", authCtrl.Logout)
			api.GET("/verify-email", authCtrl.VerifyEmail)
		}
		if profileCtrl != nil {
			// Profile routes require authentication
			profile := api.Group("/profile")
			if deps.SessionCache != nil {
				profile.Use(middleware.AuthRequired(deps.SessionCache))
			}
			{
				profile.GET("", profileCtrl.GetProfile)
				profile.PUT("", profileCtrl.UpdateProfile)
				profile.GET("/check-auth", profileCtrl.CheckAuth)
			}
		}
		if passwordCtrl != nil {
			api.POST("/password/reset/request", passwordCtrl.RequestReset)
			api.POST("/password/reset/confirm", passwordCtrl.ConfirmReset)
			api.POST("/password/change", passwordCtrl.ChangePassword)
		}
		if mfaCtrl != nil {
			api.POST("/mfa/setup", mfaCtrl.Setup)
			api.POST("/mfa/verify", mfaCtrl.Verify)
			api.POST("/mfa/disable", mfaCtrl.Disable)
			api.GET("/mfa/methods", mfaCtrl.Methods)
		}
		if sessionCtrl != nil {
			api.GET("/sessions", sessionCtrl.List)
			api.DELETE("/sessions/:id", sessionCtrl.Delete)
			api.POST("/sessions/revoke-all", sessionCtrl.RevokeAll)
		}
		if contentCtrl != nil {
			content := api.Group("/content")
			{
				content.POST("/graphql", contentCtrl.ProxyGraphQL)
			}
		}
		if usersCtrl != nil {
			api.GET("/users", usersCtrl.ListUsersWithProgress)
		}
	}

	return r
}

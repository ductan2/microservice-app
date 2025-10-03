package server

import (
	"bff-services/internal/api/controllers"
	"bff-services/internal/services"

	"github.com/gin-gonic/gin"
)

type Deps struct {
	UserService services.UserService
}

func NewRouter(deps Deps) *gin.Engine {
	r := gin.New()

	// Middlewares
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/health", controllers.Health)

	var (
		authCtrl     *controllers.AuthController
		profileCtrl  *controllers.ProfileController
		passwordCtrl *controllers.PasswordController
		mfaCtrl      *controllers.MFAController
		sessionCtrl  *controllers.SessionController
	)
	if deps.UserService != nil {
		authCtrl = controllers.NewAuthController(deps.UserService)
		profileCtrl = controllers.NewProfileController(deps.UserService)
		passwordCtrl = controllers.NewPasswordController(deps.UserService)
		mfaCtrl = controllers.NewMFAController(deps.UserService)
		sessionCtrl = controllers.NewSessionController(deps.UserService)
	}

	api := r.Group("/api/v1")
	{
		api.GET("/health", controllers.Health)
		if authCtrl != nil {
			api.POST("/register", authCtrl.Register)
			api.POST("/login", authCtrl.Login)
			api.POST("/logout", authCtrl.Logout)
			api.GET("/verify-email", authCtrl.VerifyEmail)

			// Backwards compatibility for existing clients
			api.POST("/user/register", authCtrl.Register)
			api.POST("/user/login", authCtrl.Login)
		}
		if profileCtrl != nil {
			api.GET("/profile", profileCtrl.GetProfile)
			api.PUT("/profile", profileCtrl.UpdateProfile)
			api.GET("/profile/check-auth", profileCtrl.CheckAuth)
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
	}

	return r
}

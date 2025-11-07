package server

import (
	"bff-services/internal/api/controllers"
)

// initControllers initializes all controllers based on available services
func initControllers(deps Deps) *controllers.Controllers {
	ctrl := &controllers.Controllers{}

	// Initialize user-related controllers
	if deps.UserService != nil {
		ctrl.Password = controllers.NewPasswordController(deps.UserService)
		ctrl.MFA = controllers.NewMFAController(deps.UserService)
		ctrl.Session = controllers.NewSessionController(deps.UserService)
		ctrl.ActivitySession = controllers.NewActivitySessionController(deps.UserService)
	}

	// Initialize user controller (requires both UserService and LessonService)
	if deps.UserService != nil && deps.LessonService != nil {
		ctrl.User = controllers.NewUserController(deps.UserService, deps.LessonService)
	}

	// Initialize other service controllers
	if deps.ContentService != nil {
		ctrl.Content = controllers.NewContentController(deps.ContentService)
	}

	if deps.NotificationService != nil {
		ctrl.Notification = controllers.NewNotificationController(deps.NotificationService)
	}

	if deps.LessonService != nil {
		ctrl.Lesson = controllers.NewLessonController(deps.LessonService)
	}

	if deps.QuizAttemptService != nil {
		ctrl.QuizAttempt = controllers.NewQuizAttemptController(deps.QuizAttemptService)
	}

	return ctrl
}

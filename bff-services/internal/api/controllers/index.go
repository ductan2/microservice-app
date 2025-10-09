package controllers	

import (
)

// Controllers holds all initialized controllers
type Controllers struct {
	User         *UserController
	Password     *PasswordController
	MFA          *MFAController
	Session      *SessionController
	Content      *ContentController
	Notification *NotificationController
}

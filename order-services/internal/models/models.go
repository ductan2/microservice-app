package models

// This file imports all models to ensure they're registered with GORM

// Order models
//go:generate go run -mod=mod github.com/golang/mock/mockgen -source=order_model.go -destination=../api/mocks/order_model_mock.go -package=mocks

// Payment models
//go:generate go run -mod=mod github.com/golang/mock/mockgen -source=payment_model.go -destination=../api/mocks/payment_model_mock.go -package=mocks

// Coupon models
//go:generate go run -mod=mod github.com/golang/mock/mockgen -source=coupon_model.go -destination=../api/mocks/coupon_model_mock.go -package=mocks

// Advanced feature models
//go:generate go run -mod=mod github.com/golang/mock/mockgen -source=advanced_model.go -destination=../api/mocks/advanced_model_mock.go -package=mocks

// Outbox model
//go:generate go run -mod=mod github.com/golang/mock/mockgen -source=outbox_model.go -destination=../api/mocks/outbox_model_mock.go -package=mocks

// AllModels returns all model instances for auto-migration
func AllModels() []interface{} {
	return []interface{}{
		// Core models
		&Order{},
		&OrderItem{},
		&Payment{},
		&WebhookEvent{},

		// Coupon models
		&Coupon{},
		&CouponRedemption{},

		// Advanced feature models
		&Invoice{},
		&RefundRequest{},
		&FraudLog{},

		// Event publishing
		&Outbox{},
	}
}
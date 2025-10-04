package resolver

import (
	"content-services/internal/repository"
	"content-services/internal/service"
	"content-services/internal/taxonomy"

	"go.mongodb.org/mongo-driver/mongo"
)

// Resolver serves as dependency injection for your app, add any dependencies you require here.
type Resolver struct {
	DB               *mongo.Database
	Taxonomy         *taxonomy.Store
	Media            service.MediaService
	LessonService    service.LessonService
	QuizService      service.QuizService
	FlashcardService service.FlashcardService
	TagRepo          repository.TagRepository
}

package repository

import (
	"content-services/internal/models"
	"content-services/internal/types"
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CourseFilter struct {
	TopicID      *uuid.UUID
	LevelID      *uuid.UUID
	InstructorID *uuid.UUID
	IsPublished  *bool
	IsFeatured   *bool
	Search       string
}

type CourseRepository interface {
	Create(ctx context.Context, course *models.Course) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Course, error)
	List(ctx context.Context, filter *CourseFilter, sort *SortOption, limit, offset int) ([]models.Course, int64, error)
	Update(ctx context.Context, course *models.Course) error
	Publish(ctx context.Context, id uuid.UUID, publishedAt time.Time) error
	Unpublish(ctx context.Context, id uuid.UUID, updatedAt time.Time) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type CourseLessonFilter struct {
	IsRequired *bool
}

type CourseLessonRepository interface {
	Create(ctx context.Context, lesson *models.CourseLesson) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.CourseLesson, error)
	ListByCourseID(ctx context.Context, courseID uuid.UUID, filter *CourseLessonFilter, sort *SortOption, limit, offset int) ([]models.CourseLesson, int64, error)
	Update(ctx context.Context, lesson *models.CourseLesson) error
	Reorder(ctx context.Context, courseID uuid.UUID, lessonIDs []uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByCourseID(ctx context.Context, courseID uuid.UUID) error
}

type courseRepository struct {
	collection *mongo.Collection
}

type courseLessonRepository struct {
	collection *mongo.Collection
}

func NewCourseRepository(db *mongo.Database) CourseRepository {
	return &courseRepository{collection: db.Collection("courses")}
}

func NewCourseLessonRepository(db *mongo.Database) CourseLessonRepository {
	return &courseLessonRepository{collection: db.Collection("course_lessons")}
}

type courseDoc struct {
	ID            string     `bson:"_id"`
	Title         string     `bson:"title"`
	Description   string     `bson:"description"`
	TopicID       *string    `bson:"topic_id,omitempty"`
	LevelID       *string    `bson:"level_id,omitempty"`
	InstructorID  *string    `bson:"instructor_id,omitempty"`
	ThumbnailURL  string     `bson:"thumbnail_url,omitempty"`
	IsPublished   bool       `bson:"is_published"`
	IsFeatured    bool       `bson:"is_featured"`
	Price         float64    `bson:"price,omitempty"`
	DurationHours int        `bson:"duration_hours,omitempty"`
	AverageRating float64    `bson:"average_rating,omitempty"`
	ReviewCount   int64      `bson:"review_count,omitempty"`
	CreatedAt     time.Time  `bson:"created_at"`
	UpdatedAt     time.Time  `bson:"updated_at"`
	PublishedAt   *time.Time `bson:"published_at,omitempty"`
}

func courseDocFromModel(course *models.Course) *courseDoc {
	doc := &courseDoc{
		ID:            course.ID.String(),
		Title:         course.Title,
		Description:   course.Description,
		ThumbnailURL:  course.ThumbnailURL,
		IsPublished:   course.IsPublished,
		IsFeatured:    course.IsFeatured,
		Price:         course.Price,
		DurationHours: course.DurationHours,
		AverageRating: course.AverageRating,
		ReviewCount:   int64(course.ReviewCount),
		CreatedAt:     course.CreatedAt,
		UpdatedAt:     course.UpdatedAt,
	}

	if course.TopicID != nil {
		id := course.TopicID.String()
		doc.TopicID = &id
	}
	if course.LevelID != nil {
		id := course.LevelID.String()
		doc.LevelID = &id
	}
	if course.InstructorID != nil {
		id := course.InstructorID.String()
		doc.InstructorID = &id
	}
	if course.PublishedAt.Valid {
		t := course.PublishedAt.Time
		doc.PublishedAt = &t
	}

	return doc
}

func (d *courseDoc) toModel() *models.Course {
	course := &models.Course{
		ID:            uuid.MustParse(d.ID),
		Title:         d.Title,
		Description:   d.Description,
		ThumbnailURL:  d.ThumbnailURL,
		IsPublished:   d.IsPublished,
		IsFeatured:    d.IsFeatured,
		Price:         d.Price,
		DurationHours: d.DurationHours,
		AverageRating: d.AverageRating,
		ReviewCount:   int(d.ReviewCount),
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}

	if d.TopicID != nil && *d.TopicID != "" {
		if id, err := uuid.Parse(*d.TopicID); err == nil {
			course.TopicID = &id
		}
	}

	if d.LevelID != nil && *d.LevelID != "" {
		if id, err := uuid.Parse(*d.LevelID); err == nil {
			course.LevelID = &id
		}
	}

	if d.InstructorID != nil && *d.InstructorID != "" {
		if id, err := uuid.Parse(*d.InstructorID); err == nil {
			course.InstructorID = &id
		}
	}

	if d.PublishedAt != nil {
		course.PublishedAt = sqlNullTime(*d.PublishedAt)
	}

	return course
}

func (r *courseRepository) Create(ctx context.Context, course *models.Course) error {
	doc := courseDocFromModel(course)
	_, err := r.collection.InsertOne(ctx, doc)
	return err
}

func (r *courseRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Course, error) {
	var doc courseDoc
	err := r.collection.FindOne(ctx, bson.M{"_id": id.String()}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, types.ErrCourseNotFound
		}
		return nil, err
	}
	return doc.toModel(), nil
}

func (r *courseRepository) List(ctx context.Context, filter *CourseFilter, sort *SortOption, limit, offset int) ([]models.Course, int64, error) {
	filterDoc := bson.M{}

	if filter != nil {
		if filter.TopicID != nil {
			filterDoc["topic_id"] = filter.TopicID.String()
		}
		if filter.LevelID != nil {
			filterDoc["level_id"] = filter.LevelID.String()
		}
		if filter.InstructorID != nil {
			filterDoc["instructor_id"] = filter.InstructorID.String()
		}
		if filter.IsPublished != nil {
			filterDoc["is_published"] = *filter.IsPublished
		}
		if filter.IsFeatured != nil {
			filterDoc["is_featured"] = *filter.IsFeatured
		}
		if filter.Search != "" {
			search := strings.TrimSpace(filter.Search)
			if search != "" {
				filterDoc["$or"] = []bson.M{
					{"title": bson.M{"$regex": search, "$options": "i"}},
					{"description": bson.M{"$regex": search, "$options": "i"}},
				}
			}
		}
	}

	sortField, sortDir := "created_at", SortDescending
	if sort != nil {
		sortField, sortDir = sort.apply(sortField, sortDir)
	}

	findOpts := options.Find().
		SetSort(bson.D{{Key: sortField, Value: int(sortDir)}})

	if offset > 0 {
		findOpts.SetSkip(int64(offset))
	}
	if limit > 0 {
		findOpts.SetLimit(int64(limit))
	}

	cursor, err := r.collection.Find(ctx, filterDoc, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var items []models.Course
	for cursor.Next(ctx) {
		var doc courseDoc
		if err := cursor.Decode(&doc); err != nil {
			return nil, 0, err
		}
		items = append(items, *doc.toModel())
	}
	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}

	total, err := r.collection.CountDocuments(ctx, filterDoc)
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *courseRepository) Update(ctx context.Context, course *models.Course) error {
	set := bson.M{
		"title":          course.Title,
		"description":    course.Description,
		"thumbnail_url":  course.ThumbnailURL,
		"is_featured":    course.IsFeatured,
		"price":          course.Price,
		"duration_hours": course.DurationHours,
		"updated_at":     course.UpdatedAt,
	}

	unset := bson.M{}

	if course.TopicID != nil {
		set["topic_id"] = course.TopicID.String()
	} else {
		unset["topic_id"] = ""
	}

	if course.LevelID != nil {
		set["level_id"] = course.LevelID.String()
	} else {
		unset["level_id"] = ""
	}

	if course.InstructorID != nil {
		set["instructor_id"] = course.InstructorID.String()
	} else {
		unset["instructor_id"] = ""
	}

	update := bson.M{"$set": set}
	if len(unset) > 0 {
		update["$unset"] = unset
	}

	res, err := r.collection.UpdateOne(ctx, bson.M{"_id": course.ID.String()}, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return types.ErrCourseNotFound
	}
	return nil
}

func (r *courseRepository) Publish(ctx context.Context, id uuid.UUID, publishedAt time.Time) error {
	update := bson.M{
		"$set": bson.M{
			"is_published": true,
			"published_at": publishedAt,
			"updated_at":   publishedAt,
		},
	}
	res, err := r.collection.UpdateOne(ctx, bson.M{"_id": id.String()}, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return types.ErrCourseNotFound
	}
	return nil
}

func (r *courseRepository) Unpublish(ctx context.Context, id uuid.UUID, updatedAt time.Time) error {
	update := bson.M{
		"$set": bson.M{
			"is_published": false,
			"updated_at":   updatedAt,
		},
		"$unset": bson.M{
			"published_at": "",
		},
	}
	res, err := r.collection.UpdateOne(ctx, bson.M{"_id": id.String()}, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return types.ErrCourseNotFound
	}
	return nil
}

func (r *courseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id.String()})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return types.ErrCourseNotFound
	}
	return nil
}

type courseLessonDoc struct {
	ID         string    `bson:"_id"`
	CourseID   string    `bson:"course_id"`
	LessonID   string    `bson:"lesson_id"`
	Ord        int       `bson:"ord"`
	IsRequired bool      `bson:"is_required"`
	CreatedAt  time.Time `bson:"created_at"`
}

func courseLessonDocFromModel(lesson *models.CourseLesson) *courseLessonDoc {
	return &courseLessonDoc{
		ID:         lesson.ID.String(),
		CourseID:   lesson.CourseID.String(),
		LessonID:   lesson.LessonID.String(),
		Ord:        lesson.Ord,
		IsRequired: lesson.IsRequired,
		CreatedAt:  lesson.CreatedAt,
	}
}

func (d *courseLessonDoc) toModel() *models.CourseLesson {
	return &models.CourseLesson{
		ID:         uuid.MustParse(d.ID),
		CourseID:   uuid.MustParse(d.CourseID),
		LessonID:   uuid.MustParse(d.LessonID),
		Ord:        d.Ord,
		IsRequired: d.IsRequired,
		CreatedAt:  d.CreatedAt,
	}
}

func (r *courseLessonRepository) Create(ctx context.Context, lesson *models.CourseLesson) error {
	doc := courseLessonDocFromModel(lesson)
	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return types.ErrCourseLessonExists
		}
		return err
	}
	return nil
}

func (r *courseLessonRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.CourseLesson, error) {
	var doc courseLessonDoc
	err := r.collection.FindOne(ctx, bson.M{"_id": id.String()}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, types.ErrCourseLessonNotFound
		}
		return nil, err
	}
	return doc.toModel(), nil
}

func (r *courseLessonRepository) ListByCourseID(ctx context.Context, courseID uuid.UUID, filter *CourseLessonFilter, sort *SortOption, limit, offset int) ([]models.CourseLesson, int64, error) {
	filterDoc := bson.M{"course_id": courseID.String()}
	if filter != nil && filter.IsRequired != nil {
		filterDoc["is_required"] = *filter.IsRequired
	}

	sortField, sortDir := "ord", SortAscending
	if sort != nil {
		sortField, sortDir = sort.apply(sortField, sortDir)
	}

	findOpts := options.Find().
		SetSort(bson.D{{Key: sortField, Value: int(sortDir)}})

	if offset > 0 {
		findOpts.SetSkip(int64(offset))
	}
	if limit > 0 {
		findOpts.SetLimit(int64(limit))
	}

	cursor, err := r.collection.Find(ctx, filterDoc, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var items []models.CourseLesson
	for cursor.Next(ctx) {
		var doc courseLessonDoc
		if err := cursor.Decode(&doc); err != nil {
			return nil, 0, err
		}
		items = append(items, *doc.toModel())
	}
	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}

	total, err := r.collection.CountDocuments(ctx, filterDoc)
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *courseLessonRepository) Update(ctx context.Context, lesson *models.CourseLesson) error {
	update := bson.M{
		"$set": bson.M{
			"ord":         lesson.Ord,
			"is_required": lesson.IsRequired,
		},
	}

	res, err := r.collection.UpdateOne(ctx, bson.M{"_id": lesson.ID.String()}, update)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return types.ErrCourseLessonExists
		}
		return err
	}
	if res.MatchedCount == 0 {
		return types.ErrCourseLessonNotFound
	}
	return nil
}

func (r *courseLessonRepository) Reorder(ctx context.Context, courseID uuid.UUID, lessonIDs []uuid.UUID) error {
	for index, id := range lessonIDs {
		res, err := r.collection.UpdateOne(ctx, bson.M{"_id": id.String(), "course_id": courseID.String()}, bson.M{"$set": bson.M{"ord": index + 1}})
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				return types.ErrCourseLessonExists
			}
			return err
		}
		if res.MatchedCount == 0 {
			return types.ErrCourseLessonNotFound
		}
	}
	return nil
}

func (r *courseLessonRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id.String()})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return types.ErrCourseLessonNotFound
	}
	return nil
}

func (r *courseLessonRepository) DeleteByCourseID(ctx context.Context, courseID uuid.UUID) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"course_id": courseID.String()})
	return err
}

func sqlNullTime(t time.Time) sql.NullTime {
	return sql.NullTime{Time: t, Valid: true}
}

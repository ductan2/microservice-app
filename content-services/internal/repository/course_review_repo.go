package repository

import (
	"content-services/internal/models"
	"content-services/internal/types"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CourseReviewRepository defines persistence operations for course reviews.
type CourseReviewRepository interface {
	Upsert(ctx context.Context, review *models.CourseReview) error
	GetByCourseAndUser(ctx context.Context, courseID, userID uuid.UUID) (*models.CourseReview, error)
	ListByCourse(ctx context.Context, courseID uuid.UUID, limit, offset int) ([]models.CourseReview, int64, error)
	DeleteByCourseAndUser(ctx context.Context, courseID, userID uuid.UUID) error
}

// CourseEnrollmentRepository exposes read access to course enrollment data.
type CourseEnrollmentRepository interface {
	IsUserEnrolled(ctx context.Context, courseID, userID uuid.UUID) (bool, error)
}

type courseReviewRepository struct {
	reviewsCollection *mongo.Collection
	courseCollection  *mongo.Collection
}

type courseEnrollmentRepository struct {
	collection *mongo.Collection
}

// NewCourseReviewRepository constructs a CourseReviewRepository backed by MongoDB.
func NewCourseReviewRepository(db *mongo.Database) CourseReviewRepository {
	return &courseReviewRepository{
		reviewsCollection: db.Collection("course_reviews"),
		courseCollection:  db.Collection("courses"),
	}
}

// NewCourseEnrollmentRepository constructs a CourseEnrollmentRepository backed by MongoDB.
func NewCourseEnrollmentRepository(db *mongo.Database) CourseEnrollmentRepository {
	return &courseEnrollmentRepository{collection: db.Collection("course_enrollments")}
}

type courseReviewDoc struct {
	ID        string    `bson:"_id"`
	CourseID  string    `bson:"course_id"`
	UserID    string    `bson:"user_id"`
	Rating    int       `bson:"rating"`
	Comment   string    `bson:"comment"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

func courseReviewDocFromModel(review *models.CourseReview) *courseReviewDoc {
	return &courseReviewDoc{
		ID:        review.ID.String(),
		CourseID:  review.CourseID.String(),
		UserID:    review.UserID.String(),
		Rating:    review.Rating,
		Comment:   review.Comment,
		CreatedAt: review.CreatedAt,
		UpdatedAt: review.UpdatedAt,
	}
}

func (d *courseReviewDoc) toModel() *models.CourseReview {
	return &models.CourseReview{
		ID:        uuid.MustParse(d.ID),
		CourseID:  uuid.MustParse(d.CourseID),
		UserID:    uuid.MustParse(d.UserID),
		Rating:    d.Rating,
		Comment:   d.Comment,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func (r *courseReviewRepository) Upsert(ctx context.Context, review *models.CourseReview) error {
	if review == nil {
		return errors.New("course review is nil")
	}

	doc := courseReviewDocFromModel(review)
	filter := bson.M{"course_id": doc.CourseID, "user_id": doc.UserID}
	update := bson.M{
		"$set": bson.M{
			"rating":     doc.Rating,
			"comment":    doc.Comment,
			"updated_at": doc.UpdatedAt,
		},
		"$setOnInsert": bson.M{
			"_id":        doc.ID,
			"course_id":  doc.CourseID,
			"user_id":    doc.UserID,
			"created_at": doc.CreatedAt,
		},
	}

	opts := options.Update().SetUpsert(true)
	if _, err := r.reviewsCollection.UpdateOne(ctx, filter, update, opts); err != nil {
		return err
	}

	return r.recalculateAggregates(ctx, review.CourseID)
}

func (r *courseReviewRepository) GetByCourseAndUser(ctx context.Context, courseID, userID uuid.UUID) (*models.CourseReview, error) {
	filter := bson.M{"course_id": courseID.String(), "user_id": userID.String()}
	var doc courseReviewDoc
	err := r.reviewsCollection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, types.ErrCourseReviewNotFound
		}
		return nil, err
	}
	return doc.toModel(), nil
}

func (r *courseReviewRepository) ListByCourse(ctx context.Context, courseID uuid.UUID, limit, offset int) ([]models.CourseReview, int64, error) {
	filter := bson.M{"course_id": courseID.String()}
	findOptions := options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}})
	if limit > 0 {
		findOptions.SetLimit(int64(limit))
	}
	if offset > 0 {
		findOptions.SetSkip(int64(offset))
	}

	cursor, err := r.reviewsCollection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	reviews := make([]models.CourseReview, 0)
	for cursor.Next(ctx) {
		var doc courseReviewDoc
		if err := cursor.Decode(&doc); err != nil {
			return nil, 0, err
		}
		reviews = append(reviews, *doc.toModel())
	}
	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}

	count, err := r.reviewsCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return reviews, count, nil
}

func (r *courseReviewRepository) DeleteByCourseAndUser(ctx context.Context, courseID, userID uuid.UUID) error {
	filter := bson.M{"course_id": courseID.String(), "user_id": userID.String()}
	res, err := r.reviewsCollection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return types.ErrCourseReviewNotFound
	}
	return r.recalculateAggregates(ctx, courseID)
}

func (r *courseReviewRepository) recalculateAggregates(ctx context.Context, courseID uuid.UUID) error {
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"course_id": courseID.String()}}},
		bson.D{{Key: "$group", Value: bson.M{
			"_id":   "$course_id",
			"avg":   bson.M{"$avg": "$rating"},
			"count": bson.M{"$sum": 1},
		}}},
	}

	var avg float64
	var count int64

	cursor, err := r.reviewsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		var result struct {
			Avg   float64 `bson:"avg"`
			Count int64   `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			return err
		}
		avg = result.Avg
		count = result.Count
	}
	if err := cursor.Err(); err != nil {
		return err
	}

	update := bson.M{
		"average_rating": avg,
		"review_count":   count,
	}
	if count == 0 {
		update["average_rating"] = 0.0
	}

	_, err = r.courseCollection.UpdateOne(ctx, bson.M{"_id": courseID.String()}, bson.M{"$set": update})
	return err
}

func (r *courseEnrollmentRepository) IsUserEnrolled(ctx context.Context, courseID, userID uuid.UUID) (bool, error) {
	filter := bson.M{"course_id": courseID.String(), "user_id": userID.String()}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

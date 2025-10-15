package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"

	// Internal packages
	"content-services/internal/db"
	"content-services/internal/repository"
	"content-services/internal/taxonomy"
	"content-services/internal/types"

	// domain models (struct shapes used for JSON)
	"content-services/internal/models"
)

// ImportData structure to hold all data from JSON
type ImportData struct {
	Topics          []models.Topic          `json:"topics"`
	Levels          []models.Level          `json:"levels"`
	Tags            []models.Tag            `json:"tags"`
	Folders         []models.Folder         `json:"folders"`
	MediaAssets     []models.MediaAsset     `json:"media_assets"`
	Lessons         []models.Lesson         `json:"lessons"`
	LessonSections  []models.LessonSection  `json:"lesson_sections"`
	Courses         []models.Course         `json:"courses"`
	CourseLessons   []models.CourseLesson   `json:"course_lessons"`
	FlashcardSets   []models.FlashcardSet   `json:"flashcard_sets"`
	Flashcards      []models.Flashcard      `json:"flashcards"`
	Quizzes         []models.Quiz           `json:"quizzes"`
	QuizQuestions   []models.QuizQuestion   `json:"quiz_questions"`
	QuestionOptions []models.QuestionOption `json:"question_options"`
	ContentTags     []models.ContentTag     `json:"content_tags"`
}

func main() {
	jsonFile := flag.String("file", "/scripts/data.json", "Path to JSON file")
	userID := flag.String("user-id", "", "User ID to associate with content")
	dryRun := flag.Bool("dry-run", false, "Preview changes without saving")

	flag.Parse()

	// Validate inputs
	if *userID == "" {
		log.Fatal("--user-id is required")
	}

	if _, err := uuid.Parse(*userID); err != nil {
		log.Fatalf("Invalid user ID format: %v", err)
	}

	// Read JSON file
	fileContent, err := os.ReadFile(*jsonFile)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	// Fix invalid UUID strings in known ID fields before parsing
	fixedJSON, replaced := fixInvalidUUIDs(fileContent)
	if len(replaced) > 0 {
		fmt.Printf("Fixed %d invalid UUID(s) in input (consistent across references)\n", len(replaced))
	}

	// Parse JSON
	var importData ImportData
	if err := json.Unmarshal(fixedJSON, &importData); err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	// Convert user ID to UUID
	userUUID := uuid.MustParse(*userID)

	// Inject user_id into content
	for i := range importData.Lessons {
		importData.Lessons[i].CreatedBy = &userUUID
	}

	for i := range importData.FlashcardSets {
		importData.FlashcardSets[i].CreatedBy = &userUUID
	}

	for i := range importData.MediaAssets {
		importData.MediaAssets[i].UploadedBy = &userUUID
	}

	fmt.Printf("Import Summary (User: %s)\n", *userID)
	fmt.Printf("Topics:            %d\n", len(importData.Topics))
	fmt.Printf("Levels:            %d\n", len(importData.Levels))
	fmt.Printf("Tags:              %d\n", len(importData.Tags))
	fmt.Printf("Folders:           %d\n", len(importData.Folders))
	fmt.Printf("Media Assets:      %d\n", len(importData.MediaAssets))
	fmt.Printf("Lessons:           %d\n", len(importData.Lessons))
	fmt.Printf("Lesson Sections:   %d\n", len(importData.LessonSections))
	fmt.Printf("Courses:           %d\n", len(importData.Courses))
	fmt.Printf("Course Lessons:    %d\n", len(importData.CourseLessons))
	fmt.Printf("Flashcard Sets:    %d\n", len(importData.FlashcardSets))
	fmt.Printf("Flashcards:        %d\n", len(importData.Flashcards))
	fmt.Printf("Quizzes:           %d\n", len(importData.Quizzes))
	fmt.Printf("Quiz Questions:    %d\n", len(importData.QuizQuestions))
	fmt.Printf("Question Options:  %d\n", len(importData.QuestionOptions))
	fmt.Printf("Content Tags:      %d\n", len(importData.ContentTags))

	if *dryRun {
		fmt.Println("\n[DRY RUN MODE] - No data will be saved to database")
		return
	}

	// Connect to MongoDB
	mongoClient, err := db.NewMongoClient(context.Background())
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}
	mongoDB := db.GetDatabase(mongoClient)

	// Initialize repositories and taxonomy store
	ctx := context.Background()
	store, err := taxonomy.NewStore(ctx, mongoDB)
	if err != nil {
		log.Fatalf("Error initializing taxonomy store: %v", err)
	}
	folderRepo := repository.NewFolderRepository(mongoDB)
	mediaRepo := repository.NewMediaRepository(mongoDB)
	lessonRepo := repository.NewLessonRepository(mongoDB)
	lessonSectionRepo := repository.NewLessonSectionRepository(mongoDB)
	courseRepo := repository.NewCourseRepository(mongoDB)
	courseLessonRepo := repository.NewCourseLessonRepository(mongoDB)
	flashcardSetRepo := repository.NewFlashcardSetRepository(mongoDB)
	flashcardRepo := repository.NewFlashcardRepository(mongoDB)
	quizRepo := repository.NewQuizRepository(mongoDB)
	quizQuestionRepo := repository.NewQuizQuestionRepository(mongoDB)
	optionRepo := repository.NewQuestionOptionRepository(mongoDB)

	fmt.Println("\nInserting data into MongoDB...")

	// 1. Insert Topics (taxonomy store)
	if len(importData.Topics) > 0 {
		inserted := 0
		for _, t := range importData.Topics {
			if _, err := store.CreateTopic(ctx, t.Slug, t.Name); err != nil {
				if isDuplicateErr(err) || errors.Is(err, taxonomy.ErrDuplicate) {
					continue
				}
				log.Fatalf("Error inserting topic %q: %v", t.Slug, err)
			}
			inserted++
		}
		fmt.Printf("✓ Inserted %d topics\n", inserted)
	}

	// 2. Insert Levels (taxonomy store)
	if len(importData.Levels) > 0 {
		inserted := 0
		for _, l := range importData.Levels {
			if _, err := store.CreateLevel(ctx, l.Code, l.Name); err != nil {
				if isDuplicateErr(err) || errors.Is(err, taxonomy.ErrDuplicate) {
					continue
				}
				log.Fatalf("Error inserting level %q: %v", l.Code, err)
			}
			inserted++
		}
		fmt.Printf("✓ Inserted %d levels\n", inserted)
	}

	// 3. Insert Tags (taxonomy store)
	if len(importData.Tags) > 0 {
		inserted := 0
		for _, tag := range importData.Tags {
			if _, err := store.CreateTag(ctx, tag.Slug, tag.Name); err != nil {
				if isDuplicateErr(err) || errors.Is(err, taxonomy.ErrDuplicate) {
					continue
				}
				log.Fatalf("Error inserting tag %q: %v", tag.Slug, err)
			}
			inserted++
		}
		fmt.Printf("✓ Inserted %d tags\n", inserted)
	}

	// 4. Insert Folders
	if len(importData.Folders) > 0 {
		inserted := 0
		for i := range importData.Folders {
			ensureUUID(&importData.Folders[i].ID)
			if err := folderRepo.Create(ctx, &importData.Folders[i]); err != nil {
				if isDuplicateErr(err) {
					continue
				}
				log.Fatalf("Error inserting folder: %v", err)
			}
			inserted++
		}
		fmt.Printf("✓ Inserted %d folders\n", inserted)
	}

	// 5. Insert Media Assets
	if len(importData.MediaAssets) > 0 {
		inserted := 0
		for i := range importData.MediaAssets {
			ensureUUID(&importData.MediaAssets[i].ID)
			if err := mediaRepo.Create(ctx, &importData.MediaAssets[i]); err != nil {
				if isDuplicateErr(err) {
					continue
				}
				log.Fatalf("Error inserting media asset: %v", err)
			}
			inserted++
		}
		fmt.Printf("✓ Inserted %d media assets\n", inserted)
	}

	// 6. Insert Lessons
	if len(importData.Lessons) > 0 {
		inserted := 0
		for i := range importData.Lessons {
			ensureUUID(&importData.Lessons[i].ID)
			if err := lessonRepo.Create(ctx, &importData.Lessons[i]); err != nil {
				if isDuplicateErr(err) || errors.Is(err, types.ErrDuplicateCode) {
					continue
				}
				log.Fatalf("Error inserting lesson: %v", err)
			}
			inserted++
		}
		fmt.Printf("✓ Inserted %d lessons\n", inserted)
	}

	// 7. Insert Lesson Sections
	if len(importData.LessonSections) > 0 {
		inserted := 0
		for i := range importData.LessonSections {
			ensureUUID(&importData.LessonSections[i].ID)
			if err := lessonSectionRepo.Create(ctx, &importData.LessonSections[i]); err != nil {
				if isDuplicateErr(err) {
					continue
				}
				log.Fatalf("Error inserting lesson section: %v", err)
			}
			inserted++
		}
		fmt.Printf("✓ Inserted %d lesson sections\n", inserted)
	}

	// 8. Insert Courses
	if len(importData.Courses) > 0 {
		inserted := 0
		for i := range importData.Courses {
			ensureUUID(&importData.Courses[i].ID)
			if err := courseRepo.Create(ctx, &importData.Courses[i]); err != nil {
				if isDuplicateErr(err) {
					continue
				}
				log.Fatalf("Error inserting course: %v", err)
			}
			inserted++
		}
		fmt.Printf("✓ Inserted %d courses\n", inserted)
	}

	// 9. Insert Course Lessons
	if len(importData.CourseLessons) > 0 {
		inserted := 0
		for i := range importData.CourseLessons {
			ensureUUID(&importData.CourseLessons[i].ID)
			if err := courseLessonRepo.Create(ctx, &importData.CourseLessons[i]); err != nil {
				if isDuplicateErr(err) {
					continue
				}
				log.Fatalf("Error inserting course lesson: %v", err)
			}
			inserted++
		}
		fmt.Printf("✓ Inserted %d course lessons\n", inserted)
	}

	// 10. Insert Flashcard Sets
	if len(importData.FlashcardSets) > 0 {
		inserted := 0
		for i := range importData.FlashcardSets {
			ensureUUID(&importData.FlashcardSets[i].ID)
			if err := flashcardSetRepo.Create(ctx, &importData.FlashcardSets[i]); err != nil {
				if isDuplicateErr(err) {
					continue
				}
				log.Fatalf("Error inserting flashcard set: %v", err)
			}
			inserted++
		}
		fmt.Printf("✓ Inserted %d flashcard sets\n", inserted)
	}

	// 11. Insert Flashcards
	if len(importData.Flashcards) > 0 {
		inserted := 0
		for i := range importData.Flashcards {
			ensureUUID(&importData.Flashcards[i].ID)
			if err := flashcardRepo.Create(ctx, &importData.Flashcards[i]); err != nil {
				if isDuplicateErr(err) {
					continue
				}
				log.Fatalf("Error inserting flashcard: %v", err)
			}
			inserted++
		}
		fmt.Printf("✓ Inserted %d flashcards\n", inserted)
	}

	// 12. Insert Quizzes
	if len(importData.Quizzes) > 0 {
		inserted := 0
		for i := range importData.Quizzes {
			ensureUUID(&importData.Quizzes[i].ID)
			if importData.Quizzes[i].CreatedAt.IsZero() {
				importData.Quizzes[i].CreatedAt = importData.Flashcards[0].CreatedAt // any non-zero, or use time.Now() but keep deterministic
			}
			if err := quizRepo.Create(ctx, &importData.Quizzes[i]); err != nil {
				if isDuplicateErr(err) {
					continue
				}
				log.Fatalf("Error inserting quiz: %v", err)
			}
			inserted++
		}
		fmt.Printf("✓ Inserted %d quizzes\n", inserted)
	}

	// 13. Insert Quiz Questions
	if len(importData.QuizQuestions) > 0 {
		inserted := 0
		for i := range importData.QuizQuestions {
			ensureUUID(&importData.QuizQuestions[i].ID)
			if err := quizQuestionRepo.Create(ctx, &importData.QuizQuestions[i]); err != nil {
				if isDuplicateErr(err) {
					continue
				}
				log.Fatalf("Error inserting quiz question: %v", err)
			}
			inserted++
		}
		fmt.Printf("✓ Inserted %d quiz questions\n", inserted)
	}

	// 14. Insert Question Options
	if len(importData.QuestionOptions) > 0 {
		inserted := 0
		for i := range importData.QuestionOptions {
			ensureUUID(&importData.QuestionOptions[i].ID)
			if err := optionRepo.Create(ctx, &importData.QuestionOptions[i]); err != nil {
				if isDuplicateErr(err) {
					continue
				}
				log.Fatalf("Error inserting question option: %v", err)
			}
			inserted++
		}
		fmt.Printf("✓ Inserted %d question options\n", inserted)
	}

	// 15. Content Tags (skipped in Mongo path)
	if len(importData.ContentTags) > 0 {
		fmt.Printf("- Skipped %d content tags (not supported in Mongo path)\n", len(importData.ContentTags))
	}

	fmt.Println("✓ Import completed successfully!")
}

// ensureUUID sets a new UUID if the current value is zero.
func ensureUUID(id *uuid.UUID) {
	if id == nil {
		return
	}
	if *id == uuid.Nil {
		*id = uuid.New()
	}
}

// fixInvalidUUIDs scans the raw JSON document and replaces invalid UUID strings for
// known identifier keys with generated UUIDs. The same invalid string is replaced
// consistently everywhere it appears under those keys.
func fixInvalidUUIDs(input []byte) ([]byte, map[string]string) {
	var data any
	if err := json.Unmarshal(input, &data); err != nil {
		return input, nil
	}

	replacements := map[string]string{}
	keys := map[string]struct{}{
		"id":             {},
		"lesson_id":      {},
		"topic_id":       {},
		"level_id":       {},
		"parent_id":      {},
		"folder_id":      {},
		"set_id":         {},
		"course_id":      {},
		"quiz_id":        {},
		"question_id":    {},
		"media_id":       {},
		"front_media_id": {},
		"back_media_id":  {},
		"tag_id":         {},
		"object_id":      {},
		"instructor_id":  {},
		"prompt_media":   {},
	}

	var walk func(node any) any
	walk = func(node any) any {
		switch v := node.(type) {
		case map[string]any:
			for k, val := range v {
				if _, ok := keys[k]; ok {
					if s, ok2 := val.(string); ok2 && s != "" {
						if newS, found := replacements[s]; found {
							v[k] = newS
						} else {
							if _, err := uuid.Parse(s); err != nil {
								gen := uuid.New().String()
								replacements[s] = gen
								v[k] = gen
							}
						}
					}
				}
				v[k] = walk(v[k])
			}
			return v
		case []any:
			for i := range v {
				v[i] = walk(v[i])
			}
			return v
		default:
			return v
		}
	}

	fixed := walk(data)
	out, err := json.MarshalIndent(fixed, "", "    ")
	if err != nil {
		return input, replacements
	}
	return out, replacements
}

// isDuplicateErr returns true if the error is a Mongo duplicate key error or taxonomy duplicate.
func isDuplicateErr(err error) bool {
	if err == nil {
		return false
	}
	// Check common mongo duplicate error
	if mongo.IsDuplicateKeyError(err) {
		return true
	}
	// Fallback: text contains duplicate key (for some drivers / wrapped errors)
	if msg := err.Error(); msg != "" {
		if contains(msg, "E11000 duplicate key") || contains(msg, "duplicate key error") {
			return true
		}
	}
	return false
}

func contains(s, sub string) bool {
	return len(sub) > 0 && len(s) >= len(sub) && (func() bool { return stringIndex(s, sub) >= 0 })()
}

// stringIndex is a tiny helper to avoid importing strings just for Contains.
func stringIndex(s, sub string) int {
	// naive search is fine for small error messages
	n, m := len(s), len(sub)
	if m == 0 {
		return 0
	}
	if m > n {
		return -1
	}
	for i := 0; i <= n-m; i++ {
		if s[i:i+m] == sub {
			return i
		}
	}
	return -1
}

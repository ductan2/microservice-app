package types

import "errors"

var (
        // ErrCourseNotFound is returned when a course cannot be located.
        ErrCourseNotFound = errors.New("course: not found")
        // ErrCourseLessonNotFound is returned when a course lesson cannot be located.
        ErrCourseLessonNotFound = errors.New("course lesson: not found")
        // ErrCourseLessonExists is returned when attempting to add a duplicate lesson to a course.
        ErrCourseLessonExists = errors.New("course lesson: already exists")
)

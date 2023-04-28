package models

type Schedule struct {
	Lesson_id       int
	Institute       string
	Course          int
	Group_number    int
	Lesson_name     string
	Lesson_type     string
	Date_range      string
	Day_of_the_week string
	Audience        string
	Lesson_number   int
	Week_type       string
}

type Group struct {
	CountGroup int
}

type Course struct {
	CountCourse int
}

type Institute struct {
	CountInstitute int
}

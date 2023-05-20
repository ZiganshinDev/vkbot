package models

type Schedule struct {
	Lesson_name string
	Lesson_type string
	Date_range  string
	Audience    string
}

type User struct {
	Institute    string
	Course       string
	Group_number string
	Peer_id      string
}

type IsUser struct {
	UserCount int
}

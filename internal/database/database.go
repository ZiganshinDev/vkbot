package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ZiganshinDev/scheduleVKBot/internal/models"
	_ "github.com/lib/pq"
)

const (
	host   = "localhost"
	port   = 5432
	user   = "postgres"
	dbname = "VKbot"
)

func createConnection() *sql.DB {
	password := os.Getenv("DB_PASSWORD")

	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	log.Println("Successfully connected")

	return db
}

func DBShowSchedule(institute string, course string, group_number string, week_type string, lesson_day string) string {
	db := createConnection()
	defer db.Close()

	var schedules []models.Schedule

	rows, err := db.Query(fmt.Sprintf("select * from schedule where institute = '%v' AND course = %v AND group_number = %v AND week_type = '%v' AND day_of_the_week = '%v' order by lesson_number asc;", institute, course, group_number, week_type, lesson_day))
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	for rows.Next() {
		var s models.Schedule

		err := rows.Scan(&s.Lesson_id, &s.Institute, &s.Course, &s.Group_number, &s.Lesson_name, &s.Lesson_type, &s.Date_range, &s.Day_of_the_week, &s.Audience, &s.Lesson_number, &s.Week_type)
		if err != nil {
			fmt.Println(err)
			continue
		}

		schedules = append(schedules, s)
	}

	var result string

	for _, s := range schedules {
		result += s.Audience + "_" + s.Lesson_name + "_" + s.Lesson_type + "_" + s.Date_range + " "
	}

	result = strings.Replace(result, " ", "\n", -1)
	result = strings.Replace(result, "_", " ", -1)

	return result
}

func IsInstitute(institute_id string) bool {
	db := createConnection()
	defer db.Close()

	rows, err := db.Query(fmt.Sprintf("select COUNT(institute) from schedule where institute = '%v'", institute_id))
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	for rows.Next() {
		var i models.Institute
		err := rows.Scan(&i.CountInstitute)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if i.CountInstitute == 0 {
			return false
		}
	}

	return true
}

func IsCourse(course_id string) bool {
	db := createConnection()
	defer db.Close()

	rows, err := db.Query(fmt.Sprintf("select COUNT(course) from schedule where course = %v", course_id))
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	for rows.Next() {
		var c models.Course
		err := rows.Scan(&c.CountCourse)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if c.CountCourse == 0 {
			return false
		}
	}

	return true
}

func IsGroup(group_number string) bool {
	db := createConnection()
	defer db.Close()

	rows, err := db.Query(fmt.Sprintf("select COUNT(group_number) from schedule where group_number = %v", group_number))
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	for rows.Next() {
		var g models.Group
		err := rows.Scan(&g.CountGroup)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if g.CountGroup == 0 {
			return false
		}
	}

	return true
}

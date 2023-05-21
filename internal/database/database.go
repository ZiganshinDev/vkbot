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

func DBShowSchedule(day_of_the_week, user_peer_id string) string {
	db := createConnection()
	defer db.Close()

	var schedules []models.Schedule

	rows, err := db.Query(fmt.Sprintf("SELECT schedule.audience, schedule.lesson_name, schedule.lesson_type, schedule.date_range FROM schedule JOIN users ON users.institute = schedule.institute AND users.course = schedule.course AND users.group_number = schedule.group_number AND users.week_type = schedule.week_type WHERE schedule.day_of_the_week = '%v' AND users.user_peer_id = %v ORDER BY lesson_number ASC;", day_of_the_week, user_peer_id))
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	for rows.Next() {
		var s models.Schedule

		err := rows.Scan(&s.Lesson_name, &s.Lesson_type, &s.Date_range, &s.Audience)
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

func AddUser(institute, course, group_number, user_peer_id string) {
	db := createConnection()
	defer db.Close()

	sqlStatement := `INSERT INTO users (institute, course, group_number, user_peer_id) VALUES ($1, $2, $3, $4) RETURNING user_id`

	var id int

	err := db.QueryRow(sqlStatement, institute, course, group_number, user_peer_id).Scan(&id)
	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}

	log.Printf("Inserted a single record %v", id)

}

func UpdateUser() {

}

func DeleteUser(user_peer_id string) int64 {
	db := createConnection()
	defer db.Close()

	sqlStatement := `DELETE FROM users WHERE user_peer_id = $1`

	res, err := db.Exec(sqlStatement, user_peer_id)
	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Fatalf("Error while checking the affected rows. %v", err)
	}

	log.Printf("Total rows/record affected %v", rowsAffected)

	return rowsAffected
}

func CheckSchedule(institute, course, group_number string) bool {
	db := createConnection()
	defer db.Close()

	rows, err := db.Query(fmt.Sprintf("SELECT COUNT(lesson_id) FROM schedule WHERE institute = '%v' AND course = %v AND group_number = %v;", institute, course, group_number))
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	for rows.Next() {
		var c models.IsUser
		err := rows.Scan(&c.UserCount)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if c.UserCount == 0 {
			return false
		}
	}

	return true
}

func CheckUser(user_peer_id string) bool {
	db := createConnection()
	defer db.Close()

	rows, err := db.Query(fmt.Sprintf("SELECT COUNT(user_peer_id) FROM users WHERE user_peer_id = %v;", user_peer_id))
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	for rows.Next() {
		var c models.IsUser
		err := rows.Scan(&c.UserCount)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if c.UserCount == 0 {
			return false
		}
	}

	return true
}

func CheckUserWithWeekType(user_peer_id string) bool {
	db := createConnection()
	defer db.Close()

	rows, err := db.Query(fmt.Sprintf("SELECT COUNT(user_peer_id) FROM users WHERE user_peer_id = %v AND week_type <> '';", user_peer_id))
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	for rows.Next() {
		var c models.IsUser
		err := rows.Scan(&c.UserCount)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if c.UserCount == 0 {
			return false
		}
	}

	return true
}
func AddWeekToUser(week_type, user_peer_id string) {
	db := createConnection()
	defer db.Close()

	sqlStatement := `UPDATE users SET week_type = $1 WHERE user_peer_id = $2`

	res, err := db.Exec(sqlStatement, week_type, user_peer_id)
	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Fatalf("Error while checking the affected rows. %v", err)
	}

	log.Printf("Total rows/record affected %v", rowsAffected)
}

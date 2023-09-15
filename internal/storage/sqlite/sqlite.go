package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/ZiganshinDev/scheduleVKBot/internal/storage"
	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS users(
		user_id INTEGER PRIMARY KEY,
		institute TEXT,
		course INTEGER,
		group_number INTEGER,
		peer_id INTEGER,
		week TEXT);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err = db.Prepare(`
	CREATE TABLE IF NOT EXISTS schedule(
		lesson_id INTEGER PRIMARY KEY,
		institute TEXT,
		course INTEGER,
		group_number INTEGER,
		lesson_name TEXT,
		lesson_type TEXT,
		date_range TEXT,
		day TEXT,
		audience TEXT,
		lesson_number INTEGER,
		week TEXT);
	`)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) GetSchedule(day string, peerId int) (string, error) {
	const op = "storage.sqlite.GetSchedule"

	stmt, err := s.db.Prepare("SELECT schedule.lesson_name, schedule.lesson_type, schedule.date_range, schedule.audience, schedule.lesson_number FROM schedule INNER JOIN users ON schedule.institute = users.institute AND schedule.course = users.course AND schedule.group_number = users.group_number AND schedule.week = users.week WHERE schedule.day = ? AND users.peer_id = ? ORDER BY schedule.lesson_number;")
	if err != nil {
		return "", fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	rows, err := stmt.Query(day, peerId)
	if err != nil {
		return "", fmt.Errorf("%s: execute statement: %w", op, err)
	}
	defer rows.Close()

	var result string

	for rows.Next() {
		var lesson_name, lesson_type, date_range, audience, lesson_number string
		if err := rows.Scan(&lesson_name, &lesson_type, &date_range, &audience, &lesson_number); err != nil {
			return "", fmt.Errorf("%s: scan row: %w", op, err)
		}

		result += lesson_number + " " + lesson_type + ". " + lesson_name + " " + date_range + " " + audience + "\n"
	}

	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("%s: iterate rows: %w", op, err)
	}

	return result, nil
}

func (s *Storage) AddUser(institute string, course string, groupNumber string, peerId int) error {
	const op = "storage.sqlite.AddUser"

	stmt, err := s.db.Prepare("INSERT INTO users(institute, course, group_number, peer_id) VALUES(?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec(institute, course, groupNumber, peerId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) CheckSchedule(institute string, course string, groupNumber string) (bool, error) {
	const op = "storage.sqlite.CheckSchedule"

	stmt, err := s.db.Prepare("SELECT COUNT(lesson_name) FROM schedule WHERE institute = ? AND course = ? AND group_number = ?")
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	var count int

	err = stmt.QueryRow(institute, course, groupNumber).Scan(&count)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, storage.ErrNotFound
		}

		return false, fmt.Errorf("%s: execute statement: %w", op, err)
	}

	if count == 0 {
		return false, nil
	}

	return true, nil
}

func (s *Storage) UserAddWeek(week string, peerId int) error {
	const op = "storage.sqlite.UserAddWeek"

	stmt, err := s.db.Prepare("UPDATE users SET week = ? WHERE peer_id = ?")
	if err != nil {
		return fmt.Errorf("%s: execute statement: %w", op, err)
	}

	_, err = stmt.Exec(week, peerId)
	if err != nil {
		return fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return nil
}

func (s *Storage) DeleteUser(peerId int) error {
	const op = "storage.sqlite.DeleteUser"

	stmt, err := s.db.Prepare("DELETE FROM users WHERE peer_id = ?")
	if err != nil {
		return fmt.Errorf("%s: execute statement: %w", op, err)
	}

	_, err = stmt.Exec(peerId)
	if err != nil {
		return fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return nil
}

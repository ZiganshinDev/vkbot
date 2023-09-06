package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"

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

	//TODO change to response struct
	var lesson_name, lesson_type, date_range, audience, lesson_number string

	err = stmt.QueryRow(day, peerId).Scan(&lesson_name, &lesson_type, &date_range, &audience, &lesson_number)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrNotFound
		}

		return "", fmt.Errorf("%s: execute statement: %w", op, err)
	}

	var result string

	//TODO change to response struct
	result = result + lesson_number + " " + lesson_type + ". " + lesson_name + " " + date_range + " " + audience

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

func (s *Storage) CheckUser(peerId int) (bool, error) {
	const op = "storage.sqlite.CheckUser"

	stmt, err := s.db.Prepare("SELECT COUNT(user_id) FROM users WHERE peer_id = ?")
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	var count string

	err = stmt.QueryRow(peerId).Scan(&count)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, storage.ErrNotFound
		}

		return false, fmt.Errorf("%s: execute statement: %w", op, err)
	}

	if count, err := strconv.Atoi(count); count == 0 && err == nil {
		return true, nil
	}

	return false, nil
}

func (s *Storage) UserCheckWeek(peerId int) (bool, error) {
	const op = "storage.sqlite.UserCheckWeek"

	stmt, err := s.db.Prepare("SELECT COUNT(week) FROM users WHERE peer_id = ?")
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	var count string

	err = stmt.QueryRow(peerId).Scan(&count)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, storage.ErrNotFound
		}

		return false, fmt.Errorf("%s: execute statement: %w", op, err)
	}

	if count, err := strconv.Atoi(count); count == 0 && err == nil {
		return true, nil
	}

	return false, nil
}

func (s *Storage) UserAddWeek(week string, peerId int) {
	const op = "storage.sqlite.UserAddWeek"

	stmt, err := s.db.Prepare("UPDATE users SET week = ? WHERE peer_id = ?")
	if err != nil {
		log.Printf("%s: %s", op, err)
	}

	_, err = stmt.Exec(week, peerId)
	if err != nil {
		log.Printf("%s: %s", op, err)
	}
}

func (s *Storage) DeleteUser(peerId int) {
	const op = "storage.sqlite.DeleteUser"

	stmt, err := s.db.Prepare("DELETE FROM users WHERE peer_id = ?")
	if err != nil {
		log.Printf("%s: %s", op, err)
	}

	_, err = stmt.Exec(peerId)
	if err != nil {
		log.Printf("%s: %s", op, err)
	}
}

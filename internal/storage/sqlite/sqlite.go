package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/ZiganshinDev/scheduleVKBot/internal/storage"
	"github.com/mattn/go-sqlite3"
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
	CREATE TABLE IF NOT EXISTS user(
		user_id INTEGER PRIMARY KEY,
		institute TEXT,
		course INTEGER,
		group INTEGER
		peer_id TEXT UNIQUE,
		week TEXT);
	CREATE TABLE IF NOT EXISTS schedule(
		lesson_id INTEGER PRIMARY KEY,
		institute TEXT,
		course INTEGER,
		group INTEGER,
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
	//TODO CHANGE WITH JOIN
	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?")
	if err != nil {
		return "", fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	var schedule string

	err = stmt.QueryRow(day, peerId).Scan(&schedule)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrNotFound
		}

		return "", fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return schedule, nil
}

func (s *Storage) AddUser(institute string, course string, group string, peerId int) error {
	const op = "storage.sqlite.AddUser"

	stmt, err := s.db.Prepare("INSERT INTO user(institute, course, group) VALUES(?, ?, ?) WHERE peer_id = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec(institute, course, group, peerId)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return fmt.Errorf("%s: %w", op, storage.ErrExists)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) CheckSchedule(institute string, course string, group string) bool {
	const op = "storage.sqlite.CheckSchedule"

	return true
}

func (s *Storage) CheckUser(peerId int) bool {
	const op = "storage.sqlite.CheckUser"
	//TODO
	return true
}

func (s *Storage) UserCheckWeek(peerId int) bool {
	const op = "storage.sqlite.UserCheckWeek"
	//TODO
	return true
}

func (s *Storage) UserAddWeek(week string, peerId int) {
	const op = "storage.sqlite.UserAddWeek"

	stmt, err := s.db.Prepare("INSERT INTO user(week) VALUES(?) WHERE peer_id = ?")
	if err != nil {
		log.Fatalf("%s: %s", op, err)
	}

	_, err = stmt.Exec(week, peerId)
	if err != nil {
		log.Fatalf("%s: %s", op, err)
	}
}

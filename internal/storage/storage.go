package storage

import (
	"errors"
)

var (
	ErrNotFound = errors.New("not found")
	ErrExists   = errors.New("already exists")
)

type Storage interface {
	GetSchedule(institute string, peerId int) (string, error)
	CheckSchedule(institute string, course string, groupNumber string) (bool, error)
	AddUser(institute string, course string, groupNumber string, peerId int) error
	UserAddWeek(week string, peerId int) error
	DeleteUser(peerId int) error
}

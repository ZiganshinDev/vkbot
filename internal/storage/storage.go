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
	AddUser(institute string, course string, groupNumber string, peerId int) error
	CheckSchedule(institute string, course string, groupNumber string) (bool, error)
	CheckUser(peerId int) (bool, error)
	UserCheckWeek(peerId int) (bool, error)
	UserAddWeek(week string, peerId int)
}

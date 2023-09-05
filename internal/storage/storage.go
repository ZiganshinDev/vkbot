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
	AddUser(institute string, course string, group string, peerId int) error
	CheckSchedule(institute string, course string, group string) bool
	CheckUser(peerId int) bool
	UserCheckWeek(peerId int) bool
	UserAddWeek(week string, peerId int)
}

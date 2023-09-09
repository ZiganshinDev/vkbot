package vkbot

import "sync"

type User struct {
	mu     sync.Mutex
	states map[int]int
}

func NewUserState() *User {
	return &User{
		states: make(map[int]int),
	}
}

func (us *User) SetState(peerId int, state int) {
	us.mu.Lock()
	defer us.mu.Unlock()
	us.states[peerId] = state
}

func (us *User) GetState(peerId int) (int, bool) {
	us.mu.Lock()
	defer us.mu.Unlock()
	state, ok := us.states[peerId]
	return state, ok
}

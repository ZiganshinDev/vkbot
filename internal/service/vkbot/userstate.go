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

func (us *User) SetState(peerID int, state int) {
	us.mu.Lock()
	defer us.mu.Unlock()
	us.states[peerID] = state
}

func (us *User) GetState(peerID int) (int, bool) {
	us.mu.Lock()
	defer us.mu.Unlock()
	state, ok := us.states[peerID]
	return state, ok
}

package main

import (
	"errors"
	"math/rand"
	"strconv"
	"time"
)

type Session struct {
	uuid         string
	value        string
	creationDate int64
}

type Sessions struct {
	storage []Session
}

func NewSessions() *Sessions {
	sessions := Sessions{}
	sessions.storage = make([]Session, 0)
	return &sessions
}

func (s *Sessions) Create(value string) (string, error) {
	uuid := strconv.FormatUint(rand.Uint64(), 10)

	s.storage = append(s.storage, Session{uuid: uuid, value: value, creationDate: time.Now().Unix()})
	s.Clear()
	return uuid, nil
}

func (s *Sessions) Clear() {
	if time.Now().Unix()-s.storage[0].creationDate > 15*60 {
		s.storage = s.storage[1:]
	}
}

func (s *Sessions) IsCorrectSession(name string) bool {
	_, err := s.GetValue(name)
	return err == nil
}

func (s *Sessions) GetValue(name string) (string, error) {
	for _, el := range s.storage {
		if el.uuid == name {
			return el.value, nil
		}
	}
	return "", errors.New("element not found")
}

func (s *Sessions) UpdateValue(name string, value string) bool {
	for n, _ := range s.storage {
		session := &s.storage[n]
		if session.uuid == name {
			session.value = value
			return true
		}
	}
	return false
}

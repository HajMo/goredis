package main

import "time"

type Storage struct {
	data map[string]ValueWithExpiry
}

type ValueWithExpiry struct {
	value     string
	expiresAt time.Time
}

func (value ValueWithExpiry) expired() bool {
	if value.expiresAt.IsZero() {
		return false
	}

	return value.expiresAt.Before(time.Now())
}

func NewStorage() *Storage {
	return &Storage{
		data: make(map[string]ValueWithExpiry),
	}
}

func (storage *Storage) Get(key string) (string, bool) {
	valueWithExpiration, ok := storage.data[key]
	if !ok {
		return "", false
	}

	if valueWithExpiration.expired() {
		delete(storage.data, key)
		return "", false
	}

	return valueWithExpiration.value, true
}

func (storage *Storage) Set(key string, value string) {
	storage.data[key] = ValueWithExpiry{value: value}
}

func (storage *Storage) SetWithExpiry(key string, value string, expiry time.Duration) {
	storage.data[key] = ValueWithExpiry{
		value:     value,
		expiresAt: time.Now().Add(expiry),
	}
}

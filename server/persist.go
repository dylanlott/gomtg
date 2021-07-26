package server

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/zeebo/errs"
)

// Persistence defines the persistence interface for the server.
// This interface stores Game and BoardStates for realtime interaction.
type KeyValue interface {
	Set(key string, value interface{}) error
	Get(key string, dest interface{}) error
}

// Force Server to fulfill KeyValue
var _ KeyValue = (&graphQLServer{})

// Set will set a value into the Redis client and returns an error, if any
func (s *graphQLServer) Set(key string, value interface{}) error {
	// TODO: Need to set this to an env variable
	exp, err := time.ParseDuration("12h")
	if err != nil {
		exp = 0
	}
	p, err := json.Marshal(value)
	if err != nil {
		log.Printf("failed to marshal value in Set: %s", err)
		return err
	}

	return s.redisClient.Set(key, p, exp).Err()
}

// Get returns a value from Redis client to `dest` and returns an error, if any.
// If they key is not found, it will return an error.
// If the key exists but is empty, it will return empty.
func (s *graphQLServer) Get(key string, dest interface{}) error {
	exists, err := s.redisClient.Exists(key).Result()
	if err != nil {
		return errs.Wrap(err)
	}
	if exists < 1 {
		return fmt.Errorf("key [%s] does not exist", key)
	}
	p, err := s.redisClient.Get(key).Result()
	if err != nil {
		log.Printf("failed to get key from redis: %s", err)
		return err
	}
	return json.Unmarshal([]byte(p), dest)
}

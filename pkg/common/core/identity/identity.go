package identity

import (
	"github.com/google/uuid"
)

type Generator struct {
}

func NewGenerator() *Generator {
	return &Generator{}
}

func UUIDv4() string {
	return uuid.NewString()
}

func UUIDv7() string {
	uuidV7, _ := uuid.NewV7()
	return uuidV7.String()
}

package encrypt

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type HashHelperInterface interface {
	HashPassword(password string) (string, error)
	CheckPasswordHash(hash, password string) bool
}

type HashHelper struct {
}

func NewHashService() HashHelperInterface {
	return &HashHelper{}
}

func (hs *HashHelper) HashPassword(password string) (string, error) {

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", fmt.Errorf("failed to generate hashing password: %w", err)
	}

	return string(bytes), nil
}

func (hs *HashHelper) CheckPasswordHash(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

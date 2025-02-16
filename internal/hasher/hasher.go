package hasher

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Hasher interface {
	Generate(password string) (string, error)
	Compare(hash, password string) error
}

const (
	// maxConcurrentBcryptOps represents a max bcrypt operation count.
	maxConcurrentBcryptOps = 20
)

var (
	bcryptSemaphore chan struct{}

	ErrToLong      = errors.New("providen password too long")
	ErrDontCompare = errors.New("passwords dont match")
)

type BcryptHasher struct{}

func init() {
	bcryptSemaphore = make(chan struct{}, maxConcurrentBcryptOps)
}

func (b *BcryptHasher) Generate(password string) (string, error) {
	bcryptSemaphore <- struct{}{}
	defer func() { <-bcryptSemaphore }()

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return "", ErrToLong
		}
		return "", err
	}
	return string(bytes), err

	// return password, nil
}

func (b *BcryptHasher) Compare(hash, password string) error {
	bcryptSemaphore <- struct{}{}
	defer func() { <-bcryptSemaphore }()

	res := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if errors.Is(res, bcrypt.ErrMismatchedHashAndPassword) {
		return ErrDontCompare
	}

	return res

	// if hash == password {
	// 	return nil
	// }
	// return ErrDontCompare
}

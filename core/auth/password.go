package auth

import "golang.org/x/crypto/bcrypt"

type PasswordService interface {
	Hash(password string) (string, error)
	Verify(hash string, password string) bool
}

type BcryptPasswordService struct {
	Cost int
}

func NewPasswordService(cost int) BcryptPasswordService {
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}
	return BcryptPasswordService{Cost: cost}
}

func (s BcryptPasswordService) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.Cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (s BcryptPasswordService) Verify(hash string, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

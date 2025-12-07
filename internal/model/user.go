package model

import "golang.org/x/crypto/bcrypt"

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (u *User) Equals(other *User) bool {
	if u == nil && other == nil {
		return true
	}
	if u == nil || other == nil {
		return false
	}
	return u.Login == other.Login && u.CheckPassword(other.Password)
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func (u *User) HashPassword() error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {

		return err
	}
	u.Password = string(hash)
	return nil
}

func (u *User) Empty() bool {
	if u.Login == "" || u.Password == "" {
		return true
	}
	return false
}

package service

import (
	"errors"
	"strings"
)

// LuhnValid проверяет строку number по алгоритму Луна.
func (as *AccumulationSystem) LuhnValid(number string) bool {
	sum := 0
	alt := false // флаг: удваивать ли текущую цифру

	digits := 0 // чтобы отсеять пустую строку / только пробелы

	for i := len(number) - 1; i >= 0; i-- {
		ch := number[i]

		// пропускаем пробелы
		if ch == ' ' {
			continue
		}

		if ch < '0' || ch > '9' {
			return false // недопустимый символ
		}

		d := int(ch - '0')
		digits++

		if alt {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}

		sum += d
		alt = !alt
	}

	if digits == 0 {
		return false
	}

	return sum%10 == 0
}

// ValidateLogin проверяет формат логина.
// Правила:
//   - длина 3–32 символа
//   - только латинские буквы, цифры, точки, подчёркивания и дефисы
func (as *AccumulationSystem) LoginValid(login string) error {
	login = strings.TrimSpace(login)
	if login == "" {
		return errors.New("login is required")
	}

	if !loginRegexp.MatchString(login) {
		return errors.New("login must be 3-32 chars and contain only letters, digits, '.', '_' or '-'")
	}

	return nil
}

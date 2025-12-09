package service

import (
	"fmt"
	"regexp"
	"strings"
)

var loginRegexp = regexp.MustCompile(`^[a-zA-Z0-9_.-]{3,32}$`)

func (as *AccumulationSystem) LuhnValid(number string) bool {
	sum := 0
	alt := false

	digits := 0

	for i := len(number) - 1; i >= 0; i-- {
		ch := number[i]

		if ch == ' ' {
			continue
		}

		if ch < '0' || ch > '9' {
			return false
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
		msg := "Логин пустой"
		return fmt.Errorf("%s", msg)
	}

	if !loginRegexp.MatchString(login) {
		msg := "Логин может быть 3-32 символа и содержать буквы, цифры, '.', '_' и '-'"
		return fmt.Errorf("%s", msg)
	}

	return nil
}

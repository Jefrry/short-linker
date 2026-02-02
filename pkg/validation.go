package pkg

import (
	"regexp"
	"unicode"
)

const (
	MinPasswordLength = 8
	MaxPasswordLength = 72
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func ValidateEmail(email string) bool {
	return len(email) > 0 && len(email) <= 254 && emailRegex.MatchString(email)
}

func ValidatePassword(password string) (bool, string) {
	if len(password) < MinPasswordLength {
		return false, "password must be at least 8 characters"
	}
	if len(password) > MaxPasswordLength {
		return false, "password must be at most 72 characters"
	}
	var hasLetter, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsNumber(r):
			hasDigit = true
		}
		if hasLetter && hasDigit {
			break
		}
	}
	if !hasLetter || !hasDigit {
		return false, "password must contain at least one letter and one digit"
	}
	return true, ""
}

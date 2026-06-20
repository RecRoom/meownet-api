package utils

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

const allowedNamePunct = " .,'!?-_&()#@:+"
const MaxNameLength = 16
const MinUsernameLength = 3

func IsValidNameLength(name string) bool {
	return len([]rune(strings.TrimSpace(name))) <= MaxNameLength
}

func IsValidUsernameLength(name string) bool {
	n := len([]rune(strings.TrimSpace(name)))
	return n >= MinUsernameLength && n <= MaxNameLength
}

func IsValidName(name string) bool {
	for _, r := range norm.NFC.String(name) {
		switch {
		case r >= '0' && r <= '9':
			continue
		case unicode.Is(unicode.Latin, r):
			continue
		case strings.ContainsRune(allowedNamePunct, r):
			continue
		default:
			return false
		}
	}
	return true
}

func IsValidUsername(name string) bool {
	for _, r := range norm.NFC.String(name) {
		switch {
		case r >= '0' && r <= '9':
			continue
		case unicode.Is(unicode.Latin, r):
			continue
		case r == '_':
			continue
		default:
			return false
		}
	}
	return true
}

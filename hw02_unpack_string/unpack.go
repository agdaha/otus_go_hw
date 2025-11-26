package hw02unpackstring

import (
	"errors"
	"strings"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(s string) (string, error) {
	var b strings.Builder
	var prev rune
	var isEscaped bool

	for _, cur := range s {
		switch {
		case cur == '\\' && !isEscaped:
			isEscaped = true
		case unicode.IsDigit(cur) && !isEscaped:
			if prev == 0 {
				return "", ErrInvalidString
			}
			count := int(cur - '0')
			for i := 0; i < count; i++ {
				b.WriteRune(prev)
			}
			prev = 0
		default:
			if prev != 0 {
				b.WriteRune(prev)
			}
			prev = cur
			isEscaped = false
		}
	}

	if prev != 0 {
		if isEscaped {
			return "", ErrInvalidString
		}
		b.WriteRune(prev)
	}

	return b.String(), nil
}

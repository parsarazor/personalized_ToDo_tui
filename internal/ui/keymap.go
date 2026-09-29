package ui

import "strings"

// Letters of the standard Persian layout mapped to the QWERTY key in the
// same physical position, so shortcuts keep working with a Persian keyboard.
var persianKeys = map[string]string{
	"ض": "q", "ش": "a", "ث": "e", "ی": "d", "ا": "h", "ت": "j", "ن": "k", "م": "l",
	"غ": "y", "ج": "[", "چ": "]",
}

// normDigits converts Persian (۰-۹) and Arabic-Indic (٠-٩) digits to ASCII.
func normDigits(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= '۰' && r <= '۹':
			return '0' + (r - '۰')
		case r >= '٠' && r <= '٩':
			return '0' + (r - '٠')
		}
		return r
	}, s)
}

// normKey canonicalises a key press for the shortcut switch.
func normKey(s string) string {
	s = normDigits(s)
	if v, ok := persianKeys[s]; ok {
		return v
	}
	return s
}

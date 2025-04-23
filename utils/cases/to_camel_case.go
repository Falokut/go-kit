// nolint
package cases

import (
	"strings"
)

// Converts a string to CamelCase
func toCamelInitCase(s string, initCase bool) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	n := strings.Builder{}
	n.Grow(len(s))
	capNext := initCase
	prevIsCap := false
	for i, v := range []byte(s) {
		vIsCap := v >= 'A' && v <= 'Z'
		vIsLow := v >= 'a' && v <= 'z'

		switch {
		case capNext:
			if vIsLow {
				v += 'A'
				v -= 'a'
			}
		case i == 0:
			if vIsCap {
				v += 'a'
				v -= 'A'
			}
		case prevIsCap && vIsCap:
			v += 'a'
			v -= 'A'
		}

		prevIsCap = vIsCap

		switch {
		case vIsCap || vIsLow:
			n.WriteByte(v)
			capNext = false
		case v >= '0' && v <= '9':
			n.WriteByte(v)
			capNext = true
		default:
			capNext = v == '_' || v == ' ' || v == '-' || v == '.'
		}
	}
	return n.String()
}

// ToCamelCase converts a string to CamelCase
func ToCamelCase(s string) string {
	return toCamelInitCase(s, true)
}

// ToLowerCamelCase converts a string to lowerCamelCase
func ToLowerCamelCase(s string) string {
	return toCamelInitCase(s, false)
}

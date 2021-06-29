package utils

import (
	"strings"
)

const (
	// D is .
	D = 46
	// L is {
	L = 123
	// R is }
	R = 125
)

// Render ...
func Render(template string, parameters map[string]string) string {
	s := ""
	t := []string{}
	i := 0
	l := 0
	r := 0

	for j, c := range template {
		switch c {
		case L:
			l++

		case R:
			r++

			if l == 2 && r == 2 {
				if v, ok := parameters[template[i:j-1]]; ok {
					t = append(t, v)
				}

				l = 0
				r = 0
				i = j + 1
			}

		case D:
			if l == 2 {
				t = append(t, template[i:j-2])
				i = j + 1
			}

		default:
			if l == 1 {
				l = 0
			}

			if j == len(template)-1 {
				s = strings.Join(t, "") + template[i:]
			}
		}
	}

	return s
}

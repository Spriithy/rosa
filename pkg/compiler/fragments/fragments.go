package fragments

import "strings"

type Fragment func(rune) bool

func Rune(r rune) Fragment {
	return func(c rune) bool { return c == r }
}

func Any(runes ...rune) Fragment {
	return AnyString(string(runes))
}

func AnyString(str string) Fragment {
	return func(r rune) bool { return strings.IndexRune(str, r) >= 0 }
}

func Range(lo, hi rune) Fragment {
	runes := make([]rune, hi-lo)
	for r := lo; r < hi; r++ {
		runes[r-lo] = r
	}
	return Any(runes...)
}

func Or(fragments ...Fragment) Fragment {
	return func(r rune) bool {
		for _, f := range fragments {
			if f(r) {
				return true
			}
		}
		return false
	}
}

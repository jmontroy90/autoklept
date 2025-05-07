package autoklept

import (
	"fmt"
	"golang.org/x/exp/constraints"
	"regexp"
	"strings"
)

type IntStringer interface {
	fmt.Stringer
	constraints.Integer
}

func Validate[T IntStringer](s string) (T, error) {
	for i := 0; ; i++ {
		p := T(i)
		if !isValidEnum(p) {
			return 0, fmt.Errorf("%T(\"%s\"): %w", p, s, ErrInvalidEnumMember)
		}
		if strings.ToLower(s) == strings.ToLower(p.String()) {
			return p, nil
		}
	}
}

var (
	// This looks for what Stringer returns when an iota passed is not part of the enum.
	rStringerValidator = regexp.MustCompile(`\w+\(\d+\)`)
)

func isValidEnum(s fmt.Stringer) bool {
	return !rStringerValidator.Match([]byte(s.String()))
}

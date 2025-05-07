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

// Validate attempts to validate a raw input string as a member of the given generic `T`.
// WARNING: This is meant only to be used on `stringer`-generated implementations of String().
// It FULLY depends on the actual implementation of String() that `stringer` generates.
// Namely, it depends on String() generating the string "MyIntStringer(n)" for any n that is not in the enum.
// In short - this is meant to be a convenience method for basically one scenario, nothing more. Buyer beware.
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

package autoklept

import (
	"errors"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    PromptOutputTag
		expectedErr error
	}{
		{name: "text", input: "text", expected: PromptOutputText},
		{name: "Markdown", input: "Markdown", expected: PromptOutputMarkdown},
		{name: "sImple", input: "sImple", expected: PromptOutputSimple},
		{name: "simpletext", input: "sImpletext", expectedErr: ErrInvalidEnumMember},
		{name: "doitall", input: "doitall", expectedErr: ErrInvalidEnumMember},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := Validate[PromptOutputTag](tt.input)
			if tt.expectedErr != nil {
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("expected err %v, got %v", tt.expectedErr, err)
				}
			}
			if actual != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, actual)
			}
		})
	}
}

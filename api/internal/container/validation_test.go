package container

import (
	"strings"
	"testing"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		// Valid names
		{"single char", "a", false},
		{"lowercase", "my-stack", false},
		{"with numbers", "abc-123", false},
		{"63 chars", strings.Repeat("a", 63), false},
		{"starts with number", "1stack", false},
		{"ends with number", "stack1", false},

		// Invalid names
		{"empty", "", true},
		{"uppercase", "MyStack", true},
		{"leading hyphen", "-bad", true},
		{"trailing hyphen", "bad-", true},
		{"64+ chars", strings.Repeat("a", 64), true},
		{"reserved default", "default", true},
		{"reserved devarch", "devarch", true},
		{"reserved system", "system", true},
		{"reserved none", "none", true},
		{"reserved all", "all", true},
		{"with spaces", "my stack", true},
		{"with underscore", "my_stack", true},
		{"with special chars", "my@stack", true},
		{"with dots", "my.stack", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateName(tc.input)
			if tc.wantError && err == nil {
				t.Errorf("ValidateName(%q) expected error, got nil", tc.input)
			}
			if !tc.wantError && err != nil {
				t.Errorf("ValidateName(%q) expected nil, got error: %v", tc.input, err)
			}
		})
	}
}

func TestValidateNamePrescriptive(t *testing.T) {
	// Test that error messages include suggestions
	tests := []struct {
		input             string
		shouldContainText string
	}{
		{"My App", "my-app"},
		{"foo_bar", "foo-bar"},
		{"BAD NAME", "bad-name"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			err := ValidateName(tc.input)
			if err == nil {
				t.Errorf("ValidateName(%q) expected error with suggestion", tc.input)
				return
			}
			if !strings.Contains(err.Error(), tc.shouldContainText) {
				t.Errorf("ValidateName(%q) error should suggest %q, got: %v",
					tc.input, tc.shouldContainText, err)
			}
		})
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"My App", "my-app"},
		{"foo_bar", "foo-bar"},
		{"  leading  ", "leading"},
		{"foo--bar", "foo-bar"},
		{"FOO BAR", "foo-bar"},
		{"a b c", "a-b-c"},
		{"test_name_123", "test-name-123"},
		{"---bad---", "bad"},
		{"special!@#chars$%^", "specialchars"},
		{"mixed_CASE-name", "mixed-case-name"},
		{strings.Repeat("a", 70), strings.Repeat("a", 63)},
		{"", "unnamed"},
		{"   ", "unnamed"},
		{"---", "unnamed"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := Slugify(tc.input)
			if result != tc.expected {
				t.Errorf("Slugify(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

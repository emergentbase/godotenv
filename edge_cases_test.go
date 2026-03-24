package godotenv

import (
	"testing"
)

func TestEdgeCases(t *testing.T) {
	envs, err := ReadNoExpand("testdata/edge_cases.env")
	if err != nil {
		t.Fatalf("ReadNoExpand failed: %v", err)
	}

	tests := []struct {
		key      string
		expected string
	}{
		// Basic values
		{"SIMPLE", "hello"},
		{"EMPTY_VALUE", ""},
		{"WITH_SPACES", "hello world"},

		// The bug: values starting with #
		{"HASH_VALUE", "#value"},
		{"COLOR", "#ff0000"},
		{"HEX_COLOR", "#333"},
		{"ANCHOR", "#section-1"},
		{"CHANNEL", "#general"},

		// Hash inside values
		{"URL", "https://example.com/page#anchor"},
		{"HASH_MID", "before#after"},
		{"HASH_WITH_SPACE", "value"},  // " # this is a comment" stripped
		{"MULTI_HASH", "foo#bar#baz"}, // no space before #, so not a comment

		// Quoted values with hashes
		{"SINGLE_QUOTED", "#not-a-comment"},
		{"DOUBLE_QUOTED", "#also-not-a-comment"},
		{"QUOTED_WITH_SPACE", "value # not stripped"},

		// Numeric and special values
		{"PORT", "8080"},
		{"FLOAT", "3.14"},
		{"NEGATIVE", "-1"},
		{"ZERO", "0"},
		{"BOOLEAN_TRUE", "true"},
		{"BOOLEAN_FALSE", "false"},

		// URLs and connection strings
		{"DATABASE_URL", "postgres://user:pass@host:5432/db"},
		{"MONGO_URL", "mongodb+srv://user:pass@cluster.mongodb.net/"},
		{"REDIS_URL", "redis://localhost:6379"},
		{"API_URL", "https://api.example.com/v1"},

		// Special characters
		{"WITH_EQUALS", "key=value"},
		{"WITH_COLON", "host:port"},
		{"WITH_AT", "user@domain.com"},
		{"WITH_BANG", "hello!"},
		{"WITH_DOLLAR", "price$100"},
		{"WITH_PERCENT", "100%"},
		{"WITH_AMPERSAND", "a&b"},
		{"WITH_PIPE", "a|b"},
		{"WITH_PARENS", "(hello)"},
		{"WITH_BRACKETS", "[1,2,3]"},
		{"WITH_BRACES", "{key:val}"},
		{"WITH_BACKTICK", "`code`"},
		{"WITH_TILDE", "~/path"},
		{"WITH_CARET", "a^b"},

		// Whitespace edge cases
		{"LEADING_SPACE", "hello"},
		{"TRAILING_SPACE", "hello"},
		{"TABS", "tabbed\tvalue"},

		// Empty-ish values
		{"JUST_HASH", "#"},
		{"JUST_SPACE", ""},
		{"DOUBLE_HASH", "##double"},

		// Long values
		{"LONG_VALUE", "abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz"},

		// Multi-word with hash comment
		{"SENTENCE", "the quick brown fox"}, // " # jumps over" is a comment
		{"NOSPACE_HASH", "foo#bar"},         // no space before #, not a comment
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got, ok := envs[tt.key]
			if !ok {
				t.Fatalf("key %q not found in parsed envs", tt.key)
			}
			if got != tt.expected {
				t.Errorf("key %q: got %q, want %q", tt.key, got, tt.expected)
			}
		})
	}

	// Also verify total count to catch unexpected extra keys
	t.Logf("Total keys parsed: %d", len(envs))
	for k, v := range envs {
		t.Logf("  %s=%q", k, v)
	}
}

package godotenv

import (
	"testing"
)

func TestNoExpandDollarSigns(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			"preserves dollar signs in unquoted values",
			"TWILIO_SECRET=$12345$vrimcongri$^%$#$12345",
			map[string]string{"TWILIO_SECRET": "$12345$vrimcongri$^%$#$12345"},
		},
		{
			"preserves dollar signs in double quoted values",
			`TWILIO_SECRET="$12345$vrimcongri$^%$#$12345"`,
			map[string]string{"TWILIO_SECRET": "$12345$vrimcongri$^%$#$12345"},
		},
		{
			"preserves dollar signs in single quoted values",
			"TWILIO_SECRET='$12345$vrimcongri$^%$#$12345'",
			map[string]string{"TWILIO_SECRET": "$12345$vrimcongri$^%$#$12345"},
		},
		{
			"preserves $VAR references",
			"BAR=$FOO",
			map[string]string{"BAR": "$FOO"},
		},
		{
			"preserves ${VAR} references",
			"BAR=${FOO}bar",
			map[string]string{"BAR": "${FOO}bar"},
		},
		{
			"preserves $VAR in double quoted values",
			`BAR="quote $FOO"`,
			map[string]string{"BAR": "quote $FOO"},
		},
		{
			"does not cross-reference variables",
			"FOO=test\nBAR=$FOO",
			map[string]string{"FOO": "test", "BAR": "$FOO"},
		},
		{
			"preserves multiple variable patterns",
			"DATABASE_URL=$HOST:$PORT/$DBNAME",
			map[string]string{"DATABASE_URL": "$HOST:$PORT/$DBNAME"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := make(map[string]string)
			if err := parseBytes([]byte(tt.input), out, true); err != nil {
				t.Fatalf("Error: %s", err.Error())
			}
			for k, v := range tt.expected {
				if out[k] != v {
					t.Errorf("Key %s: expected %q, got %q", k, v, out[k])
				}
			}
		})
	}
}

func TestNoExpandPreservesOtherBehavior(t *testing.T) {
	parse := func(input, key, expected string) {
		t.Helper()
		out := make(map[string]string)
		if err := parseBytes([]byte(input), out, true); err != nil {
			t.Errorf("Input %q errored: %v", input, err)
			return
		}
		if out[key] != expected {
			t.Errorf("Input %q: expected %q=%q, got %q", input, key, expected, out[key])
		}
	}

	parse("FOO=bar", "FOO", "bar")
	parse(`FOO="bar"`, "FOO", "bar")
	parse("FOO='bar'", "FOO", "bar")
	parse(`FOO="escaped\"bar"`, "FOO", `escaped"bar`)
	parse("FOO=bar ", "FOO", "bar")
	parse("KEY=value value", "KEY", "value value")
	parse("FOO=bar # comment", "FOO", "bar")
	parse(`FOO="bar#baz"`, "FOO", "bar#baz")
	parse("export OPTION_A=2", "OPTION_A", "2")
	parse(`FOO="bar\nbaz"`, "FOO", "bar\nbaz")
	parse("FOO.BAR=foobar", "FOO.BAR", "foobar")
	parse("FOO=foobar=", "FOO", "foobar=")
	parse("FOO=", "FOO", "")
}

func TestNoExpandWithFile(t *testing.T) {
	envMap, err := ReadNoExpand("fixtures/noexpand.env")
	if err != nil {
		t.Fatalf("Error reading file: %v", err)
	}

	expected := map[string]string{
		"SIMPLE":              "value",
		"SINGLE_QUOTED":       "single quoted value",
		"DOUBLE_QUOTED":       "double quoted value",
		"EMPTY_VALUE":         "",
		"DOLLAR_UNQUOTED":     "$12345$vrimcongri$^%$#$12345",
		"DOLLAR_DOUBLE":       "$12345$vrimcongri$^%$#$12345",
		"DOLLAR_SINGLE":       "$12345$vrimcongri$^%$#$12345",
		"VAR_REF":             "$FOO",
		"VAR_BRACKET":         "${FOO}",
		"MONGO_URL":           "mongodb+srv://user:p@ss%40w0rd@cluster.mongodb.net/dbname",
		"API_KEY":             "C1FVs#$%NAQQQ@J",
		"CONNECTION_STRING":   "postgresql://user:pass@host:5432/db?sslmode=disable",
		"MULTILINE_ESCAPED":   "line1\nline2\nline3",
		"PRIVATE_KEY":         "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA\n-----END RSA PRIVATE KEY-----",
		"EXPORTED_VAR":        "exported_value",
		"DATABASE_URL":        "postgres://localhost:5432/database?sslmode=disable",
		"VALUE_WITH_HASH":     "value#nospace",
		"VALUE_WITH_HASH_SPACE": "value",
	}

	for key, want := range expected {
		got, ok := envMap[key]
		if !ok {
			t.Errorf("Key %q not found", key)
			continue
		}
		if got != want {
			t.Errorf("Key %q: expected %q, got %q", key, want, got)
		}
	}
}

func TestNoExpandOriginalBugScenario(t *testing.T) {
	input := []byte(`TWILIO_SECRET=$12345$vrimcongri$^%$#$12345`)

	// Default (expansion ON) — corrupts the value
	expanded := make(map[string]string)
	parseBytes(input, expanded, false)
	if expanded["TWILIO_SECRET"] != "$vrimcongri$^%$#" {
		t.Errorf("Default parse: expected corrupted value, got %q", expanded["TWILIO_SECRET"])
	}

	// NoExpand — preserves the value
	raw := make(map[string]string)
	parseBytes(input, raw, true)
	if raw["TWILIO_SECRET"] != "$12345$vrimcongri$^%$#$12345" {
		t.Errorf("NoExpand parse: expected literal value, got %q", raw["TWILIO_SECRET"])
	}
}

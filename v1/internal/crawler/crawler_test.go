package crawler

import (
	"testing"
)

func TestGetJSONStringFromInterface(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"nil", nil, ""},
		{"empty array", []interface{}{}, ""},
		{"empty string", "", `""`},
		{"simple string", "hello", `"hello"`},
		{"number", 42, `42`},
		{"map", map[string]interface{}{"key": "val"}, `{"key":"val"}`},
		{"non-empty array", []interface{}{1, 2}, `[1,2]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getJSONStringFromInterface(tt.input)
			if result != tt.expected {
				t.Errorf("getJSONStringFromInterface(%v) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetStringField(t *testing.T) {
	data := map[string]interface{}{
		"name":  "test",
		"empty": "",
	}

	if getStringField(data, "name") != "test" {
		t.Errorf("getStringField(name) != test")
	}
	if getStringField(data, "empty") != "" {
		t.Errorf("getStringField(empty) != empty string")
	}
	if getStringField(data, "missing") != "" {
		t.Errorf("getStringField(missing) != empty string")
	}
}

func TestGetStringFieldNonString(t *testing.T) {
	data := map[string]interface{}{
		"num": 42,
	}
	if getStringField(data, "num") != "" {
		t.Errorf("getStringField(num) should return empty for non-string value")
	}
}

func TestGetIntField(t *testing.T) {
	data := map[string]interface{}{
		"float": float64(42.0),
		"int":   int(100),
	}

	if getIntField(data, "float") != 42 {
		t.Errorf("getIntField(float) = %d, want 42", getIntField(data, "float"))
	}
	if getIntField(data, "int") != 100 {
		t.Errorf("getIntField(int) = %d, want 100", getIntField(data, "int"))
	}
	if getIntField(data, "missing") != 0 {
		t.Errorf("getIntField(missing) = %d, want 0", getIntField(data, "missing"))
	}
}

func TestGetIntFieldNonNumeric(t *testing.T) {
	data := map[string]interface{}{
		"str": "hello",
	}
	if getIntField(data, "str") != 0 {
		t.Errorf("getIntField(str) should return 0 for non-numeric value")
	}
}

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package to

import (
	"testing"
)

func TestString(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{
			name:  "nil value",
			input: nil,
			want:  "",
		},
		{
			name:  "string value",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "int value",
			input: 42,
			want:  "42",
		},
		{
			name:  "negative int",
			input: -100,
			want:  "-100",
		},
		{
			name:  "zero int",
			input: 0,
			want:  "0",
		},
		{
			name:  "bool true",
			input: true,
			want:  "true",
		},
		{
			name:  "bool false",
			input: false,
			want:  "false",
		},
		{
			name:  "struct value",
			input: struct{ Name string }{Name: "test"},
			want:  `{"Name":"test"}`,
		},
		{
			name:  "map value",
			input: map[string]string{"key": "value"},
			want:  `{"key":"value"}`,
		},
		{
			name:  "slice value",
			input: []string{"a", "b", "c"},
			want:  `["a","b","c"]`,
		},
		{
			name:  "float value",
			input: 3.14,
			want:  "3.14",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := String(tt.input)
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFloat(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  float64
	}{
		{"nil", nil, 0},
		{"float64", float64(3.14), 3.14},
		{"float32", float32(2.5), float64(float32(2.5))},
		{"int", int(7), 7.0},
		{"int32", int32(7), 7.0},
		{"int64", int64(7), 7.0},
		{"string valid", "1.5", 1.5},
		{"string invalid", "abc", 0},
		{"unsupported type", true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Float(tt.input)
			if got != tt.want {
				t.Errorf("Float() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInt(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  int
	}{
		{"nil", nil, 0},
		{"int", int(5), 5},
		{"int32", int32(5), 5},
		{"int64", int64(5), 5},
		{"float64", float64(5.9), 5},
		{"float32", float32(3.0), 3},
		{"unsupported string", "10", 0},
		{"unsupported bool", true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Int(tt.input)
			if got != tt.want {
				t.Errorf("Int() = %v, want %v", got, tt.want)
			}
		})
	}
}

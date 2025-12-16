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

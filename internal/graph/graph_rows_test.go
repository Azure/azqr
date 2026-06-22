// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package graph

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestUnmarshalRows(t *testing.T) {
	type row struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	tests := []struct {
		name     string
		data     []json.RawMessage
		expected []row
	}{
		{
			name:     "nil data",
			data:     nil,
			expected: []row{},
		},
		{
			name:     "empty data",
			data:     []json.RawMessage{},
			expected: []row{},
		},
		{
			name: "valid rows",
			data: []json.RawMessage{
				json.RawMessage(`{"name":"a","count":1}`),
				json.RawMessage(`{"name":"b","count":2}`),
			},
			expected: []row{
				{Name: "a", Count: 1},
				{Name: "b", Count: 2},
			},
		},
		{
			name: "malformed rows are skipped, valid kept",
			data: []json.RawMessage{
				json.RawMessage(`{"name":"a","count":1}`),
				json.RawMessage(`{not json}`),
				json.RawMessage(`{"name":"c","count":3}`),
			},
			expected: []row{
				{Name: "a", Count: 1},
				{Name: "c", Count: 3},
			},
		},
		{
			name: "type mismatch row is skipped",
			data: []json.RawMessage{
				json.RawMessage(`{"name":"a","count":"not-an-int"}`),
				json.RawMessage(`{"name":"b","count":2}`),
			},
			expected: []row{
				{Name: "b", Count: 2},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UnmarshalRows[row](tt.data, "test")
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("UnmarshalRows() = %+v, want %+v", got, tt.expected)
			}
		})
	}
}

func TestRawMessageToString(t *testing.T) {
	tests := []struct {
		name     string
		input    json.RawMessage
		expected string
	}{
		{name: "nil", input: nil, expected: ""},
		{name: "empty", input: json.RawMessage(``), expected: ""},
		{name: "null literal", input: json.RawMessage(`null`), expected: ""},
		{name: "quoted string is unwrapped", input: json.RawMessage(`"hello"`), expected: "hello"},
		{name: "quoted string with escapes", input: json.RawMessage(`"a \"b\" c"`), expected: `a "b" c`},
		{name: "number kept as text", input: json.RawMessage(`42`), expected: "42"},
		{name: "bool kept as text", input: json.RawMessage(`true`), expected: "true"},
		{name: "object kept as raw json", input: json.RawMessage(`{"k":"v"}`), expected: `{"k":"v"}`},
		{name: "array kept as raw json", input: json.RawMessage(`["a","b"]`), expected: `["a","b"]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rawMessageToString(tt.input); got != tt.expected {
				t.Errorf("rawMessageToString(%q) = %q, want %q", string(tt.input), got, tt.expected)
			}
		})
	}
}

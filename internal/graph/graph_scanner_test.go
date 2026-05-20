// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package graph

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
)

func TestShouldSkipUnsupportedGraphLogicalTableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name: "DisallowedLogicalTableName",
			err: &azcore.ResponseError{
				ErrorCode: "DisallowedLogicalTableName",
			},
			expected: true,
		},
		{
			name: "different response error code",
			err: &azcore.ResponseError{
				ErrorCode: "BadRequest",
			},
			expected: false,
		},
		{
			name: "nested details DisallowedLogicalTableName",
			err: &azcore.ResponseError{
				ErrorCode: "BadRequest",
				RawResponse: &http.Response{
					Body: io.NopCloser(bytes.NewReader([]byte(`{"error":{"code":"BadRequest","message":"support info","details":[{"code":"DisallowedLogicalTableName","message":"Table appserviceresources is invalid, unsupported or disallowed."}]}}`))),
				},
			},
			expected: true,
		},
		{
			name: "nested details unrelated error",
			err: &azcore.ResponseError{
				ErrorCode: "BadRequest",
				RawResponse: &http.Response{
					Body: io.NopCloser(bytes.NewReader([]byte(`{"error":{"code":"BadRequest","message":"support info","details":[{"code":"OtherCode","message":"some other failure"}]}}`))),
				},
			},
			expected: false,
		},
		{
			name:     "generic error with disallowed logical table code",
			err:      errors.New(`recommendation 0b80b67c-afbe-4988-ad58-a85a146b681e query failed: failed to run resource graph query: HTTP 400 from https://management.chinacloudapi.cn/providers/Microsoft.ResourceGraph/resources?api-version=2024-04-01: {"error":{"code":"BadRequest","message":"support info","details":[{"code":"DisallowedLogicalTableName","message":"Table appserviceresources is invalid, unsupported or disallowed."}]}}`),
			expected: true,
		},
		{
			name:     "generic error with unsupported-table message only",
			err:      errors.New("Table appserviceresources is invalid, unsupported or disallowed."),
			expected: false,
		},
		{
			name:     "generic error with logical table message",
			err:      errors.New("Logical table appserviceresources is invalid, unsupported or disallowed."),
			expected: true,
		},
		{
			name:     "non-response error",
			err:      errors.New("generic error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldSkipUnsupportedGraphLogicalTableError(tt.err)
			if got != tt.expected {
				t.Errorf("ShouldSkipUnsupportedGraphLogicalTableError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

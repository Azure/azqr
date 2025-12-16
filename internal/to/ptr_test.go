// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package to

import (
	"testing"
)

func TestPtr(t *testing.T) {
	t.Run("string pointer", func(t *testing.T) {
		val := "test"
		ptr := Ptr(val)
		if ptr == nil {
			t.Fatal("Ptr() returned nil")
		}
		if *ptr != val {
			t.Errorf("Ptr() = %v, want %v", *ptr, val)
		}
	})

	t.Run("int pointer", func(t *testing.T) {
		val := 42
		ptr := Ptr(val)
		if ptr == nil {
			t.Fatal("Ptr() returned nil")
		}
		if *ptr != val {
			t.Errorf("Ptr() = %v, want %v", *ptr, val)
		}
	})

	t.Run("bool pointer", func(t *testing.T) {
		val := true
		ptr := Ptr(val)
		if ptr == nil {
			t.Fatal("Ptr() returned nil")
		}
		if *ptr != val {
			t.Errorf("Ptr() = %v, want %v", *ptr, val)
		}
	})

	t.Run("struct pointer", func(t *testing.T) {
		type TestStruct struct {
			Name string
			Age  int
		}
		val := TestStruct{Name: "test", Age: 30}
		ptr := Ptr(val)
		if ptr == nil {
			t.Fatal("Ptr() returned nil")
		}
		if ptr.Name != val.Name || ptr.Age != val.Age {
			t.Errorf("Ptr() = %v, want %v", *ptr, val)
		}
	})

	t.Run("empty string pointer", func(t *testing.T) {
		val := ""
		ptr := Ptr(val)
		if ptr == nil {
			t.Fatal("Ptr() returned nil")
		}
		if *ptr != val {
			t.Errorf("Ptr() = %v, want %v", *ptr, val)
		}
	})

	t.Run("zero int pointer", func(t *testing.T) {
		val := 0
		ptr := Ptr(val)
		if ptr == nil {
			t.Fatal("Ptr() returned nil")
		}
		if *ptr != val {
			t.Errorf("Ptr() = %v, want %v", *ptr, val)
		}
	})
}

package examples

import (
	"testing"
)

// This test file does NOT use testify
func TestGoodExample(t *testing.T) {
	result := 2 + 2
	if result != 4 {
		t.Errorf("Expected 4, got %d", result)
	}
}

func TestAnotherGoodExample(t *testing.T) {
	if true != true {
		t.Fatal("Something is very wrong")
	}
}

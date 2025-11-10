package examples

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This test file DOES use testify - should be flagged
func TestBadExample(t *testing.T) {
	result := 2 + 2
	assert.Equal(t, 4, result, "2+2 should equal 4")
}

func TestAnotherBadExample(t *testing.T) {
	require.NotNil(t, "something")
}

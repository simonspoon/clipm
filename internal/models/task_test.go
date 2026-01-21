package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidStatus(t *testing.T) {
	// Valid statuses
	assert.True(t, IsValidStatus(StatusTodo))
	assert.True(t, IsValidStatus(StatusInProgress))
	assert.True(t, IsValidStatus(StatusDone))

	// Invalid statuses
	assert.False(t, IsValidStatus(""))
	assert.False(t, IsValidStatus("invalid"))
	assert.False(t, IsValidStatus("DONE"))        // case sensitive
	assert.False(t, IsValidStatus("TODO"))        // case sensitive
	assert.False(t, IsValidStatus("in_progress")) // wrong format
}

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

func TestHasStructuredFields(t *testing.T) {
	// All three set → true
	task := &Task{Action: "do X", Verify: "check Y", Result: "report Z"}
	assert.True(t, task.HasStructuredFields())

	// Missing Action → false
	task = &Task{Action: "", Verify: "check Y", Result: "report Z"}
	assert.False(t, task.HasStructuredFields())

	// Missing Verify → false
	task = &Task{Action: "do X", Verify: "", Result: "report Z"}
	assert.False(t, task.HasStructuredFields())

	// Missing Result → false
	task = &Task{Action: "do X", Verify: "check Y", Result: ""}
	assert.False(t, task.HasStructuredFields())

	// All empty → false (legacy task)
	task = &Task{}
	assert.False(t, task.HasStructuredFields())
}

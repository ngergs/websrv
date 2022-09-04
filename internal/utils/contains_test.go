package utils_test

import (
	"testing"

	"github.com/ngergs/websrv/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	assert.True(t, utils.Contains([]string{"abc", "test", "123"}, "test"))
	assert.False(t, utils.Contains([]string{"abc", "test", "123"}, "test1"))
	assert.False(t, utils.Contains(nil, "test"))
}

func TestContainsAfterSplit(t *testing.T) {
	assert.True(t, utils.ContainsAfterSplit([]string{"abc", "test", "123"}, ";", "test"))
	assert.True(t, utils.ContainsAfterSplit([]string{"abc", "abc;test;123", "123"}, ";", "test"))
	assert.False(t, utils.ContainsAfterSplit([]string{"abc", "test", "123"}, ";", "test1"))
	assert.False(t, utils.ContainsAfterSplit([]string{"abc", "abc;test;123", "123"}, ";", "test1"))
	assert.False(t, utils.ContainsAfterSplit(nil, ";", "test"))
}

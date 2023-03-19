package utils_test

import (
	"testing"

	"github.com/ngergs/websrv/v2/internal/utils"
	"github.com/stretchr/testify/require"
)

func TestContains(t *testing.T) {
	require.True(t, utils.Contains([]string{"abc", "test", "123"}, "test"))
	require.False(t, utils.Contains([]string{"abc", "test", "123"}, "test1"))
	require.False(t, utils.Contains(nil, "test"))
}

func TestContainsAfterSplit(t *testing.T) {
	require.True(t, utils.ContainsAfterSplit([]string{"abc", "test", "123"}, ";", "test"))
	require.True(t, utils.ContainsAfterSplit([]string{"abc", "abc;test;123", "123"}, ";", "test"))
	require.False(t, utils.ContainsAfterSplit([]string{"abc", "test", "123"}, ";", "test1"))
	require.False(t, utils.ContainsAfterSplit([]string{"abc", "abc;test;123", "123"}, ";", "test1"))
	require.False(t, utils.ContainsAfterSplit(nil, ";", "test"))
}

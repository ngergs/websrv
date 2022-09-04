package utils_test

import (
	"testing"

	"github.com/ngergs/websrv/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetRandomId(t *testing.T) {
	n := 32
	randId := utils.GetRandomId(n)
	assert.Equal(t, n, len(randId))
	randId2 := utils.GetRandomId(n)
	assert.Equal(t, n, len(randId2))
	assert.NotEqual(t, randId, randId2)
}

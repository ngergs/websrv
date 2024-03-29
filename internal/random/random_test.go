package random

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetRandomId(t *testing.T) {
	n := 32
	gen := NewBufferedRandomIdGenerator(n, 16)
	defer func() {
		err := gen.Close()
		require.NoError(t, err)
	}()
	randId := gen.GetRandomId()
	require.Len(t, randId, n)
	randId2 := gen.GetRandomId()
	require.Len(t, randId2, n)
	require.NotEqual(t, randId, randId2)
}

func BenchmarkRandomId(b *testing.B) {
	n := 32
	parallel := 10
	var wg sync.WaitGroup
	wg.Add(parallel)
	for i := 0; i < parallel; i++ {
		go func() {
			for j := 0; j < b.N; j++ {
				getRandomIdWithMutex(n)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func BenchmarkBufferedRandomId(b *testing.B) {
	n := 32
	parallel := 10
	var wg sync.WaitGroup
	wg.Add(parallel)
	gen := NewBufferedRandomIdGenerator(n, 16)
	defer gen.Close()
	for i := 0; i < parallel; i++ {
		go func() {
			for j := 0; j < b.N; j++ {
				gen.GetRandomId()
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

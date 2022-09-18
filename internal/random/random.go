package random

import (
	"math/rand"
	"time"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// BufferedRandomIdGenerator is used to fetch random generator ids while avoiding a mutex.
// The random ids are prefetched ina single goroutine in the background.
// You have to call close to stop this goroutine.
type BufferedRandomIdGenerator struct {
	idLength int
	rand     *rand.Rand
	ch       chan string
	closed   chan struct{}
}

// GetRandomId returns a prefetched random id. Blocks till one is received.
func (gen *BufferedRandomIdGenerator) GetRandomId() string {
	return <-gen.ch
}

// NewBufferedRandomIdGenerator instantiates a new generator with the given buffer size.
// The BufferedRandomIdGenerator has to be closed to avoid leaking the prefetch go routine.
func NewBufferedRandomIdGenerator(idLength int, bufferSize int) *BufferedRandomIdGenerator {
	gen := &BufferedRandomIdGenerator{
		idLength: idLength,
		// use a non default source to avoid automatic mutex via the rand default source
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
		ch:     make(chan string, bufferSize),
		closed: make(chan struct{}),
	}
	go gen.prefetchRandomIds()
	return gen
}

// prefetchRandomIds is called internally to prefetch random ids.
// as we want to avoid mutexes only one version will be called per BufferedRandomIdGenerator.
func (gen *BufferedRandomIdGenerator) prefetchRandomIds() {
	for true {
		select {
		case <-gen.closed:
			close(gen.ch)
			return
		default:
			id := make([]rune, gen.idLength)
			for i := range id {
				id[i] = letters[gen.rand.Intn(len(letters))]
			}
			gen.ch <- string(id)
		}
	}
}

// Close stops the background prefetch process. Does not error.
func (gen *BufferedRandomIdGenerator) Close() error {
	close(gen.closed)
	return nil
}

// getRandomIdWithMutex generates a random string id using [a-zA-Z0-9] with the given length idLength.
// Uses a mutex via the default stdlib rand source.
func getRandomIdWithMutex(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

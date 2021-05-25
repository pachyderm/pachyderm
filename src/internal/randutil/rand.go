package randutil

import "math/rand"

var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// Bytes generates random bytes (n is number of bytes)
func Bytes(random *rand.Rand, n int) []byte {
	bs := make([]byte, n)
	for i := range bs {
		bs[i] = letters[random.Intn(len(letters))]
	}
	return bs
}

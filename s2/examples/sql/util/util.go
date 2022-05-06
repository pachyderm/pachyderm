package util

import "math/rand"

const randomStringOptions = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = randomStringOptions[rand.Intn(len(randomStringOptions))]
	}
	return string(b)
}

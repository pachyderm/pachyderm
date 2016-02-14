package main

import (
	"fmt"
	"math/rand"
	"time"
)

func item() string {
	return [3]string{"apple", "banana", "orange"}[rand.Intn(3)]
}

func amount() string {
	return fmt.Sprintf("%d", rand.Intn(9)+1)
}

func printLine() {
	fmt.Printf("%s\t%s\n", item(), amount())
}

func main() {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 100; i++ {
		printLine()
	}
}

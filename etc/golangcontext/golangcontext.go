package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"golang.org/x/net/context"
)

func main() {
	if err := do(); err != nil {
		fmt.Printf("main error: %s\n", err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func do() error {
	ctx, cancel := context.WithCancel(context.Background())
	waitGroup := &sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		i := i
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			errC := make(chan error, 1)
			go func() {
				for j := 0; j < 5; j++ {
					fmt.Printf("%d %d\n", i, j)
					time.Sleep(100 * time.Millisecond)
				}
				errC <- nil
			}()
			select {
			case <-ctx.Done():
				if err := <-errC; err != nil {
					fmt.Printf("error from errC in ctx.Done() case statement for %d: %v\n", i, err)
				}
				if err := ctx.Err(); err != nil {
					fmt.Printf("error from ctx.Err() in ctx.Done() case statement for %d: %v\n", i, err)
				}
				fmt.Printf("ctx.Done() done for %d\n", i)
			case err := <-errC:
				if err != nil {
					fmt.Printf("error from errC case statement for %d: %v\n", i, err)
				}
				fmt.Printf("errC done for %d\n", i)
			}
		}()
	}
	time.Sleep(400 * time.Millisecond)
	cancel()
	<-ctx.Done()
	fmt.Println("ctx.Done() done for do()")
	waitGroup.Wait()
	fmt.Println("waitGroup done")
	return ctx.Err()
}

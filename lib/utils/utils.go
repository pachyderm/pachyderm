package utils

import (
	"log"
	"sync"
	"time"
)

func Retry(f func() error, retries int, pause time.Duration) error {
	var err error
	for i := 0; i < retries; i++ {
		err = f()
		if err == nil {
			break
		} else {
			log.Print("Retrying due to error: ", err)
			time.Sleep(pause)
		}
	}
	return err
}

package main

import "log"

func RetryOnError(fn func() error, maxTime int, desc string) error {
	var err error
	for i := 0; i < maxTime; i++ {
		if err = fn(); err == nil {
			return nil
		}
		log.Printf("[WARN] %s failed: %v. retrying...", desc, err)
	}
	return err
}

package util

import (
	"time"
)

// Repeat executes f every interval seconds until stopc is closed.
// It executes f once right after being called.
func Repeat(interval time.Duration, stopc <-chan struct{}, f func() error) error {
	tick := time.NewTicker(interval)
	defer tick.Stop()

	for {
		if err := f(); err != nil {
			return err
		}
		select {
		case <-stopc:
			return nil
		case <-tick.C:
		}
	}
}

// Retry executes f every interval seconds until timeout or no error is returned from f.
func Retry(interval time.Duration, stopc <-chan struct{}, f func() error) error {
	return RetryWithLog(interval, stopc, f)
}

// RetryWithLog executes f every interval seconds until timeout or no error is returned from f.
// It logs an error on each f error.
func RetryWithLog(interval time.Duration, stopc <-chan struct{}, f func() error) error {
	tick := time.NewTicker(interval)
	defer tick.Stop()

	var err error
	for {
		if err = f(); err == nil {
			return nil
		}

		select {
		case <-stopc:
			return err
		case <-tick.C:
		}
	}
}

func RunTimeout(timeout time.Duration, errTimeout error, f func() error) error {
	tick := time.NewTicker(timeout)
	defer tick.Stop()

	errCh := make(chan error, 1)

	go func() {
		e := f()
		errCh <- e
		close(errCh)
	}()

	select {
	case <-tick.C:
		return errTimeout
	case err := <- errCh:
		return err
	}
}

func WaitTimeout(timeout time.Duration, errTimeout error, interval time.Duration, f func() bool) error {
	tick := time.NewTicker(timeout)
	defer tick.Stop()

	errCh := make(chan error, 1)

	go func() {
		tick := time.NewTicker(interval)
		defer tick.Stop()

		for ! f() {
			<- tick.C
		}

		errCh <- nil
		close(errCh)
	}()

	select {
	case <-tick.C:
		return errTimeout
	case err := <- errCh:
		return err
	}

	return nil
}

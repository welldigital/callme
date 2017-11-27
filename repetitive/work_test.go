package repetitive

import (
	"fmt"
	"testing"
	"time"
)

func TestThatAWorkerCanBeStopped(t *testing.T) {
	sleep := time.Second * 5
	stopper := make(chan bool)
	workDone := make(chan bool, 1024) // Allow 1024 work items.
	timeoutReached := false

	// The worker notifies a goroutine which triggers the stopper.
	worker := func() (bool, error) {
		fmt.Println("test: carrying out fake work")
		workDone <- true
		return true, nil
	}

	// Start a 5s timeout.
	go func() {
		time.Sleep(time.Second * 5)
		fmt.Println("test: timeout reached, stopping process")
		timeoutReached = true
		stopper <- true
	}()

	// Stop the work when ready.
	go func() {
		// Wait for work to be done.
		<-workDone
		fmt.Println("test: received workdone signal, stopping process")
		// Stop the worker.
		stopper <- true
	}()

	Work("test", worker, sleep, stopper)

	if timeoutReached {
		t.Errorf("expected to be able to stop the worker within 10s")
	}
}

func TestThatWorkersSleepWhenNoWorkHasBeenDone(t *testing.T) {
	var hasSlept bool
	sleep := func() { hasSlept = true }
	stopper := make(chan bool)
	workDone := make(chan bool, 1024) // Allow 1024 work items.
	timeoutReached := false

	// The worker notifies a goroutine which triggers the stopper.
	worker := func() (bool, error) {
		fmt.Println("test: carrying out fake work")
		workDone <- false
		// Return that nothing was done.
		return false, nil
	}

	// Start a 10s timeout.
	go func() {
		time.Sleep(time.Second * 10)
		fmt.Println("test: timeout reached, stopping process")
		timeoutReached = true
		stopper <- true
	}()

	// Stop the work when ready.
	go func() {
		// Wait for work to be done.
		<-workDone
		fmt.Println("test: received workdone signal, stopping process")
		// Stop the worker.
		stopper <- true
	}()

	work("test", worker, time.Nanosecond, sleep, stopper)

	if timeoutReached {
		t.Errorf("expected to be able to stop the worker within 10s")
	}
	if !hasSlept {
		t.Errorf("when no work has been done, the system should sleep before looking for new work")
	}
}

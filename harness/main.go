package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/a-h/callme/mysql"

	"github.com/a-h/callme/logger"
)

func main() {
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Create test database to work against
	dsn, dbName, err := mysql.CreateTestDatabase()
	if err != nil {
		logger.Errorf("failed to create test databse: %v", err)
		return
	}
	defer mysql.DropTestDatabase(dbName)

	// Start up a bunch of tasks in the test DB
	tasksToCreate := 100

	arn := "http://localhost:8080"
	payload := `{ "test": true }`
	jm := mysql.NewJobManager(dsn)

	taskStart := time.Now().UTC()
	for i := 0; i < tasksToCreate; i++ {
		j, err := jm.StartJob(time.Now().UTC(), arn, payload, nil)
		if err != nil {
			logger.Errorf("failed to create test job i=%v, with error: %v", i, err)
			return
		}
		logger.WithJob(j).Info("created")
	}
	taskDuration := time.Now().UTC().Sub(taskStart)

	// Start a web server to count the work.
	handler := NewCountHandler(tasksToCreate)

	s = &http.Server{
		Addr:           ":8080",
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Start serving
	go func() {
		log.Fatal(s.ListenAndServe())
	}()

	// Start the work processing
	// Start up the server
	cmd := exec.Command("../cmd/cmd")
	cmd.Env = append(cmd.Env, "CALLME_CONNECTION_STRING="+dsn)
	cmd.Env = append(cmd.Env, "CALLME_MODE=web")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	runStart := time.Now().UTC()
	go func() {
		if err := cmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
	}()

	// Wait for completion
	select {
	case <-sigs:
		logger.Infof("stop signal received")
		break
	case <-handler.Completed:
		logger.Infof("test complete")
		break
	}
	runDuration := time.Now().UTC().Sub(runStart)

	logger.Infof("shutting down web server")
	if err := s.Close(); err != nil {
		logger.Errorf("failed to shut down server with error: %v", err)
	}
	logger.Infof("killing process")
	if err := cmd.Process.Kill(); err != nil {
		logger.Errorf("failed to kill process with error: %v", err)
	}

	// Write out a summary
	fmt.Printf("created %v tasks in %v seconds\n", tasksToCreate, taskDuration.Seconds())
	fmt.Printf("web server received %v of %v messages in %v seconds\n", handler.Received, tasksToCreate, runDuration.Seconds())
}

var s *http.Server

func NewCountHandler(expected int) *CountHandler {
	ch := &CountHandler{
		c:         make(chan bool, expected),
		stopper:   make(chan bool, 1),
		Completed: make(chan bool, 1),
		Expected:  expected,
	}
	// Start receiving counts in another routine.
	go func() {
		ch.Receive()
	}()
	return ch
}

type CountHandler struct {
	// channel used to process counts sequentially
	c chan bool
	// channel used to stop everything
	stopper chan bool
	// channel returned to client to signify completion
	Completed chan bool
	// actual number of messages received
	Received int
	// expected number of messages to receive
	Expected int
}

func (h *CountHandler) Shutdown() {
	// stop the receiver
	h.stopper <- true
	// tell clients that we're finished
	h.Completed <- true
}

func (h *CountHandler) Receive() {
	go func() {
		for {
			select {
			case <-h.c:
				h.Received++
				logger.Infof("received: %v", h.Received)
			case <-h.stopper:
				return
			default:
				// stop if we've hit the expected number of messages received
				if h.Received >= h.Expected {
					h.Shutdown()
				}
			}
		}
	}()
}

func (h *CountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.c <- true
	logger.Infof("received HTTP request %v", h.Received)
	io.WriteString(w, "OK\n")
}

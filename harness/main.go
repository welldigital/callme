package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/a-h/callme/mysql"

	log "github.com/a-h/callme/logger"

	"net/http/pprof"
)

func main() {
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Start up a bunch of jobs in the test DB
	jobsToCreate := 1000
	// Create a schedule which creates a job each minute
	schedulesToCreate := 1
	// Quit after receipt of how many messages?
	quitAfter := 1000
	// Numbers of processes
	scheduleWorkerCount := 2
	jobWorkerCount := 2

	// Create test database to work against
	dsn, dbName, err := mysql.CreateTestDatabase()
	if err != nil {
		log.Errorf("failed to create test databse: %v", err)
		return
	}
	defer mysql.DropTestDatabase(dbName)

	arn := "http://localhost:8080"
	payload := `{ "test": true }`

	taskStart := time.Now().UTC()
	log.Infof("creating %v jobs", jobsToCreate)
	jm := mysql.NewJobManager(dsn)
	for i := 0; i < jobsToCreate; i++ {
		j, err := jm.StartJob(time.Now().UTC(), arn, payload, nil)
		if err != nil {
			log.Errorf("failed to create test job i=%v, with error: %v", i, err)
			return
		}
		log.WithJob(j).Info("created")
	}
	taskDuration := time.Now().UTC().Sub(taskStart)

	// Start a scheduled job.
	log.Infof("creating %v schedules", schedulesToCreate)
	sm := mysql.NewScheduleManager(dsn)
	for i := 0; i < schedulesToCreate; i++ {
		// Run every minute.
		id, err := sm.Create(time.Now().UTC(), arn, payload, []string{"* * * * *"}, "externalid", "harness")
		log.Infof("created schedule %v", id)
		if err != nil {
			log.Errorf("failed to create schedule with error: %v", err)
		}
	}

	// Start a web server to count the work.
	handler := NewCountHandler(quitAfter)

	r := http.NewServeMux()
	r.Handle("/", handler)

	// Register pprof handlers
	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)

	s = &http.Server{
		Addr:           ":8080",
		Handler:        r,
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
	cmd.Env = append(cmd.Env, "CALLME_SCHEDULE_WORKER_COUNT="+strconv.Itoa(scheduleWorkerCount))
	cmd.Env = append(cmd.Env, "CALLME_JOB_WORKER_COUNT="+strconv.Itoa(jobWorkerCount))
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
		log.Infof("stop signal received")
		break
	case <-handler.Completed:
		log.Infof("test complete")
		break
	}
	runDuration := time.Now().UTC().Sub(runStart)

	log.Infof("shutting down web server")
	if err := s.Close(); err != nil {
		log.Errorf("failed to shut down server with error: %v", err)
	}
	log.Infof("killing process")
	if err := cmd.Process.Kill(); err != nil {
		log.Errorf("failed to kill process with error: %v", err)
	}

	// Write out a summary
	log.Infof("created %v jobs in %f seconds\n", jobsToCreate, taskDuration.Seconds())
	log.Infof("web server received %v messages in %f seconds\n", handler.Received, runDuration.Seconds())
}

var s *http.Server

// NewCountHandler creates a HTTP handler which counts incoming requests.
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

// CountHandler provides a Web server to count incoming HTTP requests.
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

// Shutdown shuts down the process.
func (h *CountHandler) Shutdown() {
	// stop the receiver
	h.stopper <- true
	// tell clients that we're finished
	h.Completed <- true
}

// Receive handles incrementing the number of received messages.
func (h *CountHandler) Receive() {
	go func() {
		for {
			select {
			case <-h.c:
				h.Received++
				log.Infof("received: %v", h.Received)
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
	log.Infof("received HTTP request %v", h.Received)
	io.WriteString(w, "OK\n")
}

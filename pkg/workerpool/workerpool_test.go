package workerpool

import (
	"runtime"
	"testing"

	tassert "github.com/stretchr/testify/assert"
)

func TestNewWorkerPool(t *testing.T) {
	assert := tassert.New(t)
	wp := NewWorkerPool(0)

	assert.Equal(wp.GetWorkerNumber(), runtime.GOMAXPROCS(-1))
	wp.Stop()

	wp = NewWorkerPool(25)
	assert.Equal(wp.GetWorkerNumber(), 25)
	wp.Stop()
}

// Sample test job below for testing
type testJob struct {
	jobDone chan struct{}
	hash    uint64
}

func (tj *testJob) GetDoneCh() chan struct{} {
	return tj.jobDone
}

func (tj *testJob) Run() {
	// Just signal back we are done
	tj.jobDone <- struct{}{}
}

func (tj *testJob) JobName() string {
	return "testJob"
}

func (tj *testJob) Hash() uint64 {
	return tj.hash
}

// Uses AddJob, which relies on job hash for queue assignment
func TestAddJob(t *testing.T) {
	njobs := 10 // also worker routines
	wp := NewWorkerPool(njobs)
	joblist := make([]testJob, njobs)

	// Create and add jobs
	for i := 0; i < njobs; i++ {
		joblist[i] = testJob{
			jobDone: make(chan struct{}, 1),
			hash:    uint64(i),
		}

		wp.AddJob(&joblist[i])
	}

	// Verify all jobs ran through the workers
	for i := 0; i < njobs; i++ {
		<-joblist[i].jobDone
	}

	wp.Stop()
}

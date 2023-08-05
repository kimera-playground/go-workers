package main

import (
	"fmt"
	"sync"
	"time"
)

type Job[T, K any] struct {
	ID   int
	Fn   func(args T) (K, error)
	Args T
}

type Result[T any] struct {
	JobID int
	Err   error
	Value T
}

func (j *Job[T, K]) Execute() Result[K] {
	result, err := j.Fn(j.Args)
	if err != nil {
		return Result[K]{
			JobID: j.ID,
			Err:   err,
		}
	}

	return Result[K]{
		JobID: j.ID,
		Value: result,
	}
}

func worker[T, K any](wg *sync.WaitGroup, id int, jobs <-chan Job[T, K], results chan<- Result[K], shutdown <-chan bool) {
	defer wg.Done()

	for {
		select {
		case job, ok := <-jobs:
			if !ok {
				return
			}
			fmt.Printf("Worker %d running job %d \n", id, job.ID)
			result := job.Execute()
			results <- result
		case <-shutdown:
			fmt.Println("Shutting down`")
			return
		}
	}
}

func Process(id any) (any, error) {
	time.Sleep(1 * time.Second)

	x, _ := any(id).(int)
	return x * 10, nil
}

func runWorker() {
	const (
		concurreny = 2
		tasks      = 6
	)

	jobQueue := make(chan Job[any, any], tasks)
	resultQueue := make(chan Result[any], tasks)
	shutdown := make(chan bool)

	var wg sync.WaitGroup
	for w := 1; w <= concurreny; w++ {
		wg.Add(1)
		go worker(&wg, w, jobQueue, resultQueue, shutdown)
	}

	go func() {
		defer close(jobQueue)
		for i := 1; i <= tasks; i++ {
			jobQueue <- Job[any, any]{
				ID:   i,
				Fn:   Process,
				Args: i,
			}
		}
	}()

	go func() {
		defer close(resultQueue)

		for r := range resultQueue {
			fmt.Printf("Result %d \n", r.Value)
		}
	}()

	wg.Wait()

}

package main

import (
	"context"
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

type WorkerPool[T, K any] struct {
	workerCount int
	jobs        chan Job[T, K]
	results     chan Result[K]
	done        chan struct{}
}

func New[T, K any](numWorkers int, jobQueue chan Job[T, K]) WorkerPool[T, K] {
	return WorkerPool[T, K]{
		workerCount: numWorkers,
		jobs:        jobQueue,
		results:     make(chan Result[K]),
		done:        make(chan struct{}),
	}
}

func (wp *WorkerPool[T, K]) Run(ctx context.Context) {
	var wg sync.WaitGroup
	for i := 1; i <= wp.workerCount; i++ {
		wg.Add(1)
		go worker(ctx, &wg, i, wp.jobs, wp.results)
	}

	wg.Wait()
	close(wp.results)
	close(wp.done)
}

func (wp *WorkerPool[T, K]) Results() <-chan Result[K] {
	return wp.results
}

func worker[T, K any](ctx context.Context, wg *sync.WaitGroup, id int, jobs <-chan Job[T, K], results chan<- Result[K]) {
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
		case <-ctx.Done():
			results <- Result[K]{
				Err: ctx.Err(),
			}
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

	wp := New[any, any](concurreny, jobQueue)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*2)
	defer cancel()

	go wp.Run(ctx)
	// cancel()

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

	for {
		select {
		case result, ok := <-wp.results:
			if !ok {
				continue
			}

			if result.Err != nil {
				fmt.Printf("Got result error %s \n", result.Err.Error())
				return
			}
			fmt.Printf("Got result %d \n", result.Value.(int))
		case <-wp.done:
			return
		}
	}
}

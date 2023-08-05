package main

import (
	"fmt"
)

func test() {
	c := make(chan int, 5)
	r := make(chan int, 5)

	go func() {
		for {
			select {
			case i := <-c:
				// time.Sleep(1 * time.Second)
				r <- i
			}
		}
	}()

	go func() {
		for {
			select {
			case i := <-c:
				// time.Sleep(1 * time.Second)
				r <- i
			}
		}
	}()

	go func() {
		// defer close(c)
		for i := 1; i <= 5; i++ {
			c <- i
		}
	}()

	for i := 1; i <= 5; i++ {
		fmt.Printf("Getting %d \n", i)
		r := <-r
		fmt.Printf("Got %d", r)
	}
}

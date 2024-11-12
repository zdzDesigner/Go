package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Job struct {
	id       int
	randomno int
}
type Result struct {
	job         Job
	sumofdigits int
}

var jobs = make(chan Job, 2)
var results = make(chan Result, 2)

func digits(number int) int {
	sum := 0
	no := number
	for no != 0 {
		digit := no % 10
		sum += digit
		no /= 10
	}
	time.Sleep(time.Second)
	return sum
}

func worker(wg *sync.WaitGroup) {
	for job := range jobs {
		results <- Result{job, digits(job.randomno)}
	}
	wg.Done()
}

func createWorkerPool(n int) {
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go worker(&wg)
	}
	wg.Wait()
	close(results)
}

func allocate(n int) {
	for i := 0; i < n; i++ {
		fmt.Println("--------------")
		randomno := rand.Intn(999)
		jobs <- Job{i, randomno}
	}
	close(jobs)
}

func result(done chan bool) {
	for result := range results {
		fmt.Printf("Job id %d, input random no %d , sum of digits %d\n", result.job.id, result.job.randomno, result.sumofdigits)
	}
	done <- true
}

func result2() {
	for result := range results {
		fmt.Printf("Job id %d, input random no %d , sum of digits %d\n", result.job.id, result.job.randomno, result.sumofdigits)
	}
}

func main() {
	startTime := time.Now()
	go allocate(10)
	done := make(chan bool)
	go result(done)
	// go result2()
	createWorkerPool(5)
	<-done

	endTime := time.Now()
	diff := endTime.Sub(startTime)
	fmt.Println("total time taken ", diff.Seconds(), "seconds")
}

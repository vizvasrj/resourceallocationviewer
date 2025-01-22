package main

import (
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"
)

func stressCPUWorker(duration time.Duration, wg *sync.WaitGroup, results chan<- int) {
	defer wg.Done()
	start := time.Now()
	var count int
	for time.Since(start) < duration {
		count++
	}
	results <- count
}

func stressCPU(duration time.Duration, numWorkers int) float64 {
	var wg sync.WaitGroup
	results := make(chan int, numWorkers)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go stressCPUWorker(duration, &wg, results)
	}

	wg.Wait()
	close(results)

	var totalOps int
	for result := range results {
		totalOps += result
	}

	elapsed := duration.Seconds()
	return float64(totalOps) / elapsed
}

func getMemoryUsage() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}

func cpuUsageHandler(w http.ResponseWriter, r *http.Request) {
	duration := 5 * time.Second    // Stress CPU for 5 seconds
	numWorkers := runtime.NumCPU() // Use the number of available CPU cores
	cpuUsage := stressCPU(duration, numWorkers)

	fmt.Fprintf(w, "WORKIER %d CPU usage: %.2f operations per second\n", numWorkers, cpuUsage)
}

func memoryUsageHandler(w http.ResponseWriter, r *http.Request) {
	memUsage := getMemoryUsage()
	memUsageMB := float64(memUsage) / (1024 * 1024)
	fmt.Fprintf(w, "Memory usage: %.2f MB\n", memUsageMB)
}

func main() {

	http.HandleFunc("/cpu", cpuUsageHandler)
	http.HandleFunc("/memory", memoryUsageHandler)

	fmt.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}

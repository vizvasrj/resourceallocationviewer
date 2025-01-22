package main

import (
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"syscall"
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

func getSystemMemory() (total, free uint64) {
	var sysinfo syscall.Sysinfo_t
	err := syscall.Sysinfo(&sysinfo)
	if err != nil {
		return 0, 0
	}
	total = sysinfo.Totalram * uint64(sysinfo.Unit) / (1024 * 1024)
	free = sysinfo.Freeram * uint64(sysinfo.Unit) / (1024 * 1024)
	return total, free
}

// FormatNumberIndian formats a number according to the Indian numbering system.
func FormatNumberIndian(num float64) string {
	intPart := int64(num)
	decimalPart := strconv.FormatFloat(num-float64(intPart), 'f', 2, 64)[1:] // Decimal part with a dot
	intStr := strconv.FormatInt(intPart, 10)

	// Handle Indian number grouping
	n := len(intStr)
	if n <= 3 {
		return intStr + decimalPart
	}

	// First group of 3 digits
	result := intStr[n-3:]
	intStr = intStr[:n-3]

	// Groups of 2 digits
	for len(intStr) > 2 {
		result = intStr[len(intStr)-2:] + "," + result
		intStr = intStr[:len(intStr)-2]
	}

	// Remaining digits
	if len(intStr) > 0 {
		result = intStr + "," + result
	}

	return result + decimalPart
}

// Simulates CPU stress for a specific duration using multiple workers

// Handles the CPU usage endpoint
func cpuUsageHandler(w http.ResponseWriter, r *http.Request) {
	duration := 5 * time.Second    // Stress CPU for 5 seconds
	numWorkers := runtime.NumCPU() // Use the number of available CPU cores
	cpuUsage := stressCPU(duration, numWorkers)

	cpuUsageStr := FormatNumberIndian(cpuUsage) // Format with Indian grouping
	fmt.Fprintf(w, "WORKER %d CPU usage: %s operations per second\n", numWorkers, cpuUsageStr)
}

func memoryUsageHandler(w http.ResponseWriter, r *http.Request) {
	memUsage := getMemoryUsage()
	memUsageMB := float64(memUsage) / (1024 * 1024)

	totalMem, freeMem := getSystemMemory()

	fmt.Fprintf(w, "Memory usage: %.2f MB\n", memUsageMB)
	fmt.Fprintf(w, "Total memory: %d MB\n", totalMem)
	fmt.Fprintf(w, "Free memory: %d MB\n", freeMem)
}

func main() {
	http.HandleFunc("/cpu", cpuUsageHandler)
	http.HandleFunc("/memory", memoryUsageHandler)

	fmt.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}

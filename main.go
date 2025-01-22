package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func readFile(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func getCPULimit() (float64, error) {
	quotaStr, err := readFile("/sys/fs/cgroup/cpu/cpu.cfs_quota_us")
	if err != nil {
		return 0, err
	}
	periodStr, err := readFile("/sys/fs/cgroup/cpu/cpu.cfs_period_us")
	if err != nil {
		return 0, err
	}

	quota, err := strconv.ParseFloat(quotaStr, 64)
	if err != nil {
		return 0, err
	}
	period, err := strconv.ParseFloat(periodStr, 64)
	if err != nil {
		return 0, err
	}

	if quota == -1 {
		return -1, nil // No CPU limit
	}

	return quota / period, nil
}

func getMemoryLimit() (uint64, error) {
	memLimitStr, err := readFile("/sys/fs/cgroup/memory/memory.limit_in_bytes")
	if err != nil {
		return 0, err
	}

	memLimit, err := strconv.ParseUint(memLimitStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return memLimit, nil
}

func cpuLimitHandler(w http.ResponseWriter, r *http.Request) {
	cpuLimit, err := getCPULimit()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting CPU limit: %v", err), http.StatusInternalServerError)
		return
	} else if cpuLimit == -1 {
		fmt.Fprintln(w, "No CPU limit")
	} else {
		fmt.Fprintf(w, "CPU limit: %.2f cores\n", cpuLimit)
	}
}

func memoryLimitHandler(w http.ResponseWriter, r *http.Request) {
	memLimit, err := getMemoryLimit()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting memory limit: %v", err), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Memory limit: %d bytes\n", memLimit)
}

func main() {
	http.HandleFunc("/cpu", cpuLimitHandler)
	http.HandleFunc("/memory", memoryLimitHandler)

	fmt.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}

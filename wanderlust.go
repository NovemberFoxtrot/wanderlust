package main

import (
	"fmt"
	"image"

	_ "image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"os"
	"path/filepath"
	"runtime"
)

func process(path string) (string, bool) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, path, "panic", r)
		}
	}()

	file, err := os.Open(path)
	defer file.Close()

	if err != nil {
		fmt.Fprintln(os.Stderr, path, "error", err)
	}

	_, _, err = image.Decode(file)

	if err != nil {
		return path, false
	}

	return path, true
}

func readDir(dirname string) ([]os.FileInfo, error) {
	f, err := os.Open(dirname)

	if err != nil {
		return nil, err
	}

	list, err := f.Readdir(-1)

	f.Close()

	if err != nil {
		return nil, err
	}

	return list, nil
}

func addJobs(root string, jobs chan<- string) {
	root = filepath.Clean(root)

	fi, err := os.Lstat(root)
	isDir := fi.IsDir()
	isSymLink := fi.Mode()&os.ModeSymlink != 0

	if isDir == false || isSymLink == true || err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fileIDs, err := readDir(root)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, fileID := range fileIDs {
		if fileID.IsDir() {
			addJobs(filepath.Join(root, fileID.Name()), jobs)
		} else {
			jobs <- filepath.Join(root, fileID.Name())
		}
	}

	// close(jobs)
}

func runJobs(done chan<- struct{}, results chan<- string, jobs <-chan string) {
	for job := range jobs {
		if result, ok := process(job); ok {
			results <- result
		}
	}
	done <- struct{}{}
}

func waitAndProcessResults(done <-chan struct{}, results <-chan string) {
	for working := workers; working > 0; {
		select {
		case result := <-results:
			fmt.Println(result)
		case <-done:
			working--
		}
	}
}

var workers = runtime.NumCPU()

func main() {
	runtime.GOMAXPROCS(workers)

	root := os.Args[1]

	jobs := make(chan string, workers)
	results := make(chan string)
	done := make(chan struct{}, workers)

	go addJobs(root, jobs)

	for i := 0; i < workers; i++ {
		go runJobs(done, results, jobs)
	}

	waitAndProcessResults(done, results)
}

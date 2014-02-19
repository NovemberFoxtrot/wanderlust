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

func process(path string, f os.FileInfo, err error) bool {
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
		return false
	}

	return true
}

func addJobs(files []string, jobs chan<- string) {
	for _, filename := range files {
		jobs <- filename
	}
}

func runJobs(done chan<- struct{}, results chan<- string, jobs <-chan string) {
	for job := range jobs {
		if result, ok := process(job); ok {
			results <- result
		}
	}
	done <- struct{}{}
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

func wander(path string, paths chan os.FileInfo) {
	path = filepath.Clean(path)

	fi, err := os.Lstat(path)
	isDir := fi.IsDir()
	isSymLink := fi.Mode()&os.ModeSymlink != 0

	if isDir == false || isSymLink == true || err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fileIDs, err := readDir(path)

	fmt.Println(fileIDs, err)

	for _ = range fileIDs {
		paths <- fi
	}
}

func hunt(paths chan os.FileInfo) {
	fmt.Println("hunting")

	for {
		path := <-paths
		fmt.Println(path)
	}
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
DONE:
	for {
		select {
		case result := <-results:
			fmt.Println(results)
		default:
			break DONE
		}
	}
}

var workers = runtime.NumCPU()

func main() {

	runtime.GOMAXPROCS(workers)

	paths := make(chan os.FileInfo, workers)
	jobs := make(chan string, workers)
	results := make(chan string)
	done := make(chan struct{}, workers)

	root := os.Args[1]

	go addJobs()

	for i := 0; i < workers; i++ {
		go runJobs(done, results, jobs)
	}

	waitAndProcessResults(done, results)
}

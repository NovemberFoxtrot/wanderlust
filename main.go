package main

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	_ "image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/gosexy/exif"
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

	sourceimage, _, err := image.Decode(file)

	if err != nil {
		return path, false
	}

	reader := exif.New()

	err = reader.Open(path)

	if err != nil {
		// fmt.Println("Error: %s", err.Error())
	} else {

		for key, val := range reader.Tags {
			fmt.Printf("%s: %s\n", key, val)
		}
	}

	sourcebounds := sourceimage.Bounds()

	colors := make(map[string]int)

	for y := sourcebounds.Min.Y; y < sourcebounds.Max.Y; y++ {
		for x := sourcebounds.Min.X; x < sourcebounds.Max.X; x++ {
			sr, sg, sb, sa := sourceimage.At(x, y).RGBA()

			colors[strconv.FormatUint(uint64(sr), 10)+strconv.FormatUint(uint64(sg), 10)+strconv.FormatUint(uint64(sb), 10)+strconv.FormatUint(uint64(sa), 10)] += 1
		}
	}

	return path, true
}

func readDir(dirname string) ([]os.FileInfo, error) {
	f, err := os.Open(dirname)
	defer f.Close()

	if err != nil {
		return nil, err
	}

	list, err := f.Readdir(-1)

	if err != nil {
		return nil, err
	}

	return list, nil
}

func processDirectory(root string) []TheJob {
	root = filepath.Clean(root)

	fi, err := os.Lstat(root)

	if err != nil {
		fmt.Println(err)
		return []TheJob{}
	}

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

	var thejobs []TheJob

	for _, fileID := range fileIDs {
		thejob := TheJob{parent: root, fileinfo: fileID}
		thejobs = append(thejobs, thejob)
	}

	return thejobs
}

type TheJob struct {
	parent   string
	fileinfo os.FileInfo
}

func addJobs(root string, jobs chan<- string) {
	thejobs := processDirectory(root)

	for len(thejobs) > 0 {
		thejob := thejobs[0]

		if thejob.fileinfo.IsDir() {
			moreIDs := processDirectory(filepath.Join(thejob.parent, thejob.fileinfo.Name()))

			if len(moreIDs) > 0 {
				thejobs = append(thejobs, moreIDs...)
			}
		} else {
			jobs <- filepath.Join(thejob.parent, thejob.fileinfo.Name())
		}

		thejobs = thejobs[1:]
	}

	close(jobs)
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

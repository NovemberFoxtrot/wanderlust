package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
)

func visit(path string, f os.FileInfo, err error) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, path, "panic in f", r)
		}
	}()

	file, err := os.Open(path)
	defer file.Close()

	if err != nil {
		fmt.Fprintln(os.Stderr, path, "error", err)
	}

	if f.IsDir() {
		return nil
	}

	_, _, err = image.Decode(file)

	if err != nil {
		return nil
	}

	fmt.Println(path)

	return nil
}

func main() {
	flag.Parse()
	root := flag.Arg(0)
	filepath.Walk(root, visit)
}

package utils

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
)

// FindFile will crawl a directory to find a file
func FindFile(dir string, ext string) []string {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Fatal(err)
	}
	var fileList []string
	filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		r, err := regexp.MatchString(ext, f.Name())
		if err == nil && r {
			fileList = append(fileList, path)
		}
		return nil
	})
	return fileList
}

// TrailingSlash checks to see if a string has a `/` at the end
// it will add a trailing slash if it does not already have one.
func TrailingSlash(dir string) string {
	if string(dir[len(dir)-1]) != "/" {
		dir = dir + "/"
	}
	return dir
}

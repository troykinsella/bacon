package util

import (
	"fmt"
	"os"
)

func IsDir(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	s, err := f.Stat()
	if err != nil {
		return false, err
	}
	return s.IsDir(), nil
}

func Cls() {
	fmt.Print("\033c")
}

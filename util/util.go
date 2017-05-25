package util

import (
	"os"
	"fmt"
)

func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

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

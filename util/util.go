package util

import (
	"fmt"
	"os"
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
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
	s, err := f.Stat()
	if err != nil {
		return false, err
	}
	return s.IsDir(), nil
}

func Cls() {
	fmt.Print("\033c")
}

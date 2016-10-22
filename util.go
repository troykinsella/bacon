package main

import (
	"time"
	"fmt"
	"strings"
	"path/filepath"
	"github.com/bmatcuk/doublestar"
	"os"
	"log"
)

func cwd() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

func isDir(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	s, err := f.Stat()
	if err != nil {
		return false, err
	}
	return s.IsDir(), nil
}

func pathMatch(path string, glob string) (bool, error) {
	// special case: allow a directory to be matched when suffixed wth /**
	const sfx = "/**"
	if strings.HasSuffix(glob, sfx) &&
		filepath.Join(path, sfx) == glob {

		m, err := doublestar.PathMatch(glob[0:len(glob)-len(sfx)], path)
		if err != nil {
			return false, err
		}

		if m {
			return true, nil
		}
	}

	m, err := doublestar.PathMatch(glob, path)
	if err != nil {
		return false, err
	}

	return m, nil
}

func included(path string, includes []string) (bool, error) {
	for _, i := range includes {
		m, err := pathMatch(path, i)
		if err != nil {
			return false, err
		}
		if m {
			return true, nil
		}
	}

	return false, nil
}

func excluded(path string, excludes []string) (bool, error) {
	for _, e := range excludes {
		m, err := pathMatch(path, e)
		if err != nil {
			return false, err
		}
		if m {
			return true, nil
		}
	}
	return false, nil
}

func selected(path string, includes []string, excludes []string) (bool, error) {
	inc, err := included(path, includes)
	if err != nil {
		return false, err
	}
	if !inc {
		return false, nil
	}

	ex, err := excluded(path, excludes)
	if err != nil {
		return false, err
	}
	if ex {
		return false, nil
	}

	return true, nil
}

func normalizeGlobs(globs []string, defalt string) []string {
	if globs == nil || len(globs) == 0 {
		globs = []string{defalt}
	}

	goSrc := cwd()
	for i, g := range globs {
		if strings.HasPrefix(g, "/") {
			globs[i] = g
		} else {
			globs[i] = filepath.Join(goSrc, g)
		}
	}

	return globs
}

func expandedBaseDirs(includes []string, excludes []string) ([]string, error) {
	result := []string{}

	baseDirs, err := staticBaseDirs(includes)
	if err != nil {
		return nil, err
	}

	for _, baseDir := range baseDirs {
		err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			d, err := isDir(path)
			if err != nil {
				return err
			}
			if !d {
				return nil
			}

			s, err := selected(path, includes, excludes)
			if err != nil {
				return err
			}

			if s {
				result = append(result, path)
				return nil
			}

			return filepath.SkipDir
		})
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// Get the deepest directory path before any glob expressions occur
func staticBaseDir(glob string) (string, error) {
	parts := strings.FieldsFunc(glob, func(r rune) bool {
		switch r {
		case '*', '?', '[', '{', '\\':
			return true
		}
		return false
	})

	base := parts[0]
	d, err := isDir(base)
	if err != nil {
		return "", err
	}
	if d {
		return base, nil
	}
	return filepath.Dir(base), nil
}

func staticBaseDirs(globs []string) ([]string, error) {
	m := make(map[string]bool)
	for _, g := range globs {
		b, err := staticBaseDir(g)
		if err != nil {
			return nil, err
		}
		m[b] = true
	}

	bases :=[]string{}
	for b, _ := range m {
		bases = append(bases, b)
	}
	return bases, nil
}

func cls() {
	fmt.Print("\033c")
}

func printMessage(m string) {
	t := time.Now()
	fmt.Printf("[%s] %s\n", t.Format("15:04:05"), m)
}

func printRunning() {
	printMessage("\033[33m→ Running\033[0m")
}

func printSummary(pass bool) {
	var msg string
	if pass {
		msg = "\033[32m✓ Passed\033[0m"
	} else {
		msg = "\033[31m✗ Failed\033[0m"
	}
	printMessage(msg)
}

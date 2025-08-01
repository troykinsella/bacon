package expander

import (
	"github.com/bmatcuk/doublestar"
	"github.com/troykinsella/bacon/util"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type E struct {
	includes []string
	excludes []string
}

func New(
	dir string,
	includes []string,
	excludes []string,
) *E {
	includes = normalizeGlobs(dir, includes, "**/*", false)
	excludes = normalizeGlobs(dir, excludes, "**/.*", true)

	return &E{
		includes: includes,
		excludes: excludes,
	}
}

func (e *E) BaseDirs() ([]string, error) {
	set := make(map[string]bool)

	for _, inc := range e.includes {
		err := baseDir(inc, e.excludes, set)
		if err != nil {
			return nil, err
		}
	}

	result := []string{}
	for p := range set {
		result = append(result, p)
	}

	return result, nil
}

func baseDir(inc string, excludes []string, resultSet map[string]bool) error {
	ms, err := doublestar.Glob(inc)
	if err != nil {
		return err
	}

	for _, m := range ms {
		d, err := util.IsDir(m)
		if err != nil {
			continue
		}
		if !d {
			m = filepath.Dir(m)
		}

		ex, err := matches(m, excludes)
		if err != nil {
			return err
		}

		if !ex {
			resultSet[m] = true
		}
	}

	return nil
}

func (e *E) List() ([]string, error) {
	var result []string

	baseDirs, err := e.BaseDirs()
	if err != nil {
		return nil, err
	}

	for _, dirName := range baseDirs {
		dir, err := os.Open(dirName)
		if err != nil {
			return nil, err
		}

		fis, err := dir.Readdir(0)
		_ = dir.Close()
		if err != nil {
			return nil, err
		}

		for _, fi := range fis {
			if fi.IsDir() {
				continue
			}

			fp := filepath.Join(dirName, fi.Name())
			s, err := e.Selected(fp)
			if err != nil {
				return nil, err
			}
			if !s {
				continue
			}

			result = append(result, fp)
		}
	}

	sort.Strings(result)
	return result, nil
}

func (e *E) Selected(path string) (bool, error) {
	path = ensureRooted(path)
	return selected(path, e.includes, e.excludes)
}

func (e *E) pathSelected(path string) bool {
	sel, err := selected(path, e.includes, e.excludes)
	if err != nil {
		return false
	}
	return sel
}

func matches(path string, includes []string) (bool, error) {
	for _, i := range includes {
		if path == i {
			return true, nil
		}
		m, err := doublestar.PathMatch(i, path)
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
	inc, err := matches(path, includes)
	if err != nil {
		return false, err
	}
	if !inc {
		return false, nil
	}

	ex, err := matches(path, excludes)
	if err != nil {
		return false, err
	}
	if ex {
		return false, nil
	}

	return true, nil
}

func normalizeGlobs(dir string, globs []string, defalt string, exclude bool) []string {
	if globs == nil || len(globs) == 0 {
		globs = []string{defalt}
	}

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	if dir == "" {
		dir = cwd
	} else if !filepath.IsAbs(dir) {
		dir = filepath.Join(cwd, dir)
	}

	for i, g := range globs {
		if filepath.IsAbs(g) {
			globs[i] = g
		} else {
			globs[i] = filepath.Join(dir, g)
		}
	}

	if exclude {
		// Special case: If we exclude a directory, we must exclude all children
		for _, g := range globs {
			if !strings.HasSuffix(g, "/**") {
				globs = append(globs, g+"/**")
			}
		}
	}

	return globs
}

func ensureRooted(path string) string {
	if !filepath.IsAbs(path) {
		cwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		path = filepath.Join(cwd, path)
	}
	return path
}

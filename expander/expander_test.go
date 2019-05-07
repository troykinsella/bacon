package expander

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func prefix(s []string, p string) []string {
	r := make([]string, len(s))
	for i, e := range s {
		r[i] = filepath.Join(p, e)
	}
	return r
}

func normalizePaths(p []string) {
	for i, s := range p {
		if strings.HasSuffix(s, string(os.PathSeparator)) {
			p[i] = s[0 : len(s)-1]
		}
	}
}

func pathsEqual(a []string, b []string) bool {
	sort.Strings(a)
	sort.Strings(b)

	normalizePaths(a)
	normalizePaths(b)

	return reflect.DeepEqual(a, b)
}

func TestE_BaseDirs(t *testing.T) {
	t.Parallel()

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	cwd = filepath.Join(cwd, "testdata")

	var tests = []struct {
		inc []string
		exc []string
		exp []string
		err string
	}{
		{[]string{"a"}, []string{}, []string{"a"}, ""},
		{[]string{"a/"}, []string{}, []string{"a"}, ""},
		{[]string{"a/b"}, []string{}, []string{"a/b"}, ""},
		{[]string{"a/b/"}, []string{}, []string{"a/b"}, ""},
		{[]string{"a/b", "a/c"}, []string{}, []string{"a/b", "a/c"}, ""},
		{[]string{"a/b/d"}, []string{}, []string{"a/b/d"}, ""},
		{[]string{"a", "a/b/d"}, []string{}, []string{"a", "a/b/d"}, ""},

		{[]string{"noap"}, []string{}, []string{}, ""},
		{[]string{"noap*"}, []string{}, []string{}, ""},
		{[]string{"noap**"}, []string{}, []string{}, ""},
		{[]string{"noap/*"}, []string{}, []string{}, ""},
		{[]string{"noap/**"}, []string{}, []string{}, ""},

		{[]string{"a/*"}, []string{}, []string{"a", "a/b", "a/c"}, ""},
		{[]string{"a/**"}, []string{}, []string{"a", "a/b", "a/c", "a/b/d"}, ""},
		{[]string{"**/a"}, []string{}, []string{"a"}, ""},
		{[]string{"**/a/**"}, []string{}, []string{"a", "a/b", "a/c", "a/b/d"}, ""},
		{[]string{"a/**/d"}, []string{}, []string{"a/b/d"}, ""},

		{[]string{"a/*1"}, []string{}, []string{"a"}, ""},
		{[]string{"a/b/*1"}, []string{}, []string{"a/b"}, ""},
		{[]string{"a/b/d/*1"}, []string{}, []string{"a/b/d"}, ""},
		{[]string{"a/**/*1"}, []string{}, []string{"a", "a/b", "a/b/d", "a/c"}, ""},

		{[]string{"a/**"}, []string{"**/b"}, []string{"a", "a/c"}, ""},
		{[]string{"a/**"}, []string{"**/b/**"}, []string{"a", "a/b", "a/c"}, ""},
		{[]string{"a/**"}, []string{"**/b", "**/b/**"}, []string{"a", "a/c"}, ""},
	}

	for i, test := range tests {
		test.inc = prefix(test.inc, cwd)
		test.exc = prefix(test.exc, cwd)
		test.exp = prefix(test.exp, cwd)

		e := New("", test.inc, test.exc)
		dirs, err := e.BaseDirs()

		if test.err == "" {
			if err != nil {
				t.Errorf("%d. unexpected error: %s\n", i, err.Error())
			} else if !pathsEqual(dirs, test.exp) {
				t.Errorf("%d. unexpected result:\nexpected=%s,\nactual=%s\n", i, test.exp, dirs)
			}
		} else {
			if err == nil {
				t.Errorf("%d. expected error:\nexpected=%s,\nactual=nil\n", i, test.err)
			} else if test.err != err.Error() {
				t.Errorf("%d. unexpected error:\nexpected=%s,\nactual=%s\n", i, test.err, err.Error())
			}
		}
	}
}

func TestE_Selected(t *testing.T) {
	t.Parallel()

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	cwd = filepath.Join(cwd, "testdata")

	var tests = []struct {
		path string
		inc  []string
		exc  []string
		exp  bool
		err  string
	}{
		{"foo", []string{"foo"}, []string{}, true, ""},
		{"foo/", []string{"foo"}, []string{}, true, ""},
		{"/foo", []string{"/foo"}, []string{}, true, ""},
		{"/foo/", []string{"/foo"}, []string{}, true, ""},
		{"foo/bar", []string{"foo"}, []string{""}, false, ""},

		{"foo1", []string{"foo*"}, []string{}, true, ""},
		{"foo2", []string{"foo**"}, []string{}, true, ""},
		{"foo3", []string{"foo/*"}, []string{}, false, ""},
		{"foo4", []string{"foo/**"}, []string{}, false, ""},

		{"foo/bar1", []string{"foo/*"}, []string{}, true, ""},
		{"foo/bar2", []string{"foo/**"}, []string{}, true, ""},
		{"foo/bar/baz1", []string{"foo/*"}, []string{}, false, ""},
		{"foo/bar/baz2", []string{"foo/**"}, []string{}, true, ""},
	}

	for i, test := range tests {
		e := New("", test.inc, test.exc)
		r, err := e.Selected(test.path)
		if test.err == "" {
			if err != nil {
				t.Errorf("%d. \"%s\" unexpected error: %s\n", i, test.path, err.Error())
			} else if r != test.exp {
				t.Errorf("%d. \"%s\" unexpected result:\nexpected=%t,\nactual=%t\n", i, test.path, test.exp, r)
			}
		} else {
			if err == nil {
				t.Errorf("%d. \"%s\" expected error:\nexpected=%s,\nactual=nil\n", i, test.path, test.err)
			} else if test.err != err.Error() {
				t.Errorf("%d. \"%s\" unexpected error:\nexpected=%s,\nactual=%s\n", i, test.path, test.err, err.Error())
			}
		}
	}
}

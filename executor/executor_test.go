package executor

import (
	"bytes"
	"reflect"
	"testing"
)

func TestE_RunCommands(t *testing.T) {
	var tests = []struct {
		commands     []string
		passCommands []string
		failCommands []string
		showOutput   bool
		args         []string

		expectedResult *Result
		expectedOutput string
	}{
		// Test showOutput variations
		{ // Hide output when disabled
			[]string{"echo foo"},
			[]string{},
			[]string{},
			false,
			[]string{},
			&Result{true, true, true},
			"",
		},
		{ // Show output when enabled
			[]string{"echo foo"},
			[]string{},
			[]string{},
			true,
			[]string{},
			&Result{true, true, true},
			"foo\n",
		},
		{ // Multiple commands
			[]string{"echo foo", "echo bar"},
			[]string{},
			[]string{},
			true,
			[]string{},
			&Result{true, true, true},
			"foo\nbar\n",
		},
		{ // Show error when output disabled
			[]string{"echo foo 1>&2"},
			[]string{},
			[]string{},
			false,
			[]string{},
			&Result{true, true, true},
			"foo\n",
		},
		{ // Show error when output enabled
			[]string{"echo foo 1>&2"},
			[]string{},
			[]string{},
			true,
			[]string{},
			&Result{true, true, true},
			"foo\n",
		},
		{ // Show output and error when output enabled
			[]string{"echo foo; echo bar 1>&2"},
			[]string{},
			[]string{},
			true,
			[]string{},
			&Result{true, true, true},
			"foo\nbar\n",
		},
		{ // Output comes before error
			[]string{"echo foo; echo bar 1>&2; echo baz"},
			[]string{},
			[]string{},
			true,
			[]string{},
			&Result{true, true, true},
			"foo\nbaz\nbar\n",
		},

		// Test passing
		{ // Output comes before error
			[]string{"exit 1"},
			[]string{},
			[]string{},
			false,
			[]string{},
			&Result{false, true, true},
			"",
		},
		{ // Show output as error on failures
			[]string{"echo foo; exit 1"},
			[]string{},
			[]string{},
			false,
			[]string{},
			&Result{false, true, true},
			"foo\n",
		},
		{ // Show error as error on failures
			[]string{"echo foo; exit 1"},
			[]string{},
			[]string{},
			false,
			[]string{},
			&Result{false, true, true},
			"foo\n",
		},
		{ // Show output and error as error on failures
			[]string{"echo foo; echo bar 1>&2; exit 1"},
			[]string{},
			[]string{},
			false,
			[]string{},
			&Result{false, true, true},
			"foo\nbar\n",
		},

		// Passing/failing commands
		{ // Run pass commands on success
			[]string{"echo foo; exit 0"},
			[]string{"echo yes"},
			[]string{"echo no"},
			true,
			[]string{},
			&Result{true, true, true},
			"foo\nyes\n",
		},
		{ // Run fail commands on failure
			[]string{"echo foo; exit 1"},
			[]string{"echo yes"},
			[]string{"echo no"},
			true,
			[]string{},
			&Result{false, true, true},
			"foo\nno\n",
		},
		{ // Multiple pass commands
			[]string{"echo foo; exit 0"},
			[]string{"echo yes", "echo again"},
			[]string{"echo no"},
			true,
			[]string{},
			&Result{true, true, true},
			"foo\nyes\nagain\n",
		},
		{ // Multiple fail commands
			[]string{"echo foo; exit 1"},
			[]string{"echo yes"},
			[]string{"echo no", "echo again"},
			true,
			[]string{},
			&Result{false, true, true},
			"foo\nno\nagain\n",
		},
		{ // Pass command failure doesn't influence overall result
			[]string{"echo foo; exit 0"},
			[]string{"echo yes; exit 1"},
			[]string{"echo no"},
			true,
			[]string{},
			&Result{true, true, true},
			"foo\nyes\n",
		},
		{ // Fail command failure doesn't.. uh.. magically make the overall result success?
			[]string{"echo foo; exit 1"},
			[]string{"echo yes"},
			[]string{"echo no; exit 1"},
			true,
			[]string{},
			&Result{false, true, true},
			"foo\nno\n",
		},
		{ // Pass command doesn't output when output disabled
			[]string{"echo foo; exit 0"},
			[]string{"echo yes"},
			[]string{"echo no"},
			false,
			[]string{},
			&Result{true, true, true},
			"",
		},
		{ // Pass command errors output when output disabled
			[]string{"echo foo; exit 0"},
			[]string{"echo yes 1>&2"},
			[]string{"echo no"},
			false,
			[]string{},
			&Result{true, true, true},
			"yes\n",
		},
		{ // Fail command doesn't output when output disabled
			[]string{"echo foo; exit 1"},
			[]string{"echo yes"},
			[]string{"echo no"},
			false,
			[]string{},
			&Result{false, true, true},
			"foo\n",
		},
		{ // Fail command shows error when output disabled
			[]string{"echo foo; exit 1"},
			[]string{"echo yes"},
			[]string{"echo no 1>&2"},
			false,
			[]string{},
			&Result{false, true, true},
			"foo\nno\n",
		},
	}

	for i, test := range tests {
		var outBuf bytes.Buffer

		e := New(test.commands,
			test.passCommands,
			test.failCommands,
			"",
			"",
			test.showOutput)
		e.out = &outBuf
		e.err = &outBuf

		r := e.RunCommands("", test.args)
		outStr := outBuf.String()

		if !reflect.DeepEqual(r, test.expectedResult) {
			t.Errorf("%d. unexpected result:\nexpected=%#v,\nactual=%#v\n", i, test.expectedResult, r)
		}
		if outStr != test.expectedOutput {
			t.Errorf("%d. unexpected output/error:\nexpected=%#v,\nactual=%#v\n", i, test.expectedOutput, outStr)
		}
	}
}

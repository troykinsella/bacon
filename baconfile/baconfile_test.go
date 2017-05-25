package baconfile_test

import (
	"testing"
	"reflect"
	"github.com/troykinsella/bacon/baconfile"
)

func TestUnmarshal(t *testing.T) {
	var tests = []struct {
		b string
		exp *baconfile.B
		err string
	}{
		{
			"",
			nil,
			"Malformed Baconfile: Must supply at least one target",
		},
		{
			`--- { target: { foo: { watch: [bar], command: [echo] } } }`,
			&baconfile.B{
				Targets: map[string]*baconfile.Target{
					"foo": {
						Watch: []string{ "bar" },
						Command: []string{ "echo" },
					},
				},
			},
			"",
		},
	}

	for i, test := range tests {
		out, err := baconfile.Unmarshal([]byte(test.b))
		if test.err == "" {
			if err != nil {
				t.Errorf("%d. unexpected error: %s\n", i, err.Error())
			} else if !reflect.DeepEqual(out, test.exp) {
				t.Errorf("%d. unexpected result:\nexpected=%#v,\nactual=%#v\n", i, test.exp, out)
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
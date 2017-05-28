package baconfile

import (
	"gopkg.in/yaml.v2"
	"fmt"
)

var Version = "1.0"

type B struct {
	Version string `yaml:"version"`
	Targets map[string]*Target `yaml:"target"`
}

type Target struct {
	Watch   []string  `yaml:"watch"`
	Exclude []string  `yaml:"exclude,omitempty"`
	Dir     string    `yaml:"dir,omitempty"`
	Command []string  `yaml:"command"`
	Pass    []string  `yaml:"pass,omitempty"`
	Fail    []string  `yaml:"fail,omitempty"`
	Shell   string    `yaml:"shell,omitempty"`
}

func Unmarshal(bytes []byte) (*B, error) {
	var b B
	err := yaml.Unmarshal(bytes, &b)
	if err != nil {
		return nil, err
	}

	if err := b.Validate(); err != nil {
		return nil, err
	}

	return &b, nil
}

func errMalformed(msg string) error {
	return fmt.Errorf("Malformed Baconfile: %s", msg)
}

func (b *B) Validate() error {
	if len(b.Targets) == 0 {
		return errMalformed("Must supply at least one target")
	}

	for tName, t := range b.Targets {
		if len(t.Watch) == 0 {
			return errMalformed(fmt.Sprintf("Target '%s' must supply at least one 'watch' entry", tName))
		}
		if len(t.Command) == 0 && len(t.Pass) == 0 && len(t.Fail) == 0 {
			return errMalformed(fmt.Sprintf("Target '%s' must supply at least one 'command', 'pass', or 'fail' command", tName))
		}
	}

	return nil
}

func (b *B) Marshal() ([]byte, error) {
	if err := b.Validate(); err != nil {
		return nil, err
	}

	out, err := yaml.Marshal(b)
	if err != nil {
		return nil, err
	}

	return out, nil
}

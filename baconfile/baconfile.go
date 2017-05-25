package baconfile

import (
	"gopkg.in/yaml.v2"
	"errors"
)

var Version = "1.0"

type B struct {
	Version string `yaml:"version"`
	Targets map[string]*Target `yaml:"target"`
}

type Target struct {
	Watch []string   `yaml:"watch"`
	Exclude []string `yaml:"exclude,omitempty"`
	Run []string     `yaml:"run"`
	Pass []string    `yaml:"pass,omitempty"`
	Fail []string    `yaml:"fail,omitempty"`
	Shell string     `yaml:"shell,omitempty"`
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

func (b *B) Validate() error {
	if len(b.Targets) == 0 {
		return errors.New("Must supply at least one target")
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

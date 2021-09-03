package conf

import (
	"bytes"
	"fmt"
	"io"

	"gopkg.in/yaml.v2"
)

// ParseYaml parses yaml into the specified config struct. After the yaml is
// unmarshaled, the ParseOSArgs function is executed to apply defaults and
// overrides. Fields that are not set to their zero after the yaml is parsed
// will have the defaults ignored.
func ParseYaml(r io.Reader, prefix string, cfg interface{}) (string, error) {
	var b bytes.Buffer
	if _, err := b.ReadFrom(r); err != nil {
		return "", fmt.Errorf("reading yaml: %w", err)
	}

	err := yaml.Unmarshal(b.Bytes(), cfg)
	if err != nil {
		return "", fmt.Errorf("unmarshal yaml: %w", err)
	}

	return ParseOSArgs(prefix, cfg)
}

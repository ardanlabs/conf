package conf

import (
	"bytes"
	"fmt"
	"io"

	"gopkg.in/yaml.v2"
)

// YAML provides support for unmarshaling YAML into the applications
// config value. After the yaml is unmarshaled, the Parse function is
// executed to apply defaults and overrides. Fields that are not set to
// their zero after the yaml is parsed will have the defaults ignored.
type YAML struct {
	data []byte
}

// WithYaml accepts the yaml document as a slice of bytes.
func WithYaml(data []byte) YAML {
	return YAML{
		data: data,
	}
}

// WithYamlReader accepts a reader to read the yaml.
func WithYamlReader(r io.Reader) YAML {
	var b bytes.Buffer
	if _, err := b.ReadFrom(r); err != nil {
		return YAML{}
	}

	return WithYaml(b.Bytes())
}

// Process performs the actual processing of the yaml.
func (y YAML) Process(prefix string, cfg interface{}) error {
	err := yaml.Unmarshal(y.data, cfg)
	if err != nil {
		return fmt.Errorf("unmarshal yaml: %w", err)
	}
	return nil
}

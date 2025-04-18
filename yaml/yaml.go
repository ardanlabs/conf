// Package yaml provides yaml support for conf.
package yaml

import (
	"bytes"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// YAML provides support for unmarshalling YAML into the applications
// config value. After the yaml is unmarshalled, the Parse function is
// executed to apply defaults and overrides. Fields that are not set to
// their zero after the yaml is parsed will have the defaults ignored.
type YAML struct {
	data []byte
}

// WithData accepts the yaml document as a slice of bytes.
func WithData(data []byte) YAML {
	return YAML{
		data: data,
	}
}

// WithReader accepts a reader to read the yaml.
func WithReader(r io.Reader) YAML {
	var b bytes.Buffer
	if _, err := b.ReadFrom(r); err != nil {
		return YAML{}
	}

	return WithData(b.Bytes())
}

// Process performs the actual processing of the yaml.
func (y YAML) Process(prefix string, cfg interface{}) error {
	err := yaml.Unmarshal(y.data, cfg)
	if err != nil {
		return fmt.Errorf("unmarshal yaml: %w", err)
	}
	return nil
}

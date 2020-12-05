// Package json provides
package json

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/ardanlabs/conf"
)

type Source struct {
	m map[string]string
}

// NewSource returns a Source and, potentially, an error if a read
// error occurs or the Reader contains an invalid JSON document.
func NewSource(data []byte) (*Source, error) {
	config := make(map[string]interface{})
	err := json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("json.NewSource: %w", err)
	}

	m := make(map[string]string)
	for key, value := range config {
		switch v := value.(type) {
		case float64:
			m[key] = strings.TrimRight(fmt.Sprintf("%f", v), "0.")
		case bool:
			m[key] = fmt.Sprintf("%t", v)
		case string:
			m[key] = value.(string)
		}
	}

	return &Source{m: m}, nil
}

// SourceFrom ...
func SourceFrom(src io.Reader) (*Source, error) {
	data, err := ioutil.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("json.SourceFrom: %w", err)
	}

	return NewSource(data)
}

func (src *Source) Source(fld conf.Field) (string, bool) {
	if fld.Options.ShortFlagChar != 0 {
		flagKey := fld.Options.ShortFlagChar
		k := strings.ToLower(string(flagKey))
		if val, found := src.m[k]; found {
			return val, found
		}
	}

	k := strings.ToLower(strings.Join(fld.FlagKey, `_`))
	val, found := src.m[k]
	return val, found
}

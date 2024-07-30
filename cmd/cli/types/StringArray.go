package types

import (
	"encoding/json"
	"errors"
)

var (
	ErrUnmarshalNotSupported = errors.New("Unmarshal not supported")
)

// Custom type that accepts a String or Array of string when Unmarshalling JSON object
// It converts always to a array of string
type StringArray []string

func (sa *StringArray) UnmarshalJSON(data []byte) error {
	var array []string
	err := json.Unmarshal(data, &array)

	if err == nil {
		*sa = array
		return nil
	}

	stringValue := string(data)

	if stringValue == "null" {
		*sa = nil
		return nil
	}

	*sa = []string{stringValue}
	return nil
}

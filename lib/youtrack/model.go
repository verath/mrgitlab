package youtrack

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// Issue is a YouTrack issue as it is returned by the API.
type Issue struct {
	Fields []IssueField `json:"field"`
}

// IssueField is a field (aprox. key-value pair) attached to
// a YouTrack issue.
type IssueField struct {
	Name  string          `json:"name"`
	Value json.RawMessage `json:"value"`
}

// FieldStringValue is a helper method for extracting the value for a
// field with the provided name, where the value is expected to be
// a string. Returns an error if the field did not exist, or if
// the type of the field was not a string
func (issue *Issue) FieldStringValue(name string) (string, error) {
	for _, v := range issue.Fields {
		if v.Name == name {
			var value string
			return value, json.Unmarshal(v.Value, &value)
		}
	}
	return "", errors.Errorf("no value for Name '%s'", name)
}

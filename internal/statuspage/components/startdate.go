// Copyright 2021 Adam Chalkley
//
// https://github.com/atc0005/check-statuspage
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

package components

import (
	"encoding/json"
	"strings"
	"time"
)

// Credit:
//
// These resources were used while developing the json.Marshaler and
// json.Unmarshler interface implementations used in this file:
//
// https://stackoverflow.com/questions/31048557/assigning-null-to-json-fields-instead-of-empty-strings
// https://stackoverflow.com/questions/25087960/json-unmarshal-time-that-isnt-in-rfc-3339-format/

// ComponentStartDate represents the start_date field for a component. This
// value may potentially be null in the input JSON feed.
type ComponentStartDate struct {
	time.Time
}

// ComponentStartDateLayout is the time format/layout for the start_date field
// of a Statuspage component.
const ComponentStartDateLayout = "2006-01-02"

// MarshalJSON implements the json.Marshaler interface. This compliments the
// custom Unmarshaler implementation to handle potentially null component
// start_date field value.
func (csd *ComponentStartDate) MarshalJSON() ([]byte, error) {

	// Ensure that a zero value is mapped back to null to match the feed
	// value that we originally parsed.
	if csd.IsZero() {
		return []byte("null"), nil
	}

	return json.Marshal(csd.Time.Format(ComponentStartDateLayout))
}

// UnmarshalJSON implements the json.Unmarshaler interface to handle
// potentially null component start_date field value.
func (csd *ComponentStartDate) UnmarshalJSON(data []byte) error {

	s := strings.Trim(string(data), "\"")
	if s == "null" {
		csd.Time = time.Time{}
		return nil
	}

	var err error
	csd.Time, err = time.Parse(ComponentStartDateLayout, s)

	return err

}

// IsSet is a wrapper around the (time.Time).IsZero() method which indicates
// whether the start_date field of a component is unset or null.
func (csd ComponentStartDate) IsSet() bool {

	// Our custom implementation of the json.Unmarshaler interface uses a zero
	// value time.Time value to represent the start_date field in the JSON
	// feed having a null value, so we can use the IsZero() method to check
	// for this; if that method returns true, the start_date field was not
	// set.
	return !csd.IsZero()
}

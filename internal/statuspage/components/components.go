// Copyright 2021 Adam Chalkley
//
// https://github.com/atc0005/check-statuspage
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

package components

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/atc0005/check-statuspage/internal/statuspage"
	"github.com/atc0005/check-statuspage/internal/textutils"
	"github.com/atc0005/go-nagios"
)

// Component status enums.
// https://developer.statuspage.io/#operation/getPagesPageIdComponents
const (
	ComponentStatusDegradedPerformance string = "degraded_performance"
	ComponentStatusPartialOutage       string = "partial_outage"
	ComponentStatusMajorOutage         string = "major_outage"
	ComponentStatusUnderMaintenance    string = "under_maintenance"
	ComponentStatusOperational         string = "operational"

	// ComponentStatusUnknown is not an official component status, but rather
	// an indication that something went wrong when evaluating a component's
	// status.
	ComponentStatusUnknown string = "UNKNOWN"
)

// Component IDs are officially documented as "Identifier for component" with
// a value of type string. Neither the ID length nor the specific character
// pattern is provided in developer documentation, but after reviewing the ID
// values used by many Statuspage powered sites the following details are
// believed to accurately describe component IDs.
const (
	// ComponentIDRegex is an observed pattern for component ID values. For
	// example, the Box.com Statuspage has a component Group named "Mobile
	// Applications" with an ID of "g58jskvcnlh8" (no quotes). This is 12
	// characters long and is composed of lowercase ascii characters and the
	// numbers 0 through 9.
	ComponentIDRegex string = "[a-z0-9]{12}"

	// ComponentIDLength is the consistent length of observed component ID
	// values.
	ComponentIDLength int = 12
)

// Set represents the collection of components for a Statuspage-enabled site.
//
// Initial version generated via:
//
// wget https://status.box.com/api/v2/components.json
// json2struct -f components.json
type Set struct {
	Page struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		TimeZone  string    `json:"time_zone"`
		UpdatedAt time.Time `json:"updated_at"`
		URL       string    `json:"url"`
	} `json:"page"`

	// FilterUsed indicates what Filter values were applied to this Set (if
	// any).
	FilterUsed Filter `json:"-"`

	// Top-level collection of components.
	Components []Component `json:"components"`

	// FilterApplied indicates whether the Filter() method has been called
	// with valid Filter values.
	FilterApplied bool `json:"-"`

	// EvalAllComponents indicates whether the user has opted to skip
	// filtering entirely an evaluate all components.
	EvalAllComponents bool `json:"-"`
}

// Component represents one of the components defined for a Statuspage-enabled
// site.
type Component struct {

	// ComponentIDs is either a collection of individual subcomponent (aka,
	// "child") ID values when this component represents a component group,
	// otherwise is not present in the feed data.
	ComponentIDs []string `json:"components,omitempty"`

	// Description is additional human readable details for a component
	// intended to expand upon the component name. This is used as mouseover
	// or "hover text" for the component. This value may be null in the JSON
	// feed.
	Description statuspage.NullString `json:"description"`

	// GroupID is the component group (aka, "parent") identifier if this is a
	// subcomponent (aka, "child") or an empty string if this component is a
	// component group.
	GroupID statuspage.NullString `json:"group_id"`

	// Id is the unique identifier for the component.
	ID string `json:"id"`

	// Name is the human readable text shown on the Statuspage for a
	// component. The uniqueness of this value is not enforced, even for two
	// components at the same level, whether as members of a component group
	// or as a top-level component.
	Name string `json:"name"`

	// Status indicates the current status of the component. Uses a fixed set
	// of enum values.
	Status string `json:"status"`

	// PageID is the unique identifier for the Statuspage associated with the
	// components feed.
	PageID string `json:"page_id"`

	// Position is the order a component will appear on the page.
	Position int `json:"position"`

	// The creation time for this component. This value does not appear to be
	// updated once set.
	CreatedAt time.Time `json:"created_at"`

	// StartDate indicates the date this component started being used. By
	// default, when a new component is added the date will be set the current
	// date. This value is used to provide a representation of historical
	// uptime.
	//
	// This value may be null. See the ComponentStartDateLayout for the
	// specific format used by this date. Use the IsSet() method to determine
	// whether the component has this field set.
	StartDate ComponentStartDate `json:"start_date"`

	// UpdatedAt indicates when the component was last updated.
	UpdatedAt time.Time `json:"updated_at"`

	// Group indicates whether this component is a component group (aka,
	// "parent" or "container") for more specific subcomponents.
	Group bool `json:"group"`

	// OnlyShowIfDegraded indicates whether a component is hidden from view
	// until the status of a parent/container component is in a non-OK or
	// non-operational state.
	OnlyShowIfDegraded bool `json:"only_show_if_degraded"`

	// Showcase indicates whether this component should be showcased.
	Showcase bool `json:"showcase"`

	// Exclude indicates whether this component has been marked for exclusion
	// from overall plugin status evaluation. Until filtering is applied, all
	// components are evaluated for non-operational status.
	Exclude bool `json:"-"`
}

// ComponentGroup represents the parent or container component and all
// subcomponents which comprise a component group. This is intended as a
// "wrapper" around the parent component and associated Subcomponents and
// contains pointers to each.
type ComponentGroup struct {
	Parent        *Component
	Subcomponents []*Component
}

// Filter provides the criteria used to select a component group or
// top-level component for monitoring status evaluation. Any component not
// matched by these values is excluded from evaluation (aka, "ignored").
//
// Filter is a WIP. Need to setup logic that checks component strings for
// potential IDs, then falls back to Name match attempts. Same for Group.
type Filter struct {
	Group      string
	Components []string
}

// NewFromURL constructs a components Set by reading and decoding JSON data
// from a specified URL using the specified number of bytes as the read limit.
// If specified, unknown fields in the JSON file are ignored. An error is
// returned if there are problems reading and decoding JSON data. If provided,
// a custom user agent is supplied in place of the default Go user agent.
func NewFromURL(ctx context.Context, apiURL string, limit int64, allowUnknownFields bool, userAgent string) (*Set, error) {

	logger.Printf("Validating URL %q before attempting to read data", apiURL)
	parsedURL, err := url.Parse(apiURL)
	if err != nil {
		return &Set{}, fmt.Errorf(
			"error parsing specified URL %q: %w",
			apiURL,
			err,
		)
	}
	logger.Printf("Successfully validated URL %q", apiURL)

	c := &http.Client{}

	logger.Print("Preparing HTTP request")
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return &Set{}, fmt.Errorf(
			"error preparing request for specified URL %q: %w",
			apiURL,
			err,
		)
	}

	// Explicitly note that we want JSON content.
	request.Header.Add("Content-Type", "application/json;charset=utf-8")

	// If provided, override the default Go user agent ("Go-http-client/1.1")
	// with custom value.
	if userAgent != "" {
		request.Header.Set("User-Agent", userAgent)
	}

	logger.Print("Submitting HTTP request")
	response, err := c.Do(request)
	if err != nil {
		return &Set{}, fmt.Errorf(
			"error submitting HTTP request for specified URL %q: %w",
			apiURL,
			err,
		)
	}

	logger.Print("Successfully submitted HTTP request")

	// Make sure that we close the response body once we're done with it
	defer func() {
		if err := response.Body.Close(); err != nil {
			logger.Printf("error closing response body: %v", err)
		}
	}()

	if err := ctx.Err(); err != nil {
		logger.Print("context has expired")
		return &Set{}, fmt.Errorf("timeout reached: %w", err)
	}

	switch {
	case response.ContentLength == -1:
		logger.Printf("Response indicates unknown length of content from %q", apiURL)
	default:
		logger.Printf(
			"Response indicates %d bytes available to be read from %q",
			response.ContentLength,
			apiURL,
		)
	}

	switch {

	// Successful / expected response.
	case response.StatusCode == http.StatusOK:
		logger.Printf("Status code %d received as expected", response.StatusCode)

	// Success status range, but not in API docs for listing components.
	case response.StatusCode > 200 && response.StatusCode <= 299:
		logger.Printf(
			"Status code %d (%s) received; expected %d (%s), but received value within success range",
			response.StatusCode,
			http.StatusText(response.StatusCode),
			http.StatusOK,
			http.StatusText(http.StatusOK),
		)

	// Everything else is assumed to be an error.
	default:

		// Get the response body, then convert to string for use with extended
		// error messages
		responseData, readErr := ioutil.ReadAll(io.LimitReader(response.Body, limit))
		if readErr != nil {
			logger.Print(readErr)
			return &Set{}, readErr
		}
		responseString := string(responseData)

		statusCodeErr := fmt.Errorf(
			"unexpected response from %q API: %v (%s)",
			apiURL,
			response.Status,
			responseString,
		)

		logger.Print(statusCodeErr)

		return &Set{}, statusCodeErr

	}

	logger.Printf(
		"Decoding JSON data from %q using a limit of %d bytes",
		apiURL,
		limit,
	)

	var set Set
	err = decode(&set, response.Body, apiURL, limit, allowUnknownFields)
	if err != nil {
		return &Set{}, fmt.Errorf(
			"failed to decode JSON data: %w",
			err,
		)
	}

	logger.Printf(
		"No errors encountered while decoding JSON data from %q",
		apiURL,
	)

	return &set, nil

}

// NewFromFile constructs a components Set by reading and decoding JSON data
// from a fully-qualified path to a JSON file using the specified number of
// bytes as the read limit. If specified, unknown fields in the JSON file are
// ignored. An error is returned if there are problems reading the specified
// file.
func NewFromFile(filename string, limit int64, allowUnknownFields bool) (*Set, error) {

	logger.Printf("Opening file %s for reading", filename)
	fh, err := os.Open(filepath.Clean(filename))
	if err != nil {
		return &Set{}, fmt.Errorf("failed to open file for reading: %w", err)
	}
	logger.Printf("Successfully opened file %s for reading", filename)

	// #nosec G307
	// Believed to be a false-positive from recent gosec release
	// https://github.com/securego/gosec/issues/714
	defer func() {
		if err := fh.Close(); err != nil {
			// Ignore "file already closed" errors
			if !errors.Is(err, os.ErrClosed) {
				logger.Printf("failed to close file %s: %s",
					filename,
					err.Error(),
				)
			}
		}
	}()

	logger.Printf(
		"Decoding JSON file %s using a limit of %d bytes",
		filename,
		limit,
	)

	var set Set
	err = decode(&set, fh, filename, limit, allowUnknownFields)
	if err != nil {
		return &Set{}, fmt.Errorf(
			"failed to decode JSON data: %w",
			err,
		)
	}

	logger.Printf(
		"No errors encountered while decoding JSON file %s",
		filename,
	)

	return &set, nil
}

// decode is a helper function intended to handle the core JSON decoding tasks
// for various JSON sources (file, http body, etc.).
func decode(dst interface{}, reader io.Reader, sourceName string, limit int64, allowUnknownFields bool) error {

	logger.Printf(
		"Setting up JSON decoder for source %s with a limit of %d bytes",
		sourceName,
		limit,
	)
	dec := json.NewDecoder(io.LimitReader(reader, limit))

	switch {
	case !allowUnknownFields:
		logger.Print("Disallowing unknown JSON feed fields")
		dec.DisallowUnknownFields()
	default:
		logger.Print("Allowing unknown JSON feed fields by request")
	}

	logger.Print("Decoding JSON input")

	// Decode the first JSON object.
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf(
			"failed to decode JSON feed from source %s: %w",
			sourceName,
			err,
		)
	}
	logger.Print("Successfully decoded JSON input")

	// If there is more than one object, something is off.
	if dec.More() {
		return fmt.Errorf(
			"source %s contains multiple JSON objects; only one JSON object is supported",
			sourceName,
		)
	}

	return nil

}

// ServiceStateToComponentStatuses converts a given Nagios ServiceState to a
// collection of component statuses that are considered to be an equivalent
// value.
func ServiceStateToComponentStatuses(serviceState nagios.ServiceState) []string {

	switch serviceState.ExitCode {

	case nagios.StateCRITICALExitCode:
		return []string{
			ComponentStatusMajorOutage,
		}

	case nagios.StateWARNINGExitCode:
		return []string{
			ComponentStatusUnderMaintenance,
			ComponentStatusPartialOutage,
			ComponentStatusDegradedPerformance,
		}

	case nagios.StateOKExitCode:
		return []string{
			ComponentStatusOperational,
		}

	default:
		// this shouldn't be reached, so assume something really odd occurred
		logger.Println("unknown entity status provided, indicate UNKNOWN status")
		return []string{
			ComponentStatusUnknown,
		}
	}
}

// ComponentStatusToServiceState converts a Statuspage Status (e.g.,
// "degraded_performance", "under_maintenance") to a Nagios ServiceState.
func ComponentStatusToServiceState(componentStatus string) nagios.ServiceState {

	switch componentStatus {
	case ComponentStatusOperational:
		return nagios.ServiceState{
			Label:    nagios.StateOKLabel,
			ExitCode: nagios.StateOKExitCode,
		}

	case ComponentStatusUnderMaintenance:
		return nagios.ServiceState{
			Label:    nagios.StateWARNINGLabel,
			ExitCode: nagios.StateWARNINGExitCode,
		}

	case ComponentStatusMajorOutage:
		return nagios.ServiceState{
			Label:    nagios.StateCRITICALLabel,
			ExitCode: nagios.StateCRITICALExitCode,
		}

	case ComponentStatusPartialOutage:
		return nagios.ServiceState{
			Label:    nagios.StateWARNINGLabel,
			ExitCode: nagios.StateWARNINGExitCode,
		}

	case ComponentStatusDegradedPerformance:
		return nagios.ServiceState{
			Label:    nagios.StateWARNINGLabel,
			ExitCode: nagios.StateWARNINGExitCode,
		}

	default:
		// this shouldn't be reached, so assume something really odd occurred
		logger.Println("unknown entity status provided, indicate UNKNOWN status")
		return nagios.ServiceState{
			Label:    nagios.StateUNKNOWNLabel,
			ExitCode: nagios.StateUNKNOWNExitCode,
		}

	}

}

// String implements the Stringer interface for a components Filter.
func (f Filter) String() string {
	return fmt.Sprintf(
		`{Group: "%s", Components: "%s"}`,
		f.Group,
		strings.Join(f.Components, ", "),
	)
}

// Validate performs basic validation to assert that given filter settings are
// valid.
func (f Filter) Validate() error {

	// Assert that one of group or components list were provided.
	if f.Group == "" && len(f.Components) == 0 {
		return ErrComponentSetFilterEmpty
	}

	// Assert that group, component flags were not provided only
	// whitespace characters.
	switch {
	case f.Group != "":
		if strings.TrimSpace(f.Group) == "" {
			return ErrComponentSetFilterWhitespaceGroupField
		}

	case len(f.Components) > 0:
		for _, component := range f.Components {
			if strings.TrimSpace(component) == "" {
				return ErrComponentSetFilterWhitespaceComponentsField
			}
		}
	}

	return nil
}

// IsParent indicates whether this component represents a group of components,
// otherwise known as a component group (aka, "parent" or "container").
func (c Component) IsParent() bool {
	return c.Group
}

// IsChild indicates whether this component is a member (aka, "child" or
// "subcomponent") of a component group.
func (c Component) IsChild() bool {
	return !c.Group
}

// IsOKState indicates whether a component is in an OK or "operational" state.
func (c Component) IsOKState() bool {
	return c.Status == ComponentStatusOperational
}

// String implements the Stringer interface.
func (c Component) String() string {
	return fmt.Sprintf(
		`{Name: %q, ID: %q, Group: %t, GroupID: %q, Status: %q}`,
		c.Name,
		c.ID,
		c.Group,
		c.GroupID,
		c.Status,
	)
}

// String implements the Stringer interface.
func (cg ComponentGroup) String() string {

	quotedNames := make([]string, 0, len(cg.Subcomponents))
	for _, component := range cg.Subcomponents {
		quotedNames = append(quotedNames, fmt.Sprintf("%q", component.Name))
	}

	subComponentsFormattedList := strings.Join(quotedNames, ", ")
	return fmt.Sprintf(
		`{GroupName: %q, Subcomponents: [%s]`,
		cg.Parent.Name,
		subComponentsFormattedList,
	)
}

// Validate runs very basic validation checks on the decoded JSON input for
// fields that we either use or expect to use in the future. An error is
// returned if any validation checks fail.
func (cs *Set) Validate() error {

	// NOTE: JSON decoding/unmarshaling will not generate an error if a field
	// is missing in the source JSON stream.
	//
	// https://stackoverflow.com/questions/19633763/unmarshaling-json-in-go-required-field
	//
	// This method is responsible for ensuring that required JSON data is
	// present in the source.

	switch {
	case len(cs.Components) == 0:
		return fmt.Errorf(
			"%w: components collection empty",
			ErrComponentSetValidationFailed,
		)

	case cs.Page.ID == "":
		return fmt.Errorf(
			"%w: Page.ID field empty",
			ErrComponentSetValidationFailed,
		)

	case cs.Page.Name == "":
		return fmt.Errorf(
			"%w: Page.Name field empty",
			ErrComponentSetValidationFailed,
		)

	case cs.Page.TimeZone == "":
		return fmt.Errorf(
			"%w: Page.TimeZone field empty",
			ErrComponentSetValidationFailed,
		)

	// NOTE: This field is allowed to be null, so we shouldn't require that it
	// be present.
	//
	// case cs.Page.UpdatedAt.String() == time.Time{}.String():
	// 	return fmt.Errorf(
	// 		"%w: Page.UpdatedAt field empty",
	// 		ErrComponentSetValidationFailed,
	// 	)

	case cs.Page.URL == "":
		return fmt.Errorf(
			"%w: Page.URL field empty",
			ErrComponentSetValidationFailed,
		)
	}

	if len(cs.Components) > 0 {
		for i, component := range cs.Components {

			// Saying component 3 when the feed shows as 4th might be
			// confusing, so display using 1 as starting value.
			humanReadableComponentNumber := i + 1

			switch {
			case component.Name == "":
				return fmt.Errorf(
					"%w: Component[%d].Name field empty",
					ErrComponentSetValidationFailed,
					humanReadableComponentNumber,
				)

			case component.ID == "":
				return fmt.Errorf(
					"%w: Component[%d].ID field empty",
					ErrComponentSetValidationFailed,
					humanReadableComponentNumber,
				)

			case component.Status == "":
				return fmt.Errorf(
					"%w: Component[%d].Status field empty",
					ErrComponentSetValidationFailed,
					humanReadableComponentNumber,
				)

			case component.PageID == "":
				return fmt.Errorf(
					"%w: Component[%d].PageID field empty",
					ErrComponentSetValidationFailed,
					humanReadableComponentNumber,
				)

			case component.Group:
				// If a component is marked as a group, but does not define
				// any subcomponents.
				//
				// NOTE: This would indicate something is likely very "off"
				// with the Statuspage API itself; very unlikely scenario.
				if len(component.ComponentIDs) == 0 {
					return fmt.Errorf(
						"%w: Component[%d] is marked as group, but ComponentIDs field empty",
						ErrComponentSetValidationFailed,
						humanReadableComponentNumber,
					)
				}
			}
		}
	}

	return nil

}

// ExcludedComponents returns any components in the set which have been marked
// as excluded. The returned collection of component values may be empty. This
// method does not distinguish between top-level, sub or group component
// values; any component marked for exclusion is returned.
func (cs *Set) ExcludedComponents() []*Component {
	var excludedComponents []*Component
	for i := range cs.Components {
		if cs.Components[i].Exclude {
			excludedComponents = append(excludedComponents, &cs.Components[i])
		}
	}
	return excludedComponents
}

// NotExcludedComponents returns any components in the set which have not been
// marked as excluded. The returned collection of component values may be
// empty. This method does not distinguish between top-level, sub or group
// component values; any component not marked for exclusion is returned.
func (cs *Set) NotExcludedComponents() []*Component {
	var notExcludedComponents []*Component
	for i := range cs.Components {
		if !cs.Components[i].Exclude {
			notExcludedComponents = append(notExcludedComponents, &cs.Components[i])
		}
	}
	return notExcludedComponents
}

// AllIDsInListExcluded asserts that all component IDs in a given list have
// been excluded from the components set.
func (cs *Set) AllIDsInListExcluded(componentIDs []string) bool {

	// If we're not given a list of component IDs to evaluate, we have to
	// assume that there is a logic failure somewhere.
	if len(componentIDs) == 0 {
		return false
	}

	// Index all excluded/included components
	exclusionIdx := make(map[bool][]*Component)
	for i := range cs.Components {
		excludeStatus := cs.Components[i].Exclude
		exclusionIdx[excludeStatus] = append(
			exclusionIdx[excludeStatus],
			&cs.Components[i],
		)
	}

	// Early exit if there is a mismatch in numbers of listed IDs and excluded
	// components.
	if len(componentIDs) != len(exclusionIdx[true]) {
		return false
	}

	// Build a list of IDs for excluded components
	excludedComponentIDs := make([]string, 0, len(exclusionIdx[true]))
	for _, component := range exclusionIdx[true] {
		excludedComponentIDs = append(excludedComponentIDs, component.ID)
	}

	for _, component := range cs.Components {
		if component.Exclude {
			if !textutils.InList(component.ID, excludedComponentIDs, true) {
				return false
			}
		}
	}

	// If all conditions pass, then indicate that all listed IDs have been
	// excluded.
	return true

}

// AreOnlyIDsInListExcluded asserts that all component IDs in a given list
// have been excluded from the components set and vice versa, that all
// excluded components are in the given list.
func (cs *Set) AreOnlyIDsInListExcluded(componentIDs []string) bool {

	// If we're not given a list of component IDs to evaluate, we have to
	// assume that there is a logic failure somewhere.
	if len(componentIDs) == 0 {
		return false
	}

	// Index all excluded/included components
	exclusionIdx := make(map[bool][]*Component)
	for i := range cs.Components {
		excludeStatus := cs.Components[i].Exclude
		exclusionIdx[excludeStatus] = append(
			exclusionIdx[excludeStatus],
			&cs.Components[i],
		)
	}

	// Early exit if there is a mismatch in numbers of listed IDs and excluded
	// components.
	if len(componentIDs) != len(exclusionIdx[true]) {
		return false
	}

	// Build a list of IDs for excluded components
	excludedComponentIDs := make([]string, 0, len(exclusionIdx[true]))
	for _, component := range exclusionIdx[true] {
		excludedComponentIDs = append(excludedComponentIDs, component.ID)
	}

	for _, component := range cs.Components {
		if component.Exclude {
			if !textutils.InList(component.ID, excludedComponentIDs, true) {
				return false
			}
		}
	}

	// If all conditions pass, then indicate that all listed IDs have been
	// excluded.
	return true

}

// HasCriticalState indicates whether components in the collection have a
// specific non-operational status which maps to an CRITICAL state. A boolean
// value is accepted which indicates whether component values marked for
// exclusion (during filtering) should also be considered. The caller is
// responsible for filtering the collection prior to calling this method.
func (cs *Set) HasCriticalState(evalExcluded bool) bool {

	funcTimeStart := time.Now()

	defer func() {
		logger.Printf(
			"It took %v to execute HasCriticalState func.\n",
			time.Since(funcTimeStart),
		)
	}()

	// Validation should have caught the scenario where no components are in
	// the set; we indicate an error here since we expect something to be
	// present.
	if cs == nil {
		return false
	}

	var hasCriticalState bool

	for i := range cs.Components {

		switch {
		case cs.Components[i].Exclude && !evalExcluded:
			continue

		default:
			serviceState := ComponentStatusToServiceState(cs.Components[i].Status)
			if serviceState.ExitCode == nagios.StateCRITICALExitCode {
				hasCriticalState = true
			}
		}
	}

	return hasCriticalState
}

// NumCriticalState indicates how many components in the collection have a
// specific non-operational status which maps to a CRITICAL state. A boolean
// value is accepted which indicates whether all component values are
// evaluated or only those not marked for exclusion. The caller is responsible
// for filtering the collection prior to calling this method.
func (cs *Set) NumCriticalState(evalExcluded bool) int {

	funcTimeStart := time.Now()

	defer func() {
		logger.Printf(
			"It took %v to execute NumCriticalState func.\n",
			time.Since(funcTimeStart),
		)
	}()

	// Validation should have caught the scenario where no components are in
	// the set; we indicate an error here since we expect something to be
	// present.
	if cs == nil {
		return 0
	}

	var numCriticalState int

	for i := range cs.Components {

		switch {
		case cs.Components[i].Exclude && !evalExcluded:
			continue

		default:
			serviceState := ComponentStatusToServiceState(cs.Components[i].Status)
			if serviceState.ExitCode == nagios.StateCRITICALExitCode {
				numCriticalState++
			}
		}
	}

	return numCriticalState

}

// HasWarningState indicates whether components in the collection have a
// specific non-operational status which maps to an WARNING state. A boolean
// value is accepted which indicates whether component values marked for
// exclusion (during filtering) should also be considered. The caller is
// responsible for filtering the collection prior to calling this method.
func (cs *Set) HasWarningState(evalExcluded bool) bool {

	funcTimeStart := time.Now()

	defer func() {
		logger.Printf(
			"It took %v to execute HasWarningState func.\n",
			time.Since(funcTimeStart),
		)
	}()

	// Validation should have caught the scenario where no components are in
	// the set; we indicate an error here since we expect something to be
	// present.
	if cs == nil {
		return false
	}

	var hasWarningState bool

	for i := range cs.Components {

		switch {
		case cs.Components[i].Exclude && !evalExcluded:
			continue

		default:
			serviceState := ComponentStatusToServiceState(cs.Components[i].Status)
			if serviceState.ExitCode == nagios.StateWARNINGExitCode {
				hasWarningState = true
			}
		}
	}

	return hasWarningState
}

// NumWarningState indicates how many components in the collection have a
// specific non-operational status which maps to a WARNING state. A boolean
// value is accepted which indicates whether all component values are
// evaluated or only those not marked for exclusion. The caller is responsible
// for filtering the collection prior to calling this method.
func (cs *Set) NumWarningState(evalExcluded bool) int {

	funcTimeStart := time.Now()

	defer func() {
		logger.Printf(
			"It took %v to execute NumWarningState func.\n",
			time.Since(funcTimeStart),
		)
	}()

	// Validation should have caught the scenario where no components are in
	// the set; we indicate an error here since we expect something to be
	// present.
	if cs == nil {
		return 0
	}

	var numWarningState int

	for i := range cs.Components {

		switch {
		case cs.Components[i].Exclude && !evalExcluded:
			continue

		default:
			serviceState := ComponentStatusToServiceState(cs.Components[i].Status)
			if serviceState.ExitCode == nagios.StateWARNINGExitCode {
				numWarningState++
			}
		}
	}

	return numWarningState

}

// HasUnknownState indicates whether components in the collection have a
// specific non-operational status which maps to an UNKNOWN state. A boolean
// value is accepted which indicates whether component values marked for
// exclusion (during filtering) should also be considered. The caller is
// responsible for filtering the collection prior to calling this method.
func (cs *Set) HasUnknownState(evalExcluded bool) bool {

	funcTimeStart := time.Now()

	defer func() {
		logger.Printf(
			"It took %v to execute HasUnknownState func.\n",
			time.Since(funcTimeStart),
		)
	}()

	// Validation should have caught the scenario where no components are in
	// the set; we indicate an error here since we expect something to be
	// present.
	if cs == nil {
		return false
	}

	var hasUnknownState bool

	for i := range cs.Components {

		switch {
		case cs.Components[i].Exclude && !evalExcluded:
			continue

		default:
			serviceState := ComponentStatusToServiceState(cs.Components[i].Status)
			if serviceState.ExitCode == nagios.StateUNKNOWNExitCode {
				hasUnknownState = true
			}
		}
	}

	return hasUnknownState
}

// NumUnknownState indicates how many components in the collection have a
// specific non-operational status which maps to an UNKNOWN state. A boolean
// value is accepted which indicates whether all component values are
// evaluated or only those not marked for exclusion. The caller is responsible
// for filtering the collection prior to calling this method.
func (cs *Set) NumUnknownState(evalExcluded bool) int {

	funcTimeStart := time.Now()

	defer func() {
		logger.Printf(
			"It took %v to execute NumUnknownState func.\n",
			time.Since(funcTimeStart),
		)
	}()

	// Validation should have caught the scenario where no components are in
	// the set; we indicate an error here since we expect something to be
	// present.
	if cs == nil {
		return 0
	}

	var numUnknownState int

	for i := range cs.Components {

		switch {
		case cs.Components[i].Exclude && !evalExcluded:
			continue

		default:
			serviceState := ComponentStatusToServiceState(cs.Components[i].Status)
			if serviceState.ExitCode == nagios.StateUNKNOWNExitCode {
				numUnknownState++
			}
		}
	}

	return numUnknownState

}

// IsOKState indicates whether all components in the collection have an
// operational status which maps to an OK state. A boolean
// value is accepted which indicates whether component values marked for
// exclusion (during filtering) should also be considered. The caller is
// responsible for filtering the collection prior to calling this method.
func (cs *Set) IsOKState(evalExcluded bool) bool {

	switch {
	case cs.HasCriticalState(evalExcluded):
		return false
	case cs.HasWarningState(evalExcluded):
		return false
	case cs.HasUnknownState(evalExcluded):
		return false
	default:
		return true
	}

}

// NumOKState indicates how many components in the collection have an
// operational status which maps to an OK state. A boolean value is accepted
// which indicates whether component values marked for exclusion (during
// filtering) should also be considered. The caller is responsible for
// filtering the collection prior to calling this method.
func (cs *Set) NumOKState(evalExcluded bool) int {

	funcTimeStart := time.Now()

	defer func() {
		logger.Printf(
			"It took %v to execute NumOKState func.\n",
			time.Since(funcTimeStart),
		)
	}()

	// Validation should have caught the scenario where no components are in
	// the set; we indicate an error here since we expect something to be
	// present.
	if cs == nil {
		return 0
	}

	var numOKState int

	for i := range cs.Components {

		switch {
		case cs.Components[i].Exclude && !evalExcluded:
			continue

		default:
			serviceState := ComponentStatusToServiceState(cs.Components[i].Status)
			if serviceState.ExitCode == nagios.StateOKExitCode {
				numOKState++
			}
		}
	}

	return numOKState

}

// ServiceState returns the current Nagios ServiceState for the Set. A boolean
// value is accepted which indicates whether component values marked for
// exclusion (during filtering) should also be considered. The caller is
// responsible for filtering the collection prior to calling this method.
func (cs *Set) ServiceState(evalExcluded bool) nagios.ServiceState {

	var serviceState nagios.ServiceState

	switch {
	case cs.HasCriticalState(evalExcluded):
		serviceState = nagios.ServiceState{
			Label:    nagios.StateCRITICALLabel,
			ExitCode: nagios.StateCRITICALExitCode,
		}

	case cs.HasWarningState(evalExcluded):
		serviceState = nagios.ServiceState{
			Label:    nagios.StateWARNINGLabel,
			ExitCode: nagios.StateWARNINGExitCode,
		}

	case cs.HasUnknownState(evalExcluded):
		serviceState = nagios.ServiceState{
			Label:    nagios.StateUNKNOWNLabel,
			ExitCode: nagios.StateUNKNOWNExitCode,
		}

	case cs.IsOKState(evalExcluded):
		serviceState = nagios.ServiceState{
			Label:    nagios.StateOKLabel,
			ExitCode: nagios.StateOKExitCode,
		}

	default:
		logger.Printf("default case triggered; unable to determine ServiceState, assuming UNKNOWN")
		serviceState = nagios.ServiceState{
			Label:    nagios.StateUNKNOWNLabel,
			ExitCode: nagios.StateUNKNOWNExitCode,
		}
	}

	return serviceState

}

// NumExcluded returns the number of components from the set that have been
// excluded from evaluation. This includes both operational ("OK") components
// as well as non-operational components.
func (cs *Set) NumExcluded() int {
	var num int
	for i := range cs.Components {
		if cs.Components[i].Exclude {
			num++
		}
	}

	return num
}

// Filter applies a given component group and components list Filter against
// the set, marking any non-matching components as excluded. This allows us to
// retain the collection as a whole for later review, but focus on the
// specific components that were requested. The given Filter allows specifying
// a component group by one of case-insensitive name or ID value. The Filter
// allows specifying a list of components by a mix of name or ID values.
//
// Specifying a component group by ID results in a single group match.
// Specifying a component group by name may result in one or many matches.
//
// If specified together, components are required to be subcomponents of at
// least one of the matched component groups.
//
// If a component group was specified in the Filter, but no components, all
// subcomponents for each matched component group are exempt from exclusion.
//
// If a component group was specified as a component using its ID, an error is
// returned. If a component group was specified as a component using its name
// and the only match is a component group, an error is returned. If a
// component group is specified as a component by name and there is at least
// one component match, the component group match is ignored.
//
// NOTE: During filtering, component groups are marked as excluded due to
// their not being explicitly recorded as "matched". This is not readily
// apparent in emitted feedback and is an internal detail, though this is
// pertinent to testing the filtering behavior/logic of the Set.
func (cs *Set) Filter(filter Filter) error {

	// Unlikely, but possible scenario.
	if cs == nil {
		return fmt.Errorf(
			"unable to apply filter to empty components set;" +
				" did you use a provided constructor to build set?",
		)
	}

	// Assert valid filter was provided.
	if err := filter.Validate(); err != nil {
		return fmt.Errorf(
			"unable to apply Filter to components set: %w",
			err,
		)
	}

	logger.Printf(
		"filter.Group set: %t, filter.Components has length of %d",
		filter.Group == "",
		len(filter.Components),
	)

	// If specified, build a map of component groups based off of group id.
	// We'll use this to determine if a subcomponent is a member of one of
	// the indexed groups.
	matchedGroupComponents := make(map[string]*Component, len(filter.Group))
	if filter.Group != "" {
		err := cs.matchGroupComponents(filter, matchedGroupComponents)
		if err != nil {
			return fmt.Errorf("cs.matchGroupComponents failed: %w", err)
		}
	}

	// Early exit logic already handles empty filter, so we have at least one
	// component specified OR we have a component group index populated with
	// the implied intent to evaluate all subcomponents of each group index
	// entry.
	matchedComponents := make(map[string]*Component)
	switch {
	case len(filter.Components) > 0:
		err := cs.matchComponents(filter, matchedGroupComponents, matchedComponents)
		if err != nil {
			return fmt.Errorf("cs.matchComponents failed: %w", err)
		}
	case filter.Group != "" && len(filter.Components) == 0:
		err := cs.recordSubcomponents(filter, matchedGroupComponents, matchedComponents)
		if err != nil {
			return fmt.Errorf(
				"cs.recordSubcomponents failed to record subcomponents: %w",
				err,
			)
		}
	}

	cs.excludeUnmatchedComponents(matchedComponents)

	return nil
}

// NumComponents returns the count of all components in the set.
func (cs *Set) NumComponents() int {
	return len(cs.Components)
}

// TopLevel returns the standalone, top-level components which are not members
// or subcomponents of a component group (aka, a component "container").
func (cs *Set) TopLevel() []*Component {

	var topLevel []*Component

	for i := range cs.Components {
		if !cs.Components[i].Group && cs.Components[i].GroupID == "" {
			topLevel = append(topLevel, &cs.Components[i])
		}
	}

	return topLevel

}

// NumTopLevel returns the count of the standalone, top-level components in
// the set which are not members or subcomponents of a component Group (aka, a
// component "container").
func (cs *Set) NumTopLevel() int {

	var n int

	for i := range cs.Components {
		if !cs.Components[i].Group && cs.Components[i].GroupID == "" {
			n++
		}
	}

	return n

}

// Groups returns any available component Groups in the set as a collection of
// component values. The returned collection of component values may be empty
// and does not include subcomponents.
func (cs *Set) Groups() []*Component {

	var groups []*Component

	for i := range cs.Components {
		if cs.Components[i].Group {
			groups = append(groups, &cs.Components[i])
		}
	}

	return groups

}

// NumGroups returns the count of available component Groups in the set.
func (cs *Set) NumGroups() int {

	var n int

	for i := range cs.Components {
		if cs.Components[i].Group {
			n++
		}
	}

	return n

}

// Subcomponents returns any available subcomponents in the set as a
// collection of component values. The returned collection of component values
// may be empty.
func (cs *Set) Subcomponents() []*Component {

	var subcomponents []*Component

	for i := range cs.Components {
		if !cs.Components[i].Group {
			subcomponents = append(subcomponents, &cs.Components[i])
		}
	}

	return subcomponents

}

// NumSubcomponents returns the count of available Subcomponents in the set.
func (cs *Set) NumSubcomponents() int {

	var n int

	for i := range cs.Components {
		if !cs.Components[i].Group && cs.Components[i].GroupID != "" {
			n++
		}
	}

	return n

}

// NumProblemComponents returns the count of components in the set which are
// in a non-OK or non-operational status. component Groups are not included in
// the count since groups mirror the status of subcomponents.
//
// A boolean value is accepted which indicates whether component values marked
// for exclusion (during filtering) should also be considered. The caller is
// responsible for filtering the collection prior to calling this method.
func (cs *Set) NumProblemComponents(evalExcluded bool) int {

	var n int

	for i := range cs.Components {

		if !cs.Components[i].Group &&
			!cs.Components[i].IsOKState() &&
			(!cs.Components[i].Exclude || (cs.Components[i].Exclude && evalExcluded)) {
			n++
		}
	}

	return n

}

// ProblemComponents returns any subcomponents in the set in a non-OK or
// non-operational status as a collection of component values. The returned
// collection of component values may be empty. component groups are not
// included in the count since groups mirror the status of subcomponents.
//
// A boolean value is accepted which indicates whether component values marked
// for exclusion (during filtering) should also be considered. The caller is
// responsible for filtering the collection prior to calling this method.
func (cs *Set) ProblemComponents(evalExcluded bool) []*Component {

	probComponents := make([]*Component, 0, cs.NumProblemComponents(evalExcluded))

	for i := range cs.Components {
		if !cs.Components[i].Group &&
			!cs.Components[i].IsOKState() &&
			(!cs.Components[i].Exclude || (cs.Components[i].Exclude && evalExcluded)) {
			probComponents = append(probComponents, &cs.Components[i])
		}
	}

	return probComponents

}

// NumProblemGroups returns the count of component groups in the set which are
// in a non-OK or non-operational status. component groups mirror the status
// of subcomponents.
func (cs *Set) NumProblemGroups() int {

	var n int

	for i := range cs.Components {
		if cs.Components[i].Group &&
			cs.Components[i].Status != ComponentStatusOperational {
			n++
		}
	}

	return n

}

// ProblemComponentGroups returns ComponentGroup values from the set composed
// only of component group and subcomponent values which are in a non-OK or
// non-operational status.
//
// The returned collection of ComponentGroup values may be empty. An error is
// returned if there are problems compiling the collection of ComponentGroup
// values.
func (cs *Set) ProblemComponentGroups() ([]ComponentGroup, error) {

	var groups []*Component
	for i := range cs.Components {
		if cs.Components[i].Group && !cs.Components[i].IsOKState() {
			groups = append(groups, &cs.Components[i])
		}
	}

	componentGroups := make([]ComponentGroup, 0, len(groups))
	for _, group := range groups {
		var components []*Component
		for _, componentID := range group.ComponentIDs {
			component, err := cs.GetComponentByID(componentID)
			if err != nil {
				return []ComponentGroup{}, fmt.Errorf(
					"failed to retrieve component by ID %s for group %s: %w",
					componentID,
					group.Name,
					err,
				)
			}

			// Only include the component if it is a non-operational status.
			if !component.IsOKState() {
				components = append(components, component)
			}

		}

		componentGroups = append(componentGroups, ComponentGroup{
			Parent:        group,
			Subcomponents: components,
		})

	}

	// Return collection regardless of whether we found any components in a
	// non-operational status.
	return componentGroups, nil

}

// GetAllGroups returns all component Group values in the set along
// with their subcomponents as a collection of ComponentGroup values. An error
// is returned if one is encountered while compiling the collection.
func (cs *Set) GetAllGroups() ([]ComponentGroup, error) {

	var allGroups []*Component
	for i := range cs.Components {
		if cs.Components[i].Group {
			allGroups = append(allGroups, &cs.Components[i])
		}
	}

	componentGroups := make([]ComponentGroup, 0, len(allGroups))
	for _, group := range allGroups {
		var components []*Component
		for _, componentID := range group.ComponentIDs {
			component, err := cs.GetComponentByID(componentID)
			if err != nil {
				return []ComponentGroup{}, fmt.Errorf(
					"failed to retrieve component by ID %s for group %s: %w",
					componentID,
					group.Name,
					err,
				)
			}
			components = append(components, component)
		}

		componentGroups = append(componentGroups, ComponentGroup{
			Parent:        group,
			Subcomponents: components,
		})

	}

	if len(componentGroups) > 0 {
		return componentGroups, nil
	}

	return []ComponentGroup{}, fmt.Errorf(
		"failed to retrieve component groups: %w",
		ErrComponentGroupNotFound,
	)

}

// GetGroupsByName uses the (case-insensitive) specified component name as the
// key, returning a collection of ComponentGroup values for each match or an
// error if a match was not found in the set. If present, multiple
// ComponentGroup values may be returned for the same name (even at the same
// group level).
//
// Reminder: component names are not unique, even at the same level.
//
// component group names may be duplicated.
// Empty component groups are disallowed by the Statuspage API.
// Subcomponent names may be duplicated (even for the same parent).
// Component names may be duplicated (even at the top-level).
func (cs *Set) GetGroupsByName(key string) ([]ComponentGroup, error) {

	componentNameMatches, err := cs.GetComponentsByName(key)
	if err != nil {
		return []ComponentGroup{}, fmt.Errorf(
			"failed to retrieve components for search string %s: %w",
			key,
			err,
		)
	}

	var allGroups []*Component
	for _, component := range componentNameMatches {
		if component.Group {
			allGroups = append(allGroups, component)
		}
	}

	componentGroupMatches := make([]ComponentGroup, 0, len(cs.Components))
	for _, group := range allGroups {
		var components []*Component
		for _, componentID := range group.ComponentIDs {
			component, err := cs.GetComponentByID(componentID)
			if err != nil {
				return []ComponentGroup{}, fmt.Errorf(
					"failed to retrieve components for group %s: %w",
					group.Name,
					err,
				)
			}
			components = append(components, component)
		}

		componentGroupMatches = append(componentGroupMatches, ComponentGroup{
			Parent:        group,
			Subcomponents: components,
		})

	}

	if len(componentNameMatches) > 0 {
		return componentGroupMatches, nil
	}

	return []ComponentGroup{}, fmt.Errorf(
		"failed to retrieve component groups: %w",
		ErrComponentGroupNotFound,
	)

}

// GetGroupByID uses the (case-insensitive) specified component ID as the
// search key, returning a ComponentGroup consisting of a component serving as
// the component group and its subcomponents from the set or an error if a
// match was not found.
func (cs *Set) GetGroupByID(id string) (ComponentGroup, error) {

	group, err := cs.GetComponentByID(id)
	if err != nil {
		return ComponentGroup{}, fmt.Errorf(
			"failed to retrieve component group by id: %w",
			err,
		)
	}

	if !group.Group {
		return ComponentGroup{}, fmt.Errorf(
			"failed to retrieve component group by id: %w",
			ErrComponentIsNotComponentGroup,
		)
	}

	components := make([]*Component, 0, len(group.ComponentIDs))
	for _, componentID := range group.ComponentIDs {
		subcomponent, err := cs.GetComponentByID(componentID)
		if err != nil {
			return ComponentGroup{}, fmt.Errorf(
				"failed to retrieve subcomponents for group %s by id: %s: %w",
				group.Name,
				id,
				err,
			)
		}
		components = append(components, subcomponent)
	}

	return ComponentGroup{
			Parent:        group,
			Subcomponents: components},
		nil

}

// GetComponentsByName uses the (case-insensitive) specified component name as
// the key, returning a collection of matching components from the set or an
// error if a match was not found. If present, multiple components of the same
// name (even at the same group level) may be returned.
//
// Reminder: component names are not unique, even at the same level.
//
// Component Group names may be duplicated.
// Subcomponent names may be duplicated (even for the same parent).
// Component names may be duplicated (even at the top-level).
func (cs *Set) GetComponentsByName(searchKey string) ([]*Component, error) {

	var components []*Component

	// Play it safe and don't assume that the decoded JSON fields are devoid
	// of leading/trailing whitespace or is of a specific case.
	for i := range cs.Components {
		if strings.EqualFold(
			strings.TrimSpace(cs.Components[i].Name),
			strings.TrimSpace(searchKey),
		) {
			components = append(components, &cs.Components[i])
		}
	}

	if len(components) > 0 {
		return components, nil
	}

	return nil, ErrComponentNotFound
}

// GetComponentByID uses the (case-insensitive) specified component ID as the
// search key, returning the matching component from the set or an error if a
// match was not found.
func (cs *Set) GetComponentByID(id string) (*Component, error) {

	// Play it safe and don't assume that the decoded JSON fields are devoid
	// of leading/trailing whitespace or is of a specific case.
	for i := range cs.Components {
		if strings.EqualFold(
			strings.TrimSpace(cs.Components[i].ID),
			strings.TrimSpace(id),
		) {
			return &cs.Components[i], nil
		}
	}

	return nil, ErrComponentNotFound
}
